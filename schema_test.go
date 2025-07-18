package schema_test

import (
	"encoding/json"
	"fmt"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/assert"
)

func Example_schema_builder() {
	s, err := schema.NewBuilder().
		ID(`https://example.com/polygon`).
		Types(schema.ObjectType).
		Property("validProp", schema.New()).
		AdditionalProperties(schema.SchemaTrue()).
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
	// https://example.com/polygon
	// {"$id":"https://example.com/polygon","$schema":"https://json-schema.org/draft/2020-12/schema","additionalProperties":true,"properties":{"validProp":{"$schema":"https://json-schema.org/draft/2020-12/schema"}},"type":"object"}
}

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
