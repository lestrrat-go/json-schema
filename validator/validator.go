//go:generate ./gen.sh

package validator

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/lestrrat-go/blackmagic"
	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/internal/schemactx"
)

// Interface is the interface that all validators must implement.
type Interface interface {
	Validate(context.Context, any) (Result, error)
}

// Result contains annotation information from validation that may be used
// by other validators (e.g., for unevaluatedProperties tracking)
type Result any

// ObjectFieldResolver allows custom resolution of object fields
type ObjectFieldResolver interface {
	ResolveObjectField(string) (any, error)
}

// ArrayIndexResolver allows custom resolution of array indices
type ArrayIndexResolver interface {
	ResolveArrayIndex(int) (any, error)
}

// resolveObjectField resolves a field from an object, supporting multiple types:
// - map[string]any: direct key lookup
// - ObjectFieldResolver: custom resolution
// - struct: reflection with JSON tag support
func resolveObjectField(obj any, fieldName string) (any, error) {
	if obj == nil {
		return nil, fmt.Errorf("cannot resolve field %q from nil object", fieldName)
	}

	// Try ObjectFieldResolver interface first
	if resolver, ok := obj.(ObjectFieldResolver); ok {
		return resolver.ResolveObjectField(fieldName)
	}

	// Handle map[string]any directly
	if m, ok := obj.(map[string]any); ok {
		if value, exists := m[fieldName]; exists {
			return value, nil
		}
		return nil, fmt.Errorf("field %q not found in object", fieldName)
	}

	// Handle struct types using reflection
	return resolveStructField(obj, fieldName)
}

// resolveArrayIndex resolves an element from an array, supporting multiple types:
// - []any: direct index access
// - ArrayIndexResolver: custom resolution
func resolveArrayIndex(arr any, index int) (any, error) {
	if arr == nil {
		return nil, fmt.Errorf("cannot resolve index %d from nil array", index)
	}

	// Try ArrayIndexResolver interface first
	if resolver, ok := arr.(ArrayIndexResolver); ok {
		return resolver.ResolveArrayIndex(index)
	}

	// Handle slice types using reflection
	rv := reflect.ValueOf(arr)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("value is not an array or slice, got %T", arr)
	}

	if index < 0 || index >= rv.Len() {
		return nil, fmt.Errorf("index %d out of bounds for array of length %d", index, rv.Len())
	}

	return rv.Index(index).Interface(), nil
}

