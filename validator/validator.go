//go:generate ./gen.sh

package validator

import (
	"fmt"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
)

// Interface is the interface that all validators must implement.
type Interface interface {
	Validate(any) error
}

type Builder interface {
	Build() (Interface, error)
	MustBuild() Interface
}

func Compile(s *schema.Schema) (Interface, error) {
	var allValidators []Interface

	// Handle schema composition first
	if s.HasAllOf() {
		allOfValidators := make([]Interface, 0, len(s.AllOf()))
		for _, subSchema := range s.AllOf() {
			v, err := Compile(subSchema)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile allOf validator: %w`, err)
			}
			allOfValidators = append(allOfValidators, v)
		}
		allOfValidator := NewMultiValidator(AndMode)
		for _, v := range allOfValidators {
			allOfValidator.Append(v)
		}
		allValidators = append(allValidators, allOfValidator)
	}

	if s.HasAnyOf() {
		anyOfValidators := make([]Interface, 0, len(s.AnyOf()))
		for _, subSchema := range s.AnyOf() {
			v, err := Compile(subSchema)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile anyOf validator: %w`, err)
			}
			anyOfValidators = append(anyOfValidators, v)
		}
		anyOfValidator := NewMultiValidator(OrMode)
		for _, v := range anyOfValidators {
			anyOfValidator.Append(v)
		}
		allValidators = append(allValidators, anyOfValidator)
	}

	if s.HasOneOf() {
		oneOfValidators := make([]Interface, 0, len(s.OneOf()))
		for _, subSchema := range s.OneOf() {
			v, err := Compile(subSchema)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile oneOf validator: %w`, err)
			}
			oneOfValidators = append(oneOfValidators, v)
		}
		oneOfValidator := NewMultiValidator(OneOfMode)
		for _, v := range oneOfValidators {
			oneOfValidator.Append(v)
		}
		allValidators = append(allValidators, oneOfValidator)
	}

	if s.HasNot() {
		notValidator, err := Compile(s.Not())
		if err != nil {
			return nil, fmt.Errorf(`failed to compile not validator: %w`, err)
		}
		allValidators = append(allValidators, &NotValidator{validator: notValidator})
	}

	// Handle type-specific validators
	types := s.Types()
	var validatorsByType []Interface

	// If no types are specified but type-specific constraints are present,
	// infer the type from the constraints
	if len(types) == 0 {
		if s.HasMinLength() || s.HasMaxLength() || s.HasPattern() {
			types = append(types, schema.StringType)
		}
		if s.HasMinimum() || s.HasMaximum() || s.HasExclusiveMinimum() || s.HasExclusiveMaximum() || s.HasMultipleOf() {
			types = append(types, schema.NumberType)
		}
		if s.HasMinItems() || s.HasMaxItems() || s.HasUniqueItems() || s.HasItems() || s.HasContains() {
			types = append(types, schema.ArrayType)
		}
		if s.HasMinProperties() || s.HasMaxProperties() || s.HasRequired() || s.HasProperties() || s.HasPatternProperties() || s.HasAdditionalProperties() {
			types = append(types, schema.ObjectType)
		}
	}

	// Handle general enum/const validation when no specific type is set
	if len(types) == 0 && (s.HasEnum() || s.HasConst()) {
		validator, err := compileGeneralValidator(s)
		if err != nil {
			return nil, fmt.Errorf(`failed to compile general validator: %w`, err)
		}
		allValidators = append(allValidators, validator)
	}

	for _, typ := range types {
		// This is a placeholder code. In reality we need to
		// OR all types
		switch typ {
		case schema.StringType:
			v, err := compileStringValidator(s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile string validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.IntegerType:
			v, err := compileIntegerValidator(s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile integer validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.NumberType:
			v, err := compileNumberValidator(s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile number validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.BooleanType:
			v, err := compileBooleanValidator(s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile boolean validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.ArrayType:
			v, err := compileArrayValidator(s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile array validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.ObjectType:
			v, err := compileObjectValidator(s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile object validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.NullType:
			v, err := compileNullValidator(s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile null validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		}
	}

	// Combine type validators if multiple types
	if len(validatorsByType) > 1 {
		typeValidator := NewMultiValidator(OrMode)
		for _, v := range validatorsByType {
			typeValidator.Append(v)
		}
		allValidators = append(allValidators, typeValidator)
	} else if len(validatorsByType) == 1 {
		allValidators = append(allValidators, validatorsByType[0])
	}

	// Return the appropriate validator
	if len(allValidators) == 0 {
		// Empty schema - allows anything
		return &EmptyValidator{}, nil
	}

	if len(allValidators) == 1 {
		return allValidators[0], nil
	}

	// Multiple validators - combine with AND
	mv := NewMultiValidator(AndMode)
	for _, v := range allValidators {
		mv.Append(v)
	}

	return mv, nil
}

type EmptyValidator struct{}

func (e *EmptyValidator) Validate(v any) error {
	// Empty schema allows anything
	return nil
}

type NotValidator struct {
	validator Interface
}

func (n *NotValidator) Validate(v any) error {
	err := n.validator.Validate(v)
	if err == nil {
		return fmt.Errorf(`not validation failed: value should not validate against the schema`)
	}
	return nil
}

type NullValidator struct{}

func (n *NullValidator) Validate(v any) error {
	if v == nil {
		return nil
	}
	return fmt.Errorf(`invalid value passed to NullValidator: expected null, got %T`, v)
}

func compileNullValidator(s *schema.Schema) (Interface, error) {
	return &NullValidator{}, nil
}

// GeneralValidator handles enum and const validation for schemas without specific types
type GeneralValidator struct {
	enum     []any
	const_   any
	hasConst bool
}

func compileGeneralValidator(s *schema.Schema) (Interface, error) {
	v := &GeneralValidator{}

	if s.HasEnum() {
		v.enum = s.Enum()
	}

	if s.HasConst() {
		v.const_ = s.Const()
		v.hasConst = true
	}

	return v, nil
}

func (g *GeneralValidator) Validate(value any) error {
	// Check const first
	if g.hasConst {
		if !reflect.DeepEqual(value, g.const_) {
			return fmt.Errorf(`invalid value: must equal const value %v, got %v`, g.const_, value)
		}
		return nil
	}

	// Check enum
	if g.enum != nil {
		for _, enumVal := range g.enum {
			if reflect.DeepEqual(value, enumVal) {
				return nil
			}
		}
		return fmt.Errorf(`invalid value: %v not found in enum %v`, value, g.enum)
	}

	return nil
}

type MultiValidator struct {
	and        bool
	oneOf      bool
	validators []Interface
}

type MultiValidatorMode int

const (
	OrMode MultiValidatorMode = iota
	AndMode
	OneOfMode
	InvalidMode
)

func NewMultiValidator(mode MultiValidatorMode) *MultiValidator {
	mv := &MultiValidator{}
	if mode == AndMode {
		mv.and = true
	} else if mode == OneOfMode {
		mv.and = false
		mv.oneOf = true
	}
	return mv
}

func (v *MultiValidator) Append(in Interface) *MultiValidator {
	v.validators = append(v.validators, in)
	return v
}

func (v *MultiValidator) Validate(in any) error {
	if v.and {
		for i, subv := range v.validators {
			if err := subv.Validate(in); err != nil {
				return fmt.Errorf(`allOf validation failed: validator #%d failed: %w`, i, err)
			}
		}
		return nil
	}

	if v.oneOf {
		passedCount := 0
		for _, subv := range v.validators {
			if err := subv.Validate(in); err == nil {
				passedCount++
			}
		}
		if passedCount == 0 {
			return fmt.Errorf(`oneOf validation failed: none of the validators passed`)
		}
		if passedCount > 1 {
			return fmt.Errorf(`oneOf validation failed: more than one validator passed (%d), expected exactly one`, passedCount)
		}
		return nil
	}

	// This is for anyOf (OrMode)
	for _, subv := range v.validators {
		if err := subv.Validate(in); err == nil {
			return nil
		}
	}
	return fmt.Errorf(`anyOf validation failed: none of the validators passed`)
}
