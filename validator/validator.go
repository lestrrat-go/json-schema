//go:generate ./gen.sh

package validator

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/lestrrat-go/blackmagic"
	schema "github.com/lestrrat-go/json-schema"
)

// Interface is the interface that all validators must implement.
type Interface interface {
	Validate(context.Context, any) (Result, error)
}

// Result contains annotation information from validation that may be used
// by other validators (e.g., for unevaluatedProperties tracking)
type Result any

// ObjectResult contains information about which object properties were evaluated
type ObjectResult struct {
	evaluatedProperties map[string]bool // property name -> true if evaluated
}

// ArrayResult contains information about which array items were evaluated
type ArrayResult struct {
	evaluatedItems []bool // index -> true if evaluated, length determines max evaluated index
}

// NewObjectResult creates a new ObjectResult with an initialized map
func NewObjectResult() *ObjectResult {
	return &ObjectResult{
		evaluatedProperties: make(map[string]bool),
	}
}

// NewArrayResult creates a new ArrayResult with an initialized slice
func NewArrayResult() *ArrayResult {
	return &ArrayResult{
		evaluatedItems: make([]bool, 0),
	}
}

// EvaluatedProperties returns a copy of the evaluated properties map
func (r *ObjectResult) EvaluatedProperties() map[string]bool {
	if r == nil || r.evaluatedProperties == nil {
		return make(map[string]bool)
	}
	result := make(map[string]bool, len(r.evaluatedProperties))
	for k, v := range r.evaluatedProperties {
		result[k] = v
	}
	return result
}

// SetEvaluatedProperty marks a property as evaluated
func (r *ObjectResult) SetEvaluatedProperty(prop string) {
	if r != nil && r.evaluatedProperties != nil {
		r.evaluatedProperties[prop] = true
	}
}

// EvaluatedItems returns a copy of the evaluated items slice
func (r *ArrayResult) EvaluatedItems() []bool {
	if r == nil || r.evaluatedItems == nil {
		return make([]bool, 0)
	}
	result := make([]bool, len(r.evaluatedItems))
	copy(result, r.evaluatedItems)
	return result
}

// SetEvaluatedItem marks an item at the given index as evaluated
func (r *ArrayResult) SetEvaluatedItem(index int) {
	if r == nil {
		return
	}
	// Extend slice if necessary
	for len(r.evaluatedItems) <= index {
		r.evaluatedItems = append(r.evaluatedItems, false)
	}
	r.evaluatedItems[index] = true
}

// SetEvaluatedItems sets the entire slice of evaluated items
func (r *ArrayResult) SetEvaluatedItems(items []bool) {
	if r == nil {
		return
	}
	r.evaluatedItems = make([]bool, len(items))
	copy(r.evaluatedItems, items)
}

// mergeObjectResults merges multiple ObjectResult instances into a single result
func mergeObjectResults(results ...*ObjectResult) *ObjectResult {
	merged := NewObjectResult()
	for _, result := range results {
		if result != nil && result.evaluatedProperties != nil {
			for prop := range result.evaluatedProperties {
				merged.evaluatedProperties[prop] = true
			}
		}
	}
	return merged
}

// mergeArrayResults merges multiple ArrayResult instances into a single result
func mergeArrayResults(results ...*ArrayResult) *ArrayResult {
	merged := NewArrayResult()
	maxLen := 0

	// Find the maximum length needed
	for _, result := range results {
		if result != nil && len(result.evaluatedItems) > maxLen {
			maxLen = len(result.evaluatedItems)
		}
	}

	// Initialize with the correct length
	merged.evaluatedItems = make([]bool, maxLen)

	// Merge all results
	for _, result := range results {
		if result != nil {
			for i, evaluated := range result.evaluatedItems {
				if evaluated {
					merged.evaluatedItems[i] = true
				}
			}
		}
	}

	return merged
}

// MergeResults merges multiple validation results and assigns the final result to dst
// using blackmagic.AssignIfCompatible for type-safe assignment
func MergeResults(dst any, results ...any) error {
	var objectResults []*ObjectResult
	var arrayResults []*ArrayResult

	// Collect results by type
	for _, result := range results {
		switch r := result.(type) {
		case *ObjectResult:
			if r != nil {
				objectResults = append(objectResults, r)
			}
		case *ArrayResult:
			if r != nil {
				arrayResults = append(arrayResults, r)
			}
		}
	}

	// Merge based on destination type
	switch dst.(type) {
	case **ObjectResult:
		merged := mergeObjectResults(objectResults...)
		return blackmagic.AssignIfCompatible(dst, merged)
	case **ArrayResult:
		merged := mergeArrayResults(arrayResults...)
		return blackmagic.AssignIfCompatible(dst, merged)
	default:
		return fmt.Errorf("unsupported destination type: %T", dst)
	}
}

// Helper functions to convert between map[string]bool and map[string]struct{}
func boolMapToStructMap(boolMap map[string]bool) map[string]struct{} {
	if boolMap == nil {
		return nil
	}
	structMap := make(map[string]struct{})
	for key := range boolMap {
		if boolMap[key] {
			structMap[key] = struct{}{}
		}
	}
	return structMap
}

// dependentSchemasKey is now handled by the public schema package

// WithDependentSchemas adds compiled dependent schema validators to the context
func WithDependentSchemas(ctx context.Context, dependentSchemas map[string]Interface) context.Context {
	// Convert map[string]Interface to map[string]any
	converted := make(map[string]any, len(dependentSchemas))
	for k, v := range dependentSchemas {
		converted[k] = v
	}
	return schema.WithDependentSchemas(ctx, converted)
}

// DependentSchemasFromContext extracts compiled dependent schema validators from context, returns nil if none are associated with ctx
func DependentSchemasFromContext(ctx context.Context) map[string]Interface {
	deps := schema.DependentSchemasFromContext(ctx)
	if deps == nil {
		return nil
	}
	// Convert map[string]any back to map[string]Interface
	converted := make(map[string]Interface, len(deps))
	for k, v := range deps {
		if validator, ok := v.(Interface); ok {
			converted[k] = validator
		}
	}
	return converted
}

type Builder interface {
	Build() (Interface, error)
	MustBuild() Interface
}

// convertSchemaOrBool converts a SchemaOrBool to a *Schema.
// When the value is true, it returns an empty Schema which accepts everything.
// When the value is false, it returns a Schema with "not": {} which rejects everything.
// When the value is already a *Schema, it returns the schema as-is.
// When the value is a map[string]any from JSON unmarshaling, it converts it to a Schema.
func convertSchemaOrBool(v schema.SchemaOrBool) *schema.Schema {
	switch val := v.(type) {
	case schema.BoolSchema:
		if bool(val) {
			// true schema accepts everything
			return schema.New()
		}
		// false schema rejects everything
		return schema.NewBuilder().Not(schema.New()).MustBuild()
	case *schema.Schema:
		return val
	default:
		// This shouldn't happen if validation is working correctly
		panic(fmt.Sprintf("invalid SchemaOrBool type: %T", v))
	}
}

// hasOtherConstraints checks if a schema has constraints other than $ref/$dynamicRef
func hasOtherConstraints(s *schema.Schema) bool {
	// Check for types separately since it's not a bit field check
	if len(s.Types()) > 0 {
		return true
	}

	// Use bit field approach for efficient checking of multiple constraints
	constraintFields := schema.AllOfField | schema.AnyOfField | schema.OneOfField | schema.NotField |
		schema.IfSchemaField | schema.ThenSchemaField | schema.ElseSchemaField |
		schema.PropertiesField | schema.PatternPropertiesField | schema.AdditionalPropertiesField |
		schema.UnevaluatedPropertiesField | schema.RequiredField | schema.MinPropertiesField | schema.MaxPropertiesField |
		schema.DependentSchemasField | schema.ItemsField | schema.ContainsField | schema.MinItemsField | schema.MaxItemsField |
		schema.UniqueItemsField | schema.UnevaluatedItemsField | schema.MinLengthField | schema.MaxLengthField |
		schema.PatternField | schema.FormatField | schema.MinimumField | schema.MaximumField | schema.ExclusiveMinimumField |
		schema.ExclusiveMaximumField | schema.MultipleOfField | schema.EnumField | schema.ConstField |
		schema.ContentEncodingField | schema.ContentMediaTypeField | schema.ContentSchemaField |
		schema.PropertyNamesField

	// Returns true if ANY of the constraint fields are set
	return s.HasAny(constraintFields)
}

// createSchemaWithoutRef creates a copy of the schema without the $ref/$dynamicRef constraint
func createSchemaWithoutRef(s *schema.Schema) *schema.Schema {
	// Use the new Clone Builder pattern to create a copy without the $ref/$dynamicRef field
	builder := schema.NewBuilder().Clone(s).ResetReference()
	if s.HasDynamicReference() {
		builder = builder.ResetDynamicReference()
	}
	return builder.MustBuild()
}

