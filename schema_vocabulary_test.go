package schema_test

import (
	"encoding/json"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

// TestJSONSchemaVocabulary tests vocabulary support as per JSON Schema 2020-12
func TestJSONSchemaVocabulary(t *testing.T) {
	t.Run("Core Vocabulary Keywords", func(t *testing.T) {
		// Test that core vocabulary keywords are supported
		s, err := schema.NewBuilder().
			ID("https://example.com/test").
			Schema(schema.Version).
			Reference("#/definitions/test").
			Anchor("test-anchor").
			DynamicReference("#dynamic").
			Comment("Test comment").
			Build()
		require.NoError(t, err)

		// Verify all core vocabulary keywords are accessible
		require.Equal(t, "https://example.com/test", s.ID())
		require.Equal(t, schema.Version, s.Schema())
		require.Equal(t, "#/definitions/test", s.Reference())
		require.Equal(t, "test-anchor", s.Anchor())
		require.Equal(t, "#dynamic", s.DynamicReference())
		require.Equal(t, "Test comment", s.Comment())
	})

	t.Run("Applicator Vocabulary Keywords", func(t *testing.T) {
		// Test applicator vocabulary keywords
		itemSchema, err := schema.NewBuilder().Type(schema.StringType).Build()
		require.NoError(t, err)

		propSchema, err := schema.NewBuilder().Type(schema.NumberType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			AllOf(itemSchema).
			AnyOf(itemSchema).
			OneOf(itemSchema).
			Not(itemSchema).
			Items(itemSchema).
			Property("test", propSchema).
			PatternProperty("^test", propSchema).
			AdditionalProperties(propSchema).
			Contains(itemSchema).
			Build()
		require.NoError(t, err)

		// Verify applicator keywords work
		require.Len(t, s.AllOf(), 1)
		require.Len(t, s.AnyOf(), 1)
		require.Len(t, s.OneOf(), 1)
		require.NotNil(t, s.Not())
		require.NotNil(t, s.Items())
		require.NotNil(t, s.Properties()["test"])
		require.NotNil(t, s.PatternProperties()["^test"])
		require.NotNil(t, s.AdditionalProperties())
		require.NotNil(t, s.Contains())
	})

	t.Run("Validation Vocabulary Keywords", func(t *testing.T) {
		// Test validation vocabulary keywords
		s, err := schema.NewBuilder().
			Type(schema.StringType).
			Enum("red", "green", "blue").
			Const("constant").
			MultipleOf(2.5).
			Maximum(100.0).
			ExclusiveMaximum(99.0).
			Minimum(0.0).
			ExclusiveMinimum(1.0).
			MaxLength(50).
			MinLength(5).
			Pattern("^[a-zA-Z]+$").
			MaxItems(10).
			MinItems(1).
			UniqueItems(true).
			MaxContains(5).
			MinContains(1).
			MaxProperties(20).
			MinProperties(2).
			Build()
		require.NoError(t, err)

		// Verify validation keywords
		require.Equal(t, []schema.PrimitiveType{schema.StringType}, s.Types())
		require.Equal(t, []interface{}{"red", "green", "blue"}, s.Enum())
		require.Equal(t, "constant", s.Const())
		require.Equal(t, 2.5, s.MultipleOf())
		require.Equal(t, 100.0, s.Maximum())
		require.Equal(t, 99.0, s.ExclusiveMaximum())
		require.Equal(t, 0.0, s.Minimum())
		require.Equal(t, 1.0, s.ExclusiveMinimum())
		require.Equal(t, 50, s.MaxLength())
		require.Equal(t, 5, s.MinLength())
		require.Equal(t, "^[a-zA-Z]+$", s.Pattern())
		require.Equal(t, uint(10), s.MaxItems())
		require.Equal(t, uint(1), s.MinItems())
		require.True(t, s.UniqueItems())
		require.Equal(t, uint(5), s.MaxContains())
		require.Equal(t, uint(1), s.MinContains())
		require.Equal(t, uint(20), s.MaxProperties())
		require.Equal(t, uint(2), s.MinProperties())
	})

	t.Run("Unevaluated Vocabulary Keywords", func(t *testing.T) {
		// Test unevaluated vocabulary keywords
		propSchema, err := schema.NewBuilder().Type(schema.StringType).Build()
		require.NoError(t, err)

		itemSchema, err := schema.NewBuilder().Type(schema.NumberType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			UnevaluatedItems(itemSchema).
			UnevaluatedProperties(propSchema).
			Build()
		require.NoError(t, err)

		require.NotNil(t, s.UnevaluatedItems())
		require.NotNil(t, s.UnevaluatedProperties())
	})
}

// TestSchemaJSONSerialization tests that schemas can be properly serialized to and from JSON
func TestSchemaJSONSerialization(t *testing.T) {
	t.Run("Simple Schema Serialization", func(t *testing.T) {
		original, err := schema.NewBuilder().
			ID("https://example.com/person").
			Schema(schema.Version).
			Type(schema.ObjectType).
			Property("name", schema.NewBuilder().Type(schema.StringType).MustBuild()).
			Property("age", schema.NewBuilder().Type(schema.IntegerType).Minimum(0).MustBuild()).
			Build()
		require.NoError(t, err)

		// Serialize to JSON
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		// Verify JSON structure by unmarshaling back to Schema
		var s schema.Schema
		require.NoError(t, json.Unmarshal(jsonData, &s), `json.Unmarshal should succeed`)

		// Check core fields are present
		require.Equal(t, "https://example.com/person", s.ID())
		require.Equal(t, schema.Version, s.Schema())
		require.True(t, s.ContainsType(schema.ObjectType))
		require.NotNil(t, s.Properties())
		require.Len(t, s.Required(), 0)

		// Check properties structure
		props := s.Properties()
		require.NotNil(t, props["name"])
		require.NotNil(t, props["age"])
	})

	t.Run("Complex Schema Serialization", func(t *testing.T) {
		stringSchema, err := schema.NewBuilder().
			Type(schema.StringType).
			MinLength(1).
			MaxLength(100).
			Pattern("^[a-zA-Z ]+$").
			Build()
		require.NoError(t, err)

		numberSchema, err := schema.NewBuilder().
			Type(schema.NumberType).
			Minimum(0.0).
			Maximum(1000.0).
			MultipleOf(0.01).
			Build()
		require.NoError(t, err)

		original, err := schema.NewBuilder().
			ID("https://example.com/complex").
			Type(schema.ObjectType).
			AllOf(stringSchema, numberSchema).
			AnyOf(stringSchema).
			OneOf(numberSchema).
			Not(stringSchema).
			Enum("option1", "option2", "option3").
			Const("constant_value").
			Comment("Complex schema for testing").
			Build()
		require.NoError(t, err)

		// Serialize to JSON
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		// Verify JSON structure by unmarshaling back to Schema
		var s schema.Schema
		require.NoError(t, json.Unmarshal(jsonData, &s), `json.Unmarshal should succeed`)

		// Check complex fields are present
		require.Equal(t, "https://example.com/complex", s.ID())
		require.True(t, s.ContainsType(schema.ObjectType))
		require.True(t, s.HasAllOf())
		require.True(t, s.HasAnyOf())
		require.True(t, s.HasOneOf())
		require.True(t, s.HasNot())
		require.True(t, s.HasEnum())
		require.Equal(t, "constant_value", s.Const())
		require.Equal(t, "Complex schema for testing", s.Comment())
	})

	t.Run("Schema with References", func(t *testing.T) {
		original, err := schema.NewBuilder().
			ID("https://example.com/ref-test").
			Reference("#/definitions/person").
			DynamicReference("#person").
			Definitions("person").
			Anchor("person-anchor").
			Build()
		require.NoError(t, err)

		// Serialize to JSON
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		// Verify JSON structure by unmarshaling back to Schema
		var s schema.Schema
		require.NoError(t, json.Unmarshal(jsonData, &s), `json.Unmarshal should succeed`)

		// Check reference fields
		require.Equal(t, "https://example.com/ref-test", s.ID())
		require.Equal(t, "#/definitions/person", s.Reference())
		require.Equal(t, "#person", s.DynamicReference())
		require.Equal(t, "person", s.Definitions())
		require.Equal(t, "person-anchor", s.Anchor())
	})
}

// TestSchemaBuilderErrorHandling tests error handling in schema builder
func TestSchemaBuilderErrorHandling(t *testing.T) {
	t.Run("Invalid Pattern", func(t *testing.T) {
		// This should not cause an error at build time since pattern validation
		// is done at validation time, not build time
		s, err := schema.NewBuilder().
			Type(schema.StringType).
			Pattern("[invalid").
			Build()
		require.NoError(t, err)
		require.Equal(t, "[invalid", s.Pattern())
	})

	t.Run("Duplicate Properties", func(t *testing.T) {
		propSchema, err := schema.NewBuilder().Type(schema.StringType).Build()
		require.NoError(t, err)

		// Adding duplicate properties should result in error
		_, err = schema.NewBuilder().
			Type(schema.ObjectType).
			Property("name", propSchema).
			Property("name", propSchema).
			Build()
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate key")
	})

	t.Run("Invalid Additional Properties", func(t *testing.T) {
		// Test invalid value for additional properties
		var s schema.Schema
		err := s.Accept("invalid")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid value for additionalProperties")
	})
}

// TestMultipleTypes tests schemas with multiple types
func TestMultipleTypes(t *testing.T) {
	t.Run("Multiple Types", func(t *testing.T) {
		s, err := schema.NewBuilder().
			Type(schema.StringType).
			Type(schema.NumberType).
			Type(schema.BooleanType).
			Build()
		require.NoError(t, err)

		types := s.Types()
		require.Len(t, types, 3)
		require.Contains(t, types, schema.StringType)
		require.Contains(t, types, schema.NumberType)
		require.Contains(t, types, schema.BooleanType)
	})
}
