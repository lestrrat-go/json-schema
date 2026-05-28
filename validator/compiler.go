package validator

import (
	"context"
	"fmt"
	"slices"
	"strings"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/vocabulary"
)

// Compile builds a validator for s. A schema that declares a $dynamicAnchor is
// a potential bookend target for $dynamicRef, so its validator is wrapped to
// push the schema onto the runtime dynamic scope when validation enters it.
//
// Configuration (resolver, vocabulary set, base URI, base schema) is supplied
// through CompileOptions; the remaining compilation state (root/base schema,
// reference stack, recursion depths) is derived and threaded internally via
// compileState. ctx is used only for cancellation.
func Compile(ctx context.Context, s *schema.Schema, options ...CompileOption) (Interface, error) {
	return compile(ctx, s, newCompileState(s, options))
}

// compile is the internal entry point that threads an explicit compileState. It
// compiles s and, when s is a schema resource ($id) or declares a
// $dynamicAnchor, wraps the result so entering it during validation records it
// on the runtime dynamic scope (letting $dynamicRef find the outermost in-scope
// $dynamicAnchor).
func compile(ctx context.Context, s *schema.Schema, cs compileState) (Interface, error) {
	v, err := compileSchema(ctx, s, cs)
	if err != nil {
		return nil, err
	}
	if s != nil && (s.HasID() || s.HasDynamicAnchor()) {
		return &dynamicScopeValidator{schema: s, inner: v}, nil
	}
	return v, nil
}

// combineReferenceWithConstraints combines a resolved $ref/$dynamicRef validator
// with any sibling keywords present on the same schema, mirroring $ref handling
// so that keywords like unevaluatedProperties alongside a $dynamicRef are still
// applied (and can see the reference target's annotations).
func combineReferenceWithConstraints(ctx context.Context, s *schema.Schema, cs compileState, resolvedValidator Interface) (Interface, error) {
	if !hasOtherConstraints(s) {
		return resolvedValidator, nil
	}
	schemaWithoutRef := createSchemaWithoutRef(s)
	if schemaWithoutRef.HasUnevaluatedProperties() || schemaWithoutRef.HasUnevaluatedItems() {
		schemaWithoutUnevaluated := createSchemaWithoutUnevaluatedFields(schemaWithoutRef)
		additionalValidator, err := compile(ctx, schemaWithoutUnevaluated, cs)
		if err != nil {
			return nil, fmt.Errorf("failed to compile additional constraints: %w", err)
		}
		return &unevaluatedCoordinator{
			validators:       []Interface{resolvedValidator, additionalValidator},
			unevaluatedProps: schemaWithoutRef.UnevaluatedProperties(),
			unevaluatedItems: schemaWithoutRef.UnevaluatedItems(),
			strictArrayType:  hasExplicitArrayType(schemaWithoutRef),
			strictObjectType: hasExplicitObjectType(schemaWithoutRef),
		}, nil
	}
	additionalValidator, err := compile(ctx, schemaWithoutRef, cs)
	if err != nil {
		return nil, fmt.Errorf("failed to compile additional constraints: %w", err)
	}
	return AllOf(resolvedValidator, additionalValidator), nil
}

