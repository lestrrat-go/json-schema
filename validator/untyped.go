package validator

import (
	"context"
	"fmt"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/vocabulary"
)

var _ Interface = (*untypedValidator)(nil)

// untypedValidator handles enum and const validation for schemas without specific types
type untypedValidator struct {
	enum          []any
	constantValue *any // Pointer distinguishes nil vs no const
}

// Untyped creates a validator for schemas without explicit types that can have enum/const constraints
func Untyped() *UntypedValidatorBuilder {
	return (&UntypedValidatorBuilder{}).Reset()
}

// UntypedValidatorBuilder builds untyped validators
type UntypedValidatorBuilder struct {
	err error
	v   *untypedValidator
}

func (b *UntypedValidatorBuilder) Enum(values ...any) *UntypedValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.v.enum = make([]any, len(values))
	copy(b.v.enum, values)
	return b
}

func (b *UntypedValidatorBuilder) Const(value any) *UntypedValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.v.constantValue = &value
	return b
}

func (b *UntypedValidatorBuilder) Build() (Interface, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.v, nil
}

func (b *UntypedValidatorBuilder) MustBuild() Interface {
	if b.err != nil {
		panic(b.err)
	}
	return b.v
}

func (b *UntypedValidatorBuilder) Reset() *UntypedValidatorBuilder {
	b.err = nil
	b.v = &untypedValidator{}
	return b
}

func compileUntypedValidator(ctx context.Context, s *schema.Schema) (Interface, error) {
	v := Untyped()

	if s.HasEnum() && vocabulary.IsKeywordEnabledInContext(ctx, "enum") {
		v.Enum(s.Enum()...)
	}

	if s.HasConst() && vocabulary.IsKeywordEnabledInContext(ctx, "const") {
		v.Const(s.Const())
	}

	return v.Build()
}

func (u *untypedValidator) Validate(ctx context.Context, value any) (Result, error) {
	// Check const first (more specific)
	if u.constantValue != nil {
		if err := validateConst(ctx, value, *u.constantValue); err != nil {
			return nil, err
		}
		//nolint: nilnil
		return nil, nil
	}

	// Check enum
	if len(u.enum) > 0 {
		if err := validateEnum(ctx, value, u.enum); err != nil {
			return nil, err
		}
	}

	//nolint: nilnil
	return nil, nil
}

// validateConst checks if a value exactly matches the expected constant value
func validateConst(ctx context.Context, value any, constValue any) error {
	logger := TraceSlogFromContext(ctx)
	logger.InfoContext(ctx, "validating const constraint", "expected", constValue, "actual", value)

	if !reflect.DeepEqual(value, constValue) {
		return fmt.Errorf(`must be const value %v`, constValue)
	}
	return nil
}

// validateEnum checks if a value is found in the allowed enum values
func validateEnum(ctx context.Context, value any, enumValues []any) error {
	logger := TraceSlogFromContext(ctx)
	logger.InfoContext(ctx, "validating enum constraint", "allowed_values", enumValues, "actual", value)

	for _, enumVal := range enumValues {
		if reflect.DeepEqual(value, enumVal) {
			return nil
		}
	}
	return fmt.Errorf(`invalid value: %v not found in enum %v`, value, enumValues)
}
