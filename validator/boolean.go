package validator

import (
	"context"
	"fmt"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/vocabulary"
)

var _ Builder = (*BooleanValidatorBuilder)(nil)
var _ Interface = (*booleanValidator)(nil)

func compileBooleanValidator(ctx context.Context, s *schema.Schema) (Interface, error) {
	v := Boolean()
	if s.HasConst() && vocabulary.IsKeywordEnabledInContext(ctx, "const") {
		v.Const(s.Const())
	}
	if s.HasEnum() && vocabulary.IsKeywordEnabledInContext(ctx, "enum") {
		v.Enum(s.Enum()...)
	}
	return v.Build()
}

type booleanValidator struct {
	enum          []any
	constantValue any
}

type BooleanValidatorBuilder struct {
	err error
	c   *booleanValidator
}

// Boolean creates a new BooleanValidatorBuilder instance that can be used to build a
// Validator for boolean values according to the JSON Schema specification.
func Boolean() *BooleanValidatorBuilder {
	return (&BooleanValidatorBuilder{}).Reset()
}

func (b *BooleanValidatorBuilder) Const(v any) *BooleanValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.constantValue = v
	return b
}

func (b *BooleanValidatorBuilder) Enum(v ...any) *BooleanValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.enum = make([]any, len(v))
	copy(b.c.enum, v)
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
	logger := TraceSlogFromContext(ctx)
	logger.InfoContext(ctx, "boolean validator starting", "value", v, "type", fmt.Sprintf("%T", v))

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Bool:
		boolVal := rv.Bool()
		logger.InfoContext(ctx, "boolean validator processing boolean value", "value", boolVal)

		// Check const constraint
		if c.constantValue != nil {
			if err := validateConst(ctx, boolVal, c.constantValue); err != nil {
				return nil, fmt.Errorf(`invalid value passed to BooleanValidator: %w`, err)
			}
		}

		// Check enum constraint
		if len(c.enum) > 0 {
			if err := validateEnum(ctx, boolVal, c.enum); err != nil {
				return nil, fmt.Errorf(`invalid value passed to BooleanValidator: %w`, err)
			}
		}

		//nolint: nilnil
		return nil, nil
	default:
		logger.InfoContext(ctx, "boolean validator rejecting non-boolean", "type", fmt.Sprintf("%T", v))
		return nil, fmt.Errorf(`invalid value passed to BooleanValidator: expected boolean, got %T`, v)
	}
}