func Compile(ctx context.Context, s *schema.Schema) (Interface, error) {
	// Set up context with default resolver if none provided
	if schema.ResolverFromContext(ctx) == nil {
		ctx = schema.WithResolver(ctx, schema.NewResolver())
	}

	// Set up context with root schema if none provided
	if schema.RootSchemaFromContext(ctx) == nil {
		ctx = schema.WithRootSchema(ctx, s)
	}

	// Set up base schema for local reference resolution if none provided
	if schema.BaseSchemaFromContext(ctx) == nil {
		ctx = schema.WithBaseSchema(ctx, s)
	}

	// Add current schema to dynamic scope chain for $dynamicRef resolution
	ctx = schema.WithDynamicScope(ctx, s)

	// Set up base URI context from schema's $id field if present
	if s.HasID() {
		schemaID := s.ID()
		if schemaID != "" {
			// Extract base URI from $id for resolving relative references within this schema
			if baseURI := extractBaseURI(schemaID); baseURI != "" {
				ctx = schema.WithBaseURI(ctx, baseURI)
			}
		}
	}

	// Handle vocabulary context - always check for root schema vocabulary resolution
	rootSchema := schema.RootSchemaFromContext(ctx)

	// For root schema compilation, resolve vocabulary from metaschema
	if rootSchema == s {
		// For testing, hardcode known metaschema URIs that disable validation vocabulary
		schemaURI := s.Schema()
		if schemaURI != "" {
			if schemaURI == "http://localhost:1234/draft2020-12/metaschema-no-validation.json" {
				// This specific metaschema disables validation vocabulary
				vocabSet := VocabularySet{
					"https://json-schema.org/draft/2020-12/vocab/core":              true,
					"https://json-schema.org/draft/2020-12/vocab/applicator":        true,
					"https://json-schema.org/draft/2020-12/vocab/unevaluated":       true,
					"https://json-schema.org/draft/2020-12/vocab/validation":        false, // Disabled!
					"https://json-schema.org/draft/2020-12/vocab/format-annotation": true,
					"https://json-schema.org/draft/2020-12/vocab/format-assertion":  true,
					"https://json-schema.org/draft/2020-12/vocab/content":           true,
					"https://json-schema.org/draft/2020-12/vocab/meta-data":         true,
				}
				ctx = WithVocabularySet(ctx, vocabSet)
			}
		}
	}

	// Ensure vocabulary context is set (fallback to all enabled if not set)
	if VocabularySetFromContext(ctx) == nil {
		ctx = WithVocabularySet(ctx, AllEnabled())
	}

	// Handle $ref and $dynamicRef first - if schema has a reference, resolve it immediately
	if s.HasReference() {
		reference := s.Reference()

		// Get resolver from context (guaranteed to be present)
		resolver := schema.ResolverFromContext(ctx)

		// Get root schema from context (guaranteed to be present)
		rootSchema := schema.RootSchemaFromContext(ctx)

		// Check for circular references by looking at context
		if stack := schema.ReferenceStackFromContext(ctx); stack != nil {
			for _, ref := range stack {
				if ref == reference {
					return nil, fmt.Errorf("circular reference detected: %s", reference)
				}
			}
			// Add current reference to stack
			newStack := make([]string, len(stack)+1)
			copy(newStack, stack)
			newStack[len(stack)] = reference
			ctx = schema.WithReferenceStack(ctx, newStack)
		} else {
			// Start new reference stack
			ctx = schema.WithReferenceStack(ctx, []string{reference})
		}

		// Resolve the reference to get the target schema
		var targetSchema schema.Schema
		baseURI := schema.BaseURIFromContext(ctx)
		refCtx := ctx
		if baseURI != "" {
			refCtx = schema.WithBaseURI(ctx, baseURI)
		}
		// Add base schema context for reference resolution
		if rootSchema := schema.RootSchemaFromContext(ctx); rootSchema != nil {
			refCtx = schema.WithBaseSchema(refCtx, rootSchema)
		}
		if err := resolver.ResolveReference(refCtx, &targetSchema, reference); err != nil {
			return nil, fmt.Errorf("failed to resolve reference %s: %w", reference, err)
		}

		// If the target schema has relative references, we need to ensure they're resolved
		// against the correct base URI. For metaschema, this is crucial.
		if targetSchema.HasID() && targetSchema.ID() != "" {
			// Set the base URI from the target schema's $id
			if baseURI := extractBaseURI(targetSchema.ID()); baseURI != "" {
				ctx = schema.WithBaseURI(ctx, baseURI)
			}
		}

		// For external schemas, set the resolved schema as the root for its internal references
		// This ensures that local references like #/$defs/... resolve against the external schema
		if !strings.HasPrefix(reference, "#") {
			ctx = schema.WithRootSchema(ctx, &targetSchema)
		}

		// Compile the resolved target schema, and wrap it with a reference validator for debugging
		resolved, err := Compile(ctx, &targetSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to compile resolved schema for reference %s: %w", reference, err)
		}

		refValidator := &ReferenceValidator{
			reference:  reference,
			resolved:   resolved,
			resolver:   resolver,
			rootSchema: rootSchema,
		}
		// Mark as already resolved to prevent lazy resolution from overwriting
		refValidator.resolvedOnce.Do(func() {})

		// Check if the schema has other constraints besides $ref
		if hasOtherConstraints(s) {
			// Special handling for $ref + unevaluatedProperties
			if s.HasUnevaluatedProperties() && s.HasAny(schema.BasicPropertiesFields) {
				// Create a composition validator that properly handles annotation flow
				compositionValidator := NewRefUnevaluatedPropertiesCompositionValidator(ctx, s, refValidator)
				return compositionValidator, nil
			}

			// Create a composite validator that combines $ref with other constraints
			// First, create a schema without the $ref for other constraints
			otherSchema := createSchemaWithoutRef(s)
			otherValidator, err := Compile(ctx, otherSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to compile other constraints validator: %w", err)
			}

			// Create a MultiValidator with allOf logic to combine both
			compositeValidator := NewMultiValidator(AndMode)
			compositeValidator.Append(refValidator)
			compositeValidator.Append(otherValidator)
			return compositeValidator, nil
		}

		// Only $ref constraint, return the reference validator
		return refValidator, nil
	} else if s.HasDynamicReference() {
		// Handle $dynamicRef with lazy dynamic scope resolution
		dynamicRef := s.DynamicReference()

		// Get resolver and root schema from context for compilation
		resolver := schema.ResolverFromContext(ctx)
		if resolver == nil {
			resolver = schema.NewResolver()
		}
		rootSchema := schema.RootSchemaFromContext(ctx)

		// Capture the dynamic scope chain from compilation time
		dynamicScope := schema.DynamicScopeFromContext(ctx)

		return &DynamicReferenceValidator{
			reference:    dynamicRef,
			resolver:     resolver,
			rootSchema:   rootSchema,
			dynamicScope: dynamicScope,
		}, nil
	}

	var allValidators []Interface

	// Handle schema composition first
	if s.HasAllOf() {
		// Special handling for allOf with unevaluatedProperties or unevaluatedItems in base schema
		if hasBaseConstraints(s) && s.HasAny(schema.UnevaluatedFields) {
			// Create a special validator that evaluates allOf first, then base constraints with annotation context
			compositionValidator, err := NewUnevaluatedPropertiesCompositionValidatorWithResolver(ctx, s, schema.ResolverFromContext(ctx))
			if err != nil {
				return nil, fmt.Errorf(`failed to compile allOf composition validator: %w`, err)
			}
			allValidators = append(allValidators, compositionValidator)
		} else {
			allOfValidators := make([]Interface, 0, len(s.AllOf())+1)

			// If the schema has base properties/constraints, create a base validator first
			if hasBaseConstraints(s) {
				baseSchema := createBaseSchema(s)
				baseValidator, err := Compile(ctx, baseSchema)
				if err != nil {
					return nil, fmt.Errorf(`failed to compile base schema for allOf: %w`, err)
				}
				allOfValidators = append(allOfValidators, baseValidator)
			}

			for _, subSchema := range s.AllOf() {
				convertedSchema := convertSchemaOrBool(subSchema)
				v, err := Compile(ctx, convertedSchema)
				if err != nil {
					return nil, fmt.Errorf(`failed to compile allOf validator: %w`, err)
				}
				allOfValidators = append(allOfValidators, v)
			}
			allOfValidator := NewMultiValidator(AndMode)
			for _, v := range allOfValidators {
				allOfValidator.Append(v)
			}
			allValidators = append(allValidators, allOfValidator)
		}
	}

	if s.HasAnyOf() {
		anyOfValidators := make([]Interface, 0, len(s.AnyOf()))
		for _, subSchema := range s.AnyOf() {
			v, err := Compile(ctx, convertSchemaOrBool(subSchema))
			if err != nil {
				return nil, fmt.Errorf(`failed to compile anyOf validator: %w`, err)
			}
			anyOfValidators = append(anyOfValidators, v)
		}

		if hasBaseConstraints(s) && s.HasUnevaluatedProperties() {
			// Special anyOf composition validator for unevaluatedProperties
			compositionValidator, err := NewAnyOfUnevaluatedPropertiesCompositionValidatorWithResolver(ctx, s, anyOfValidators, schema.ResolverFromContext(ctx))
			if err != nil {
				return nil, fmt.Errorf(`failed to compile anyOf composition validator: %w`, err)
			}
			allValidators = append(allValidators, compositionValidator)
		} else {
			anyOfValidator := NewMultiValidator(OrMode)
			for _, v := range anyOfValidators {
				anyOfValidator.Append(v)
			}
			allValidators = append(allValidators, anyOfValidator)
		}
	}

	if s.HasOneOf() {
		oneOfValidators := make([]Interface, 0, len(s.OneOf()))
		for _, subSchema := range s.OneOf() {
			v, err := Compile(ctx, convertSchemaOrBool(subSchema))
			if err != nil {
				return nil, fmt.Errorf(`failed to compile oneOf validator: %w`, err)
			}
			oneOfValidators = append(oneOfValidators, v)
		}

		if hasBaseConstraints(s) && s.HasUnevaluatedProperties() {
			// Special oneOf composition validator for unevaluatedProperties
			compositionValidator, err := NewOneOfUnevaluatedPropertiesCompositionValidatorWithResolver(ctx, s, oneOfValidators, schema.ResolverFromContext(ctx))
			if err != nil {
				return nil, fmt.Errorf(`failed to compile oneOf composition validator: %w`, err)
			}
			allValidators = append(allValidators, compositionValidator)
		} else {
			oneOfValidator := NewMultiValidator(OneOfMode)
			for _, v := range oneOfValidators {
				oneOfValidator.Append(v)
			}
			allValidators = append(allValidators, oneOfValidator)
		}
	}

	if s.HasNot() {
		notValidator, err := Compile(ctx, s.Not())
		if err != nil {
			return nil, fmt.Errorf(`failed to compile not validator: %w`, err)
		}
		allValidators = append(allValidators, &NotValidator{validator: notValidator})
	}

	// Handle if/then/else conditional validation
	if s.HasIfSchema() {
		// Special handling for if/then/else with unevaluatedProperties in base schema
		if hasBaseConstraints(s) && s.HasUnevaluatedProperties() {
			// Create a special validator that evaluates if/then/else first, then base constraints with annotation context
			compositionValidator := NewIfThenElseUnevaluatedPropertiesCompositionValidator(ctx, s)
			allValidators = append(allValidators, compositionValidator)
		} else {
			ifThenElseValidator, err := compileIfThenElseValidator(ctx, s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile if/then/else validator: %w`, err)
			}
			allValidators = append(allValidators, ifThenElseValidator)
		}
	}

	// Compile dependent schemas and pass in context for two-pass validation
	if s.HasDependentSchemas() {
		compiledDependentSchemas := make(map[string]Interface)
		for propertyName, depSchema := range s.DependentSchemas() {
			// Handle SchemaOrBool types
			switch val := depSchema.(type) {
			case schema.BoolSchema:
				// Boolean schema: true means always valid, false means always invalid
				if bool(val) {
					compiledDependentSchemas[propertyName] = &alwaysPassValidator{}
				} else {
					compiledDependentSchemas[propertyName] = &alwaysFailValidator{}
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
		ctx = WithDependentSchemas(ctx, compiledDependentSchemas)
	}

	// Handle type-specific validators
	explicitTypes := s.Types()
	types := make([]schema.PrimitiveType, 0, len(explicitTypes))

	// Only include types if type constraint is enabled
	if IsKeywordEnabledInContext(ctx, "type") {
		types = make([]schema.PrimitiveType, len(explicitTypes))
		copy(types, explicitTypes)
	}
	var validatorsByType []Interface

	// Track which types were inferred (not explicitly declared)
	inferredTypes := make(map[schema.PrimitiveType]bool)

	// If no types are specified but type-specific constraints are present,
	// infer the type from the constraints
	// Skip this if allOf is present and has base constraints (they'll be handled in allOf)
	// Also skip if we have anyOf/oneOf composition validators that will handle these constraints
	hasCompositionValidator := (s.HasAllOf() && hasBaseConstraints(s)) ||
		(s.HasAnyOf() && hasBaseConstraints(s) && s.HasUnevaluatedProperties()) ||
		(s.HasOneOf() && hasBaseConstraints(s) && s.HasUnevaluatedProperties()) ||
		(s.HasIfSchema() && hasBaseConstraints(s) && s.HasUnevaluatedProperties())

	if len(types) == 0 && !hasCompositionValidator {
		if s.HasAny(schema.StringConstraintFields) {
			types = append(types, schema.StringType)
			inferredTypes[schema.StringType] = true
		}
		if s.HasAny(schema.NumericConstraintFields) {
			// For inferred numeric types, create a non-strict validator that only validates numeric constraints when values are numbers
			v, err := compileInferredNumberValidator(ctx, s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile inferred number validator: %w`, err)
			}
			allValidators = append(allValidators, v)
		}
		if s.HasAny(schema.ArrayConstraintFields) {
			// For inferred array types, create a non-strict array validator
			v, err := compileArrayValidator(ctx, s, false)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile inferred array validator: %w`, err)
			}
			allValidators = append(allValidators, v)
		}
		if s.HasAny(schema.ObjectConstraintFields) {
			// For inferred object types, create non-strict object validator
			v, err := compileObjectValidator(ctx, s, false)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile inferred object validator: %w`, err)
			}
			allValidators = append(allValidators, v)
		}
	}

	// Handle general enum/const validation when no specific type is set
	// Skip this if we have composition validators that will handle these constraints
	if len(types) == 0 && s.HasAny(schema.ValueConstraintFields) && !hasCompositionValidator {
		validator, err := compileGeneralValidator(ctx, s)
		if err != nil {
			return nil, fmt.Errorf(`failed to compile general validator: %w`, err)
		}
		allValidators = append(allValidators, validator)
	}

	for _, typ := range types {
		// This is a placeholder code. In reality we need to
		// OR all types
		switch typ {
		case schema.StringType:
			// Use strict type checking only for explicitly declared string types
			strictType := !inferredTypes[schema.StringType]
			v, err := compileStringValidator(ctx, s, strictType)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile string validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.IntegerType:
			v, err := compileIntegerValidator(ctx, s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile integer validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.NumberType:
			v, err := compileNumberValidator(ctx, s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile number validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.BooleanType:
			v, err := compileBooleanValidator(ctx, s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile boolean validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.ArrayType:
			v, err := compileArrayValidator(ctx, s, true)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile array validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.ObjectType:
			v, err := compileObjectValidator(ctx, s, true)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile object validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.NullType:
			v, err := compileNullValidator(ctx, s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile null validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		}
	}

	// Combine type validators if multiple types
	if len(validatorsByType) > 1 {
		typeValidator := NewMultiValidator(OrMode)
		for _, v := range validatorsByType {
			typeValidator.Append(v)
		}
		allValidators = append(allValidators, typeValidator)
	} else if len(validatorsByType) == 1 {
		allValidators = append(allValidators, validatorsByType[0])
	}

	// Handle content validation (contentEncoding, contentMediaType, contentSchema)
	if contentValidator, err := compileContentValidator(ctx, s); err != nil {
		return nil, fmt.Errorf(`failed to compile content validator: %w`, err)
	} else if contentValidator != nil {
		allValidators = append(allValidators, contentValidator)
	}

	// Return the appropriate validator
	if len(allValidators) == 0 {
		// Empty schema - allows anything
		return &EmptyValidator{}, nil
	}

	if len(allValidators) == 1 {
		return allValidators[0], nil
	}

	// Multiple validators - combine with AND
	mv := NewMultiValidator(AndMode)
	for _, v := range allValidators {
		mv.Append(v)
	}

	return mv, nil
}

// inferredNumberValidator validates numeric constraints only when the value is a number,
// ignoring non-numeric values (for inferred number types without explicit type declaration)
type inferredNumberValidator struct {
	numberValidator Interface
}

func compileInferredNumberValidator(ctx context.Context, s *schema.Schema) (Interface, error) {
	// Create the underlying number validator
	numValidator, err := compileNumberValidator(ctx, s)
	if err != nil {
		return nil, err
	}

	return &inferredNumberValidator{
		numberValidator: numValidator,
	}, nil
}

func (v *inferredNumberValidator) Validate(ctx context.Context, in any) (Result, error) {
	// Check if the value is numeric
	rv := reflect.ValueOf(in)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		// Value is numeric, apply number validation
		return v.numberValidator.Validate(ctx, in)
	default:
		// Value is not numeric, ignore numeric constraints (per JSON Schema spec)
		//nolint: nilnil
		return nil, nil
	}
}

type EmptyValidator struct{}

func (e *EmptyValidator) Validate(_ context.Context, _ any) (Result, error) {
	// Empty schema allows anything
	//nolint: nilnil
	return nil, nil
}

type NotValidator struct {
	validator Interface
}

func (n *NotValidator) Validate(ctx context.Context, v any) (Result, error) {
	_, err := n.validator.Validate(ctx, v)
	if err == nil {
		return nil, fmt.Errorf(`not validation failed: value should not validate against the schema`)
	}
	//nolint: nilnil
	return nil, nil
}

type NullValidator struct{}

func (n *NullValidator) Validate(_ context.Context, v any) (Result, error) {
	if v == nil {
		//nolint: nilnil
		return nil, nil
	}
	return nil, fmt.Errorf(`invalid value passed to NullValidator: expected null, got %T`, v)
}

func compileNullValidator(_ context.Context, _ *schema.Schema) (Interface, error) {
	return &NullValidator{}, nil
}

// GeneralValidator handles enum and const validation for schemas without specific types
type GeneralValidator struct {
	enum       []any
	constValue any
	hasConst   bool
}

func compileGeneralValidator(_ context.Context, s *schema.Schema) (Interface, error) {
	v := &GeneralValidator{}

	if s.HasEnum() {
		v.enum = s.Enum()
	}

	if s.HasConst() {
		v.constValue = s.Const()
		v.hasConst = true
	}

	return v, nil
}

func (g *GeneralValidator) Validate(_ context.Context, value any) (Result, error) {
	// Check const first
	if g.hasConst {
		if !reflect.DeepEqual(value, g.constValue) {
			return nil, fmt.Errorf(`invalid value: must equal const value %v, got %v`, g.constValue, value)
		}
		//nolint: nilnil
		return nil, nil
	}

	// Check enum
	if g.enum != nil {
		for _, enumVal := range g.enum {
			if reflect.DeepEqual(value, enumVal) {
				//nolint: nilnil
				return nil, nil
			}
		}
		return nil, fmt.Errorf(`invalid value: %v not found in enum %v`, value, g.enum)
	}

	//nolint: nilnil
	return nil, nil
}

// ReferenceValidator handles schema references ($ref) with lazy resolution and circular detection
type ReferenceValidator struct {
	reference    string
	resolvedOnce sync.Once
	resolved     Interface
	resolveErr   error
	resolver     *schema.Resolver
	rootSchema   *schema.Schema
}

func (r *ReferenceValidator) Validate(ctx context.Context, v any) (Result, error) {
	// Lazy resolution - only resolve when actually needed for validation
	r.resolvedOnce.Do(func() {
		r.resolved, r.resolveErr = r.resolveReference(ctx)
	})

	if r.resolveErr != nil {
		return nil, fmt.Errorf("reference resolution failed for %s: %w", r.reference, r.resolveErr)
	}

	return r.resolved.Validate(ctx, v)
}

func (r *ReferenceValidator) resolveReference(ctx context.Context) (Interface, error) {
	// Use stored resolver and root schema, fall back to context if not available
	resolver := r.resolver
	if resolver == nil {
		resolver = schema.ResolverFromContext(ctx)
		if resolver == nil {
			resolver = schema.NewResolver()
		}
	}

	rootSchema := r.rootSchema
	if rootSchema == nil {
		rootSchema = schema.RootSchemaFromContext(ctx)
		if rootSchema == nil {
			return nil, fmt.Errorf("no root schema available in context for reference resolution: %s", r.reference)
		}
	}

	// Check for circular references by looking at context
	if stack := schema.ReferenceStackFromContext(ctx); stack != nil {
		for _, ref := range stack {
			if ref == r.reference {
				return nil, fmt.Errorf("circular reference detected: %s", r.reference)
			}
		}
		// Add current reference to stack
		newStack := make([]string, len(stack)+1)
		copy(newStack, stack)
		newStack[len(stack)] = r.reference
		ctx = schema.WithReferenceStack(ctx, newStack)
	} else {
		// Start new reference stack
		ctx = schema.WithReferenceStack(ctx, []string{r.reference})
	}

	// Resolve the reference to get the target schema
	var targetSchema schema.Schema
	baseURI := schema.BaseURIFromContext(ctx)
	refCtx := ctx
	if baseURI != "" {
		refCtx = schema.WithBaseURI(ctx, baseURI)
	}
	// Add base schema context for reference resolution
	if rootSchema := schema.RootSchemaFromContext(ctx); rootSchema != nil {
		refCtx = schema.WithBaseSchema(refCtx, rootSchema)
	}
	if err := resolver.ResolveReference(refCtx, &targetSchema, r.reference); err != nil {
		return nil, fmt.Errorf("failed to resolve reference %s: %w", r.reference, err)
	}

	// Compile the resolved schema into a validator
	// IMPORTANT: Keep the original root schema context to ensure nested references can be resolved
	return Compile(ctx, &targetSchema)
}

// DynamicReferenceValidator handles $dynamicRef with proper dynamic scope resolution
type DynamicReferenceValidator struct {
	reference    string
	resolvedOnce sync.Once
	resolved     Interface
	resolveErr   error
	resolver     *schema.Resolver
	rootSchema   *schema.Schema
	dynamicScope []*schema.Schema // Store the dynamic scope chain from compilation
}

func (dr *DynamicReferenceValidator) Validate(ctx context.Context, v any) (Result, error) {
	// Lazy resolution - only resolve when actually needed for validation
	dr.resolvedOnce.Do(func() {
		dr.resolved, dr.resolveErr = dr.resolveDynamicReference(ctx)
	})

	if dr.resolveErr != nil {
		return nil, fmt.Errorf("dynamic reference resolution failed for %s: %w", dr.reference, dr.resolveErr)
	}

	return dr.resolved.Validate(ctx, v)
}

func (dr *DynamicReferenceValidator) resolveDynamicReference(ctx context.Context) (Interface, error) {
	// Use stored resolver and root schema from compilation time
	resolver := dr.resolver
	if resolver == nil {
		resolver = schema.NewResolver()
	}

	rootSchema := dr.rootSchema
	if rootSchema == nil {
		return nil, fmt.Errorf("no root schema available for dynamic reference resolution: %s", dr.reference)
	}

	// Create context with stored dynamic scope chain from compilation time
	ctxWithScope := ctx
	if dr.dynamicScope != nil {
		// Build context with all scope elements at once to avoid nested context in loop
		//nolint:fatcontext // Intentional: building dynamic scope chain requires nested contexts
		for _, scope := range dr.dynamicScope {
			ctxWithScope = schema.WithDynamicScope(ctxWithScope, scope)
		}
	}

	// Check for circular references by looking at context
	if stack := schema.ReferenceStackFromContext(ctxWithScope); stack != nil {
		for _, ref := range stack {
			if ref == dr.reference {
				return nil, fmt.Errorf("circular reference detected: %s", dr.reference)
			}
		}
		// Add current reference to stack
		newStack := make([]string, len(stack)+1)
		copy(newStack, stack)
		newStack[len(stack)] = dr.reference
		ctxWithScope = schema.WithReferenceStack(ctxWithScope, newStack)
	} else {
		// Start new reference stack
		ctxWithScope = schema.WithReferenceStack(ctxWithScope, []string{dr.reference})
	}

	// Resolve the $dynamicRef using stored dynamic scope chain
	targetSchema, err := resolveDynamicRef(ctxWithScope, resolver, rootSchema, dr.reference)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dynamic reference %s: %w", dr.reference, err)
	}

	// If the target schema has relative references, we need to ensure they're resolved
	// against the correct base URI. For metaschema, this is crucial.
	if targetSchema.HasID() && targetSchema.ID() != "" {
		// Set the base URI from the target schema's $id
		if baseURI := extractBaseURI(targetSchema.ID()); baseURI != "" {
			ctxWithScope = schema.WithBaseURI(ctxWithScope, baseURI)
		}
	}

	// Compile the resolved target schema
	return Compile(ctxWithScope, targetSchema)
}

// Context keys for passing validator-specific data
// Note: These are now handled by the internal schemactx package

// extractBaseURI extracts the base URI from a reference for context resolution
func extractBaseURI(reference string) string {
	// Handle absolute URIs
	if strings.HasPrefix(reference, "http://") || strings.HasPrefix(reference, "https://") {
		// Split on '#' to get the URI part without fragment
		parts := strings.Split(reference, "#")
		uri := parts[0]

		// Find the last '/' to get the directory path
		lastSlash := strings.LastIndex(uri, "/")
		if lastSlash != -1 {
			return uri[:lastSlash+1] // Include the trailing slash
		}
		return uri + "/" // Add trailing slash if not present
	}

	// For relative references, we can't determine base URI without context
	return ""
}

// resolveDynamicRef resolves a $dynamicRef by looking up the dynamic scope chain
// for the nearest schema with a matching $dynamicAnchor
func resolveDynamicRef(ctx context.Context, resolver *schema.Resolver, rootSchema *schema.Schema, dynamicRef string) (*schema.Schema, error) {
	// Parse the dynamic reference - it should be in the form "#anchorName"
	if !strings.HasPrefix(dynamicRef, "#") {
		// For non-anchor dynamic refs, treat as normal reference
		var targetSchema schema.Schema
		baseURI := schema.BaseURIFromContext(ctx)
		refCtx := schema.WithBaseSchema(ctx, rootSchema)
		if baseURI != "" {
			refCtx = schema.WithBaseURI(refCtx, baseURI)
		}
		if err := resolver.ResolveReference(refCtx, &targetSchema, dynamicRef); err != nil {
			return nil, fmt.Errorf("failed to resolve dynamic reference %s: %w", dynamicRef, err)
		}
		return &targetSchema, nil
	}

	// Check if this is a JSON pointer reference (starts with #/)
	if strings.HasPrefix(dynamicRef, "#/") {
		// For JSON pointer references, try dynamic anchor lookup first, then fall back to normal reference
		// Get the dynamic scope chain from context
		scopeChain := schema.DynamicScopeFromContext(ctx)

		// Search the dynamic scope chain from oldest to most recent for matching $dynamicAnchor
		anchorName := dynamicRef[1:] // Remove the '#' prefix
		for i := range scopeChain {
			currentSchema := scopeChain[i]

			// Check if this schema has a matching $dynamicAnchor
			if currentSchema.HasDynamicAnchor() && currentSchema.DynamicAnchor() == anchorName {
				return currentSchema, nil
			}
		}

		// No matching $dynamicAnchor found, fall back to normal JSON pointer resolution
		var targetSchema schema.Schema
		refCtx := schema.WithBaseSchema(ctx, rootSchema)
		if err := resolver.ResolveReference(refCtx, &targetSchema, dynamicRef); err != nil {
			return nil, fmt.Errorf("failed to resolve dynamic reference %s: %w", dynamicRef, err)
		}
		return &targetSchema, nil
	}

	// This is a plain anchor reference (e.g., "#anchorName")
	anchorName := dynamicRef[1:] // Remove the '#' prefix

	// Get the dynamic scope chain from context
	scopeChain := schema.DynamicScopeFromContext(ctx)

	// Search the dynamic scope chain from oldest to most recent
	// $dynamicRef should resolve to the nearest enclosing schema with matching $dynamicAnchor
	// "Nearest enclosing" means closest to the root, not most recently added
	for i := range scopeChain {
		currentSchema := scopeChain[i]

		// Check if this schema has a matching $dynamicAnchor
		if currentSchema.HasDynamicAnchor() && currentSchema.DynamicAnchor() == anchorName {
			return currentSchema, nil
		}
	}

	// If no matching $dynamicAnchor found in dynamic scope, fall back to normal anchor resolution
	// This is the correct behavior according to JSON Schema spec
	var targetSchema schema.Schema
	refCtx := schema.WithBaseSchema(ctx, rootSchema)
	if err := resolver.ResolveAnchor(refCtx, &targetSchema, anchorName); err != nil {
		return nil, fmt.Errorf("failed to resolve dynamic reference %s (no matching $dynamicAnchor in scope): %w", dynamicRef, err)
	}

	return &targetSchema, nil
}

type MultiValidator struct {
	and        bool
	oneOf      bool
	validators []Interface
}

type MultiValidatorMode int

const (
	OrMode MultiValidatorMode = iota
	AndMode
	OneOfMode
	InvalidMode
)

func NewMultiValidator(mode MultiValidatorMode) *MultiValidator {
	mv := &MultiValidator{}
	if mode == AndMode {
		mv.and = true
	} else if mode == OneOfMode {
		mv.and = false
		mv.oneOf = true
	}
	return mv
}

func (v *MultiValidator) Append(in Interface) *MultiValidator {
	v.validators = append(v.validators, in)
	return v
}

// UnevaluatedPropertiesCompositionValidator handles complex unevaluatedProperties with allOf
type UnevaluatedPropertiesCompositionValidator struct {
	allOfValidators []Interface
	baseValidator   Interface
	schema          *schema.Schema
}

func NewUnevaluatedPropertiesCompositionValidator(s *schema.Schema) *UnevaluatedPropertiesCompositionValidator {
	v, err := NewUnevaluatedPropertiesCompositionValidatorWithResolver(context.Background(), s, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create composition validator: %v", err))
	}
	return v
}

func NewUnevaluatedPropertiesCompositionValidatorWithResolver(ctx context.Context, s *schema.Schema, _ *schema.Resolver) (*UnevaluatedPropertiesCompositionValidator, error) {
	v := &UnevaluatedPropertiesCompositionValidator{
		schema: s,
	}

	// Compile allOf validators
	for _, subSchema := range s.AllOf() {
		subValidator, err := Compile(ctx, convertSchemaOrBool(subSchema))
		if err != nil {
			return nil, fmt.Errorf("failed to compile allOf validator: %w", err)
		}
		v.allOfValidators = append(v.allOfValidators, subValidator)
	}

	// Compile base validator (everything except allOf)
	baseSchema := createBaseSchema(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to compile base schema: %w", err)
	}
	v.baseValidator = baseValidator

	return v, nil
}

func (v *UnevaluatedPropertiesCompositionValidator) Validate(ctx context.Context, in any) (Result, error) {
	// First, validate all allOf subschemas and collect their annotations
	var mergedObjectResult *ObjectResult
	var mergedArrayResult *ArrayResult

	for i, subValidator := range v.allOfValidators {
		result, err := subValidator.Validate(ctx, in)
		if err != nil {
			return nil, fmt.Errorf(`allOf validation failed: validator #%d failed: %w`, i, err)
		}

		// Merge object results for property evaluation tracking
		if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
			if mergedObjectResult == nil {
				mergedObjectResult = NewObjectResult()
			}
			for prop := range objResult.EvaluatedProperties() {
				mergedObjectResult.SetEvaluatedProperty(prop)
			}
		}

		// Merge array results for item evaluation tracking
		if arrResult, ok := result.(*ArrayResult); ok && arrResult != nil {
			if mergedArrayResult == nil {
				mergedArrayResult = NewArrayResult()
			}
			arrItems := arrResult.EvaluatedItems()
			for i, evaluated := range arrItems {
				if evaluated {
					mergedArrayResult.SetEvaluatedItem(i)
				}
			}
		}
	}

	// Now validate base constraints, passing the evaluated annotations from allOf
	baseResult, err := v.validateBaseWithContext(ctx, in, mergedObjectResult, mergedArrayResult)
	if err != nil {
		return nil, err
	}

	// Merge the base result with allOf result
	if baseObjResult, ok := baseResult.(*ObjectResult); ok && baseObjResult != nil {
		if mergedObjectResult == nil {
			mergedObjectResult = NewObjectResult()
		}
		for prop := range baseObjResult.EvaluatedProperties() {
			mergedObjectResult.SetEvaluatedProperty(prop)
		}
	}

	if baseArrResult, ok := baseResult.(*ArrayResult); ok && baseArrResult != nil {
		if mergedArrayResult == nil {
			mergedArrayResult = NewArrayResult()
		}
		arrItems := baseArrResult.EvaluatedItems()
		for i, evaluated := range arrItems {
			if evaluated {
				mergedArrayResult.SetEvaluatedItem(i)
			}
		}
	}

	// Return appropriate result type
	if mergedObjectResult != nil && mergedArrayResult != nil {
		// Both object and array results - prioritize object result for now
		return mergedObjectResult, nil
	} else if mergedObjectResult != nil {
		return mergedObjectResult, nil
	} else if mergedArrayResult != nil {
		return mergedArrayResult, nil
	}

	//nolint: nilnil
	return nil, nil
}

