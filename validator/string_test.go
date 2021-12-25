package validator_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/assert"
)

func TestStringConstrainctSanity(t *testing.T) {
	testcases := makeSanityTestCases()
	for _, tc := range testcases {
		switch tc.Name {
		case "String":
		default:
			tc.Error = true
		}
	}

	var c validator.StringValidator
	for _, tc := range testcases {
		t.Run(tc.Name, makeSanityTestFunc(tc, &c))
	}
}

func TestStringValidator(t *testing.T) {
	testcases := []struct {
		Name      string
		Object    interface{}
		Validator func() (validator.Validator, error)
		Error     bool
	}{
		{
			Name:   "no maxLength set, no minLength set, value within range",
			Object: "Hello, World!",
			Validator: func() (validator.Validator, error) {
				s, err := schema.NewBuilder().
					Type(schema.StringType).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
		{
			Name:   "maxLength set, no minLength set, value within range",
			Object: "Hello, World!",
			Validator: func() (validator.Validator, error) {
				s, err := schema.NewBuilder().
					Type(schema.StringType).
					MaxLength(20).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
		{
			Name:   "maxLength set, minLength set, value within range",
			Object: "Hello, World!",
			Validator: func() (validator.Validator, error) {
				s, err := schema.NewBuilder().
					Type(schema.StringType).
					MinLength(1).
					MaxLength(20).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
		{
			Name:   "maxLength set, minLength set, value within range",
			Object: "Hello, World!",
			Validator: func() (validator.Validator, error) {
				s, err := schema.NewBuilder().
					Type(schema.StringType).
					MinLength(5).
					MaxLength(20).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
		{
			Name:   "maxLength set, no minLength set, value outside range (l > max)",
			Object: "Hello, World!",
			Error:  true,
			Validator: func() (validator.Validator, error) {
				s, err := schema.NewBuilder().
					Type(schema.StringType).
					MaxLength(5).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
		{
			Name:   "maxLength set, minLength set, value outside range (l > max)",
			Object: "Hello, World!",
			Error:  true,
			Validator: func() (validator.Validator, error) {
				s, err := schema.NewBuilder().
					Type(schema.StringType).
					MinLength(1).
					MaxLength(5).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
		{
			Name:   "maxLength set, minLength set, value outside range (l < min)",
			Object: "Hello, World!",
			Error:  true,
			Validator: func() (validator.Validator, error) {
				s, err := schema.NewBuilder().
					Type(schema.StringType).
					MinLength(14).
					MaxLength(20).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
		{
			Name:   "pattern set, valid value",
			Object: "Hello, World!",
			Validator: func() (validator.Validator, error) {
				s, err := schema.NewBuilder().
					Type(schema.StringType).
					Pattern(`^Hello, .+$`).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
		{
			Name:   "pattern set, invalid value",
			Object: "Hello, World!",
			Error:  true,
			Validator: func() (validator.Validator, error) {
				s, err := schema.NewBuilder().
					Type(schema.StringType).
					Pattern(`^Night, .+$`).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			c, err := tc.Validator()
			if !assert.NoError(t, err, `tc.Validator() should succeed`) {
				return
			}
			err = c.Validate(tc.Object)

			if tc.Error {
				if !assert.Error(t, err, `c.Validate should fail`) {
					return
				}
			} else {
				if !assert.NoError(t, err, `c.Validate should succeed`) {
					return
				}
			}
		})
	}
}
