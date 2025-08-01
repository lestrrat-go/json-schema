package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCloneBuilderPattern(t *testing.T) {
	// Create an original schema with multiple fields
	original := NewBuilder().
		Reference("#/definitions/MyType").
		Property("name", NewBuilder().Types(StringType).MustBuild()).
		Property("age", NewBuilder().Types(IntegerType).MustBuild()).
		Required("name").
		MinProperties(1).
		MaxProperties(10).
		MustBuild()

	// Test that original schema has all the fields
	require.True(t, original.Has(ReferenceField))
	require.Equal(t, "#/definitions/MyType", original.Reference())
	require.True(t, original.Has(PropertiesField))
	require.True(t, original.Has(RequiredField))
	require.True(t, original.Has(MinPropertiesField))
	require.True(t, original.Has(MaxPropertiesField))

	// Test Clone method - should copy all fields
	cloned := NewBuilder().Clone(original).MustBuild()

	require.True(t, cloned.Has(ReferenceField))
	require.Equal(t, "#/definitions/MyType", cloned.Reference())
	require.True(t, cloned.Has(PropertiesField))
	require.True(t, cloned.Has(RequiredField))
	require.True(t, cloned.Has(MinPropertiesField))
	require.True(t, cloned.Has(MaxPropertiesField))

	// Properties should be the same
	require.Equal(t, len(original.Properties()), len(cloned.Properties()))
	require.Equal(t, original.Required(), cloned.Required())
	require.Equal(t, original.MinProperties(), cloned.MinProperties())
	require.Equal(t, original.MaxProperties(), cloned.MaxProperties())

	// Test ResetReference method - should remove only reference
	withoutRef := NewBuilder().Clone(original).Reset(ReferenceField).MustBuild()

	require.False(t, withoutRef.Has(ReferenceField))
	require.True(t, withoutRef.Has(PropertiesField))
	require.True(t, withoutRef.Has(RequiredField))
	require.True(t, withoutRef.Has(MinPropertiesField))
	require.True(t, withoutRef.Has(MaxPropertiesField))

	// Other fields should remain unchanged
	require.Equal(t, len(original.Properties()), len(withoutRef.Properties()))
	require.Equal(t, original.Required(), withoutRef.Required())
	require.Equal(t, original.MinProperties(), withoutRef.MinProperties())
	require.Equal(t, original.MaxProperties(), withoutRef.MaxProperties())

	// Test other reset methods
	withoutProps := NewBuilder().Clone(original).Reset(PropertiesField).MustBuild()
	require.True(t, withoutProps.Has(ReferenceField))
	require.False(t, withoutProps.Has(PropertiesField))
	require.True(t, withoutProps.Has(RequiredField))

	withoutRequired := NewBuilder().Clone(original).Reset(RequiredField).MustBuild()
	require.True(t, withoutRequired.Has(ReferenceField))
	require.True(t, withoutRequired.Has(PropertiesField))
	require.False(t, withoutRequired.Has(RequiredField))
}

func TestCloneBuilderWithCompositeValidator(t *testing.T) {
	// Create a schema with both $ref and other constraints
	schemaWithRef := NewBuilder().
		Reference("#/definitions/BaseType").
		Property("extraField", NewBuilder().Types(StringType).MustBuild()).
		Required("extraField").
		MustBuild()

	// Test createSchemaWithoutRef function from validator package
	// This should create a schema with all constraints except $ref
	withoutRef := NewBuilder().Clone(schemaWithRef).Reset(ReferenceField).MustBuild()

	require.False(t, withoutRef.Has(ReferenceField))
	require.True(t, withoutRef.Has(PropertiesField))
	require.True(t, withoutRef.Has(RequiredField))
	require.Equal(t, schemaWithRef.Properties(), withoutRef.Properties())
	require.Equal(t, schemaWithRef.Required(), withoutRef.Required())
}