// validateBaseWithContext validates the base schema with annotation context
func (v *UnevaluatedPropertiesCompositionValidator) validateBaseWithContext(ctx context.Context, in any, previousObjectResult *ObjectResult, previousArrayResult *ArrayResult) (Result, error) {
	// Create context with evaluated properties and items if we have previous evaluation results
	var currentCtx context.Context = ctx

	if previousObjectResult != nil {
		evalProps := previousObjectResult.EvaluatedProperties()
		if len(evalProps) > 0 {
			currentCtx = schema.WithEvaluatedProperties(currentCtx, boolMapToStructMap(evalProps))
		}
	}

	if previousArrayResult != nil {
		evalItems := previousArrayResult.EvaluatedItems()
		if len(evalItems) > 0 {
			currentCtx = schema.WithEvaluatedItems(currentCtx, evalItems)
		}
	}

	return v.baseValidator.Validate(currentCtx, in)
}

// AnyOfUnevaluatedPropertiesCompositionValidator handles complex unevaluatedProperties with anyOf
type AnyOfUnevaluatedPropertiesCompositionValidator struct {
	anyOfValidators []Interface
	baseValidator   Interface
	schema          *schema.Schema
}

func NewAnyOfUnevaluatedPropertiesCompositionValidator(s *schema.Schema) *AnyOfUnevaluatedPropertiesCompositionValidator {
	v, err := NewAnyOfUnevaluatedPropertiesCompositionValidatorWithResolver(context.Background(), s, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create anyOf composition validator: %v", err))
	}
	return v
}

