package validator_test

import (
	"testing"

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

	var c validator.StringConstraint
	for _, tc := range testcases {
		t.Run(tc.Name, makeSanityTestFunc(tc, &c))
	}
}

func TestStringConstraint(t *testing.T) {
	testcases := []struct {
		Name       string
		Object     interface{}
		Constraint func() (*validator.StringConstraint, error)
		Error      bool
	}{
		{
			Name:   "no maxLength set, no minLength set, value within range",
			Object: "Hello, World!",
			Constraint: func() (*validator.StringConstraint, error) {
				return validator.String().Build()
			},
		},
		{
			Name:   "maxLength set, no minLength set, value within range",
			Object: "Hello, World!",
			Constraint: func() (*validator.StringConstraint, error) {
				return validator.String().MaxLength(20).Build()
			},
		},
		{
			Name:   "maxLength set, minLength set, value within range",
			Object: "Hello, World!",
			Constraint: func() (*validator.StringConstraint, error) {
				return validator.String().MinLength(1).MaxLength(20).Build()
			},
		},
		{
			Name:   "maxLength set, minLength set, value within range",
			Object: "Hello, World!",
			Constraint: func() (*validator.StringConstraint, error) {
				return validator.String().MaxLength(20).MinLength(5).Build()
			},
		},
		{
			Name:   "maxLength set, no minLength set, value outside range (l > max)",
			Object: "Hello, World!",
			Error:  true,
			Constraint: func() (*validator.StringConstraint, error) {
				return validator.String().MaxLength(5).Build()
			},
		},
		{
			Name:   "maxLength set, minLength set, value outside range (l > max)",
			Object: "Hello, World!",
			Error:  true,
			Constraint: func() (*validator.StringConstraint, error) {
				return validator.String().MinLength(1).MaxLength(5).Build()
			},
		},
		{
			Name:   "maxLength set, minLength set, value outside range (l < min)",
			Object: "Hello, World!",
			Error:  true,
			Constraint: func() (*validator.StringConstraint, error) {
				return validator.String().MinLength(14).MaxLength(20).Build()
			},
		},
		{
			Name:   "pattern set, valid value",
			Object: "Hello, World!",
			Constraint: func() (*validator.StringConstraint, error) {
				return validator.String().Pattern(`^Hello, .+$`).Build()
			},
		},
		{
			Name:   "pattern set, invalid value",
			Object: "Hello, World!",
			Error:  true,
			Constraint: func() (*validator.StringConstraint, error) {
				return validator.String().Pattern(`^Night, .+$`).Build()
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			c, err := tc.Constraint()
			if !assert.NoError(t, err, `tc.Constraint() should succeed`) {
				return
			}
			err = c.Check(tc.Object)

			if tc.Error {
				if !assert.Error(t, err, `c.Check should fail`) {
					return
				}
			} else {
				if !assert.NoError(t, err, `c.Check should succeed`) {
					return
				}
			}
		})
	}
}
