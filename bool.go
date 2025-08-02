package schema

import (
	"encoding/json"
	"fmt"
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
