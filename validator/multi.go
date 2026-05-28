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
	return v.evaluate(ctx, in, newEvalState(ctx))
}

func (v *allOfValidator) evaluate(ctx context.Context, in any, st *evalState) (Result, error) {
	// allOf subschemas are "cousins": each is evaluated independently and none
	// can see the items/properties evaluated by another. Their annotations are
	// merged upward for the parent (so a parent-level unevaluatedItems /
	// unevaluatedProperties does see them), but they are NOT shared sideways
	// between branches.
	merger, err := executeValidatorsAndMergeResults(ctx, v.validators, in, st, "allOf")
	if err != nil {
		return nil, err
	}
	return merger.FinalResult(), nil
}

type anyOfValidator struct {
	validators []Interface
}

func (v *anyOfValidator) Validate(ctx context.Context, in any) (Result, error) {
	return v.evaluate(ctx, in, newEvalState(ctx))
}

func (v *anyOfValidator) evaluate(ctx context.Context, in any, st *evalState) (Result, error) {
	var resultMerger resultMerger
	anyPassed := false

	// According to JSON Schema spec, anyOf must collect annotations from ALL passing validators
	for _, subv := range v.validators {
		result, err := evalChild(ctx, subv, in, st)
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
	return v.evaluate(ctx, in, newEvalState(ctx))
}

func (v *oneOfValidator) evaluate(ctx context.Context, in any, st *evalState) (Result, error) {
	passedCount := 0
	var validResult Result
	for _, subv := range v.validators {
		result, err := evalChild(ctx, subv, in, st)
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
