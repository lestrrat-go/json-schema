package validator

import (
	"context"
	"fmt"
)

// resultMerger handles the common pattern of merging ObjectResult and ArrayResult
// from multiple validation operations (unexported since it's an internal utility)
type resultMerger struct {
	objectResult *ObjectResult
	arrayResult  *ArrayResult
}

// mergeResult merges a validation result into the accumulated results
func (rm *resultMerger) mergeResult(result Result) {
	switch res := result.(type) {
	case *ObjectResult:
		if res != nil {
			rm.mergeObjectResult(res)
		}
	case *ArrayResult:
		if res != nil {
			rm.mergeArrayResult(res)
		}
	}
}

// mergeObjectResult merges an ObjectResult into the accumulated object result
func (rm *resultMerger) mergeObjectResult(objResult *ObjectResult) {
	if rm.objectResult == nil {
		rm.objectResult = NewObjectResult()
	}
	for prop := range objResult.EvaluatedProperties() {
		rm.objectResult.SetEvaluatedProperty(prop)
	}
}

// mergeArrayResult merges an ArrayResult into the accumulated array result
func (rm *resultMerger) mergeArrayResult(arrResult *ArrayResult) {
	if rm.arrayResult == nil {
		rm.arrayResult = NewArrayResult()
	}
	arrItems := arrResult.EvaluatedItems()
	for i, evaluated := range arrItems {
		if evaluated {
			rm.arrayResult.SetEvaluatedItem(i)
		}
	}
}

// FinalResult returns the appropriate result without losing annotations.
// This fixes the original issue where object results were arbitrarily prioritized.
func (rm *resultMerger) FinalResult() Result {
	// If we have both results, prioritize array result since array validation
	// is more context-sensitive for unevaluatedItems scenarios
	if rm.objectResult != nil && rm.arrayResult != nil {
		return rm.arrayResult
	}

	if rm.objectResult != nil {
		return rm.objectResult
	}

	if rm.arrayResult != nil {
		return rm.arrayResult
	}

	return nil
}

// ObjectResult returns the accumulated object result (may be nil)
func (rm *resultMerger) ObjectResult() *ObjectResult {
	return rm.objectResult
}

// ArrayResult returns the accumulated array result (may be nil)
func (rm *resultMerger) ArrayResult() *ArrayResult {
	return rm.arrayResult
}

// executeValidatorsAndMergeResults executes all validators and merges their results
// Returns the result merger and any error encountered
func executeValidatorsAndMergeResults(ctx context.Context, validators []Interface, input any, st *evalState, validatorType string) (*resultMerger, error) {
	var merger resultMerger

	for i, validator := range validators {
		result, err := evalChild(ctx, validator, input, st)
		if err != nil {
			return nil, fmt.Errorf(`%s validation failed: validator #%d failed: %w`, validatorType, i, err)
		}
		merger.mergeResult(result)
	}

	return &merger, nil
}
