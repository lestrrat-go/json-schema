package schema_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

func TestCloneBuilderPattern(t *testing.T) {
	// Create an original schema with multiple fields
	original := schema.NewBuilder().
		Reference("#/definitions/MyType").
		Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
		Property("age", schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
		Required("name").
		MinProperties(1).
		MaxProperties(10).
		MustBuild()

	// Test that original schema has all the fields
	require.True(t, original.Has(schema.ReferenceField))
	require.Equal(t, "#/definitions/MyType", original.Reference())
	require.True(t, original.Has(schema.PropertiesField))
	require.True(t, original.Has(schema.RequiredField))
	require.True(t, original.Has(schema.MinPropertiesField))
	require.True(t, original.Has(schema.MaxPropertiesField))

	// Test Clone method - should copy all fields
	cloned := schema.NewBuilder().Clone(original).MustBuild()

	require.True(t, cloned.Has(schema.ReferenceField))
	require.Equal(t, "#/definitions/MyType", cloned.Reference())
	require.True(t, cloned.Has(schema.PropertiesField))
	require.True(t, cloned.Has(schema.RequiredField))
	require.True(t, cloned.Has(schema.MinPropertiesField))
	require.True(t, cloned.Has(schema.MaxPropertiesField))

	// Properties should be the same
	require.Equal(t, len(original.Properties()), len(cloned.Properties()))
	require.Equal(t, original.Required(), cloned.Required())
	require.Equal(t, original.MinProperties(), cloned.MinProperties())
	require.Equal(t, original.MaxProperties(), cloned.MaxProperties())

	// Test ResetReference method - should remove only reference
	withoutRef := schema.NewBuilder().Clone(original).Reset(schema.ReferenceField).MustBuild()

	require.False(t, withoutRef.Has(schema.ReferenceField))
	require.True(t, withoutRef.Has(schema.PropertiesField))
	require.True(t, withoutRef.Has(schema.RequiredField))
	require.True(t, withoutRef.Has(schema.MinPropertiesField))
	require.True(t, withoutRef.Has(schema.MaxPropertiesField))

	// Other fields should remain unchanged
	require.Equal(t, len(original.Properties()), len(withoutRef.Properties()))
	require.Equal(t, original.Required(), withoutRef.Required())
	require.Equal(t, original.MinProperties(), withoutRef.MinProperties())
	require.Equal(t, original.MaxProperties(), withoutRef.MaxProperties())

	// Test other reset methods
	withoutProps := schema.NewBuilder().Clone(original).Reset(schema.PropertiesField).MustBuild()
	require.True(t, withoutProps.Has(schema.ReferenceField))
	require.False(t, withoutProps.Has(schema.PropertiesField))
	require.True(t, withoutProps.Has(schema.RequiredField))

	withoutRequired := schema.NewBuilder().Clone(original).Reset(schema.RequiredField).MustBuild()
	require.True(t, withoutRequired.Has(schema.ReferenceField))
	require.True(t, withoutRequired.Has(schema.PropertiesField))
	require.False(t, withoutRequired.Has(schema.RequiredField))
}

func TestCloneBuilderWithCompositeValidator(t *testing.T) {
	// Create a schema with both $ref and other constraints
	schemaWithRef := schema.NewBuilder().
		Reference("#/definitions/BaseType").
		Property("extraField", schema.NewBuilder().Types(schema.StringType).MustBuild()).
		Required("extraField").
		MustBuild()

	// Test that cloning and resetting $ref works correctly
	withoutRef := schema.NewBuilder().Clone(schemaWithRef).Reset(schema.ReferenceField).MustBuild()

	require.False(t, withoutRef.Has(schema.ReferenceField))
	require.True(t, withoutRef.Has(schema.PropertiesField))
	require.True(t, withoutRef.Has(schema.RequiredField))
	require.Equal(t, schemaWithRef.Properties(), withoutRef.Properties())
	require.Equal(t, schemaWithRef.Required(), withoutRef.Required())
}
