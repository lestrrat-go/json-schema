//go:generate ./gen.sh

package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"unicode"

	"github.com/lestrrat-go/json-schema/internal/schemactx"
)

// SchemaOrBool is an interface for types that can be either a Schema or boolean
type SchemaOrBool interface { //nolint:revive
	schemaOrBool() // internal identifier
}

// BoolSchema represents a boolean value in allOf, oneOf, anyOf, etc
type BoolSchema bool

// schemaOrBool implements the SchemaOrBool interface
func (s BoolSchema) schemaOrBool() {}

// UnmarshalJSON implements json.Unmarshaler
func (s *BoolSchema) UnmarshalJSON(data []byte) error {
	var b bool
	if err := json.Unmarshal(data, &b); err != nil {
		return fmt.Errorf("failed to unmarshal BoolSchema: %w", err)
	}
	*s = BoolSchema(b)
	return nil
}

// MarshalJSON implements json.Marshaler
func (s BoolSchema) MarshalJSON() ([]byte, error) {
	return json.Marshal(bool(s))
}

// The schema that this implementation supports. We use the name
// `Version` here because `Schema` is confusin with other types
const Version = `https://json-schema.org/draft/2020-12/schema`

// schemaOrBool implements the SchemaOrBool interface for Schema
func (s *Schema) schemaOrBool() {}

// Convenience variables and functions for BoolSchema values
var trueSchema = BoolSchema(true)
var falseSchema = BoolSchema(false)

// TrueSchema returns a BoolSchema representing true
func TrueSchema() BoolSchema {
	return trueSchema
}

// FalseSchema returns a BoolSchema representing false
func FalseSchema() BoolSchema {
	return falseSchema
}

// Predefined field groups for common bit flag checks

// StringConstraintFields groups all string-related validation fields
const StringConstraintFields = MinLengthField | MaxLengthField | PatternField | FormatField

// NumericConstraintFields groups all numeric validation fields
const NumericConstraintFields = MinimumField | MaximumField | MultipleOfField | ExclusiveMinimumField | ExclusiveMaximumField

// ObjectConstraintFields groups all object validation fields
const ObjectConstraintFields = PropertiesField | AdditionalPropertiesField | RequiredField | MinPropertiesField | MaxPropertiesField | PatternPropertiesField | PropertyNamesField | UnevaluatedPropertiesField | DependentSchemasField | DependentRequiredField

// ArrayConstraintFields groups all array validation fields
const ArrayConstraintFields = ItemsField | MinItemsField | MaxItemsField | UniqueItemsField | PrefixItemsField | ContainsField | MinContainsField | MaxContainsField | UnevaluatedItemsField

// CompositionFields groups all composition/logical fields
const CompositionFields = AllOfField | AnyOfField | OneOfField | NotField

// ConditionalFields groups all conditional validation fields
const ConditionalFields = IfSchemaField | ThenSchemaField | ElseSchemaField

// ContentFields groups all content-related validation fields
const ContentFields = ContentEncodingField | ContentMediaTypeField | ContentSchemaField

// BasicPropertiesFields groups basic property-related fields (excluding constraints like minProperties)
const BasicPropertiesFields = PropertiesField | PatternPropertiesField | AdditionalPropertiesField

// UnevaluatedFields groups unevaluated items and properties fields
const UnevaluatedFields = UnevaluatedPropertiesField | UnevaluatedItemsField

// ValueConstraintFields groups enum and const constraint fields
const ValueConstraintFields = EnumField | ConstField

// unmarshalSchemaOrBoolSlice parses a JSON array using token-based decoding
func unmarshalSchemaOrBoolSlice(dec *json.Decoder) ([]SchemaOrBool, error) {
	// We need to decode the array as raw JSON first, then handle each element
	var rawArray []json.RawMessage
	if err := dec.Decode(&rawArray); err != nil {
		return nil, fmt.Errorf("failed to decode array: %w", err)
	}

	result := make([]SchemaOrBool, 0, len(rawArray))

	for i, rawElement := range rawArray {
		// Try to decode as boolean first
		var b bool
		if err := json.Unmarshal(rawElement, &b); err == nil {
			result = append(result, BoolSchema(b))
			continue
		}

		// Try to decode as Schema object
		var schema Schema
		if err := json.Unmarshal(rawElement, &schema); err == nil {
			result = append(result, &schema)
			continue
		}

		return nil, fmt.Errorf("element at index %d is neither boolean nor valid schema object", i)
	}

	return result, nil
}

// unmarshalSchemaOrBoolMap parses a JSON map using token-based decoding
func unmarshalSchemaOrBoolMap(dec *json.Decoder) (map[string]SchemaOrBool, error) {
	// We need to decode the map as raw JSON first, then handle each value
	var rawMap map[string]json.RawMessage
	if err := dec.Decode(&rawMap); err != nil {
		return nil, fmt.Errorf("failed to decode map: %w", err)
	}

	result := make(map[string]SchemaOrBool)

	for key, rawValue := range rawMap {
		// Try to decode as boolean first
		var b bool
		if err := json.Unmarshal(rawValue, &b); err == nil {
			result[key] = BoolSchema(b)
			continue
		}

		// Try to decode as Schema object
		var schema Schema
		if err := json.Unmarshal(rawValue, &schema); err == nil {
			result[key] = &schema
			continue
		}

		return nil, fmt.Errorf("value for key %q is neither boolean nor valid schema object", key)
	}

	return result, nil
}

