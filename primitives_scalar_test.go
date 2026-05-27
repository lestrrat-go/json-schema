package schema_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

// IsScalarPrimitiveType reports whether a primitive type is a scalar (leaf)
// type as opposed to a container type (object/array). The JSON Schema spec
// distinguishes "container instances" (arrays and objects) from the rest; this
// helper is the inverse of that classification.
func TestIsScalarPrimitiveType(t *testing.T) {
	scalars := []schema.PrimitiveType{
		schema.StringType,
		schema.IntegerType,
		schema.NumberType,
		schema.BooleanType,
		schema.NullType,
	}
	for _, typ := range scalars {
		require.True(t, schema.IsScalarPrimitiveType(typ), "%s should be scalar", typ)
	}

	containers := []schema.PrimitiveType{
		schema.ObjectType,
		schema.ArrayType,
	}
	for _, typ := range containers {
		require.False(t, schema.IsScalarPrimitiveType(typ), "%s should not be scalar", typ)
	}
}