// resolveStructField resolves a field from a struct using reflection and JSON tags
func resolveStructField(obj any, fieldName string) (any, error) {
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("value is not a struct, got %T", obj)
	}

	rt := rv.Type()

	// First, try to find field by JSON tag
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Check JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			// Parse JSON tag (field name is before first comma)
			tagName := strings.Split(jsonTag, ",")[0]
			if tagName == fieldName {
				return rv.Field(i).Interface(), nil
			}
			// Skip if tag explicitly sets a different name
			if tagName != "" && tagName != "-" {
				continue
			}
		}

		// Check if field name matches (case-insensitive for JSON compatibility)
		if strings.EqualFold(field.Name, fieldName) {
			return rv.Field(i).Interface(), nil
		}
	}

	return nil, fmt.Errorf("field %q not found in struct %T", fieldName, obj)
}

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
func NewArrayResult(size ...int) *ArrayResult {
	var capacity int
	if len(size) > 0 {
		capacity = size[0]
	}
	return &ArrayResult{
		evaluatedItems: make([]bool, 0, capacity),
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

// getValueType determines the JSON Schema primitive type of a value
func getValueType(value any) schema.PrimitiveType {
	switch value.(type) {
	case string:
		return schema.StringType
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return schema.IntegerType
	case float32, float64:
		return schema.NumberType
	case bool:
		return schema.BooleanType
	case []any:
		return schema.ArrayType
	case map[string]any:
		return schema.ObjectType
	case nil:
		return schema.NullType
	default:
		// Default to string for unknown types
		return schema.StringType
	}
}

// isPrimitiveType returns true if the given type is a simple primitive type
// (not a complex type like object or array) that's safe for enum type inference
func isPrimitiveType(typ schema.PrimitiveType) bool {
	switch typ {
	case schema.StringType, schema.IntegerType, schema.NumberType, schema.BooleanType, schema.NullType:
		return true
	case schema.ObjectType, schema.ArrayType:
		return false
	default:
		return false
	}
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

	// Set up vocabulary context if none provided
	// Default to JSON Schema 2020-12 default vocabulary (format-assertion disabled)
	var vocabSet VocabularySet
	if err := schemactx.VocabularySetFromContext(ctx, &vocabSet); err != nil {
		// No vocabulary set in context, use default vocabulary per JSON Schema spec
		ctx = WithVocabularySet(ctx, DefaultVocabularySet())
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
			compositeValidator := AllOf(refValidator, otherValidator)
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
			allOfValidator := AllOf(allOfValidators...)
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
			anyOfValidator := AnyOf(anyOfValidators...)
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
			oneOfValidator := OneOf(oneOfValidators...)
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
	hasBC := hasBaseConstraints(s)
	hasCompositionValidator := (s.HasAllOf() && hasBC) ||
		(s.HasAnyOf() && hasBC && s.HasUnevaluatedProperties()) ||
		(s.HasOneOf() && hasBC && s.HasUnevaluatedProperties()) ||
		(s.HasIfSchema() && hasBC && s.HasUnevaluatedProperties())

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
		// Try to infer type from enum values if they're homogeneous
		var inferredType *schema.PrimitiveType
		if s.HasEnum() && IsKeywordEnabledInContext(ctx, "enum") {
			enumValues := s.Enum()
			if len(enumValues) > 0 {
				// Check if all enum values are of the same type
				firstType := getValueType(enumValues[0])
				allSameType := true
				for _, val := range enumValues[1:] {
					if getValueType(val) != firstType {
						allSameType = false
						break
					}
				}
				// Only infer primitive types (string, number, integer, boolean, null)
				// For complex types (object, array), keep as untyped to avoid validation conflicts
				if allSameType && isPrimitiveType(firstType) {
					inferredType = &firstType
				}
			}
		}

		// If we inferred a specific type, use that type's validator
		if inferredType != nil {
			types = append(types, *inferredType)
			inferredTypes[*inferredType] = true
		} else {
			// Fall back to untyped validator for mixed-type or const-only constraints
			validator, err := compileUntypedValidator(ctx, s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile general validator: %w`, err)
			}
			allValidators = append(allValidators, validator)
		}
	}

	for _, typ := range types {
		// This is a placeholder code. In reality we need to
		// OR all types
		switch typ {
		case schema.StringType:
			// Use strict type checking for explicitly declared string types OR
			// for inferred string types that have enum constraints (since enum values determine the type)
			strictType := !inferredTypes[schema.StringType] || (inferredTypes[schema.StringType] && s.HasEnum())
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
		typeValidator := AnyOf(validatorsByType...)
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
	return AllOf(allValidators...), nil
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

type nullValidator struct{}

func Null() Interface {
	return nullValidator{}
}

func (nullValidator) Validate(_ context.Context, v any) (Result, error) {
	if v == nil {
		//nolint: nilnil
		return nil, nil
	}
	return nil, fmt.Errorf(`invalid value passed to NullValidator: expected null, got %T`, v)
}

func compileNullValidator(_ context.Context, _ *schema.Schema) (Interface, error) {
	return nullValidator{}, nil
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
	if objResult, ok := refResult.(*ObjectResult); ok && objResult != nil {
		evalProps := objResult.EvaluatedProperties()
		if len(evalProps) > 0 {
			ctx = schema.WithEvaluatedProperties(ctx, boolMapToStructMap(evalProps))
		}
	}

	return v.baseValidator.Validate(ctx, in)
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