// compareFieldNames compares two field names with custom sorting logic:
// Character-by-character comparison where at each position:
// 1. Non-alphanumeric characters sort before alphanumeric characters
// 2. Within each category (non-alphanumeric or alphanumeric), sort lexicographically
func compareFieldNames(a, b string) bool {
	runesA := []rune(a)
	runesB := []rune(b)

	// Compare character by character up to the length of the longer string
	maxLen := len(runesA)
	if len(runesB) > maxLen {
		maxLen = len(runesB)
	}

	for i := range maxLen {
		var charA, charB rune
		var hasCharA, hasCharB bool

		if i < len(runesA) {
			charA = runesA[i]
			hasCharA = true
		}
		if i < len(runesB) {
			charB = runesB[i]
			hasCharB = true
		}

		// If one string is shorter, the longer string comes later
		if !hasCharA && hasCharB {
			return true // A is shorter, comes first
		}
		if hasCharA && !hasCharB {
			return false // B is shorter, comes first
		}
		if !hasCharA && !hasCharB {
			break // Both strings end here
		}

		isAlphaNumA := unicode.IsLetter(charA) || unicode.IsDigit(charA)
		isAlphaNumB := unicode.IsLetter(charB) || unicode.IsDigit(charB)

		// If one is non-alphanumeric and the other is alphanumeric,
		// non-alphanumeric comes first
		if !isAlphaNumA && isAlphaNumB {
			return true
		}
		if isAlphaNumA && !isAlphaNumB {
			return false
		}

		// Both are in same category (both alphanumeric or both non-alphanumeric)
		// Compare lexicographically (case-insensitive for letters)
		lowerA := unicode.ToLower(charA)
		lowerB := unicode.ToLower(charB)

		if lowerA != lowerB {
			return lowerA < lowerB
		}

		// If lowercase versions are equal, compare original case
		if charA != charB {
			return charA < charB
		}
	}

	// All characters are equal
	return false
}

// Context helper functions - these delegate to internal schemactx package

// WithBaseSchema adds a base schema to the context for reference resolution
func WithBaseSchema(ctx context.Context, baseSchema *Schema) context.Context {
	return schemactx.WithBaseSchema(ctx, baseSchema)
}

// BaseSchemaFromContext retrieves the base schema from context, returns nil if not found
func BaseSchemaFromContext(ctx context.Context) *Schema {
	var baseSchema *Schema
	if err := schemactx.BaseSchemaFromContext(ctx, &baseSchema); err != nil {
		return nil
	}
	return baseSchema
}

// WithResolver adds a resolver to the context
func WithResolver(ctx context.Context, resolver *Resolver) context.Context {
	return schemactx.WithResolver(ctx, resolver)
}

// ResolverFromContext retrieves the resolver from context, returns nil if not found
func ResolverFromContext(ctx context.Context) *Resolver {
	var resolver *Resolver
	if err := schemactx.ResolverFromContext(ctx, &resolver); err != nil {
		return nil
	}
	return resolver
}

// WithRootSchema adds a root schema to the context
func WithRootSchema(ctx context.Context, rootSchema *Schema) context.Context {
	return schemactx.WithRootSchema(ctx, rootSchema)
}

// RootSchemaFromContext retrieves the root schema from context, returns nil if not found
func RootSchemaFromContext(ctx context.Context) *Schema {
	var rootSchema *Schema
	if err := schemactx.RootSchemaFromContext(ctx, &rootSchema); err != nil {
		return nil
	}
	return rootSchema
}

// WithBaseURI adds a base URI to the context for reference resolution
func WithBaseURI(ctx context.Context, baseURI string) context.Context {
	return schemactx.WithBaseURI(ctx, baseURI)
}

// BaseURIFromContext extracts the base URI from context, returns empty string if not present
func BaseURIFromContext(ctx context.Context) string {
	var baseURI string
	if err := schemactx.BaseURIFromContext(ctx, &baseURI); err != nil {
		return ""
	}
	return baseURI
}

// WithDynamicScope adds a schema to the dynamic scope chain in the context
func WithDynamicScope(ctx context.Context, s *Schema) context.Context {
	return schemactx.WithDynamicScope(ctx, s)
}

// DynamicScopeFromContext retrieves the dynamic scope chain from context, returns nil if not present
func DynamicScopeFromContext(ctx context.Context) []*Schema {
	var scope []any
	if err := schemactx.DynamicScopeFromContext(ctx, &scope); err != nil {
		return nil
	}

	// Convert []any to []*Schema
	result := make([]*Schema, 0, len(scope))
	for _, s := range scope {
		if schema, ok := s.(*Schema); ok {
			result = append(result, schema)
		}
	}
	return result
}

// WithReferenceStack adds a reference stack to the context for circular reference detection
func WithReferenceStack(ctx context.Context, stack []string) context.Context {
	return schemactx.WithReferenceStack(ctx, stack)
}

// ReferenceStackFromContext retrieves the reference stack from context, returns nil if not present
func ReferenceStackFromContext(ctx context.Context) []string {
	var stack []string
	if err := schemactx.ReferenceStackFromContext(ctx, &stack); err != nil {
		return nil
	}
	return stack
}

// Context keys for validator-specific data
type dependentSchemasKey struct{}

// WithDependentSchemas adds compiled dependent schema validators to the context
func WithDependentSchemas(ctx context.Context, dependentSchemas map[string]any) context.Context {
	return context.WithValue(ctx, dependentSchemasKey{}, dependentSchemas)
}

// DependentSchemasFromContext extracts compiled dependent schema validators from context, returns nil if none are associated with ctx
func DependentSchemasFromContext(ctx context.Context) map[string]any {
	if deps, ok := ctx.Value(dependentSchemasKey{}).(map[string]any); ok {
		return deps
	}
	return nil
}
