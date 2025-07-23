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

func compileNullValidator(_ context.Context, _ *schema.Schema) (Interface, error) {
	return nullValidator{}, nil
}

// unevaluatedPropertiesValidator handles complex unevaluatedProperties with allOf
type unevaluatedPropertiesValidator struct {
	allOfValidators []Interface
	baseValidator   Interface
	schema          *schema.Schema
}

// unevaluatedItemsValidator handles complex unevaluatedItems with allOf
type unevaluatedItemsValidator struct {
	allOfValidators []Interface
	baseValidator   Interface
	schema          *schema.Schema
}

func compileUnevaluatedPropertiesValidator(ctx context.Context, s *schema.Schema) (*unevaluatedPropertiesValidator, error) {
	v := &unevaluatedPropertiesValidator{
		schema: s,
	}

	// Compile allOf validators
	for _, subSchema := range s.AllOf() {
		subValidator, err := Compile(ctx, convertSchemaOrBool(subSchema))
		if err != nil {
			return nil, fmt.Errorf("failed to compile allOf validator: %w", err)
		}
		v.allOfValidators = append(v.allOfValidators, subValidator)
	}

	// Compile base validator (everything except allOf)
	baseSchema := createBaseSchema(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to compile base schema: %w", err)
	}
	v.baseValidator = baseValidator

	return v, nil
}

func compileUnevaluatedItemsValidator(ctx context.Context, s *schema.Schema) (*unevaluatedItemsValidator, error) {
	v := &unevaluatedItemsValidator{
		schema: s,
	}

	// Compile allOf validators
	for _, subSchema := range s.AllOf() {
		subValidator, err := Compile(ctx, convertSchemaOrBool(subSchema))
		if err != nil {
			return nil, fmt.Errorf(`failed to compile allOf subschema: %w`, err)
		}
		v.allOfValidators = append(v.allOfValidators, subValidator)
	}

	// Create a copy of the schema without allOf for base validation
	baseSchema := createBaseSchema(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		return nil, fmt.Errorf(`failed to compile base schema: %w`, err)
	}
	v.baseValidator = baseValidator

	return v, nil
}

// User: This looks suspiciously like an allOf() validator with the base validator being
// evaluated at the end. Instead of createing yet another validator, can't we just
// use the existing allOf() validator and pass the base validator as the last one? The only
// major difference loooks like we need to pass the evaluated properties from the first
// set of allOf validators to the base validator.
//
// Perhaps, what's happening here is more of a collecting data from an allOf() validator
// and then passing the collected data to a child validator, which is the base validator.
func (v *unevaluatedPropertiesValidator) Validate(ctx context.Context, in any) (Result, error) {
	// Execute allOf validators and collect their annotations
	merger, err := executeValidatorsAndMergeResults(ctx, v.allOfValidators, in, "allOf")
	if err != nil {
		return nil, err
	}

	// Validate base constraints with annotations from allOf subschemas
	baseResult, err := v.validateBaseWithContext(ctx, in, merger.ObjectResult(), merger.ArrayResult())
	if err != nil {
		return nil, err
	}

	// Merge base result with allOf results
	merger.mergeResult(baseResult)
	return merger.FinalResult(), nil
}

func (v *unevaluatedItemsValidator) Validate(ctx context.Context, in any) (Result, error) {
	// Execute allOf validators with context flow for array items
	merger, err := executeValidatorsWithContextFlow(ctx, v.allOfValidators, in)
	if err != nil {
		return nil, err
	}

	// Create context with accumulated evaluations for base validation
	newCtx := withEvaluatedResults(ctx, merger.ObjectResult(), merger.ArrayResult())

	// Validate base constraints (including unevaluatedItems) with context from allOf
	baseResult, err := v.baseValidator.Validate(newCtx, in)
	if err != nil {
		return nil, err
	}

	// Merge base result with allOf results
	merger.mergeResult(baseResult)
	return merger.FinalResult(), nil
}

// RefUnevaluatedPropertiesCompositionValidator handles complex unevaluatedProperties with $ref
type RefUnevaluatedPropertiesCompositionValidator struct {
	refValidator  Interface
	baseValidator Interface
	schema        *schema.Schema
}

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
		evalProps := objResult.EvaluatedProperties()
		if len(evalProps) > 0 {
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
