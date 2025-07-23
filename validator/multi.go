package validator

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/internal/schemactx"
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

// validateBaseWithContext validates the base schema with annotation context
func (v *unevaluatedPropertiesValidator) validateBaseWithContext(ctx context.Context, in any, previousObjectResult *ObjectResult, previousArrayResult *ArrayResult) (Result, error) {
	// Get existing evaluation context or create a new one
	var ec *schemactx.EvaluationContext
	_ = schemactx.EvaluationContextFromContext(ctx, &ec)
	if ec == nil {
		ec = &schemactx.EvaluationContext{}
	}

	if previousObjectResult != nil {
		evalProps := previousObjectResult.EvaluatedProperties()
		if len(evalProps) > 0 {
			// Mark properties as evaluated
			for prop := range evalProps {
				if evalProps[prop] {
					ec.Properties.MarkEvaluated(prop)
				}
			}
		}
	}

	if previousArrayResult != nil {
		evalItems := previousArrayResult.EvaluatedItems()
		if len(evalItems) > 0 {
			// Copy evaluated items
			for i, evaluated := range evalItems {
				if evaluated {
					ec.Items.Set(i, true)
				}
			}
		}
	}

	ctx = schemactx.WithEvaluationContext(ctx, ec)

	return v.baseValidator.Validate(ctx, in)
}

// AnyOfUnevaluatedPropertiesCompositionValidator handles complex unevaluatedProperties with anyOf
type AnyOfUnevaluatedPropertiesCompositionValidator struct {
	anyOfValidators []Interface
	baseValidator   Interface
	schema          *schema.Schema
}

func NewAnyOfUnevaluatedPropertiesCompositionValidator(s *schema.Schema) *AnyOfUnevaluatedPropertiesCompositionValidator {
	v, err := NewAnyOfUnevaluatedPropertiesCompositionValidatorWithResolver(context.Background(), s, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create anyOf composition validator: %v", err))
	}
	return v
}

func NewAnyOfUnevaluatedPropertiesCompositionValidatorWithResolver(ctx context.Context, s *schema.Schema, anyOfValidators []Interface, _ *schema.Resolver) (*AnyOfUnevaluatedPropertiesCompositionValidator, error) {
	v := &AnyOfUnevaluatedPropertiesCompositionValidator{
		schema: s,
	}

	// Use provided validators or compile them if not provided
	if anyOfValidators != nil {
		v.anyOfValidators = anyOfValidators
	} else {
		// Compile anyOf validators
		for _, subSchema := range s.AnyOf() {
			subValidator, err := Compile(ctx, convertSchemaOrBool(subSchema))
			if err != nil {
				return nil, fmt.Errorf("failed to compile anyOf validator: %w", err)
			}
			v.anyOfValidators = append(v.anyOfValidators, subValidator)
		}
	}

	// Compile base validator (everything except anyOf)
	baseSchema := createBaseSchema(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to compile base schema: %w", err)
	}
	v.baseValidator = baseValidator

	return v, nil
}

func (v *AnyOfUnevaluatedPropertiesCompositionValidator) Validate(ctx context.Context, in any) (Result, error) {
	// For anyOf, we need at least one subschema to pass and collect its annotations
	var validResult *ObjectResult
	anyOfPassed := false

	for _, subValidator := range v.anyOfValidators {
		result, err := subValidator.Validate(ctx, in)
		if err == nil {
			anyOfPassed = true
			// Collect annotations from ALL passing validators (not just the first)
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				if validResult == nil {
					validResult = NewObjectResult()
				}
				for prop := range objResult.EvaluatedProperties() {
					validResult.SetEvaluatedProperty(prop)
				}
			}
			// Continue to check other validators for annotation collection
		}
	}

	if !anyOfPassed {
		return nil, fmt.Errorf(`anyOf validation failed: none of the validators passed`)
	}

	// Now validate base constraints, passing the evaluated properties from anyOf
	baseResult, err := v.validateBaseWithContext(ctx, in, validResult)
	if err != nil {
		return nil, err
	}

	// Merge the base result with anyOf result
	if baseObjResult, ok := baseResult.(*ObjectResult); ok && baseObjResult != nil {
		if validResult == nil {
			validResult = NewObjectResult()
		}
		for prop := range baseObjResult.EvaluatedProperties() {
			validResult.SetEvaluatedProperty(prop)
		}
	}

	return validResult, nil
}