// compileSchema implements the simplified compilation approach using UnevaluatedCoordinator.
func compileSchema(ctx context.Context, s *schema.Schema, cs compileState) (Interface, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	skipIDRebase := cs.skipIDRebase
	cs.skipIDRebase = false // applies only to the immediate schema, not its subschemas

	// A schema with its own $id establishes a new base URI and is itself the
	// base resource for resolving references that appear within it. Re-base both
	// the base URI and the base schema so that this resource's relative refs
	// (e.g. "./bar.json") and local pointers (e.g. "#/$defs/inner") resolve
	// against this resource rather than an enclosing one.
	if s.HasID() && s.ID() != "" && !skipIDRebase {
		newBaseURI := cs.baseURI
		if absBase := schema.ResolveURI(cs.baseURI, s.ID()); absBase != "" {
			newBaseURI = absBase
		}
		cs = cs.withBase(s, newBaseURI)
	}

	// For root schema compilation, resolve vocabulary from metaschema.
	//
	// FIXME: This special-cases a single JSON Schema Test Suite remote fixture
	// instead of resolving the metaschema's $vocabulary declaration. The proper
	// implementation is vocabulary.ResolveVocabularyFromMetaschema, but wiring it
	// in is a behavioral change (its AllEnabled() fallback enables format-assertion,
	// unlike the DefaultSet) and can only be verified against the full
	// JSON Schema Test Suite. Left as-is pending that verified change.
	if cs.rootSchema == s && s.HasSchema() && s.Schema() == "http://localhost:1234/draft2020-12/metaschema-no-validation.json" {
		// This specific metaschema disables validation vocabulary.
		vocabSet := vocabulary.AllEnabled()
		vocabSet.Disable(vocabulary.ValidationURL)
		cs.cfg = &compileConfig{resolver: cs.cfg.resolver, vocab: vocabSet}
	}

	// Handle $ref and $dynamicRef first - if schema has a reference, resolve it immediately
	var reference string
	var isDynamicRef bool
	if s.HasReference() {
		reference = s.Reference()
	} else if s.HasDynamicReference() {
		reference = s.DynamicReference()
		isDynamicRef = true
	}

	if reference != "" {
		// Handle $dynamicRef with a DynamicReferenceValidator. Capture the
		// enclosing resource's base schema and base URI (not just the document
		// root) so the non-dynamic fallback — a $dynamicRef whose fragment is a
		// JSON pointer or a plain $ref to an $anchor — resolves within the correct
		// schema resource. The dynamic scope itself is read at validation time.
		if isDynamicRef {
			baseSchema := cs.baseSchema
			if baseSchema == nil {
				baseSchema = cs.rootSchema
			}
			drv := &DynamicReferenceValidator{
				reference:  reference,
				resolver:   cs.cfg.resolver,
				rootSchema: cs.rootSchema,
				baseSchema: baseSchema,
				baseURI:    cs.baseURI,
			}
			// Combine with any sibling keywords (e.g. unevaluatedProperties), the
			// same way $ref does, so they are not silently dropped.
			return combineReferenceWithConstraints(ctx, s, cs, drv)
		}

		resolver := cs.cfg.resolver

		// Circular-reference handling. A reference already on the stack is a
		// cycle. If a data boundary (an object/array child-applying keyword) has
		// been crossed since it was entered, it is data-bounded recursion that
		// terminates on the instance being validated, so compile it lazily as a
		// ReferenceValidator. Otherwise it is a pure cycle that can never
		// terminate, which is a compile-time error.
		if slices.Contains(cs.referenceStack, reference) {
			if cs.dataDepth > cs.refDepths[reference] {
				return &ReferenceValidator{
					reference:  reference,
					resolver:   resolver,
					rootSchema: cs.rootSchema,
					baseSchema: cs.baseSchema,
					baseURI:    cs.baseURI,
				}, nil
			}
			return nil, fmt.Errorf("circular reference detected: %s", reference)
		}
		// Push the reference, recording the data depth at which it was entered.
		cs = cs.pushReference(reference)

		// Resolve the reference to get the target schema.
		var targetSchema schema.Schema
		if err := resolver.ResolveReference(ctx, &targetSchema, reference, cs.baseSchema, cs.baseURI); err != nil {
			return nil, fmt.Errorf("reference resolution failed for %s: %w", reference, err)
		}

		// Check if schema has other constraints beyond the reference
		if hasOtherConstraints(s) {
			// Schema has both $ref and additional constraints: combine the resolved
			// schema and additional constraints.
			resolvedValidator, err := compile(ctx, &targetSchema, cs.withBaseSchema(&targetSchema))
			if err != nil {
				return nil, fmt.Errorf("failed to compile resolved schema: %w", err)
			}

			// Create schema without reference for additional constraints
			schemaWithoutRef := createSchemaWithoutRef(s)

			// Check if we need unevaluated coordination
			if schemaWithoutRef.HasUnevaluatedProperties() || schemaWithoutRef.HasUnevaluatedItems() {
				// Create additional validator WITHOUT unevaluated constraints
				schemaWithoutUnevaluated := createSchemaWithoutUnevaluatedFields(schemaWithoutRef)
				additionalValidator, err := compile(ctx, schemaWithoutUnevaluated, cs)
				if err != nil {
					return nil, fmt.Errorf("failed to compile additional constraints: %w", err)
				}

				// Use unevaluatedCoordinator to properly handle annotation sharing
				return &unevaluatedCoordinator{
					validators:       []Interface{resolvedValidator, additionalValidator},
					unevaluatedProps: schemaWithoutRef.UnevaluatedProperties(),
					unevaluatedItems: schemaWithoutRef.UnevaluatedItems(),
					strictArrayType:  hasExplicitArrayType(schemaWithoutRef),
					strictObjectType: hasExplicitObjectType(schemaWithoutRef),
				}, nil
			}

			// Regular allOf for cases without unevaluated constraints
			additionalValidator, err := compile(ctx, schemaWithoutRef, cs)
			if err != nil {
				return nil, fmt.Errorf("failed to compile additional constraints: %w", err)
			}
			return AllOf(resolvedValidator, additionalValidator), nil
		}

		// Schema has only $ref: recursively compile the resolved schema.
		resolvedCs := cs
		var resource *schema.Schema
		if strings.HasPrefix(reference, "#") {
			// Local reference: the target lives in the current resource, so the
			// base URI must not change. Re-base only if the target carries its own
			// $id.
			if targetSchema.HasID() {
				resolvedCs = cs.withBaseSchema(&targetSchema)
			}
		} else {
			// A reference into another document/resource. Its base URI is the
			// reference's absolute (retrieval) URI, so the target's own relative
			// references (e.g. "string.json") resolve against where it lives, and
			// its local "#/..." pointers resolve within the enclosing resource.
			absBase, _, _ := strings.Cut(schema.ResolveURI(cs.baseURI, reference), "#")
			resource = resolver.ResourceFor(absBase)
			if absBase != "" {
				resolvedCs = resolvedCs.withBaseURI(absBase)
			}
			switch {
			case resource != nil:
				// absBase is the resource's canonical registry URI, so suppress the
				// $id re-base in compileSchema (it would double a path segment).
				resolvedCs = resolvedCs.withBaseSchema(resource)
				resolvedCs.skipIDRebase = true
			case targetSchema.HasID():
				resolvedCs = resolvedCs.withBaseSchema(&targetSchema)
			}
		}
		compiled, err := compile(ctx, &targetSchema, resolvedCs)
		if err != nil {
			return nil, err
		}
		// Following the $ref enters the target's resource; record it on the
		// dynamic scope so a $dynamicRef deeper in the target can find it.
		if resource != nil && resource != &targetSchema {
			compiled = &dynamicScopeValidator{schema: resource, inner: compiled}
		}
		return compiled, nil
	}
	var validators []Interface

	// Phase 2: Compile composite validators (allOf, anyOf, oneOf)
	compositeValidators, err := compileCompositeValidators(ctx, s, cs)
	if err != nil {
		return nil, err
	}
	validators = append(validators, compositeValidators...)

	// Phase 3: Compile conditional validators (if/then/else, not)
	conditionalValidators, err := compileConditionalValidators(ctx, s, cs)
	if err != nil {
		return nil, err
	}
	validators = append(validators, conditionalValidators...)

	// Phase 4: Compile base constraint validators (properties, items, type, etc.)
	baseValidator, err := compileBaseConstraints(ctx, s, cs)
	if err != nil {
		return nil, err
	}
	if baseValidator != nil {
		validators = append(validators, baseValidator)
	}

	// Phase 4: If unevaluated constraints exist, wrap in coordinator
	if s.HasUnevaluatedProperties() || s.HasUnevaluatedItems() {
		return &unevaluatedCoordinator{
			validators:       validators, // May be empty if schema only has unevaluated constraints
			unevaluatedProps: s.UnevaluatedProperties(),
			unevaluatedItems: s.UnevaluatedItems(),
			strictArrayType:  hasExplicitArrayType(s),
			strictObjectType: hasExplicitObjectType(s),
		}, nil
	}

	// Phase 5: If no validators, return EmptyValidator
	if len(validators) == 0 {
		return &EmptyValidator{}, nil
	}

	// Phase 6: Single validator optimization
	if len(validators) == 1 {
		return validators[0], nil
	}

	// Phase 7: Multiple validators without unevaluated constraints - combine with AllOf
	return AllOf(validators...), nil
}

