package schema_test

import (
	"encoding/json"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

// TestMetaSchemaCompliance tests meta-schema compliance for JSON Schema 2020-12
func TestMetaSchemaCompliance(t *testing.T) {
	t.Run("Default Meta-Schema", func(t *testing.T) {
		// Test that new schemas use the correct default meta-schema
		s := schema.New()
		require.Equal(t, schema.Version, s.Schema())
		require.Equal(t, "https://json-schema.org/draft/2020-12/schema", s.Schema())
	})

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
			Type(schema.StringType).
			Build()
		require.NoError(t, err)
		
		// Should have default meta-schema
		require.NotEmpty(t, s.Schema())
		require.Equal(t, schema.Version, s.Schema())
	})
}

// TestJSONSchemaMetaValidation tests that schemas can be validated against meta-schemas
func TestJSONSchemaMetaValidation(t *testing.T) {
	t.Run("Valid Schema Structure", func(t *testing.T) {
		// Create a valid schema
		s, err := schema.NewBuilder().
			ID("https://example.com/valid").
			Schema("https://json-schema.org/draft/2020-12/schema").
			Type(schema.ObjectType).
			Property("name", schema.NewBuilder().Type(schema.StringType).MustBuild()).
			Property("age", schema.NewBuilder().Type(schema.IntegerType).Minimum(0).MustBuild()).
			Required(true).
			Build()
		require.NoError(t, err)
		
		// Serialize to JSON to check structure
		jsonData, err := json.Marshal(s)
		require.NoError(t, err)
		
		// Parse back to verify structure
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)
		
		// Check that all fields are properly structured
		require.Equal(t, "https://example.com/valid", jsonMap["$id"])
		require.Equal(t, "https://json-schema.org/draft/2020-12/schema", jsonMap["$schema"])
		require.Equal(t, "object", jsonMap["type"])
		require.NotNil(t, jsonMap["properties"])
		require.Equal(t, true, jsonMap["required"])
	})

	t.Run("Schema with All Core Keywords", func(t *testing.T) {
		// Test schema with all supported core keywords
		s, err := schema.NewBuilder().
			ID("https://example.com/comprehensive").
			Schema("https://json-schema.org/draft/2020-12/schema").
			Reference("#/definitions/base").
			Anchor("main").
			DynamicReference("#meta").
			Comment("Comprehensive test schema").
			Build()
		require.NoError(t, err)
		
		// Serialize and verify structure
		jsonData, err := json.Marshal(s)
		require.NoError(t, err)
		
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)
		
		// Verify all core keywords are present
		require.Equal(t, "https://example.com/comprehensive", jsonMap["$id"])
		require.Equal(t, "https://json-schema.org/draft/2020-12/schema", jsonMap["$schema"])
		require.Equal(t, "#/definitions/base", jsonMap["$ref"])
		require.Equal(t, "main", jsonMap["$anchor"])
		require.Equal(t, "#meta", jsonMap["$dynamicRef"])
		require.Equal(t, "Comprehensive test schema", jsonMap["$comment"])
	})

	t.Run("Schema with Composition Keywords", func(t *testing.T) {
		// Test schema with composition (allOf, anyOf, oneOf, not)
		stringSchema, err := schema.NewBuilder().Type(schema.StringType).Build()
		require.NoError(t, err)
		
		numberSchema, err := schema.NewBuilder().Type(schema.NumberType).Build()
		require.NoError(t, err)
		
		s, err := schema.NewBuilder().
			AllOf([]*schema.Schema{stringSchema}).
			AnyOf([]*schema.Schema{stringSchema, numberSchema}).
			OneOf([]*schema.Schema{stringSchema}).
			Not(numberSchema).
			Build()
		require.NoError(t, err)
		
		// Serialize and verify structure
		jsonData, err := json.Marshal(s)
		require.NoError(t, err)
		
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)
		
		// Verify composition keywords
		require.NotNil(t, jsonMap["allOf"])
		require.NotNil(t, jsonMap["anyOf"])
		require.NotNil(t, jsonMap["oneOf"])
		require.NotNil(t, jsonMap["not"])
		
		// Check array lengths
		allOf, ok := jsonMap["allOf"].([]interface{})
		require.True(t, ok)
		require.Len(t, allOf, 1)
		
		anyOf, ok := jsonMap["anyOf"].([]interface{})
		require.True(t, ok)
		require.Len(t, anyOf, 2)
		
		oneOf, ok := jsonMap["oneOf"].([]interface{})
		require.True(t, ok)
		require.Len(t, oneOf, 1)
	})
}

// TestSchemaVocabularyDeclaration tests vocabulary declaration
func TestSchemaVocabularyDeclaration(t *testing.T) {
	t.Run("Core Vocabulary Support", func(t *testing.T) {
		// Test that the implementation supports core vocabulary
		s := schema.New()
		
		// Should have the correct version which implies core vocabulary support
		require.Equal(t, "https://json-schema.org/draft/2020-12/schema", s.Schema())
	})

	t.Run("Schema Version Compatibility", func(t *testing.T) {
		// Test version string matches expected format
		require.Equal(t, "https://json-schema.org/draft/2020-12/schema", schema.Version)
		
		// Test that schemas declare this version by default
		s := schema.New()
		require.Equal(t, schema.Version, s.Schema())
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

// TestBooleanSchemas tests boolean schema values
func TestBooleanSchemas(t *testing.T) {
	t.Run("True Schema", func(t *testing.T) {
		// true schema should accept everything
		var s schema.Schema
		err := s.Accept(true)
		require.NoError(t, err)
		
		// Should be an empty schema (accepts everything)
		require.Nil(t, s.Not())
	})

	t.Run("False Schema", func(t *testing.T) {
		// false schema should reject everything
		var s schema.Schema
		err := s.Accept(false)
		require.NoError(t, err)
		
		// Should have a 'not' constraint with empty schema
		require.NotNil(t, s.Not())
	})

	t.Run("Invalid Schema Value", func(t *testing.T) {
		// Test invalid value for schema
		var s schema.Schema
		err := s.Accept("invalid")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid value for additionalProperties")
	})
}