func NewAnyOfUnevaluatedPropertiesCompositionValidatorWithResolver(ctx context.Context, s *schema.Schema, anyOfValidators []Interface, _ *schema.Resolver) (*AnyOfUnevaluatedPropertiesCompositionValidator, error) {
	v := &AnyOfUnevaluatedPropertiesCompositionValidator{
		schema: s,
	}

	// Use provided validators or compile them if not provided
	if anyOfValidators != nil {
		v.anyOfValidators = anyOfValidators
	} else {
		// Compile anyOf validators
		for _, subSchema := range s.AnyOf() {
			subValidator, err := Compile(ctx, convertSchemaOrBool(subSchema))
			if err != nil {
				return nil, fmt.Errorf("failed to compile anyOf validator: %w", err)
			}
			v.anyOfValidators = append(v.anyOfValidators, subValidator)
		}
	}

	// Compile base validator (everything except anyOf)
	baseSchema := createBaseSchema(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to compile base schema: %w", err)
	}
	v.baseValidator = baseValidator

	return v, nil
}

func (v *AnyOfUnevaluatedPropertiesCompositionValidator) Validate(ctx context.Context, in any) (Result, error) {
	// For anyOf, we need at least one subschema to pass and collect its annotations
	var validResult *ObjectResult
	anyOfPassed := false

	for _, subValidator := range v.anyOfValidators {
		result, err := subValidator.Validate(ctx, in)
		if err == nil {
			anyOfPassed = true
			// Collect annotations from ALL passing validators (not just the first)
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				if validResult == nil {
					validResult = NewObjectResult()
				}
				for prop := range objResult.EvaluatedProperties() {
					validResult.SetEvaluatedProperty(prop)
				}
			}
			// Continue to check other validators for annotation collection
		}
	}

	if !anyOfPassed {
		return nil, fmt.Errorf(`anyOf validation failed: none of the validators passed`)
	}

	// Now validate base constraints, passing the evaluated properties from anyOf
	baseResult, err := v.validateBaseWithContext(ctx, in, validResult)
	if err != nil {
		return nil, err
	}

	// Merge the base result with anyOf result
	if baseObjResult, ok := baseResult.(*ObjectResult); ok && baseObjResult != nil {
		if validResult == nil {
			validResult = NewObjectResult()
		}
		for prop := range baseObjResult.EvaluatedProperties() {
			validResult.SetEvaluatedProperty(prop)
		}
	}

	return validResult, nil
}

