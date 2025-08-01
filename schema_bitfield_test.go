package schema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBitFieldFunctionality(t *testing.T) {
	t.Parallel()
	t.Run("Has method returns correct bit flags", func(t *testing.T) {
		t.Parallel()
		// Test with an empty schema
		schema := New()
		require.False(t, schema.Has(AnchorField), "Expected empty schema to have no fields set")

		// Test with a builder
		builder := NewBuilder()
		schema = builder.
			Anchor("test-anchor").
			Maximum(100.0).
			MinLength(5).
			MustBuild()

		// Test individual field checks still work
		require.True(t, schema.Has(AnchorField), "Expected HasAnchor() to return true")
		require.True(t, schema.Has(MaximumField), "Expected HasMaximum() to return true")
		require.True(t, schema.Has(MinLengthField), "Expected HasMinLength() to return true")
		require.False(t, schema.Has(MinimumField), "Expected HasMinimum() to return false")

		// Test the new Has method with multiple fields
		require.True(t, schema.Has(AnchorField|MaximumField|MinLengthField), "Expected Has() to return true for all set fields")

		// Test with a missing field
		require.False(t, schema.Has(AnchorField|MaximumField|MinimumField), "Expected Has() to return false when one field is missing")
	})

	t.Run("Bit field operations work correctly", func(t *testing.T) {
		t.Parallel()
		schema := NewBuilder().
			Anchor("test").
			Property("foo", New()).
			AllOf(BoolSchema(true)).
			MustBuild()

		// Test the new Has() method for combined bit field operations
		requiredFields := AnchorField | PropertiesField | AllOfField
		require.True(t, schema.Has(requiredFields), "Expected all three fields to be set")

		// Test that combination with an unset field returns false
		requiredFieldsWithMissing := AnchorField | PropertiesField | MinimumField
		require.False(t, schema.Has(requiredFieldsWithMissing), "Expected combination with unset field to fail")
	})

	t.Run("JSON unmarshaling sets bit fields", func(t *testing.T) {
		t.Parallel()
		jsonData := `{
			"$anchor": "test-anchor",
			"type": "string",
			"minLength": 10,
			"properties": {
				"name": {"type": "string"}
			}
		}`

		var schema Schema
		err := json.Unmarshal([]byte(jsonData), &schema)
		require.NoError(t, err, "Failed to unmarshal JSON")

		// Check that the appropriate bit fields are set
		require.True(t, schema.Has(AnchorField), "Expected HasAnchor() to return true after JSON unmarshal")
		require.True(t, len(schema.Types()) > 0, "Expected HasTypes() to return true after JSON unmarshal")
		require.True(t, schema.Has(MinLengthField), "Expected HasMinLength() to return true after JSON unmarshal")
		require.True(t, schema.Has(PropertiesField), "Expected HasProperties() to return true after JSON unmarshal")

		// Check that unset fields return false
		require.False(t, schema.Has(MaxLengthField), "Expected HasMaxLength() to return false")

		// Verify the bit field contains the expected flags using Has method
		expectedFields := AnchorField | TypesField | MinLengthField | PropertiesField
		require.True(t, schema.Has(expectedFields), "Expected all bit fields to be set")
	})

	t.Run("Clone preserves bit fields", func(t *testing.T) {
		t.Parallel()
		original := NewBuilder().
			Anchor("original").
			Maximum(50.0).
			Required("field1", "field2").
			MustBuild()

		cloned := NewBuilder().Clone(original).MustBuild()

		// Check that the cloned schema has the same populated fields
		expectedFields := AnchorField | MaximumField | RequiredField
		require.True(t, cloned.Has(expectedFields) && original.Has(expectedFields), "Expected both original and cloned schema to have all expected fields set")

		// Test specific fields
		require.True(t, cloned.Has(AnchorField) && cloned.Anchor() == "original", "Expected cloned schema to have anchor 'original'")
		require.True(t, cloned.Has(MaximumField) && cloned.Maximum() == 50.0, "Expected cloned schema to have maximum 50.0")
		require.True(t, cloned.Has(RequiredField), "Expected cloned schema to have required fields")
	})

	t.Run("Reset methods clear bit fields", func(t *testing.T) {
		t.Parallel()
		schema := NewBuilder().
			Anchor("test").
			Maximum(100.0).
			ResetAnchor().
			MustBuild()

		// Should have Maximum but not Anchor
		require.False(t, schema.Has(AnchorField), "Expected HasAnchor() to return false after ResetAnchor()")
		require.True(t, schema.Has(MaximumField), "Expected HasMaximum() to return true")

		// Bit field should only have MaximumField set
		require.True(t, schema.Has(MaximumField) && !schema.Has(AnchorField), "Expected only MaximumField to be set after ResetAnchor()")
	})
}

func TestBitFieldEfficiency(t *testing.T) {
	t.Parallel()
	// This demonstrates the efficiency improvement mentioned in the spec
	schema := NewBuilder().
		Anchor("test").
		Property("field1", New()).
		AllOf(BoolSchema(true)).
		MustBuild()

	// The old way (this is what we're replacing):
	// if schema.HasAnchor() && schema.HasProperties() && schema.HasAllOf() { ... }

	// The new efficient way using bit fields with Has() method:
	requiredFlags := AnchorField | PropertiesField | AllOfField
	require.True(t, schema.Has(requiredFlags), "Expected all three fields to be populated")

	// Test that missing field makes the check fail
	schema2 := NewBuilder().
		Anchor("test").
		Property("field1", New()).
		// Note: no AllOf
		MustBuild()

	require.False(t, schema2.Has(requiredFlags), "Expected check to fail when AllOf is not set")
}
