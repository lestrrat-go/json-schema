//go:generate ./gen.sh

package schema

import (
	"encoding/json"
	"fmt"
	"unicode"
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

// Version is the schema that this implementation supports. We use the name
// Version here because Schema is confusing with other types.
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
	maxLen := max(len(runesA), len(runesB))

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
