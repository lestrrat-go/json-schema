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

// Bool represents a "boolean" value in a JSON Schema, such as
// "exclusiveMinimum", "exclusiveMaximum", etc.
type Bool struct {
	val          bool
	defaultValue bool
	initialized  bool
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

// NewPrimitiveType creates a PrimitiveType from its string representation.
// It accepts standard JSON Schema type names such as "null", "integer", "string", "object", "array", "boolean", and "number".
func NewPrimitiveType(s string) (PrimitiveType, error) {
	switch s {
	case "null":
		return NullType, nil
	case "integer":
		return IntegerType, nil
	case "string":
		return StringType, nil
	case "object":
		return ObjectType, nil
	case "array":
		return ArrayType, nil
	case "boolean":
		return BooleanType, nil
	case "number":
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
		v = "null"
	case IntegerType:
		v = "integer"
	case StringType:
		v = "string"
	case ObjectType:
		v = "object"
	case ArrayType:
		v = "array"
	case BooleanType:
		v = "boolean"
	case NumberType:
		v = "number"
	default:
		v = "<invalid>"
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

// Bool returns the underlying boolean value for the
// primitive boolean type
func (b Bool) Bool() bool {
	if b.initialized {
		return b.val
	}
	return b.defaultValue
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
