package validator

import (
	"context"
	"fmt"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/internal/schemactx"
)

// inferredNumberValidator validates numeric constraints only when the value is a number,
// ignoring non-numeric values (for inferred number types without explicit type declaration)
type inferredNumberValidator struct {
	numberValidator Interface
}

func compileInferredNumberValidator(ctx context.Context, s *schema.Schema) (Interface, error) {
	// Create the underlying number validator
	numValidator, err := compileNumberValidator(ctx, s)
	if err != nil {
		return nil, err
	}

	return &inferredNumberValidator{
		numberValidator: numValidator,
	}, nil
}

func (v *inferredNumberValidator) Validate(ctx context.Context, in any) (Result, error) {
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

func (e *EmptyValidator) Validate(_ context.Context, _ any) (Result, error) {
	// Empty schema allows anything
	//nolint: nilnil
	return nil, nil
}

type NotValidator struct {
	validator Interface
}

func (n *NotValidator) Validate(ctx context.Context, v any) (Result, error) {
	_, err := n.validator.Validate(ctx, v)
	if err == nil {
		return nil, fmt.Errorf(`not validation failed: value should not validate against the schema`)
	}
	//nolint: nilnil
	return nil, nil
}

type nullValidator struct{}

// Null creates a validator that accepts only null values.
func Null() Interface {
	return nullValidator{}
}

func (nullValidator) Validate(_ context.Context, v any) (Result, error) {
	if v == nil {
		//nolint: nilnil
		return nil, nil
	}
	return nil, fmt.Errorf(`invalid value passed to NullValidator: expected null, got %T`, v)
}

// RefUnevaluatedPropertiesCompositionValidator handles complex unevaluatedProperties with $ref
type RefUnevaluatedPropertiesCompositionValidator struct {
	refValidator  Interface
	baseValidator Interface
	schema        *schema.Schema
}

// NewRefUnevaluatedPropertiesCompositionValidator creates a new RefUnevaluatedPropertiesCompositionValidator instance.
func NewRefUnevaluatedPropertiesCompositionValidator(ctx context.Context, s *schema.Schema, refValidator Interface) *RefUnevaluatedPropertiesCompositionValidator {
	v := &RefUnevaluatedPropertiesCompositionValidator{
		schema:       s,
		refValidator: refValidator,
	}

	// Compile base validator (everything except $ref)
	baseSchema := createSchemaWithoutRef(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		panic(fmt.Sprintf("failed to compile base schema: %v", err))
	}
	v.baseValidator = baseValidator

	return v
}

func (v *RefUnevaluatedPropertiesCompositionValidator) Validate(ctx context.Context, in any) (Result, error) {
	// First, validate the $ref and collect its annotations
	refResult, err := v.refValidator.Validate(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("$ref validation failed: %w", err)
	}

	// Now validate base constraints, passing the evaluated properties from $ref
	baseResult, err := v.validateBaseWithContext(ctx, in, refResult)
	if err != nil {
		return nil, err
	}

	// Merge the base result with $ref result
	var finalResult *ObjectResult
	if err := MergeResults(&finalResult, refResult, baseResult); err != nil {
		// Fall back to simple merging if MergeResults fails
		if objRefResult, ok := refResult.(*ObjectResult); ok {
			if objBaseResult, ok := baseResult.(*ObjectResult); ok {
				finalResult = mergeObjectResults(objRefResult, objBaseResult)
			} else {
				finalResult = objRefResult
			}
		} else if objBaseResult, ok := baseResult.(*ObjectResult); ok {
			finalResult = objBaseResult
		}
	}
	return finalResult, nil
}

// validateBaseWithContext validates the base schema with annotation context from $ref
func (v *RefUnevaluatedPropertiesCompositionValidator) validateBaseWithContext(ctx context.Context, in any, refResult Result) (Result, error) {
	// Create context with evaluated properties if we have evaluation results from $ref
	if objResult, ok := refResult.(*ObjectResult); ok && objResult != nil {
		if evalProps := objResult.EvaluatedProperties(); len(evalProps) > 0 {
			// Get existing evaluation context or create a new one
			var ec *schemactx.EvaluationContext
			_ = schemactx.EvaluationContextFromContext(ctx, &ec)
			if ec == nil {
				ec = &schemactx.EvaluationContext{}
			}

			// Mark properties as evaluated
			for prop := range evalProps {
				if evalProps[prop] {
					ec.Properties.MarkEvaluated(prop)
				}
			}

			ctx = schemactx.WithEvaluationContext(ctx, ec)
		}
	}

	return v.baseValidator.Validate(ctx, in)
}