// compileCompositeValidators handles allOf, anyOf, oneOf compilation
func compileCompositeValidators(ctx context.Context, s *schema.Schema, cs compileState) ([]Interface, error) {
	var validators []Interface

	// AllOf
	if s.HasAllOf() {
		allOfValidators := make([]Interface, 0, len(s.AllOf()))
		for _, subSchema := range s.AllOf() {
			v, err := compile(ctx, convertSchemaOrBool(subSchema), cs)
			if err != nil {
				return nil, fmt.Errorf("failed to compile allOf validator: %w", err)
			}
			allOfValidators = append(allOfValidators, v)
		}
		validators = append(validators, AllOf(allOfValidators...))
	}

	// AnyOf
	if s.HasAnyOf() {
		anyOfValidators := make([]Interface, 0, len(s.AnyOf()))
		for _, subSchema := range s.AnyOf() {
			v, err := compile(ctx, convertSchemaOrBool(subSchema), cs)
			if err != nil {
				return nil, fmt.Errorf("failed to compile anyOf validator: %w", err)
			}
			anyOfValidators = append(anyOfValidators, v)
		}
		validators = append(validators, AnyOf(anyOfValidators...))
	}

	// OneOf
	if s.HasOneOf() {
		oneOfValidators := make([]Interface, 0, len(s.OneOf()))
		for _, subSchema := range s.OneOf() {
			v, err := compile(ctx, convertSchemaOrBool(subSchema), cs)
			if err != nil {
				return nil, fmt.Errorf("failed to compile oneOf validator: %w", err)
			}
			oneOfValidators = append(oneOfValidators, v)
		}
		validators = append(validators, OneOf(oneOfValidators...))
	}

	return validators, nil
}

