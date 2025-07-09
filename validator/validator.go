//go:generate ./gen.sh

package validator

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

func Compile(s *schema.Schema) (Validator, error) {
	types := s.Types()
	var validatorsByType []Validator
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
		}
	}

	if len(validatorsByType) == 0 {
		// Empty schema - allows anything
		return &EmptyValidator{}, nil
	}

	if len(validatorsByType) == 1 {
		return validatorsByType[0], nil
	}

	mv := NewMultiValidator(OrMode)
	for _, v := range validatorsByType {
		mv.Append(v)
	}

	return mv, nil
}

type Validator interface {
	Validate(interface{}) error
}

type EmptyValidator struct{}

func (e *EmptyValidator) Validate(v interface{}) error {
	// Empty schema allows anything
	return nil
}

type MultiValidator struct {
	and        bool
	validators []Validator
}

type MultiValidatorMode int

const (
	OrMode MultiValidatorMode = iota
	AndMode
	InvalidMode
)

func NewMultiValidator(mode MultiValidatorMode) *MultiValidator {
	mv := &MultiValidator{}
	if mode == AndMode {
		mv.and = true
	}
	return mv
}

func (v *MultiValidator) Append(in Validator) *MultiValidator {
	v.validators = append(v.validators, in)
	return v
}

func (v *MultiValidator) Validate(in interface{}) error {
	if v.and {
		for i, subv := range v.validators {
			if err := subv.Validate(in); err != nil {
				return fmt.Errorf(`validator #%d failed: %w`, i, err)
			}
		}
		return nil
	}

	for _, subv := range v.validators {
		if err := subv.Validate(in); err == nil {
			return nil
		}
	}
	return fmt.Errorf(`none of the validators passed`)
}
