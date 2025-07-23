package validator

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/json-schema/internal/schemactx"
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

	//nolint:nilnil
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

// withEvaluatedProperties creates a new context with evaluated properties
func withEvaluatedProperties(ctx context.Context, evaluatedProps map[string]bool) context.Context {
	if len(evaluatedProps) == 0 {
		return ctx
	}

	// Get existing evaluation context or create a new one
	var ec *schemactx.EvaluationContext
	_ = schemactx.EvaluationContextFromContext(ctx, &ec)
	if ec == nil {
		ec = &schemactx.EvaluationContext{}
	}

	// Mark properties as evaluated
	for prop := range evaluatedProps {
		if evaluatedProps[prop] {
			ec.Properties.MarkEvaluated(prop)
		}
	}

	return schemactx.WithEvaluationContext(ctx, ec)
}

// withEvaluatedItems creates a new context with evaluated items
func withEvaluatedItems(ctx context.Context, evaluatedItems []bool) context.Context {
	if len(evaluatedItems) == 0 {
		return ctx
	}

	// Get existing evaluation context or create a new one
	var ec *schemactx.EvaluationContext
	_ = schemactx.EvaluationContextFromContext(ctx, &ec)
	if ec == nil {
		ec = &schemactx.EvaluationContext{}
	}

	// Copy evaluated items
	for i, evaluated := range evaluatedItems {
		if evaluated {
			ec.Items.Set(i, true)
		}
	}

	return schemactx.WithEvaluationContext(ctx, ec)
}

// withEvaluatedResults creates a new context with evaluated properties and items from results
func withEvaluatedResults(ctx context.Context, objResult *ObjectResult, arrResult *ArrayResult) context.Context {
	// Get existing evaluation context or create a new one
	var ec *schemactx.EvaluationContext
	_ = schemactx.EvaluationContextFromContext(ctx, &ec)
	if ec == nil {
		ec = &schemactx.EvaluationContext{}
	}

	// Add evaluated properties
	if objResult != nil {
		evalProps := objResult.EvaluatedProperties()
		for prop := range evalProps {
			if evalProps[prop] {
				ec.Properties.MarkEvaluated(prop)
			}
		}
	}

	// Add evaluated items
	if arrResult != nil {
		evalItems := arrResult.EvaluatedItems()
		for i, evaluated := range evalItems {
			if evaluated {
				ec.Items.Set(i, true)
			}
		}
	}

	return schemactx.WithEvaluationContext(ctx, ec)
}

// executeValidatorsAndMergeResults executes all validators and merges their results
// Returns the result merger and any error encountered
func executeValidatorsAndMergeResults(ctx context.Context, validators []Interface, input any, validatorType string) (*resultMerger, error) {
	var merger resultMerger

	for i, validator := range validators {
		result, err := validator.Validate(ctx, input)
		if err != nil {
			return nil, fmt.Errorf(`%s validation failed: validator #%d failed: %w`, validatorType, i, err)
		}
		merger.mergeResult(result)
	}

	return &merger, nil
}

// executeValidatorsWithContextFlow executes validators with context flow for array items
// (items annotations flow between validators)
func executeValidatorsWithContextFlow(ctx context.Context, validators []Interface, input any) (*resultMerger, error) {
	var merger resultMerger
	currentCtx := ctx

	for i, validator := range validators {
		// Update context with accumulated array results for context flow
		if merger.ArrayResult() != nil {
			evalItems := merger.ArrayResult().EvaluatedItems()
			if len(evalItems) > 0 {
				currentCtx = withEvaluatedItems(currentCtx, evalItems)
			}
		}

		result, err := validator.Validate(currentCtx, input)
		if err != nil {
			return nil, fmt.Errorf(`validator #%d failed: %w`, i, err)
		}

		merger.mergeResult(result)
	}

	return &merger, nil
}
