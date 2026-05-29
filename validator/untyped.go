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
	hasEnum       bool // Distinguishes an empty enum (rejects all) from no enum constraint
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
	b.v.hasEnum = true
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

func compileUntypedValidator(s *schema.Schema, vocab *vocabulary.VocabularySet) (Interface, error) {
	v := Untyped()

	if s.HasEnum() && vocab.IsKeywordEnabled("enum") {
		v.Enum(s.Enum()...)
	}

	if s.HasConst() && vocab.IsKeywordEnabled("const") {
		v.Const(s.Const())
	}

	return v.Build()
}

func (u *untypedValidator) Validate(ctx context.Context, value any, _ ...ValidateOption) (Result, error) {
	// Check const first (more specific)
	if u.constantValue != nil {
		if err := validateConst(ctx, value, *u.constantValue); err != nil {
			return nil, err
		}
		//nolint: nilnil
		return nil, nil
	}

	// Check enum. An empty enum is a valid constraint that rejects every value,
	// so gate on whether enum was set rather than on its length.
	if u.hasEnum {
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

	if !jsonSchemaEqual(value, constValue) {
		return fmt.Errorf(`must be const value %v`, constValue)
	}
	return nil
}

// validateEnum checks if a value is found in the allowed enum values
func validateEnum(ctx context.Context, value any, enumValues []any) error {
	logger := TraceSlogFromContext(ctx)
	logger.InfoContext(ctx, "validating enum constraint", "allowed_values", enumValues, "actual", value)

	for _, enumVal := range enumValues {
		if jsonSchemaEqual(value, enumVal) {
			return nil
		}
	}
	return fmt.Errorf(`invalid value: %v not found in enum %v`, value, enumValues)
}

// jsonSchemaEqual compares two values according to JSON Schema equality rules
// This handles numeric type equivalence (5 == 5.0) as required by JSON Schema spec
func jsonSchemaEqual(a, b any) bool {
	// First try direct equality (handles same types efficiently)
	if reflect.DeepEqual(a, b) {
		return true
	}

	// Handle numeric comparisons specially
	aNum, aIsNum := convertToNumber(a)
	bNum, bIsNum := convertToNumber(b)

	if aIsNum && bIsNum {
		// Both are numbers - compare their mathematical values
		return aNum == bNum
	}

	// For non-numeric types, fall back to reflect.DeepEqual
	return false
}

// convertToNumber converts a value to float64 if it's a numeric type. It
// recognizes native numeric kinds and json.Number (see validator/numeric.go),
// so enum/const equality treats 5, 5.0, and json.Number("5") as equal.
func convertToNumber(v any) (float64, bool) {
	f, ok, err := numericFloat(v)
	if err != nil || !ok {
		return 0, false
	}
	return f, true
}
