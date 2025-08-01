//go:generate ./gen.sh

package validator

import (
	"context"
	"fmt"
	"reflect"
	"strings"

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
	for i := range rt.NumField() {
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