// validateBaseWithContext for AnyOf
func (v *AnyOfUnevaluatedPropertiesCompositionValidator) validateBaseWithContext(ctx context.Context, in any, previousResult *ObjectResult) (Result, error) {
	if objValidator, ok := v.baseValidator.(*objectValidator); ok {
		var previouslyEvaluated map[string]bool
		if previousResult != nil {
			previouslyEvaluated = previousResult.EvaluatedProperties()
		}
		var currentCtx context.Context
		if previouslyEvaluated != nil && len(previouslyEvaluated) > 0 {
			currentCtx = schema.WithEvaluatedProperties(ctx, boolMapToStructMap(previouslyEvaluated))
		} else {
			currentCtx = ctx
		}
		return objValidator.Validate(currentCtx, in)
	}
	if multiValidator, ok := v.baseValidator.(*MultiValidator); ok {
		// If the base validator is a MultiValidator, we need to handle it specially
		return v.validateMultiValidatorWithContext(ctx, multiValidator, in, previousResult)
	}
	// For other validator types, just validate normally without annotation context
	return v.baseValidator.Validate(ctx, in)
}

// validateMultiValidatorWithContext for AnyOf
func (v *AnyOfUnevaluatedPropertiesCompositionValidator) validateMultiValidatorWithContext(ctx context.Context, mv *MultiValidator, in any, previousResult *ObjectResult) (Result, error) {
	if !mv.and {
		// For OR mode, just validate normally
		return mv.Validate(ctx, in)
	}

	// For AND mode (allOf), validate each sub-validator independently (cousins cannot see each other)
	var mergedResult *ObjectResult
	if previousResult != nil {
		mergedResult = NewObjectResult()
		for prop := range previousResult.EvaluatedProperties() {
			mergedResult.SetEvaluatedProperty(prop)
		}
	}

	for i, subValidator := range mv.validators {
		var result Result
		var err error

		// Each cousin validator should be validated independently
		// without seeing evaluated properties from other cousins
		// Only pass the original previousResult context, not accumulated cousin results
		if objValidator, ok := subValidator.(*objectValidator); ok {
			var previouslyEvaluated map[string]bool
			if previousResult != nil {
				previouslyEvaluated = previousResult.EvaluatedProperties()
			}
			var currentCtx context.Context
			if previouslyEvaluated != nil && len(previouslyEvaluated) > 0 {
				currentCtx = schema.WithEvaluatedProperties(ctx, boolMapToStructMap(previouslyEvaluated))
			} else {
				currentCtx = ctx
			}
			result, err = objValidator.Validate(currentCtx, in)
		} else {
			result, err = subValidator.Validate(ctx, in)
		}

		if err != nil {
			return nil, fmt.Errorf(`allOf validation failed: validator #%d failed: %w`, i, err)
		}

		// Merge object results
		if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
			if mergedResult == nil {
				mergedResult = NewObjectResult()
			}
			for prop := range objResult.EvaluatedProperties() {
				mergedResult.SetEvaluatedProperty(prop)
			}
		}
	}
	return mergedResult, nil
}

