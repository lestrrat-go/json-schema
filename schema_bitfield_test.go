package schema

import (
	"encoding/json"
	"testing"
)

func TestBitFieldFunctionality(t *testing.T) {
	t.Parallel()
	t.Run("Has method returns correct bit flags", func(t *testing.T) {
		t.Parallel()
		// Test with an empty schema
		schema := New()
		if schema.Has(AnchorField) {
			t.Error("Expected empty schema to have no fields set")
		}

		// Test with a builder
		builder := NewBuilder()
		schema = builder.
			Anchor("test-anchor").
			Maximum(100.0).
			MinLength(5).
			MustBuild()

		// Test individual field checks still work
		if !schema.HasAnchor() {
			t.Error("Expected HasAnchor() to return true")
		}
		if !schema.HasMaximum() {
			t.Error("Expected HasMaximum() to return true")
		}
		if !schema.HasMinLength() {
			t.Error("Expected HasMinLength() to return true")
		}
		if schema.HasMinimum() {
			t.Error("Expected HasMinimum() to return false")
		}

		// Test the new Has method with multiple fields
		if !schema.Has(AnchorField | MaximumField | MinLengthField) {
			t.Error("Expected Has() to return true for all set fields")
		}

		// Test with a missing field
		if schema.Has(AnchorField | MaximumField | MinimumField) {
			t.Error("Expected Has() to return false when one field is missing")
		}
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
		if schema.Has(requiredFields) {
			t.Log("All three fields are set correctly using Has() method")
		} else {
			t.Error("Expected all three fields to be set")
		}

		// Test that combination with an unset field returns false
		requiredFieldsWithMissing := AnchorField | PropertiesField | MinimumField
		if schema.Has(requiredFieldsWithMissing) {
			t.Error("Expected combination with unset field to fail")
		} else {
			t.Log("Correctly detected unset field using Has() method")
		}
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
		if err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		// Check that the appropriate bit fields are set
		if !schema.HasAnchor() {
			t.Error("Expected HasAnchor() to return true after JSON unmarshal")
		}
		if !schema.HasTypes() {
			t.Error("Expected HasTypes() to return true after JSON unmarshal")
		}
		if !schema.HasMinLength() {
			t.Error("Expected HasMinLength() to return true after JSON unmarshal")
		}
		if !schema.HasProperties() {
			t.Error("Expected HasProperties() to return true after JSON unmarshal")
		}

		// Check that unset fields return false
		if schema.HasMaxLength() {
			t.Error("Expected HasMaxLength() to return false")
		}

		// Verify the bit field contains the expected flags using Has method
		expectedFields := AnchorField | TypesField | MinLengthField | PropertiesField
		if !schema.Has(expectedFields) {
			t.Error("Expected all bit fields to be set")
		}
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
		if !cloned.Has(expectedFields) || !original.Has(expectedFields) {
			t.Error("Expected both original and cloned schema to have all expected fields set")
		}

		// Test specific fields
		if !cloned.HasAnchor() || cloned.Anchor() != "original" {
			t.Error("Expected cloned schema to have anchor 'original'")
		}
		if !cloned.HasMaximum() || cloned.Maximum() != 50.0 {
			t.Error("Expected cloned schema to have maximum 50.0")
		}
		if !cloned.HasRequired() {
			t.Error("Expected cloned schema to have required fields")
		}
	})

	t.Run("Reset methods clear bit fields", func(t *testing.T) {
		t.Parallel()
		schema := NewBuilder().
			Anchor("test").
			Maximum(100.0).
			ResetAnchor().
			MustBuild()

		// Should have Maximum but not Anchor
		if schema.HasAnchor() {
			t.Error("Expected HasAnchor() to return false after ResetAnchor()")
		}
		if !schema.HasMaximum() {
			t.Error("Expected HasMaximum() to return true")
		}

		// Bit field should only have MaximumField set
		if !schema.Has(MaximumField) || schema.Has(AnchorField) {
			t.Error("Expected only MaximumField to be set after ResetAnchor()")
		}
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
	if schema.Has(requiredFlags) {
		t.Log("All three fields are populated - efficient bit field check passed")
	} else {
		t.Error("Expected all three fields to be populated")
	}

	// Test that missing field makes the check fail
	schema2 := NewBuilder().
		Anchor("test").
		Property("field1", New()).
		// Note: no AllOf
		MustBuild()

	if schema2.Has(requiredFlags) {
		t.Error("Expected check to fail when AllOf is not set")
	} else {
		t.Log("Bit field check correctly failed when field is missing")
	}
}
