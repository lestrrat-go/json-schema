package validator

import (
	"context"
	"fmt"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
)

var _ Builder = (*BooleanValidatorBuilder)(nil)
var _ Interface = (*booleanValidator)(nil)

func compileBooleanValidator(ctx context.Context, s *schema.Schema) (Interface, error) {
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

type booleanValidator struct {
	enum          []bool
	constantValue *bool
}

type BooleanValidatorBuilder struct {
	err error
	c   *booleanValidator
}

func Boolean() *BooleanValidatorBuilder {
	return (&BooleanValidatorBuilder{}).Reset()
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

func (b *BooleanValidatorBuilder) Build() (Interface, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.c, nil
}

func (b *BooleanValidatorBuilder) MustBuild() Interface {
	if b.err != nil {
		panic(b.err)
	}
	return b.c
}

func (b *BooleanValidatorBuilder) Reset() *BooleanValidatorBuilder {
	b.err = nil
	b.c = &booleanValidator{}
	return b
}

func (c *booleanValidator) Validate(ctx context.Context, v any) (Result, error) {
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
				return nil, fmt.Errorf(`invalid value passed to BooleanValidator: must be const value %t, got %t`, *c.constantValue, boolVal)
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
				return nil, fmt.Errorf(`invalid value passed to BooleanValidator: %t not found in enum`, boolVal)
			}
		}

		return nil, nil
	default:
		return nil, fmt.Errorf(`invalid value passed to BooleanValidator: expected boolean, got %T`, v)
	}
}
