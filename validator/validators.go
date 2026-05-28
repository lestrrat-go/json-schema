package validator

import (
	"context"
	"fmt"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/vocabulary"
)

// inferredNumberValidator validates numeric constraints only when the value is a number,
// ignoring non-numeric values (for inferred number types without explicit type declaration)
type inferredNumberValidator struct {
	numberValidator Interface
}

func compileInferredNumberValidator(s *schema.Schema, vocab *vocabulary.VocabularySet) (Interface, error) {
	// Create the underlying number validator
	numValidator, err := compileNumberValidator(s, vocab)
	if err != nil {
		return nil, err
	}

	return &inferredNumberValidator{
		numberValidator: numValidator,
	}, nil
}

func (v *inferredNumberValidator) Validate(ctx context.Context, in any, _ ...ValidateOption) (Result, error) {
	// Check if the value is numeric
	rv := reflect.ValueOf(in)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		// Value is numeric, apply number validation
		return v.numberValidator.Validate(ctx, in)
	default:
		// Value is not numeric, ignore numeric constraints (per JSON Schema spec)
		//nolint: nilnil
		return nil, nil
	}
}

type EmptyValidator struct{}

func (e *EmptyValidator) Validate(_ context.Context, _ any, _ ...ValidateOption) (Result, error) {
	// Empty schema allows anything
	//nolint: nilnil
	return nil, nil
}

type NotValidator struct {
	validator Interface
}

func (n *NotValidator) Validate(ctx context.Context, v any, options ...ValidateOption) (Result, error) {
	return n.evaluate(ctx, v, newEvalState(ctx, options))
}

func (n *NotValidator) evaluate(ctx context.Context, v any, st *evalState) (Result, error) {
	_, err := evalChild(ctx, n.validator, v, st)
	if err == nil {
		return nil, fmt.Errorf(`not validation failed: value should not validate against the schema`)
	}
	//nolint: nilnil
	return nil, nil
}

type nullValidator struct{}

func Null() Interface {
	return nullValidator{}
}

func (nullValidator) Validate(_ context.Context, v any, _ ...ValidateOption) (Result, error) {
	if v == nil {
		//nolint: nilnil
		return nil, nil
	}
	return nil, fmt.Errorf(`invalid value passed to NullValidator: expected null, got %T`, v)
}
