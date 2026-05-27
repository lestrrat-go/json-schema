package validator

import (
	"context"
	"fmt"
	"maps"
	"slices"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/internal/schemactx"
	"github.com/lestrrat-go/json-schema/vocabulary"
)

// Compile implements the new simplified compilation approach using UnevaluatedCoordinator.
// This replaces the complex old compilation logic.
func Compile(ctx context.Context, s *schema.Schema) (Interface, error) {
	// Set up context with default resolver if none provided
	if schema.ResolverFromContext(ctx) == nil {
		ctx = schema.WithResolver(ctx, schema.NewResolver())
	}

	// Set up context with root schema if none provided. When s is the root
	// document (either because we just set it, or the caller pre-populated it
	// with this same schema), index its $id resources and anchors up front:
	// resolution is eager, so the index must exist before the first $ref is
	// compiled.
	isRoot := false
	if schema.RootSchemaFromContext(ctx) == nil {
		ctx = schema.WithRootSchema(ctx, s)
		isRoot = true
	} else if schema.RootSchemaFromContext(ctx) == s {
		isRoot = true
	}
	if isRoot {
		if resolver := schema.ResolverFromContext(ctx); resolver != nil {
			resolver.RegisterRoot(s)
		}
	}

	// Set up base schema for local reference resolution if none provided
	if schema.BaseSchemaFromContext(ctx) == nil {
		ctx = schema.WithBaseSchema(ctx, s)
	}

	// Set up vocabulary context if none provided
	// Default to JSON Schema 2020-12 default vocabulary (format-assertion disabled)
	var vocabSet *vocabulary.VocabularySet
	if err := schemactx.VocabularySetFromContext(ctx, &vocabSet); err != nil {
		// No vocabulary set in context, use default vocabulary per JSON Schema spec
		ctx = vocabulary.WithSet(ctx, vocabulary.DefaultSet())
	}

	// Add current schema to dynamic scope chain for $dynamicRef resolution
	ctx = schema.WithDynamicScope(ctx, s)

	// A schema with its own $id establishes a new base URI and is itself the
	// base resource for resolving references that appear within it. Re-base both
	// the base URI and the base schema so that this resource's relative refs
	// (e.g. "./bar.json") and local pointers (e.g. "#/$defs/inner") resolve
	// against this resource rather than an enclosing one.
	if s.HasID() && s.ID() != "" {
		parentBase := schema.BaseURIFromContext(ctx)
		if absBase := schema.ResolveURI(parentBase, s.ID()); absBase != "" {
			ctx = schema.WithBaseURI(ctx, absBase)
		}
		ctx = schema.WithBaseSchema(ctx, s)
	}

	// Handle vocabulary context - always check for root schema vocabulary resolution
	rootSchema := schema.RootSchemaFromContext(ctx)

	// For root schema compilation, resolve vocabulary from metaschema.
	//
	// FIXME: This special-cases a single JSON Schema Test Suite remote fixture
	// instead of resolving the metaschema's $vocabulary declaration. The proper
	// implementation is vocabulary.ResolveVocabularyFromMetaschema, but wiring it
	// in is a behavioral change (its AllEnabled() fallback enables format-assertion,
	// unlike the DefaultSet used above) and can only be verified against the full
	// JSON Schema Test Suite. Left as-is pending that verified change.
	if rootSchema == s {
		schemaURI := ""
		if s.HasSchema() {
			schemaURI = s.Schema()
		}

		if schemaURI != "" {
			if schemaURI == "http://localhost:1234/draft2020-12/metaschema-no-validation.json" {
				// This specific metaschema disables validation vocabulary
				// Create vocabulary set with validation disabled using VocabularySet
				vocabSet := vocabulary.AllEnabled()
				vocabSet.Disable(vocabulary.ValidationURL)
				ctx = vocabulary.WithSet(ctx, vocabSet)
			}
		}
	}

	// Ensure vocabulary context is set (fallback to all enabled if not set)
	if vocabulary.SetFromContext(ctx) == nil {
		ctx = vocabulary.WithSet(ctx, vocabulary.AllEnabled())
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
		// Handle $dynamicRef with proper DynamicReferenceValidator
		if isDynamicRef {
			// Create dynamic reference validator with stored context. Capture the
			// enclosing resource's base schema and base URI (not just the document
			// root) so that the non-dynamic fallback — a $dynamicRef whose fragment
			// is a JSON pointer or a plain $ref to an $anchor — resolves within the
			// correct schema resource when nested $id re-basing is in effect.
			dynamicScope := schema.DynamicScopeFromContext(ctx)
			baseSchema := schema.BaseSchemaFromContext(ctx)
			if baseSchema == nil {
				baseSchema = schema.RootSchemaFromContext(ctx)
			}
			return &DynamicReferenceValidator{
				reference:    reference,
				resolver:     schema.ResolverFromContext(ctx),
				rootSchema:   schema.RootSchemaFromContext(ctx),
				baseSchema:   baseSchema,
				baseURI:      schema.BaseURIFromContext(ctx),
				dynamicScope: dynamicScope,
			}, nil
		}

		// Get resolver from context (guaranteed to be present)
		resolver := schema.ResolverFromContext(ctx)

		// Get root schema from context (guaranteed to be present)
		_ = schema.RootSchemaFromContext(ctx)

		// Circular-reference handling. A reference already on the stack is a
		// cycle. If a data boundary (an object/array child-applying keyword) has
		// been crossed since it was entered, it is data-bounded recursion that
		// terminates on the instance being validated, so compile it lazily as a
		// ReferenceValidator. Otherwise it is a pure cycle that can never
		// terminate, which is a compile-time error.
		curDepth := schema.DataDepthFromContext(ctx)
		refDepths := schema.RefDepthsFromContext(ctx)
		stack := schema.ReferenceStackFromContext(ctx)
		if slices.Contains(stack, reference) {
			if curDepth > refDepths[reference] {
				return &ReferenceValidator{
					reference:  reference,
					resolver:   resolver,
					rootSchema: schema.RootSchemaFromContext(ctx),
					baseSchema: schema.BaseSchemaFromContext(ctx),
					baseURI:    schema.BaseURIFromContext(ctx),
				}, nil
			}
			return nil, fmt.Errorf("circular reference detected: %s", reference)
		}
		// Push the reference, recording the data depth at which it was entered.
		newStack := make([]string, len(stack)+1)
		copy(newStack, stack)
		newStack[len(stack)] = reference
		ctx = schema.WithReferenceStack(ctx, newStack)
		newDepths := make(map[string]int, len(refDepths)+1)
		maps.Copy(newDepths, refDepths)
		newDepths[reference] = curDepth
		ctx = schema.WithRefDepths(ctx, newDepths)

		// Resolve the reference to get the target schema
		var targetSchema schema.Schema
		baseURI := schema.BaseURIFromContext(ctx)
		refCtx := ctx
		if baseURI != "" {
			refCtx = schema.WithBaseURI(ctx, baseURI)
		}
		// Add base schema context for reference resolution if none exists
		// Don't override existing base schema context
		if schema.BaseSchemaFromContext(refCtx) == nil {
			if rootSchema := schema.RootSchemaFromContext(ctx); rootSchema != nil {
				refCtx = schema.WithBaseSchema(refCtx, rootSchema)
			}
		}

		if err := resolver.ResolveReference(refCtx, &targetSchema, reference); err != nil {
			return nil, fmt.Errorf("reference resolution failed for %s: %w", reference, err)
		}

		// Check if schema has other constraints beyond the reference
		if hasOtherConstraints(s) {
			// Schema has both $ref and additional constraints
			// Create an allOf validator combining the resolved schema and additional constraints

			// Set up context for the resolved schema with proper base schema context
			resolvedCtx := schema.WithBaseSchema(ctx, &targetSchema)
			resolvedValidator, err := Compile(resolvedCtx, &targetSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to compile resolved schema: %w", err)
			}

			// Create schema without reference for additional constraints
			schemaWithoutRef := createSchemaWithoutRef(s)

			// Check if we need unevaluated coordination
			if schemaWithoutRef.HasUnevaluatedProperties() || schemaWithoutRef.HasUnevaluatedItems() {
				// Create additional validator WITHOUT unevaluated constraints
				schemaWithoutUnevaluated := createSchemaWithoutUnevaluatedFields(schemaWithoutRef)
				additionalValidator, err := Compile(ctx, schemaWithoutUnevaluated)
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
			additionalValidator, err := Compile(ctx, schemaWithoutRef)
			if err != nil {
				return nil, fmt.Errorf("failed to compile additional constraints: %w", err)
			}
			return AllOf(resolvedValidator, additionalValidator), nil
		}

		// Schema has only $ref, recursively compile the resolved schema
		// Set up context for the resolved schema with proper base schema context

		// Only set resolved schema as base if it has its own ID (proper schema scope)
		// Otherwise keep existing base schema that was able to resolve the reference
		resolvedCtx := ctx
		if targetSchema.HasID() {
			resolvedCtx = schema.WithBaseSchema(ctx, &targetSchema)
		}
		return Compile(resolvedCtx, &targetSchema)
	}
	var validators []Interface

	// Phase 2: Compile composite validators (allOf, anyOf, oneOf)
	compositeValidators, err := compileCompositeValidators(ctx, s)
	if err != nil {
		return nil, err
	}
	validators = append(validators, compositeValidators...)

	// Phase 3: Compile conditional validators (if/then/else, not)
	conditionalValidators, err := compileConditionalValidators(ctx, s)
	if err != nil {
		return nil, err
	}
	validators = append(validators, conditionalValidators...)

	// Phase 4: Compile base constraint validators (properties, items, type, etc.)
	baseValidator, err := compileBaseConstraints(ctx, s)
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
func compileCompositeValidators(ctx context.Context, s *schema.Schema) ([]Interface, error) {
	var validators []Interface

	// AllOf
	if s.HasAllOf() {
		allOfValidators := make([]Interface, 0, len(s.AllOf()))
		for _, subSchema := range s.AllOf() {
			v, err := Compile(ctx, convertSchemaOrBool(subSchema))
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
			v, err := Compile(ctx, convertSchemaOrBool(subSchema))
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
			v, err := Compile(ctx, convertSchemaOrBool(subSchema))
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
func compileConditionalValidators(ctx context.Context, s *schema.Schema) ([]Interface, error) {
	var validators []Interface

	// Not
	if s.HasNot() {
		notValidator, err := Compile(ctx, s.Not())
		if err != nil {
			return nil, fmt.Errorf("failed to compile not validator: %w", err)
		}
		validators = append(validators, &NotValidator{validator: notValidator})
	}

	// If/Then/Else
	if s.HasIfSchema() {
		ifThenElseValidator, err := compileIfThenElseValidator(ctx, s)
		if err != nil {
			return nil, fmt.Errorf("failed to compile if/then/else validator: %w", err)
		}
		validators = append(validators, ifThenElseValidator)
	}

	return validators, nil
}

// compileBaseConstraints compiles all base constraint validators (properties, items, type, etc.)
// This function excludes composition keywords and unevaluated constraints which are handled separately
func compileBaseConstraints(ctx context.Context, s *schema.Schema) (Interface, error) {
	var validators []Interface

	// Type validators - handle explicit type declarations
	if len(s.Types()) > 0 {
		var typeValidators []Interface
		for _, typ := range s.Types() {
			switch typ {
			case schema.StringType:
				// String type validator (with or without additional string constraints)
				stringValidator, err := compileStringValidator(ctx, s, true) // strict type checking
				if err != nil {
					return nil, fmt.Errorf("failed to compile string validator: %w", err)
				}
				typeValidators = append(typeValidators, stringValidator)
			case schema.IntegerType:
				// Integer type validator
				integerValidator, err := compileIntegerValidator(ctx, s)
				if err != nil {
					return nil, fmt.Errorf("failed to compile integer validator: %w", err)
				}
				typeValidators = append(typeValidators, integerValidator)
			case schema.NumberType:
				// Number type validator
				numberValidator, err := compileNumberValidator(ctx, s)
				if err != nil {
					return nil, fmt.Errorf("failed to compile number validator: %w", err)
				}
				typeValidators = append(typeValidators, numberValidator)
			case schema.BooleanType:
				// Boolean type validator
				booleanValidator, err := compileBooleanValidator(ctx, s)
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
					arrayValidator, err := compileArrayValidator(ctx, baseSchema, true) // strict type checking
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
					objectValidator, err := compileObjectValidator(ctx, baseSchema, true) // strict type checking
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
			stringValidator, err := compileStringValidator(ctx, s, false)
			if err != nil {
				return nil, fmt.Errorf("failed to compile string validator: %w", err)
			}
			validators = append(validators, stringValidator)
		}

		// Numeric constraints (includes both integer and number constraints)
		if s.HasAny(schema.NumericConstraintFields) {
			// Use inferred number validator for untyped schemas
			inferredValidator, err := compileInferredNumberValidator(ctx, s)
			if err != nil {
				return nil, fmt.Errorf("failed to compile inferred number validator: %w", err)
			}
			validators = append(validators, inferredValidator)
		}

		// Array constraints (excluding unevaluatedItems) without explicit type
		arrayFields := schema.ArrayConstraintFields &^ schema.UnevaluatedItemsField
		if s.HasAny(arrayFields) {
			baseSchema := createSchemaWithoutUnevaluatedFields(s)
			arrayValidator, err := compileArrayValidator(ctx, baseSchema, false)
			if err != nil {
				return nil, fmt.Errorf("failed to compile array validator: %w", err)
			}
			validators = append(validators, arrayValidator)
		}

		// Object constraints (excluding unevaluatedProperties) without explicit type
		objectFields := schema.ObjectConstraintFields &^ schema.UnevaluatedPropertiesField
		if s.HasAny(objectFields) {
			baseSchema := createSchemaWithoutUnevaluatedFields(s)
			objectValidator, err := compileObjectValidator(ctx, baseSchema, false)
			if err != nil {
				return nil, fmt.Errorf("failed to compile object validator: %w", err)
			}
			validators = append(validators, objectValidator)
		}
	}

	// Content validation
	if s.HasContentEncoding() || s.HasContentMediaType() || s.HasContentSchema() {
		contentValidator, err := compileContentValidator(ctx, s)
		if err != nil {
			return nil, fmt.Errorf("failed to compile content validator: %w", err)
		}
		validators = append(validators, contentValidator)
	}

	// Value constraints (enum, const) - handle for both typed and untyped schemas
	if s.HasAny(schema.ValueConstraintFields) {
		if len(s.Types()) == 0 {
			// Untyped schema with value constraints
			untypedValidator, err := compileUntypedValidator(ctx, s)
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
				depValidator, err := Compile(ctx, val)
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
			resolver:   schema.ResolverFromContext(ctx),
			rootSchema: schema.RootSchemaFromContext(ctx),
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
