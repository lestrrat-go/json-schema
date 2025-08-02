package schema

import (
	"encoding/json"
	"errors"
	"fmt"
)

// PrimitiveType represents a JSON Schema primitive type
type PrimitiveType int

const (
	// InvalidType represents an invalid or unknown type
	InvalidType PrimitiveType = iota
	// NullType represents the JSON null type
	NullType
	// IntegerType represents the JSON integer type
	IntegerType
	// StringType represents the JSON string type
	StringType
	// ObjectType represents the JSON object type
	ObjectType
	// ArrayType represents the JSON array type
	ArrayType
	// BooleanType represents the JSON boolean type
	BooleanType
	// NumberType represents the JSON number type
	NumberType
	maxPrimitiveType
)

// unexported string constants for primitive type names
const (
	nullTypeString    = "null"
	integerTypeString = "integer"
	stringTypeString  = "string"
	objectTypeString  = "object"
	arrayTypeString   = "array"
	booleanTypeString = "boolean"
	numberTypeString  = "number"
	invalidTypeString = "<invalid>"
)

var knwonPrimitives [int(maxPrimitiveType)]string
var availablePrimitives map[string]PrimitiveType

func init() {
	knwonPrimitives[0] = invalidTypeString
	knwonPrimitives[1] = nullTypeString
	knwonPrimitives[2] = integerTypeString
	knwonPrimitives[3] = stringTypeString
	knwonPrimitives[4] = objectTypeString
	knwonPrimitives[5] = arrayTypeString
	knwonPrimitives[6] = booleanTypeString
	knwonPrimitives[7] = numberTypeString

	availablePrimitives = make(map[string]PrimitiveType)
	for i := 1; i < int(maxPrimitiveType); i++ {
		availablePrimitives[knwonPrimitives[i]] = PrimitiveType(i)
	}
}

// UnmarshalJSON initializes the primitive type from
// a JSON string.
func (t *PrimitiveType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	x, err := NewPrimitiveType(s)
	if err != nil {
		return err
	}
	*t = x
	return nil
}

// IsScalarPrimitiveType returns true if the given type is a scalar type (not object or array)
func IsScalarPrimitiveType(typ PrimitiveType) bool {
	switch typ {
	case StringType, IntegerType, NumberType, BooleanType, NullType:
		return true
	default:
		return false
	}
}

// NewPrimitiveType creates a PrimitiveType from its string representation.
// It accepts standard JSON Schema type names such as "null", "integer", "string", "object", "array", "boolean", and "number".
func NewPrimitiveType(s string) (PrimitiveType, error) {
	if pt, ok := availablePrimitives[s]; ok {
		return pt, nil
	}
	return InvalidType, fmt.Errorf(`unknown primitive type %q`, s)
}

// String returns the string representation of this primitive type
func (t PrimitiveType) String() string {
	if t < NullType || t >= maxPrimitiveType {
		return invalidTypeString
	}
	return knwonPrimitives[t]
}

// MarshalJSON seriealises the primitive type into a JSON string
func (t PrimitiveType) MarshalJSON() ([]byte, error) {
	if t < NullType || t >= maxPrimitiveType {
		return nil, errors.New("unknown primitive type")
	}

	v := knwonPrimitives[t]
	return json.Marshal(v)
}

type PrimitiveTypes []PrimitiveType

// UnmarshalJSON initializes the list of primitive types
func (pt *PrimitiveTypes) UnmarshalJSON(data []byte) error {
	if data[0] != '[' {
		var t PrimitiveType
		if err := json.Unmarshal(data, &t); err != nil {
			return err
		}

		*pt = PrimitiveTypes{t}
		return nil
	}

	var list []PrimitiveType
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	*pt = PrimitiveTypes(list)
	return nil
}

// MarshalJSON serializes the list of primitive types
func (pt PrimitiveTypes) MarshalJSON() ([]byte, error) {
	if len(pt) == 1 {
		return json.Marshal(pt[0])
	}
	return json.Marshal([]PrimitiveType(pt))
}

// Contains returns true if the list of primitive types
// contains `p`
func (pt PrimitiveTypes) Contains(p PrimitiveType) bool {
	for _, v := range pt {
		if p == v {
			return true
		}
	}
	return false
}
