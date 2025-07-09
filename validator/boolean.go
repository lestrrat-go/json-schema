package validator

import (
	"fmt"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
)

func compileBooleanValidator(s *schema.Schema) (Validator, error) {
	v := Boolean()
	if s.HasConst() {
		c, ok := s.Const().(bool)
		if !ok {
			return nil, fmt.Errorf(`invalid element in const: expected boolean element, got %T`, s.Const())
		}
		v.Const(c)
	}
	if s.HasEnum() {
		enums := s.Enum()
		l := make([]bool, 0, len(enums))
		for i, e := range s.Enum() {
			b, ok := e.(bool)
			if !ok {
				return nil, fmt.Errorf(`invalid element in enum: expected boolean element, got %T for element %d`, e, i)
			}
			l = append(l, b)
		}
		v.Enum(l)
	}
	return v.Build()
}

type BooleanValidator struct {
	enum          []bool
	constantValue *bool
}

type BooleanValidatorBuilder struct {
	err error
	c   *BooleanValidator
}

func Boolean() *BooleanValidatorBuilder {
	return &BooleanValidatorBuilder{
		c: &BooleanValidator{},
	}
}

func (b *BooleanValidatorBuilder) Const(v bool) *BooleanValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.constantValue = &v
	return b
}

func (b *BooleanValidatorBuilder) Enum(v []bool) *BooleanValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.enum = v
	return b
}

func (b *BooleanValidatorBuilder) Build() (Validator, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.c, nil
}

func (c *BooleanValidator) Validate(v interface{}) error {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Bool:
		boolVal := rv.Bool()
		
		// Check const constraint
		if c.constantValue != nil {
			if boolVal != *c.constantValue {
				return fmt.Errorf(`invalid value passed to BooleanValidator: must be const value %t, got %t`, *c.constantValue, boolVal)
			}
		}
		
		// Check enum constraint
		if len(c.enum) > 0 {
			found := false
			for _, allowed := range c.enum {
				if boolVal == allowed {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf(`invalid value passed to BooleanValidator: %t not found in enum`, boolVal)
			}
		}
		
		return nil
	default:
		return fmt.Errorf(`invalid value passed to BooleanValidator: expected boolean, got %T`, v)
	}
}