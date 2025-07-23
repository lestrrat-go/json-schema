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
	require.True(t, original.HasReference())
	require.Equal(t, "#/definitions/MyType", original.Reference())
	require.True(t, original.HasProperties())
	require.True(t, original.HasRequired())
	require.True(t, original.HasMinProperties())
	require.True(t, original.HasMaxProperties())

	// Test Clone method - should copy all fields
	cloned := NewBuilder().Clone(original).MustBuild()

	require.True(t, cloned.HasReference())
	require.Equal(t, "#/definitions/MyType", cloned.Reference())
	require.True(t, cloned.HasProperties())
	require.True(t, cloned.HasRequired())
	require.True(t, cloned.HasMinProperties())
	require.True(t, cloned.HasMaxProperties())

	// Properties should be the same
	require.Equal(t, len(original.Properties()), len(cloned.Properties()))
	require.Equal(t, original.Required(), cloned.Required())
	require.Equal(t, original.MinProperties(), cloned.MinProperties())
	require.Equal(t, original.MaxProperties(), cloned.MaxProperties())

	// Test ResetReference method - should remove only reference
	withoutRef := NewBuilder().Clone(original).ResetReference().MustBuild()

	require.False(t, withoutRef.HasReference())
	require.True(t, withoutRef.HasProperties())
	require.True(t, withoutRef.HasRequired())
	require.True(t, withoutRef.HasMinProperties())
	require.True(t, withoutRef.HasMaxProperties())

	// Other fields should remain unchanged
	require.Equal(t, len(original.Properties()), len(withoutRef.Properties()))
	require.Equal(t, original.Required(), withoutRef.Required())
	require.Equal(t, original.MinProperties(), withoutRef.MinProperties())
	require.Equal(t, original.MaxProperties(), withoutRef.MaxProperties())

	// Test other reset methods
	withoutProps := NewBuilder().Clone(original).ResetProperties().MustBuild()
	require.True(t, withoutProps.HasReference())
	require.False(t, withoutProps.HasProperties())
	require.True(t, withoutProps.HasRequired())

	withoutRequired := NewBuilder().Clone(original).ResetRequired().MustBuild()
	require.True(t, withoutRequired.HasReference())
	require.True(t, withoutRequired.HasProperties())
	require.False(t, withoutRequired.HasRequired())
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
	withoutRef := NewBuilder().Clone(schemaWithRef).ResetReference().MustBuild()

	require.False(t, withoutRef.HasReference())
	require.True(t, withoutRef.HasProperties())
	require.True(t, withoutRef.HasRequired())
	require.Equal(t, schemaWithRef.Properties(), withoutRef.Properties())
	require.Equal(t, schemaWithRef.Required(), withoutRef.Required())
}
