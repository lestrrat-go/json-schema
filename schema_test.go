package schema_test

import (
	"encoding/json"
	"fmt"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/assert"
)

func ExampleSchemaBuilder() {
	s, err := schema.NewBuilder().
		ID(`https://example.com/polygon`).
		Type(schema.ObjectType).
		Property("validProp", schema.New()).
		//		AdditionalProperties(true).
		Build()
	if err != nil {
		fmt.Println(err)
	}
	_ = s
	fmt.Println(s.ID())
	buf, err := json.Marshal(s)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("%s\n", buf)
	// OUTPUT:
}

func TestPrimitiveType(t *testing.T) {
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
		tc := tc
		t.Run(tc.Input, func(t *testing.T) {
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