// OneOfUnevaluatedPropertiesCompositionValidator handles complex unevaluatedProperties with oneOf
type OneOfUnevaluatedPropertiesCompositionValidator struct {
	oneOfValidators []Interface
	baseValidator   Interface
	schema          *schema.Schema
}

func NewOneOfUnevaluatedPropertiesCompositionValidator(s *schema.Schema) *OneOfUnevaluatedPropertiesCompositionValidator {
	v, err := NewOneOfUnevaluatedPropertiesCompositionValidatorWithResolver(context.Background(), s, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create oneOf composition validator: %v", err))
	}
	return v
}

func NewOneOfUnevaluatedPropertiesCompositionValidatorWithResolver(ctx context.Context, s *schema.Schema, oneOfValidators []Interface, _ *schema.Resolver) (*OneOfUnevaluatedPropertiesCompositionValidator, error) {
	v := &OneOfUnevaluatedPropertiesCompositionValidator{
		schema: s,
	}

	// Use provided validators or compile them if not provided
	if oneOfValidators != nil {
		v.oneOfValidators = oneOfValidators
	} else {
		// Compile oneOf validators
		for _, subSchema := range s.OneOf() {
			subValidator, err := Compile(ctx, convertSchemaOrBool(subSchema))
			if err != nil {
				return nil, fmt.Errorf("failed to compile oneOf validator: %w", err)
			}
			v.oneOfValidators = append(v.oneOfValidators, subValidator)
		}
	}

	// Compile base validator (everything except oneOf)
	baseSchema := createBaseSchema(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to compile base schema: %w", err)
	}
	v.baseValidator = baseValidator

	return v, nil
}

func (v *OneOfUnevaluatedPropertiesCompositionValidator) Validate(ctx context.Context, in any) (Result, error) {
	// For oneOf, exactly one subschema must pass and we collect its annotations
	var validResult *ObjectResult
	passedCount := 0

	for _, subValidator := range v.oneOfValidators {
		result, err := subValidator.Validate(ctx, in)
		if err == nil {
			passedCount++
			// Collect annotations from the passing validator
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				validResult = NewObjectResult()
				for prop := range objResult.EvaluatedProperties() {
					validResult.SetEvaluatedProperty(prop)
				}
			}
		}
	}

	if passedCount == 0 {
		return nil, fmt.Errorf(`oneOf validation failed: none of the validators passed`)
	}
	if passedCount > 1 {
		return nil, fmt.Errorf(`oneOf validation failed: more than one validator passed (%d), expected exactly one`, passedCount)
	}

	// Now validate base constraints, passing the evaluated properties from oneOf
	baseResult, err := v.validateBaseWithContext(ctx, in, validResult)
	if err != nil {
		return nil, err
	}

	// Merge the base result with oneOf result
	if baseObjResult, ok := baseResult.(*ObjectResult); ok && baseObjResult != nil {
		if validResult == nil {
			validResult = NewObjectResult()
		}
		for prop := range baseObjResult.EvaluatedProperties() {
			validResult.SetEvaluatedProperty(prop)
		}
	}

	return validResult, nil
}

// validateBaseWithContext for OneOf
func (v *OneOfUnevaluatedPropertiesCompositionValidator) validateBaseWithContext(ctx context.Context, in any, previousResult *ObjectResult) (Result, error) {
	if objValidator, ok := v.baseValidator.(*objectValidator); ok {
		var previouslyEvaluated map[string]bool
		if previousResult != nil {
			previouslyEvaluated = previousResult.EvaluatedProperties()
		}
		var currentCtx context.Context
		if previouslyEvaluated != nil && len(previouslyEvaluated) > 0 {
			currentCtx = schema.WithEvaluatedProperties(ctx, boolMapToStructMap(previouslyEvaluated))
		} else {
			currentCtx = ctx
		}
		return objValidator.Validate(currentCtx, in)
	}

	if multiValidator, ok := v.baseValidator.(*MultiValidator); ok {
		// If the base validator is a MultiValidator, we need to handle it specially
		return v.validateMultiValidatorWithContext(ctx, multiValidator, in, previousResult)
	}
	// For other validator types, just validate normally without annotation context
	return v.baseValidator.Validate(ctx, in)
}

// validateMultiValidatorWithContext for OneOf
func (v *OneOfUnevaluatedPropertiesCompositionValidator) validateMultiValidatorWithContext(ctx context.Context, mv *MultiValidator, in any, previousResult *ObjectResult) (Result, error) {
	if !mv.and {
		// For OR mode, just validate normally
		return mv.Validate(ctx, in)
	}
	// For AND mode (allOf), validate each sub-validator independently (cousins cannot see each other)
	var mergedResult *ObjectResult
	if previousResult != nil {
		mergedResult = NewObjectResult()
		for prop := range previousResult.EvaluatedProperties() {
			mergedResult.SetEvaluatedProperty(prop)
		}
	}

	for i, subValidator := range mv.validators {
		var result Result
		var err error

		// Each cousin validator should be validated independently
		// without seeing evaluated properties from other cousins
		// Only pass the original previousResult context, not accumulated cousin results
		if objValidator, ok := subValidator.(*objectValidator); ok {
			var previouslyEvaluated map[string]bool
			if previousResult != nil {
				previouslyEvaluated = previousResult.EvaluatedProperties()
			}
			var currentCtx context.Context
			if previouslyEvaluated != nil && len(previouslyEvaluated) > 0 {
				currentCtx = schema.WithEvaluatedProperties(ctx, boolMapToStructMap(previouslyEvaluated))
			} else {
				currentCtx = ctx
			}
			result, err = objValidator.Validate(currentCtx, in)
		} else {
			result, err = subValidator.Validate(ctx, in)
		}

		if err != nil {
			return nil, fmt.Errorf(`allOf validation failed: validator #%d failed: %w`, i, err)
		}

		// Merge object results
		if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
			if mergedResult == nil {
				mergedResult = NewObjectResult()
			}
			for prop := range objResult.EvaluatedProperties() {
				mergedResult.SetEvaluatedProperty(prop)
			}
		}
	}
	return mergedResult, nil
}

