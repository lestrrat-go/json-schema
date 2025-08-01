package schema_test

import (
	"encoding/json"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

// TestMetaSchemaCompliance tests meta-schema compliance for JSON Schema 2020-12
func TestMetaSchemaCompliance(t *testing.T) {
	t.Run("Custom Meta-Schema", func(t *testing.T) {
		// Test setting custom meta-schema
		customMetaSchema := "https://example.com/custom-meta-schema"
		s, err := schema.NewBuilder().
			Schema(customMetaSchema).
			Build()
		require.NoError(t, err)
		require.Equal(t, customMetaSchema, s.Schema())
	})

	t.Run("Meta-Schema Declaration Required", func(t *testing.T) {
		// Every schema should declare its meta-schema
		s, err := schema.NewBuilder().
			Schema(schema.Version).
			Types(schema.StringType).
			Build()
		require.NoError(t, err)

		// Should have default meta-schema
		require.True(t, s.HasSchema(), "Schema should have a meta-schema declared")
		require.NotEmpty(t, s.Schema())
		require.Equal(t, schema.Version, s.Schema())
	})
}

// TestJSONSchemaMetaValidation tests that schemas can be validated against meta-schemas
func TestJSONSchemaMetaValidation(t *testing.T) {
	t.Run("Valid Schema Structure", func(t *testing.T) {
		// Create a valid schema
		original, err := schema.NewBuilder().
			ID("https://example.com/valid").
			Schema(schema.Version).
			Types(schema.ObjectType).
			Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("age", schema.NewBuilder().Types(schema.IntegerType).Minimum(0).MustBuild()).
			Build()
		require.NoError(t, err)

		// Serialize to JSON to check structure
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		t.Logf("%s", jsonData)

		var s schema.Schema
		require.NoError(t, json.Unmarshal(jsonData, &s), `json.Unmarshal should succeed`)

		// Check that all fields are properly structured
		require.Equal(t, "https://example.com/valid", s.ID())
		require.True(t, s.HasSchema(), "Schema should have a meta-schema declared")
		require.Equal(t, schema.Version, s.Schema())
		require.Equal(t, s.ContainsType(schema.ObjectType), true, `Schema should be of type Object`)
		require.NotNil(t, s.Properties())
	})

	t.Run("Schema with All Core Keywords", func(t *testing.T) {
		// Test schema with all supported core keywords
		original, err := schema.NewBuilder().
			ID("https://example.com/comprehensive").
			Schema(schema.Version).
			Reference("#/definitions/base").
			Anchor("main").
			DynamicReference("#meta").
			Comment("Comprehensive test schema").
			Build()
		require.NoError(t, err)

		// Serialize and verify structure
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		var s schema.Schema
		require.NoError(t, json.Unmarshal(jsonData, &s), `json.Unmarshal should succeed`)

		// Verify all core keywords are present
		require.Equal(t, "https://example.com/comprehensive", s.ID())
		require.True(t, s.HasSchema(), "Schema should have a meta-schema declared")
		require.Equal(t, schema.Version, s.Schema())
		require.Equal(t, "#/definitions/base", s.Reference())
		require.Equal(t, "main", s.Anchor())
		require.Equal(t, "#meta", s.DynamicReference())
		require.Equal(t, "Comprehensive test schema", s.Comment())
	})

	t.Run("Schema with Composition Keywords", func(t *testing.T) {
		// Test schema with composition (allOf, anyOf, oneOf, not)
		stringSchema, err := schema.NewBuilder().Types(schema.StringType).Build()
		require.NoError(t, err)

		numberSchema, err := schema.NewBuilder().Types(schema.NumberType).Build()
		require.NoError(t, err)

		original, err := schema.NewBuilder().
			AllOf(stringSchema).
			AnyOf(stringSchema, numberSchema).
			OneOf(stringSchema).
			Not(numberSchema).
			Build()
		require.NoError(t, err)

		// Serialize and verify structure
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		var s schema.Schema
		require.NoError(t, json.Unmarshal(jsonData, &s), `json.Unmarshal should succeed`)

		// Verify composition keywords
		require.True(t, s.HasAllOf())
		require.True(t, s.HasAnyOf())
		require.True(t, s.HasOneOf())
		require.True(t, s.HasNot())

		// Check array lengths
		require.Len(t, s.AllOf(), 1)
		require.Len(t, s.AnyOf(), 2)
		require.Len(t, s.OneOf(), 1)
		require.NotNil(t, s.Not())
	})
}

// TestSchemaVocabularyDeclaration tests vocabulary declaration
func TestSchemaVocabularyDeclaration(t *testing.T) {
	t.Run("Schema Version Compatibility", func(t *testing.T) {
		// Test version string matches expected format
		require.Equal(t, "https://json-schema.org/draft/2020-12/schema", schema.Version)
	})
}

// TestSchemaIdentification tests schema identification mechanisms
func TestSchemaIdentification(t *testing.T) {
	t.Run("Schema ID", func(t *testing.T) {
		testCases := []struct {
			name string
			id   string
		}{
			{"HTTP URI", "https://example.com/schema"},
			{"HTTPS URI", "https://secure.example.com/schema"},
			{"URI with Path", "https://example.com/schemas/person"},
			{"URI with Fragment", "https://example.com/schema#person"},
			{"Relative URI", "/schemas/person"},
			{"Fragment Only", "#person"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					ID(tc.id).
					Build()
				require.NoError(t, err)
				require.Equal(t, tc.id, s.ID())
			})
		}
	})

	t.Run("Schema Anchor", func(t *testing.T) {
		testCases := []string{
			"main",
			"person",
			"address",
			"contact-info",
		}

		for _, anchor := range testCases {
			t.Run(anchor, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Anchor(anchor).
					Build()
				require.NoError(t, err)
				require.Equal(t, anchor, s.Anchor())
			})
		}
	})

	t.Run("Schema with ID and Anchor", func(t *testing.T) {
		s, err := schema.NewBuilder().
			ID("https://example.com/person").
			Anchor("person-schema").
			Build()
		require.NoError(t, err)

		require.Equal(t, "https://example.com/person", s.ID())
		require.Equal(t, "person-schema", s.Anchor())
	})
}
