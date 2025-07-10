package validator_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/assert"
)

func TestIntegerConstrainctSanity(t *testing.T) {
	testcases := makeSanityTestCases()
	for _, tc := range testcases {
		switch tc.Name {
		case "Integer":
		default:
			tc.Error = true
		}
	}

	c := validator.Integer().MustBuild()
	for _, tc := range testcases {
		t.Run(tc.Name, makeSanityTestFunc(tc, c))
	}
}

func TestIntegerValidator(t *testing.T) {
	testcases := []struct {
		Name      string
		Object    interface{}
		Validator func() (validator.Interface, error)
		Error     bool
	}{
		{
			Name:   "multipleOf set, valid value",
			Object: 36,
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Type(schema.IntegerType).
					MultipleOf(6).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
		{
			Name:   "multipleOf set, invalid value",
			Object: 36,
			Error:  true,
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Type(schema.IntegerType).
					MultipleOf(5).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
		{
			Name:   "maximum set, no minimum set, valid value",
			Object: 36,
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Type(schema.IntegerType).
					Maximum(40).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
		{
			Name:   "maximum set, no minimum set, invalid value",
			Object: 36,
			Error:  true,
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Type(schema.IntegerType).
					Maximum(30).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
		{
			Name:   "no maximum set, minimum set, valid value",
			Object: 36,
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Type(schema.IntegerType).
					Minimum(30).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
		{
			Name:   "no maximum set, minimum set, invalid value",
			Object: 36,
			Error:  true,
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Type(schema.IntegerType).
					Minimum(40).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
		{
			Name:   "maximum set, minimum set, valid value",
			Object: 36,
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Type(schema.IntegerType).
					Minimum(30).
					Maximum(40).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(s)
			},
		},
		{
			Name:   "maximum set, minimum set, invalid value",
			Object: 36,
			Error:  true,
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Type(schema.IntegerType).
					Minimum(39).
					Maximum(40).
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
