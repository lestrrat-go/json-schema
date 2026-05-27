package validator

import (
	"context"
	"fmt"
)

// AllOf is a convnience function to create a Validator that can handle allOf validation.
func AllOf(validators ...Interface) Interface {
	return &allOfValidator{
		validators: validators,
	}
}

func AnyOf(validators ...Interface) Interface {
	return &anyOfValidator{
		validators: validators,
	}
}

func OneOf(validators ...Interface) Interface {
	return &oneOfValidator{
		validators: validators,
	}
}

type allOfValidator struct {
	validators []Interface
}

func (v *allOfValidator) Validate(ctx context.Context, in any) (Result, error) {
	// Use executeValidatorsWithContextFlow with context flow for array items
	// NOTE: We do NOT pass evaluated properties between allOf subschemas
	// This implements the "cousin" semantics where properties evaluated by one
	// subschema are not visible to other subschemas in the same allOf
	merger, err := executeValidatorsWithContextFlow(ctx, v.validators, in)
	if err != nil {
		return nil, fmt.Errorf(`allOf validation failed: %w`, err)
	}
	return merger.FinalResult(), nil
}

type anyOfValidator struct {
	validators []Interface
}

func (v *anyOfValidator) Validate(ctx context.Context, in any) (Result, error) {
	var resultMerger resultMerger
	anyPassed := false

	// According to JSON Schema spec, anyOf must collect annotations from ALL passing validators
	for _, subv := range v.validators {
		result, err := subv.Validate(ctx, in)
		if err == nil {
			anyPassed = true
			resultMerger.mergeResult(result)
			// Continue checking other validators to collect all annotations
		}
	}

	if !anyPassed {
		return nil, fmt.Errorf(`anyOf validation failed: none of the validators passed`)
	}

	return resultMerger.FinalResult(), nil
}

type oneOfValidator struct {
	validators []Interface
}

func (v *oneOfValidator) Validate(ctx context.Context, in any) (Result, error) {
	passedCount := 0
	var validResult Result
	for _, subv := range v.validators {
		result, err := subv.Validate(ctx, in)
		if err == nil {
			passedCount++
			validResult = result
		}
	}
	if passedCount == 0 {
		return nil, fmt.Errorf(`oneOf validation failed: none of the validators passed`)
	}
	if passedCount > 1 {
		return nil, fmt.Errorf(`oneOf validation failed: more than one validator passed (%d), expected exactly one`, passedCount)
	}
	return validResult, nil
}
