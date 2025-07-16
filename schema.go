//go:generate ./gen.sh

package schema

import (
	"encoding/json"
	"fmt"
	"unicode"
)

// SchemaOrBool is an interface for types that can be either a Schema or boolean
type SchemaOrBool interface {
	schemaOrBool() // internal identifier
}

// SchemaBool represents a boolean value in allOf, oneOf, anyOf, etc
type SchemaBool bool

// schemaOrBool implements the SchemaOrBool interface
func (s SchemaBool) schemaOrBool() {}

// UnmarshalJSON implements json.Unmarshaler
func (s *SchemaBool) UnmarshalJSON(data []byte) error {
	var b bool
	if err := json.Unmarshal(data, &b); err != nil {
		return fmt.Errorf("failed to unmarshal SchemaBool: %w", err)
	}
	*s = SchemaBool(b)
	return nil
}

// MarshalJSON implements json.Marshaler
func (s SchemaBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(bool(s))
}

// The schema that this implementation supports. We use the name
// `Version` here because `Schema` is confusin with other types
const Version = `https://json-schema.org/draft/2020-12/schema`

// schemaOrBool implements the SchemaOrBool interface for Schema
func (s *Schema) schemaOrBool() {}

// Convenience variables and functions for SchemaBool values
var schemaTrue = SchemaBool(true)
var schemaFalse = SchemaBool(false)

// SchemaTrue returns a SchemaBool representing true
func SchemaTrue() SchemaBool {
	return schemaTrue
}

// SchemaFalse returns a SchemaBool representing false
func SchemaFalse() SchemaBool {
	return schemaFalse
}

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
			result = append(result, SchemaBool(b))
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

// validateSchemaOrBool checks if a value is either a bool, SchemaBool, or *Schema
func validateSchemaOrBool(v any) error {
	switch v.(type) {
	case bool, SchemaBool, *Schema:
		return nil
	default:
		return fmt.Errorf(`expected bool, SchemaBool, or *Schema, got %T`, v)
	}
}

type propPair struct {
	Name   string
	Schema *Schema
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