// validateBaseWithContext for AnyOf
func (v *AnyOfUnevaluatedPropertiesCompositionValidator) validateBaseWithContext(ctx context.Context, in any, previousResult *ObjectResult) (Result, error) {
	if objValidator, ok := v.baseValidator.(*objectValidator); ok {
		var previouslyEvaluated map[string]bool
		if previousResult != nil {
			previouslyEvaluated = previousResult.EvaluatedProperties()
		}
		if len(previouslyEvaluated) > 0 {
			// Get existing evaluation context or create a new one
			var ec *schemactx.EvaluationContext
			_ = schemactx.EvaluationContextFromContext(ctx, &ec)
			if ec == nil {
				ec = &schemactx.EvaluationContext{}
			}

			// Mark properties as evaluated
			for prop := range previouslyEvaluated {
				if previouslyEvaluated[prop] {
					ec.Properties.MarkEvaluated(prop)
				}
			}

			ctx = schemactx.WithEvaluationContext(ctx, ec)
		}
		return objValidator.Validate(ctx, in)
	}

	switch mv := v.baseValidator.(type) {
	case *anyOfValidator, *oneOfValidator:
		return mv.Validate(ctx, in)
	case *allOfValidator:
		return v.validateMultiValidatorWithContext(ctx, mv, in, previousResult)
	default:
		// For other validator types, just validate normally without annotation context
		return v.baseValidator.Validate(ctx, in)
	}
}

// validateMultiValidatorWithContext for AnyOf
func (v *AnyOfUnevaluatedPropertiesCompositionValidator) validateMultiValidatorWithContext(ctx context.Context, mv *allOfValidator, in any, previousResult *ObjectResult) (Result, error) {
	// For AND mode (allOf), validate each sub-validator independently (cousins cannot see each other)
	var mergedResult *ObjectResult
	if previousResult != nil {
		mergedResult = NewObjectResult()
		for prop := range previousResult.EvaluatedProperties() {
			mergedResult.SetEvaluatedProperty(prop)
		}
	}

	for i, subValidator := range mv.validators {
		// Each cousin validator should be validated independently
		// without seeing evaluated properties from other cousins
		// Only pass the original previousResult context, not accumulated cousin results
		if _, ok := subValidator.(*objectValidator); ok {
			var previouslyEvaluated map[string]bool
			if previousResult != nil {
				previouslyEvaluated = previousResult.EvaluatedProperties()
			}
			if len(previouslyEvaluated) > 0 {
				// Get existing evaluation context or create a new one
				var ec *schemactx.EvaluationContext
				_ = schemactx.EvaluationContextFromContext(ctx, &ec)
				if ec == nil {
					ec = &schemactx.EvaluationContext{}
				}

				// Mark properties as evaluated
				for prop := range previouslyEvaluated {
					if previouslyEvaluated[prop] {
						ec.Properties.MarkEvaluated(prop)
					}
				}

				ctx = schemactx.WithEvaluationContext(ctx, ec)
			}
		}
		result, err := subValidator.Validate(ctx, in)
		if err != nil {
			return nil, fmt.Errorf(`allOf validation failed: validator #%d failed: %w`, i, err)
		}

		// Merge object results
		if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
			if mergedResult == nil {
				mergedResult = NewObjectResult()
			}
			for prop := range objResult.EvaluatedProperties() {
				mergedResult.SetEvaluatedProperty(prop)
			}
		}
	}
	return mergedResult, nil
}

// OneOfUnevaluatedPropertiesCompositionValidator handles complex unevaluatedProperties with oneOf
type OneOfUnevaluatedPropertiesCompositionValidator struct {
	oneOfValidators []Interface
	baseValidator   Interface
	schema          *schema.Schema
}

func NewOneOfUnevaluatedPropertiesCompositionValidator(s *schema.Schema) *OneOfUnevaluatedPropertiesCompositionValidator {
	v, err := NewOneOfUnevaluatedPropertiesCompositionValidatorWithResolver(context.Background(), s, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create oneOf composition validator: %v", err))
	}
	return v
}

func NewOneOfUnevaluatedPropertiesCompositionValidatorWithResolver(ctx context.Context, s *schema.Schema, oneOfValidators []Interface, _ *schema.Resolver) (*OneOfUnevaluatedPropertiesCompositionValidator, error) {
	v := &OneOfUnevaluatedPropertiesCompositionValidator{
		schema: s,
	}

	// Use provided validators or compile them if not provided
	if oneOfValidators != nil {
		v.oneOfValidators = oneOfValidators
	} else {
		// Compile oneOf validators
		for _, subSchema := range s.OneOf() {
			subValidator, err := Compile(ctx, convertSchemaOrBool(subSchema))
			if err != nil {
				return nil, fmt.Errorf("failed to compile oneOf validator: %w", err)
			}
			v.oneOfValidators = append(v.oneOfValidators, subValidator)
		}
	}

	// Compile base validator (everything except oneOf)
	baseSchema := createBaseSchema(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to compile base schema: %w", err)
	}
	v.baseValidator = baseValidator

	return v, nil
}

