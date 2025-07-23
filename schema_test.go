package schema_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/assert"
)


func TestPrimitiveType(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		Input    string
		Expected schema.PrimitiveType
		Error    bool
	}{
		{
			Input:    "null",
			Expected: schema.NullType,
		},
		{
			Input:    "integer",
			Expected: schema.IntegerType,
		},
		{
			Input:    "string",
			Expected: schema.StringType,
		},
		{
			Input:    "object",
			Expected: schema.ObjectType,
		},
		{
			Input:    "array",
			Expected: schema.ArrayType,
		},
		{
			Input:    "boolean",
			Expected: schema.BooleanType,
		},
		{
			Input:    "number",
			Expected: schema.NumberType,
		},
		{
			Input: "foo",
			Error: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Input, func(t *testing.T) {
			t.Parallel()
			pt, err := schema.NewPrimitiveType(tc.Input)
			if tc.Error {
				if !assert.Error(t, err, `schema.NewPrimitiveType should fail`) {
					return
				}
			} else {
				if !assert.NoError(t, err, `schema.NewPrimitiveType should succeed`) {
					return
				}

				if !assert.Equal(t, tc.Expected, pt, `values should match`) {
					return
				}
			}
		})
	}
}
