package schema

import (
	"encoding/json"
	"errors"
	"fmt"
)

type PrimitiveType int

const (
	InvalidType PrimitiveType = iota
	NullType
	IntegerType
	StringType
	ObjectType
	ArrayType
	BooleanType
	NumberType
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

// NewPrimitiveType creates a PrimitiveType from its string representation.
// It accepts standard JSON Schema type names such as "null", "integer", "string", "object", "array", "boolean", and "number".
func NewPrimitiveType(s string) (PrimitiveType, error) {
	switch s {
	case nullTypeString:
		return NullType, nil
	case integerTypeString:
		return IntegerType, nil
	case stringTypeString:
		return StringType, nil
	case objectTypeString:
		return ObjectType, nil
	case arrayTypeString:
		return ArrayType, nil
	case booleanTypeString:
		return BooleanType, nil
	case numberTypeString:
		return NumberType, nil
	default:
		return InvalidType, fmt.Errorf(`unknown primitive type %q`, s)
	}
}

// String returns the string representation of this primitive type
func (t PrimitiveType) String() string {
	var v string
	switch t {
	case NullType:
		v = nullTypeString
	case IntegerType:
		v = integerTypeString
	case StringType:
		v = stringTypeString
	case ObjectType:
		v = objectTypeString
	case ArrayType:
		v = arrayTypeString
	case BooleanType:
		v = booleanTypeString
	case NumberType:
		v = numberTypeString
	default:
		v = invalidTypeString
	}
	return v
}

// MarshalJSON seriealises the primitive type into a JSON string
func (t PrimitiveType) MarshalJSON() ([]byte, error) {
	switch t {
	case NullType, IntegerType, StringType, ObjectType, ArrayType, BooleanType, NumberType:
		return json.Marshal(t.String())
	default:
		return nil, errors.New("unknown primitive type")
	}
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
