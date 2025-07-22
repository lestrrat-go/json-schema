package validator

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
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
	// For allOf, collect all results and merge them while passing context between validators
	var mergedObjectResult *ObjectResult
	var mergedArrayResult *ArrayResult

	for i, subv := range v.validators {
		// Create context with accumulated annotations for this validator
		var currentCtx context.Context = ctx

		// Add evaluated items if we have them (items annotations flow between allOf subschemas)
		if mergedArrayResult != nil {
			evalItems := mergedArrayResult.EvaluatedItems()
			if len(evalItems) > 0 {
				currentCtx = schema.WithEvaluatedItems(ctx, evalItems)
			}
		}

		// NOTE: We do NOT pass evaluated properties between allOf subschemas
		// This implements the "cousin" semantics where properties evaluated by one
		// subschema are not visible to other subschemas in the same allOf

		result, err := subv.Validate(currentCtx, in)
		if err != nil {
			return nil, fmt.Errorf(`allOf validation failed: validator #%d failed: %w`, i, err)
		}
		// Merge object results for property evaluation tracking
		if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
			if mergedObjectResult == nil {
				mergedObjectResult = NewObjectResult()
			}
			for prop := range objResult.EvaluatedProperties() {
				mergedObjectResult.SetEvaluatedProperty(prop)
			}
		}

		// Merge array results for item evaluation tracking
		if arrResult, ok := result.(*ArrayResult); ok && arrResult != nil {
			if mergedArrayResult == nil {
				mergedArrayResult = NewArrayResult()
			}
			arrItems := arrResult.EvaluatedItems()
			for i, evaluated := range arrItems {
				if evaluated {
					mergedArrayResult.SetEvaluatedItem(i)
				}
			}
		}
	}

	// Return appropriate result type based on what we merged
	if mergedObjectResult != nil && mergedArrayResult != nil {
		// Both object and array results - this shouldn't happen in normal validation
		// but prioritize object result for now
		return mergedObjectResult, nil
	}

	if mergedObjectResult != nil {
		return mergedObjectResult, nil
	}

	if mergedArrayResult != nil {
		return mergedArrayResult, nil
	}

	//nolint:nilnil
	return nil, nil
}

type anyOfValidator struct {
	validators []Interface
}

func (v *anyOfValidator) Validate(ctx context.Context, in any) (Result, error) {
	for _, subv := range v.validators {
		result, err := subv.Validate(ctx, in)
		if err == nil {
			return result, nil
		}
	}
	return nil, fmt.Errorf(`anyOf validation failed: none of the validators passed`)
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

// hasBaseConstraints checks if a schema has base-level constraints that need validation
// when used with allOf/anyOf/oneOf
func hasBaseConstraints(s *schema.Schema) bool {
	// Check for types separately since it's not a bit field check
	if len(s.Types()) > 0 {
		return true
	}

	// Use bit field approach for efficient checking of multiple constraints
	baseConstraintFields := schema.MinLengthField | schema.MaxLengthField | schema.PatternField |
		schema.MinimumField | schema.MaximumField | schema.ExclusiveMinimumField | schema.ExclusiveMaximumField | schema.MultipleOfField |
		schema.MinItemsField | schema.MaxItemsField | schema.UniqueItemsField | schema.ItemsField | schema.ContainsField | schema.UnevaluatedItemsField |
		schema.MinPropertiesField | schema.MaxPropertiesField | schema.RequiredField | schema.PropertiesField | schema.PatternPropertiesField | schema.AdditionalPropertiesField | schema.UnevaluatedPropertiesField | schema.DependentSchemasField | schema.PropertyNamesField |
		schema.EnumField | schema.ConstField

	// Returns true if ANY of the base constraint fields are set
	return s.HasAny(baseConstraintFields)
}
