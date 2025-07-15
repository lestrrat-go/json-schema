package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/assert"
)

func makeSanityTestFunc(tc *sanityTestCase, c validator.Interface) func(*testing.T) {
	return func(t *testing.T) {
		if tc.Error {
			_, err := c.Validate(context.Background(), tc.Object)
			if !assert.Error(t, err, `c.check should fail`) {
				return
			}
		} else {
			_, err := c.Validate(context.Background(), tc.Object)
			if !assert.NoError(t, err, `c.Validate should succeed`) {
				return
			}
		}
	}
}

// Some default set of objects used for sanity checking
type sanityTestCase struct {
	Object interface{}
	Name   string
	Error  bool
}

func makeSanityTestCases() []*sanityTestCase {
	return []*sanityTestCase{
		{
			Name:   "Empty Map",
			Object: make(map[string]interface{}),
		},
		{
			Name:   "Empty Object",
			Object: struct{}{},
		},
		{
			Name:   "Integer",
			Object: 1,
		},
	}
}

func TestValidator(t *testing.T) {
	s, err := schema.NewBuilder().
		Types(schema.ObjectType).
		Build()
	if !assert.NoError(t, err, `schema.NewBuilder should succeed`) {
		return
	}
	_ = s
	/*
		v, err := validator.Compile(context.Background(), s)
		if !assert.NoError(t, err, `validator.Build should succeed`) {
			return
		}
		_ = v
	*/
}