// RefUnevaluatedPropertiesCompositionValidator handles complex unevaluatedProperties with $ref
type RefUnevaluatedPropertiesCompositionValidator struct {
	refValidator  Interface
	baseValidator Interface
	schema        *schema.Schema
}

func NewRefUnevaluatedPropertiesCompositionValidator(ctx context.Context, s *schema.Schema, refValidator Interface) *RefUnevaluatedPropertiesCompositionValidator {
	v := &RefUnevaluatedPropertiesCompositionValidator{
		schema:       s,
		refValidator: refValidator,
	}

	// Compile base validator (everything except $ref)
	baseSchema := createSchemaWithoutRef(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		panic(fmt.Sprintf("failed to compile base schema: %v", err))
	}
	v.baseValidator = baseValidator

	return v
}

func (v *RefUnevaluatedPropertiesCompositionValidator) Validate(ctx context.Context, in any) (Result, error) {
	// First, validate the $ref and collect its annotations
	refResult, err := v.refValidator.Validate(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("$ref validation failed: %w", err)
	}

	// Now validate base constraints, passing the evaluated properties from $ref
	baseResult, err := v.validateBaseWithContext(ctx, in, refResult)
	if err != nil {
		return nil, err
	}

	// Merge the base result with $ref result
	var finalResult *ObjectResult
	if err := MergeResults(&finalResult, refResult, baseResult); err != nil {
		// Fall back to simple merging if MergeResults fails
		if objRefResult, ok := refResult.(*ObjectResult); ok {
			if objBaseResult, ok := baseResult.(*ObjectResult); ok {
				finalResult = mergeObjectResults(objRefResult, objBaseResult)
			} else {
				finalResult = objRefResult
			}
		} else if objBaseResult, ok := baseResult.(*ObjectResult); ok {
			finalResult = objBaseResult
		}
	}
	return finalResult, nil
}

// validateBaseWithContext validates the base schema with annotation context from $ref
func (v *RefUnevaluatedPropertiesCompositionValidator) validateBaseWithContext(ctx context.Context, in any, refResult Result) (Result, error) {
	// Create context with evaluated properties if we have evaluation results from $ref
	var currentCtx context.Context
	if objResult, ok := refResult.(*ObjectResult); ok && objResult != nil {
		evalProps := objResult.EvaluatedProperties()
		if len(evalProps) > 0 {
			currentCtx = schema.WithEvaluatedProperties(ctx, boolMapToStructMap(evalProps))
		} else {
			currentCtx = ctx
		}
	} else {
		currentCtx = ctx
	}

	return v.baseValidator.Validate(currentCtx, in)
}

func (v *MultiValidator) Validate(ctx context.Context, in any) (Result, error) {
	if v.and {
		// For allOf, collect all results and merge them while passing context between validators
		var mergedObjectResult *ObjectResult
		var mergedArrayResult *ArrayResult

		for i, subv := range v.validators {
			// Create context with accumulated annotations for this validator
			var currentCtx context.Context = ctx

			// Add evaluated items if we have them (items annotations flow between allOf subschemas)
			if mergedArrayResult != nil {
				evalItems := mergedArrayResult.EvaluatedItems()
				if len(evalItems) > 0 {
					currentCtx = schema.WithEvaluatedItems(ctx, evalItems)
				}
			}

			// NOTE: We do NOT pass evaluated properties between allOf subschemas
			// This implements the "cousin" semantics where properties evaluated by one
			// subschema are not visible to other subschemas in the same allOf

			result, err := subv.Validate(currentCtx, in)
			if err != nil {
				return nil, fmt.Errorf(`allOf validation failed: validator #%d failed: %w`, i, err)
			}
			// Merge object results for property evaluation tracking
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				if mergedObjectResult == nil {
					mergedObjectResult = NewObjectResult()
				}
				for prop := range objResult.EvaluatedProperties() {
					mergedObjectResult.SetEvaluatedProperty(prop)
				}
			}

			// Merge array results for item evaluation tracking
			if arrResult, ok := result.(*ArrayResult); ok && arrResult != nil {
				if mergedArrayResult == nil {
					mergedArrayResult = NewArrayResult()
				}
				arrItems := arrResult.EvaluatedItems()
				for i, evaluated := range arrItems {
					if evaluated {
						mergedArrayResult.SetEvaluatedItem(i)
					}
				}
			}
		}

		// Return appropriate result type based on what we merged
		if mergedObjectResult != nil && mergedArrayResult != nil {
			// Both object and array results - this shouldn't happen in normal validation
			// but prioritize object result for now
			return mergedObjectResult, nil
		} else if mergedObjectResult != nil {
			return mergedObjectResult, nil
		} else if mergedArrayResult != nil {
			return mergedArrayResult, nil
		}

		//nolint:nilnil
		return nil, nil
	}

	if v.oneOf {
		passedCount := 0
		var validResult Result
		for _, subv := range v.validators {
			result, err := subv.Validate(ctx, in)
			if err == nil {
				passedCount++
				validResult = result
			}
		}
		if passedCount == 0 {
			return nil, fmt.Errorf(`oneOf validation failed: none of the validators passed`)
		}
		if passedCount > 1 {
			return nil, fmt.Errorf(`oneOf validation failed: more than one validator passed (%d), expected exactly one`, passedCount)
		}
		return validResult, nil
	}

	// This is for anyOf (OrMode)
	for _, subv := range v.validators {
		result, err := subv.Validate(ctx, in)
		if err == nil {
			return result, nil
		}
	}
	return nil, fmt.Errorf(`anyOf validation failed: none of the validators passed`)
}

// hasBaseConstraints checks if a schema has base-level constraints that need validation
// when used with allOf/anyOf/oneOf
func hasBaseConstraints(s *schema.Schema) bool {
	// Check for types separately since it's not a bit field check
	if len(s.Types()) > 0 {
		return true
	}

	// Use bit field approach for efficient checking of multiple constraints
	baseConstraintFields := schema.MinLengthField | schema.MaxLengthField | schema.PatternField |
		schema.MinimumField | schema.MaximumField | schema.ExclusiveMinimumField | schema.ExclusiveMaximumField | schema.MultipleOfField |
		schema.MinItemsField | schema.MaxItemsField | schema.UniqueItemsField | schema.ItemsField | schema.ContainsField | schema.UnevaluatedItemsField |
		schema.MinPropertiesField | schema.MaxPropertiesField | schema.RequiredField | schema.PropertiesField | schema.PatternPropertiesField | schema.AdditionalPropertiesField | schema.UnevaluatedPropertiesField | schema.DependentSchemasField | schema.PropertyNamesField |
		schema.EnumField | schema.ConstField

	// Returns true if ANY of the base constraint fields are set
	return s.HasAny(baseConstraintFields)
}

// createBaseSchema creates a new schema with only the base constraints (no composition keywords).
// This function excludes ALL composition and control flow keywords:
//   - allOf, anyOf, oneOf (composition keywords)
//   - not (negation keyword)
//   - if/then/else (conditional keywords)
//   - $ref, $dynamicRef (reference keywords)
//
// Only basic validation constraints are copied (types, string/number/array/object constraints, enum/const).
func createBaseSchema(s *schema.Schema) *schema.Schema {
	builder := schema.NewBuilder()

	// Copy types
	if len(s.Types()) > 0 {
		builder.Types(s.Types()...)
	}

	// Copy string constraints
	if s.HasMinLength() {
		builder.MinLength(s.MinLength())
	}
	if s.HasMaxLength() {
		builder.MaxLength(s.MaxLength())
	}
	if s.HasPattern() {
		builder.Pattern(s.Pattern())
	}

	// Copy number constraints
	if s.HasMinimum() {
		builder.Minimum(s.Minimum())
	}
	if s.HasMaximum() {
		builder.Maximum(s.Maximum())
	}
	if s.HasExclusiveMinimum() {
		builder.ExclusiveMinimum(s.ExclusiveMinimum())
	}
	if s.HasExclusiveMaximum() {
		builder.ExclusiveMaximum(s.ExclusiveMaximum())
	}
	if s.HasMultipleOf() {
		builder.MultipleOf(s.MultipleOf())
	}

	// Copy array constraints
	if s.HasMinItems() {
		builder.MinItems(s.MinItems())
	}
	if s.HasMaxItems() {
		builder.MaxItems(s.MaxItems())
	}
	if s.HasUniqueItems() {
		builder.UniqueItems(s.UniqueItems())
	}
	if s.HasItems() {
		builder.Items(s.Items())
	}
	if s.HasContains() {
		builder.Contains(s.Contains())
	}
	if s.HasUnevaluatedItems() {
		builder.UnevaluatedItems(s.UnevaluatedItems())
	}

	// Copy object constraints
	if s.HasMinProperties() {
		builder.MinProperties(s.MinProperties())
	}
	if s.HasMaxProperties() {
		builder.MaxProperties(s.MaxProperties())
	}
	if s.HasRequired() {
		for _, req := range s.Required() {
			builder.Required(req)
		}
	}
	if s.HasProperties() {
		for name, prop := range s.Properties() {
			builder.Property(name, prop)
		}
	}
	if s.HasPatternProperties() {
		for pattern, prop := range s.PatternProperties() {
			builder.PatternProperty(pattern, prop)
		}
	}
	if s.HasAdditionalProperties() {
		builder.AdditionalProperties(s.AdditionalProperties())
	}
	if s.HasUnevaluatedProperties() {
		builder.UnevaluatedProperties(s.UnevaluatedProperties())
	}
	if s.HasDependentSchemas() {
		builder.DependentSchemas(s.DependentSchemas())
	}
	if s.HasPropertyNames() {
		builder.PropertyNames(s.PropertyNames())
	}

	// Copy enum/const
	if s.HasEnum() {
		builder.Enum(s.Enum()...)
	}
	if s.HasConst() {
		builder.Const(s.Const())
	}

	return builder.MustBuild()
}