func (v *OneOfUnevaluatedPropertiesCompositionValidator) Validate(ctx context.Context, in any) (Result, error) {
	// For oneOf, exactly one subschema must pass and we collect its annotations
	var validResult *ObjectResult
	passedCount := 0

	for _, subValidator := range v.oneOfValidators {
		result, err := subValidator.Validate(ctx, in)
		if err == nil {
			passedCount++
			// Collect annotations from the passing validator
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				validResult = NewObjectResult()
				for prop := range objResult.EvaluatedProperties() {
					validResult.SetEvaluatedProperty(prop)
				}
			}
		}
	}

	if passedCount == 0 {
		return nil, fmt.Errorf(`oneOf validation failed: none of the validators passed`)
	}
	if passedCount > 1 {
		return nil, fmt.Errorf(`oneOf validation failed: more than one validator passed (%d), expected exactly one`, passedCount)
	}

	// Now validate base constraints, passing the evaluated properties from oneOf
	baseResult, err := v.validateBaseWithContext(ctx, in, validResult)
	if err != nil {
		return nil, err
	}

	// Merge the base result with oneOf result
	if baseObjResult, ok := baseResult.(*ObjectResult); ok && baseObjResult != nil {
		if validResult == nil {
			validResult = NewObjectResult()
		}
		for prop := range baseObjResult.EvaluatedProperties() {
			validResult.SetEvaluatedProperty(prop)
		}
	}

	return validResult, nil
}

// validateBaseWithContext for OneOf
func (v *OneOfUnevaluatedPropertiesCompositionValidator) validateBaseWithContext(ctx context.Context, in any, previousResult *ObjectResult) (Result, error) {
	if objValidator, ok := v.baseValidator.(*objectValidator); ok {
		var previouslyEvaluated map[string]bool
		if previousResult != nil {
			previouslyEvaluated = previousResult.EvaluatedProperties()
		}
		if len(previouslyEvaluated) > 0 {
			// Get existing evaluation context or create a new one
			var ec *schemactx.EvaluationContext
			_ = schemactx.EvaluationContextFromContext(ctx, &ec)
			if ec == nil {
				ec = &schemactx.EvaluationContext{}
			}

			// Mark properties as evaluated
			for prop := range previouslyEvaluated {
				if previouslyEvaluated[prop] {
					ec.Properties.MarkEvaluated(prop)
				}
			}

			ctx = schemactx.WithEvaluationContext(ctx, ec)
		}
		return objValidator.Validate(ctx, in)
	}

	switch mv := v.baseValidator.(type) {
	case *anyOfValidator, *oneOfValidator:
		return mv.Validate(ctx, in)
	case *allOfValidator:
		// If the base validator is a allOfValidator, we need to handle it specially
		return v.validateMultiValidatorWithContext(ctx, mv, in, previousResult)
	default:
		// For other validator types, just validate normally without annotation context
		return v.baseValidator.Validate(ctx, in)
	}
}

// validateMultiValidatorWithContext for OneOf
func (v *OneOfUnevaluatedPropertiesCompositionValidator) validateMultiValidatorWithContext(ctx context.Context, mv *allOfValidator, in any, previousResult *ObjectResult) (Result, error) {
	// For AND mode (allOf), validate each sub-validator independently (cousins cannot see each other)
	var mergedResult *ObjectResult
	if previousResult != nil {
		mergedResult = NewObjectResult()
		for prop := range previousResult.EvaluatedProperties() {
			mergedResult.SetEvaluatedProperty(prop)
		}
	}

	for i, subValidator := range mv.validators {
		var result Result
		var err error

		// Each cousin validator should be validated independently
		// without seeing evaluated properties from other cousins
		// Only pass the original previousResult context, not accumulated cousin results
		if objValidator, ok := subValidator.(*objectValidator); ok {
			var previouslyEvaluated map[string]bool
			if previousResult != nil {
				previouslyEvaluated = previousResult.EvaluatedProperties()
			}
			if len(previouslyEvaluated) > 0 {
				// Get existing evaluation context or create a new one
				var ec *schemactx.EvaluationContext
				_ = schemactx.EvaluationContextFromContext(ctx, &ec)
				if ec == nil {
					ec = &schemactx.EvaluationContext{}
				}

				// Mark properties as evaluated
				for prop := range previouslyEvaluated {
					if previouslyEvaluated[prop] {
						ec.Properties.MarkEvaluated(prop)
					}
				}

				ctx = schemactx.WithEvaluationContext(ctx, ec)
			}
			result, err = objValidator.Validate(ctx, in)
		} else {
			result, err = subValidator.Validate(ctx, in)
		}

		if err != nil {
			return nil, fmt.Errorf(`allOf validation failed: validator #%d failed: %w`, i, err)
		}

		// Merge object results
		if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
			if mergedResult == nil {
				mergedResult = NewObjectResult()
			}
			for prop := range objResult.EvaluatedProperties() {
				mergedResult.SetEvaluatedProperty(prop)
			}
		}
	}
	return mergedResult, nil
}
