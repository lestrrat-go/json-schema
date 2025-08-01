package validator

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// IfThenElseValidator handles if/then/else conditional validation
type IfThenElseValidator struct {
	ifValidator   Interface
	thenValidator Interface
	elseValidator Interface
}

func compileIfThenElseValidator(ctx context.Context, s *schema.Schema) (Interface, error) {
	v := &IfThenElseValidator{}

	// Compile 'if' validator (required)
	ifSchema := convertSchemaOrBool(s.IfSchema())
	ifValidator, err := Compile(ctx, ifSchema)
	if err != nil {
		return nil, fmt.Errorf(`failed to compile if validator: %w`, err)
	}
	v.ifValidator = ifValidator

	// Compile 'then' validator (optional)
	if s.Has(schema.ThenSchemaField) {
		thenSchema := convertSchemaOrBool(s.ThenSchema())
		thenValidator, err := Compile(ctx, thenSchema)
		if err != nil {
			return nil, fmt.Errorf(`failed to compile then validator: %w`, err)
		}
		v.thenValidator = thenValidator
	}

	// Compile 'else' validator (optional)
	if s.Has(schema.ElseSchemaField) {
		elseSchema := convertSchemaOrBool(s.ElseSchema())
		elseValidator, err := Compile(ctx, elseSchema)
		if err != nil {
			return nil, fmt.Errorf(`failed to compile else validator: %w`, err)
		}
		v.elseValidator = elseValidator
	}

	return v, nil
}

func (v *IfThenElseValidator) Validate(ctx context.Context, in any) (Result, error) {
	// First, check the 'if' condition and collect its annotations
	ifResult, ifErr := v.ifValidator.Validate(ctx, in)

	// The 'if' schema contributes annotations regardless of whether it passes or fails
	var conditionalResult Result

	if ifErr == nil {
		// 'if' condition passed, validate against 'then' if it exists
		if v.thenValidator != nil {
			thenResult, err := v.thenValidator.Validate(ctx, in)
			if err != nil {
				return nil, err
			}
			// Merge 'if' and 'then' results
			conditionalResult = mergeGenericResults(ifResult, thenResult)
		} else {
			// Only 'if' result
			conditionalResult = ifResult
		}
	} else {
		// 'if' condition failed, validate against 'else' if it exists
		if v.elseValidator != nil {
			elseResult, err := v.elseValidator.Validate(ctx, in)
			if err != nil {
				return nil, err
			}
			// Merge 'if' and 'else' results
			conditionalResult = mergeGenericResults(ifResult, elseResult)
		} else {
			// Only 'if' result (even though it failed validation, it may have annotations)
			conditionalResult = ifResult
		}
	}

	return conditionalResult, nil
}

// IfThenElseUnevaluatedPropertiesCompositionValidator handles complex unevaluatedProperties with if/then/else
type IfThenElseUnevaluatedPropertiesCompositionValidator struct {
	ifValidator   Interface
	thenValidator Interface
	elseValidator Interface
	baseValidator Interface
	schema        *schema.Schema
}

// NewIfThenElseUnevaluatedPropertiesCompositionValidator creates a new IfThenElseUnevaluatedPropertiesCompositionValidator instance.
func NewIfThenElseUnevaluatedPropertiesCompositionValidator(ctx context.Context, s *schema.Schema) *IfThenElseUnevaluatedPropertiesCompositionValidator {
	v := &IfThenElseUnevaluatedPropertiesCompositionValidator{
		schema: s,
	}

	// Compile if validator
	ifSchema := convertSchemaOrBool(s.IfSchema())
	ifValidator, err := Compile(ctx, ifSchema)
	if err != nil {
		panic(fmt.Sprintf("failed to compile if validator: %v", err))
	}
	v.ifValidator = ifValidator

	// Compile then validator if it exists
	if s.Has(schema.ThenSchemaField) {
		thenSchema := convertSchemaOrBool(s.ThenSchema())
		thenValidator, err := Compile(ctx, thenSchema)
		if err != nil {
			panic(fmt.Sprintf("failed to compile then validator: %v", err))
		}
		v.thenValidator = thenValidator
	}

	// Compile else validator if it exists
	if s.Has(schema.ElseSchemaField) {
		elseSchema := convertSchemaOrBool(s.ElseSchema())
		elseValidator, err := Compile(ctx, elseSchema)
		if err != nil {
			panic(fmt.Sprintf("failed to compile else validator: %v", err))
		}
		v.elseValidator = elseValidator
	}

	// Compile base validator (everything except if/then/else)
	baseSchema := createBaseSchema(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		panic(fmt.Sprintf("failed to compile base schema: %v", err))
	}
	v.baseValidator = baseValidator

	return v
}

func (v *IfThenElseUnevaluatedPropertiesCompositionValidator) Validate(ctx context.Context, in any) (Result, error) {
	// First, evaluate if/then/else and collect annotations
	var conditionalResult *ObjectResult

	// Check the 'if' condition and collect its annotations
	ifResult, ifErr := v.ifValidator.Validate(ctx, in)

	// Collect annotations from 'if' schema (contributes regardless of outcome)
	if ifObjResult, ok := ifResult.(*ObjectResult); ok && ifObjResult != nil {
		conditionalResult = NewObjectResult()
		for prop := range ifObjResult.EvaluatedProperties() {
			conditionalResult.SetEvaluatedProperty(prop)
		}
	}

	if ifErr == nil {
		// 'if' condition passed, validate against 'then' if it exists
		if v.thenValidator != nil {
			result, err := v.thenValidator.Validate(ctx, in)
			if err != nil {
				return nil, fmt.Errorf(`if/then validation failed: %w`, err)
			}
			// Merge annotations from 'then' with 'if' annotations
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				if conditionalResult == nil {
					conditionalResult = NewObjectResult()
				}
				for prop := range objResult.EvaluatedProperties() {
					conditionalResult.SetEvaluatedProperty(prop)
				}
			}
		}
	} else {
		// 'if' condition failed, validate against 'else' if it exists
		if v.elseValidator != nil {
			result, err := v.elseValidator.Validate(ctx, in)
			if err != nil {
				return nil, fmt.Errorf(`if/else validation failed: %w`, err)
			}
			// Merge annotations from 'else' with 'if' annotations
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				if conditionalResult == nil {
					conditionalResult = NewObjectResult()
				}
				for prop := range objResult.EvaluatedProperties() {
					conditionalResult.SetEvaluatedProperty(prop)
				}
			}
		}
	}

	// Now validate base constraints, passing the evaluated properties from if/then/else
	baseResult, err := v.validateBaseWithContext(ctx, in, conditionalResult)
	if err != nil {
		return nil, err
	}

	// Merge the base result with if/then/else result
	if baseObjResult, ok := baseResult.(*ObjectResult); ok && baseObjResult != nil {
		if conditionalResult == nil {
			conditionalResult = NewObjectResult()
		}
		for prop := range baseObjResult.EvaluatedProperties() {
			conditionalResult.SetEvaluatedProperty(prop)
		}
	}

	return conditionalResult, nil
}

// validateBaseWithContext for if/then/else
func (v *IfThenElseUnevaluatedPropertiesCompositionValidator) validateBaseWithContext(ctx context.Context, in any, previousResult *ObjectResult) (Result, error) {
	// Create context with evaluated properties if we have previous evaluation results
	if previousResult != nil {
		if evalProps := previousResult.EvaluatedProperties(); len(evalProps) > 0 {
			ctx = withEvaluatedProperties(ctx, evalProps)
		}
	}

	return v.baseValidator.Validate(ctx, in)
}