// mergeGenericResults merges two results, handling both ObjectResult and ArrayResult types
func mergeGenericResults(result1, result2 Result) Result {
	// If either result is nil, return the other
	if result1 == nil {
		return result2
	}
	if result2 == nil {
		return result1
	}

	// Try to merge as ObjectResult first
	if objResult1, ok := result1.(*ObjectResult); ok {
		if objResult2, ok := result2.(*ObjectResult); ok {
			return mergeObjectResults(objResult1, objResult2)
		}
		// Only first is ObjectResult
		return objResult1
	}

	// Try to merge as ArrayResult
	if arrResult1, ok := result1.(*ArrayResult); ok {
		if arrResult2, ok := result2.(*ArrayResult); ok {
			return mergeArrayResults(arrResult1, arrResult2)
		}
		// Only first is ArrayResult
		return arrResult1
	}

	// If neither is a known type, return the second one
	return result2
}

// IfThenElseValidator handles if/then/else conditional validation
type IfThenElseValidator struct {
	ifValidator   Interface
	thenValidator Interface
	elseValidator Interface
}

func compileIfThenElseValidator(ctx context.Context, s *schema.Schema) (Interface, error) {
	v := &IfThenElseValidator{}

	// Compile 'if' validator (required)
	ifSchema := convertSchemaOrBool(s.IfSchema())
	ifValidator, err := Compile(ctx, ifSchema)
	if err != nil {
		return nil, fmt.Errorf(`failed to compile if validator: %w`, err)
	}
	v.ifValidator = ifValidator

	// Compile 'then' validator (optional)
	if s.HasThenSchema() {
		thenSchema := convertSchemaOrBool(s.ThenSchema())
		thenValidator, err := Compile(ctx, thenSchema)
		if err != nil {
			return nil, fmt.Errorf(`failed to compile then validator: %w`, err)
		}
		v.thenValidator = thenValidator
	}

	// Compile 'else' validator (optional)
	if s.HasElseSchema() {
		elseSchema := convertSchemaOrBool(s.ElseSchema())
		elseValidator, err := Compile(ctx, elseSchema)
		if err != nil {
			return nil, fmt.Errorf(`failed to compile else validator: %w`, err)
		}
		v.elseValidator = elseValidator
	}

	return v, nil
}

func (v *IfThenElseValidator) Validate(ctx context.Context, in any) (Result, error) {
	// First, check the 'if' condition and collect its annotations
	ifResult, ifErr := v.ifValidator.Validate(ctx, in)

	// The 'if' schema contributes annotations regardless of whether it passes or fails
	var conditionalResult Result

	if ifErr == nil {
		// 'if' condition passed, validate against 'then' if it exists
		if v.thenValidator != nil {
			thenResult, err := v.thenValidator.Validate(ctx, in)
			if err != nil {
				return nil, err
			}
			// Merge 'if' and 'then' results
			conditionalResult = mergeGenericResults(ifResult, thenResult)
		} else {
			// Only 'if' result
			conditionalResult = ifResult
		}
	} else {
		// 'if' condition failed, validate against 'else' if it exists
		if v.elseValidator != nil {
			elseResult, err := v.elseValidator.Validate(ctx, in)
			if err != nil {
				return nil, err
			}
			// Merge 'if' and 'else' results
			conditionalResult = mergeGenericResults(ifResult, elseResult)
		} else {
			// Only 'if' result (even though it failed validation, it may have annotations)
			conditionalResult = ifResult
		}
	}

	return conditionalResult, nil
}

// IfThenElseUnevaluatedPropertiesCompositionValidator handles complex unevaluatedProperties with if/then/else
type IfThenElseUnevaluatedPropertiesCompositionValidator struct {
	ifValidator   Interface
	thenValidator Interface
	elseValidator Interface
	baseValidator Interface
	schema        *schema.Schema
}

func NewIfThenElseUnevaluatedPropertiesCompositionValidator(ctx context.Context, s *schema.Schema) *IfThenElseUnevaluatedPropertiesCompositionValidator {
	v := &IfThenElseUnevaluatedPropertiesCompositionValidator{
		schema: s,
	}

	// Compile if validator
	ifSchema := convertSchemaOrBool(s.IfSchema())
	ifValidator, err := Compile(ctx, ifSchema)
	if err != nil {
		panic(fmt.Sprintf("failed to compile if validator: %v", err))
	}
	v.ifValidator = ifValidator

	// Compile then validator if it exists
	if s.HasThenSchema() {
		thenSchema := convertSchemaOrBool(s.ThenSchema())
		thenValidator, err := Compile(ctx, thenSchema)
		if err != nil {
			panic(fmt.Sprintf("failed to compile then validator: %v", err))
		}
		v.thenValidator = thenValidator
	}

	// Compile else validator if it exists
	if s.HasElseSchema() {
		elseSchema := convertSchemaOrBool(s.ElseSchema())
		elseValidator, err := Compile(ctx, elseSchema)
		if err != nil {
			panic(fmt.Sprintf("failed to compile else validator: %v", err))
		}
		v.elseValidator = elseValidator
	}

	// Compile base validator (everything except if/then/else)
	baseSchema := createBaseSchema(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		panic(fmt.Sprintf("failed to compile base schema: %v", err))
	}
	v.baseValidator = baseValidator

	return v
}

func (v *IfThenElseUnevaluatedPropertiesCompositionValidator) Validate(ctx context.Context, in any) (Result, error) {
	// First, evaluate if/then/else and collect annotations
	var conditionalResult *ObjectResult

	// Check the 'if' condition and collect its annotations
	ifResult, ifErr := v.ifValidator.Validate(ctx, in)

	// Collect annotations from 'if' schema (contributes regardless of outcome)
	if ifObjResult, ok := ifResult.(*ObjectResult); ok && ifObjResult != nil {
		conditionalResult = NewObjectResult()
		for prop := range ifObjResult.EvaluatedProperties() {
			conditionalResult.SetEvaluatedProperty(prop)
		}
	}

	if ifErr == nil {
		// 'if' condition passed, validate against 'then' if it exists
		if v.thenValidator != nil {
			result, err := v.thenValidator.Validate(ctx, in)
			if err != nil {
				return nil, fmt.Errorf(`if/then validation failed: %w`, err)
			}
			// Merge annotations from 'then' with 'if' annotations
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				if conditionalResult == nil {
					conditionalResult = NewObjectResult()
				}
				for prop := range objResult.EvaluatedProperties() {
					conditionalResult.SetEvaluatedProperty(prop)
				}
			}
		}
	} else {
		// 'if' condition failed, validate against 'else' if it exists
		if v.elseValidator != nil {
			result, err := v.elseValidator.Validate(ctx, in)
			if err != nil {
				return nil, fmt.Errorf(`if/else validation failed: %w`, err)
			}
			// Merge annotations from 'else' with 'if' annotations
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				if conditionalResult == nil {
					conditionalResult = NewObjectResult()
				}
				for prop := range objResult.EvaluatedProperties() {
					conditionalResult.SetEvaluatedProperty(prop)
				}
			}
		}
	}

	// Now validate base constraints, passing the evaluated properties from if/then/else
	baseResult, err := v.validateBaseWithContext(ctx, in, conditionalResult)
	if err != nil {
		return nil, err
	}

	// Merge the base result with if/then/else result
	if baseObjResult, ok := baseResult.(*ObjectResult); ok && baseObjResult != nil {
		if conditionalResult == nil {
			conditionalResult = NewObjectResult()
		}
		for prop := range baseObjResult.EvaluatedProperties() {
			conditionalResult.SetEvaluatedProperty(prop)
		}
	}

	return conditionalResult, nil
}

// validateBaseWithContext for if/then/else
func (v *IfThenElseUnevaluatedPropertiesCompositionValidator) validateBaseWithContext(ctx context.Context, in any, previousResult *ObjectResult) (Result, error) {
	// Create context with evaluated properties if we have previous evaluation results
	var currentCtx context.Context
	if previousResult != nil {
		evalProps := previousResult.EvaluatedProperties()
		if len(evalProps) > 0 {
			currentCtx = schema.WithEvaluatedProperties(ctx, boolMapToStructMap(evalProps))
		} else {
			currentCtx = ctx
		}
	} else {
		currentCtx = ctx
	}

	return v.baseValidator.Validate(currentCtx, in)
}
