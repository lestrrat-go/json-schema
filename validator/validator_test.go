package validator_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/assert"
)

func makeSanityTestFunc(tc *sanityTestCase, c validator.Constraint) func(*testing.T) {
	return func(t *testing.T) {
		if tc.Error {
			if !assert.Error(t, c.Check(tc.Object), `c.check should fail`) {
				return
			}
		} else {
			if !assert.NoError(t, c.Check(tc.Object), `c.Check should succeed`) {
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
		Type(schema.ObjectType).
		Build()
	if !assert.NoError(t, err, `schema.NewBuilder should succeed`) {
		return
	}
	v, err := validator.Build(s)
	if !assert.NoError(t, err, `validator.Build should succeed`) {
		return
	}
	_ = v
}