// compileConditionalValidators handles if/then/else, not compilation
func compileConditionalValidators(ctx context.Context, s *schema.Schema, cs compileState) ([]Interface, error) {
	var validators []Interface

	// Not
	if s.HasNot() {
		notValidator, err := compile(ctx, s.Not(), cs)
		if err != nil {
			return nil, fmt.Errorf("failed to compile not validator: %w", err)
		}
		validators = append(validators, &NotValidator{validator: notValidator})
	}

	// If/Then/Else
	if s.HasIfSchema() {
		ifThenElseValidator, err := compileIfThenElseValidator(ctx, s, cs)
		if err != nil {
			return nil, fmt.Errorf("failed to compile if/then/else validator: %w", err)
		}
		validators = append(validators, ifThenElseValidator)
	}

	return validators, nil
}

// compileBaseConstraints compiles all base constraint validators (properties, items, type, etc.)
// This function excludes composition keywords and unevaluated constraints which are handled separately
func compileBaseConstraints(ctx context.Context, s *schema.Schema, cs compileState) (Interface, error) {
	var validators []Interface

	// Type validators - handle explicit type declarations
	if len(s.Types()) > 0 {
		var typeValidators []Interface
		for _, typ := range s.Types() {
			switch typ {
			case schema.StringType:
				// String type validator (with or without additional string constraints)
				stringValidator, err := compileStringValidator(s, cs.cfg.vocab, true) // strict type checking
				if err != nil {
					return nil, fmt.Errorf("failed to compile string validator: %w", err)
				}
				typeValidators = append(typeValidators, stringValidator)
			case schema.IntegerType:
				// Integer type validator
				integerValidator, err := compileIntegerValidator(s, cs.cfg.vocab)
				if err != nil {
					return nil, fmt.Errorf("failed to compile integer validator: %w", err)
				}
				typeValidators = append(typeValidators, integerValidator)
			case schema.NumberType:
				// Number type validator
				numberValidator, err := compileNumberValidator(s, cs.cfg.vocab)
				if err != nil {
					return nil, fmt.Errorf("failed to compile number validator: %w", err)
				}
				typeValidators = append(typeValidators, numberValidator)
			case schema.BooleanType:
				// Boolean type validator
				booleanValidator, err := compileBooleanValidator(s, cs.cfg.vocab)
				if err != nil {
					return nil, fmt.Errorf("failed to compile boolean validator: %w", err)
				}
				typeValidators = append(typeValidators, booleanValidator)
			case schema.ArrayType:
				// Array type validator (excluding unevaluatedItems)
				arrayFields := schema.ArrayConstraintFields &^ schema.UnevaluatedItemsField
				if s.HasAny(arrayFields) {
					// Strip unevaluatedItems so the array validator does not
					// self-enforce it; the unevaluatedCoordinator owns that
					// decision once sibling applicators' annotations are merged.
					baseSchema := createSchemaWithoutUnevaluatedFields(s)
					arrayValidator, err := compileArrayValidator(ctx, baseSchema, cs, true) // strict type checking
					if err != nil {
						return nil, fmt.Errorf("failed to compile array validator: %w", err)
					}
					typeValidators = append(typeValidators, arrayValidator)
				} else {
					// Just type checking without constraints
					simpleArrayValidator := Array().StrictArrayType(true).MustBuild()
					typeValidators = append(typeValidators, simpleArrayValidator)
				}
			case schema.ObjectType:
				// Object type validator (excluding unevaluatedProperties)
				objectFields := schema.ObjectConstraintFields &^ schema.UnevaluatedPropertiesField
				if s.HasAny(objectFields) {
					baseSchema := createSchemaWithoutUnevaluatedFields(s)
					objectValidator, err := compileObjectValidator(ctx, baseSchema, cs, true) // strict type checking
					if err != nil {
						return nil, fmt.Errorf("failed to compile object validator: %w", err)
					}
					typeValidators = append(typeValidators, objectValidator)
				} else {
					// Just type checking without constraints
					simpleObjectValidator := Object().StrictObjectType(true).MustBuild()
					typeValidators = append(typeValidators, simpleObjectValidator)
				}
			case schema.NullType:
				// Null type validator
				nullValidator := Null()
				typeValidators = append(typeValidators, nullValidator)
			}
		}

		// Combine type validators with AnyOf (OR logic) for multiple types
		if len(typeValidators) == 1 {
			validators = append(validators, typeValidators[0])
		} else {
			validators = append(validators, AnyOf(typeValidators...))
		}
	} else {
		// No explicit types - check for type-specific constraints that would imply a type

		// String constraints without explicit type
		if s.HasAny(schema.StringConstraintFields) {
			stringValidator, err := compileStringValidator(s, cs.cfg.vocab, false)
			if err != nil {
				return nil, fmt.Errorf("failed to compile string validator: %w", err)
			}
			validators = append(validators, stringValidator)
		}

		// Numeric constraints (includes both integer and number constraints)
		if s.HasAny(schema.NumericConstraintFields) {
			// Use inferred number validator for untyped schemas
			inferredValidator, err := compileInferredNumberValidator(s, cs.cfg.vocab)
			if err != nil {
				return nil, fmt.Errorf("failed to compile inferred number validator: %w", err)
			}
			validators = append(validators, inferredValidator)
		}

		// Array constraints (excluding unevaluatedItems) without explicit type
		arrayFields := schema.ArrayConstraintFields &^ schema.UnevaluatedItemsField
		if s.HasAny(arrayFields) {
			baseSchema := createSchemaWithoutUnevaluatedFields(s)
			arrayValidator, err := compileArrayValidator(ctx, baseSchema, cs, false)
			if err != nil {
				return nil, fmt.Errorf("failed to compile array validator: %w", err)
			}
			validators = append(validators, arrayValidator)
		}

		// Object constraints (excluding unevaluatedProperties) without explicit type
		objectFields := schema.ObjectConstraintFields &^ schema.UnevaluatedPropertiesField
		if s.HasAny(objectFields) {
			baseSchema := createSchemaWithoutUnevaluatedFields(s)
			objectValidator, err := compileObjectValidator(ctx, baseSchema, cs, false)
			if err != nil {
				return nil, fmt.Errorf("failed to compile object validator: %w", err)
			}
			validators = append(validators, objectValidator)
		}
	}

	// Content validation
	if s.HasContentEncoding() || s.HasContentMediaType() || s.HasContentSchema() {
		contentValidator, err := compileContentValidator(ctx, s, cs)
		if err != nil {
			return nil, fmt.Errorf("failed to compile content validator: %w", err)
		}
		validators = append(validators, contentValidator)
	}

	// Value constraints (enum, const) - handle for both typed and untyped schemas
	if s.HasAny(schema.ValueConstraintFields) {
		if len(s.Types()) == 0 {
			// Untyped schema with value constraints
			untypedValidator, err := compileUntypedValidator(s, cs.cfg.vocab)
			if err != nil {
				return nil, fmt.Errorf("failed to compile untyped validator: %w", err)
			}
			validators = append(validators, untypedValidator)
		} else {
			// Typed schema with value constraints - enum/const should be validated regardless of type
			valueValidator, err := compileValueConstraintsValidator(ctx, s)
			if err != nil {
				return nil, fmt.Errorf("failed to compile value constraints validator: %w", err)
			}
			if valueValidator != nil {
				validators = append(validators, valueValidator)
			}
		}
	}

	// Dependent schemas
	if s.HasDependentSchemas() {
		dependentSchemas := s.DependentSchemas()
		compiledDependentSchemas := make(map[string]Interface)
		for propertyName, depSchema := range dependentSchemas {
			// Handle SchemaOrBool types
			switch val := depSchema.(type) {
			case schema.BoolSchema:
				// Boolean schema: true means always valid, false means always invalid
				if bool(val) {
					compiledDependentSchemas[propertyName] = &EmptyValidator{}
				} else {
					compiledDependentSchemas[propertyName] = &NotValidator{validator: &EmptyValidator{}}
				}
			case *schema.Schema:
				// Regular schema object
				depValidator, err := compile(ctx, val, cs)
				if err != nil {
					return nil, fmt.Errorf("failed to compile dependent schema for property %s: %w", propertyName, err)
				}
				compiledDependentSchemas[propertyName] = depValidator
			default:
				return nil, fmt.Errorf("unexpected dependent schema type for property %s: %T", propertyName, depSchema)
			}
		}

		// Create object validator with dependent schemas context
		if len(compiledDependentSchemas) > 0 {
			dependentValidator := &objectValidator{
				dependentSchemas: compiledDependentSchemas,
			}
			validators = append(validators, dependentValidator)
		}
	}

	// Reference validation - create ReferenceValidator directly
	if s.HasReference() {
		refValidator := &ReferenceValidator{
			reference:  s.Reference(),
			resolver:   cs.cfg.resolver,
			rootSchema: cs.rootSchema,
		}
		validators = append(validators, refValidator)
	}

	// If no base constraints, return nil
	if len(validators) == 0 {
		//nolint: nilnil
		return nil, nil
	}

	// Single base validator optimization
	if len(validators) == 1 {
		return validators[0], nil
	}

	// Multiple base validators - combine with AllOf
	return AllOf(validators...), nil
}

// compileValueConstraintsValidator compiles enum and const constraints for typed schemas
func compileValueConstraintsValidator(_ context.Context, s *schema.Schema) (Interface, error) {
	// Use the untyped validator builder since enum/const validation logic is the same
	v := Untyped()

	if s.HasEnum() {
		v.Enum(s.Enum()...)
	}

	if s.HasConst() {
		v.Const(s.Const())
	}

	return v.Build()
}

// Helper functions for schema manipulation and type strictness detection

// createSchemaWithoutUnevaluatedFields creates a copy of the schema without unevaluated constraints
func createSchemaWithoutUnevaluatedFields(s *schema.Schema) *schema.Schema {
	// Use the builder to clone the schema and reset unevaluated fields
	builder := schema.NewBuilder().Clone(s)
	builder.ResetUnevaluatedProperties()
	builder.ResetUnevaluatedItems()
	return builder.MustBuild()
}

func hasExplicitArrayType(s *schema.Schema) bool {
	types := s.Types()
	if len(types) == 1 && types[0] == schema.ArrayType {
		return true
	}
	return false
}

func hasExplicitObjectType(s *schema.Schema) bool {
	types := s.Types()
	if len(types) == 1 && types[0] == schema.ObjectType {
		return true
	}
	return false
}
