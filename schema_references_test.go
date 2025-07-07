package schema_test

import (
	"encoding/json"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

// TestSchemaReferences tests schema reference resolution as per JSON Schema 2020-12
func TestSchemaReferences(t *testing.T) {
	t.Run("Basic Reference", func(t *testing.T) {
		// Test basic $ref functionality
		s, err := schema.NewBuilder().
			Reference("#/definitions/person").
			Build()
		require.NoError(t, err)
		require.Equal(t, "#/definitions/person", s.Reference())
	})

	t.Run("Absolute URI Reference", func(t *testing.T) {
		// Test absolute URI reference
		s, err := schema.NewBuilder().
			Reference("https://example.com/schemas/person.json").
			Build()
		require.NoError(t, err)
		require.Equal(t, "https://example.com/schemas/person.json", s.Reference())
	})

	t.Run("Relative Reference", func(t *testing.T) {
		// Test relative reference
		s, err := schema.NewBuilder().
			Reference("person.json").
			Build()
		require.NoError(t, err)
		require.Equal(t, "person.json", s.Reference())
	})

	t.Run("Fragment Reference", func(t *testing.T) {
		// Test fragment-only reference
		s, err := schema.NewBuilder().
			Reference("#person").
			Build()
		require.NoError(t, err)
		require.Equal(t, "#person", s.Reference())
	})

	t.Run("Dynamic Reference", func(t *testing.T) {
		// Test $dynamicRef functionality
		s, err := schema.NewBuilder().
			DynamicReference("#person").
			Build()
		require.NoError(t, err)
		require.Equal(t, "#person", s.DynamicReference())
	})

	t.Run("Dynamic Reference with URI", func(t *testing.T) {
		// Test $dynamicRef with full URI
		s, err := schema.NewBuilder().
			DynamicReference("https://example.com/schemas/person.json#person").
			Build()
		require.NoError(t, err)
		require.Equal(t, "https://example.com/schemas/person.json#person", s.DynamicReference())
	})
}

// TestSchemaDefinitions tests schema definitions and references
func TestSchemaDefinitions(t *testing.T) {
	t.Run("Schema with Definitions", func(t *testing.T) {
		// Test $defs keyword
		s, err := schema.NewBuilder().
			Definitions("person").
			Build()
		require.NoError(t, err)
		require.Equal(t, "person", s.Definitions())
	})

	t.Run("Schema with ID and Definitions", func(t *testing.T) {
		// Test schema with both ID and definitions
		s, err := schema.NewBuilder().
			ID("https://example.com/schemas/main").
			Definitions("person").
			Build()
		require.NoError(t, err)
		require.Equal(t, "https://example.com/schemas/main", s.ID())
		require.Equal(t, "person", s.Definitions())
	})

	t.Run("Complex Schema with Multiple References", func(t *testing.T) {
		// Test schema with multiple reference types
		s, err := schema.NewBuilder().
			ID("https://example.com/schemas/complex").
			Reference("#/definitions/base").
			DynamicReference("#/definitions/dynamic").
			Definitions("definitions").
			Anchor("main").
			Build()
		require.NoError(t, err)
		
		require.Equal(t, "https://example.com/schemas/complex", s.ID())
		require.Equal(t, "#/definitions/base", s.Reference())
		require.Equal(t, "#/definitions/dynamic", s.DynamicReference())
		require.Equal(t, "definitions", s.Definitions())
		require.Equal(t, "main", s.Anchor())
	})
}

// TestSchemaReferenceResolution tests reference resolution behavior
func TestSchemaReferenceResolution(t *testing.T) {
	t.Run("Self-Reference", func(t *testing.T) {
		// Test schema that references itself
		s, err := schema.NewBuilder().
			ID("https://example.com/recursive").
			Type(schema.ObjectType).
			Property("name", schema.NewBuilder().Type(schema.StringType).MustBuild()).
			Property("child", schema.NewBuilder().Reference("#").MustBuild()).
			Build()
		require.NoError(t, err)
		
		// Verify structure
		require.Equal(t, "https://example.com/recursive", s.ID())
		require.NotNil(t, s.Properties()["name"])
		require.NotNil(t, s.Properties()["child"])
		require.Equal(t, "#", s.Properties()["child"].Reference())
	})

	t.Run("Cross-Reference", func(t *testing.T) {
		// Test schema that references another schema
		personSchema, err := schema.NewBuilder().
			ID("https://example.com/person").
			Type(schema.ObjectType).
			Property("name", schema.NewBuilder().Type(schema.StringType).MustBuild()).
			Build()
		require.NoError(t, err)
		
		addressSchema, err := schema.NewBuilder().
			ID("https://example.com/address").
			Type(schema.ObjectType).
			Property("street", schema.NewBuilder().Type(schema.StringType).MustBuild()).
			Property("resident", schema.NewBuilder().Reference("https://example.com/person").MustBuild()).
			Build()
		require.NoError(t, err)
		
		// Verify cross-reference
		require.Equal(t, "https://example.com/person", personSchema.ID())
		require.Equal(t, "https://example.com/address", addressSchema.ID())
		require.Equal(t, "https://example.com/person", addressSchema.Properties()["resident"].Reference())
	})

	t.Run("Nested References", func(t *testing.T) {
		// Test schema with nested references
		s, err := schema.NewBuilder().
			ID("https://example.com/nested").
			Type(schema.ObjectType).
			Property("items", schema.NewBuilder().
				Type(schema.ArrayType).
				Items(schema.NewBuilder().Reference("#/definitions/item").MustBuild()).
				MustBuild()).
			Build()
		require.NoError(t, err)
		
		// Verify nested reference structure
		require.Equal(t, "https://example.com/nested", s.ID())
		itemsProp := s.Properties()["items"]
		require.NotNil(t, itemsProp)
		require.NotNil(t, itemsProp.Items())
		require.Equal(t, "#/definitions/item", itemsProp.Items().Reference())
	})
}

// TestSchemaReferencesSerialization tests that references serialize correctly
func TestSchemaReferencesSerialization(t *testing.T) {
	t.Run("Reference Serialization", func(t *testing.T) {
		s, err := schema.NewBuilder().
			ID("https://example.com/test").
			Reference("#/definitions/person").
			DynamicReference("#/definitions/dynamic").
			Definitions("person").
			Anchor("main").
			Build()
		require.NoError(t, err)
		
		// Serialize to JSON
		jsonData, err := json.Marshal(s)
		require.NoError(t, err)
		
		// Parse back to verify
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)
		
		// Verify reference fields in JSON
		require.Equal(t, "https://example.com/test", jsonMap["$id"])
		require.Equal(t, "#/definitions/person", jsonMap["$ref"])
		require.Equal(t, "#/definitions/dynamic", jsonMap["$dynamicRef"])
		require.Equal(t, "person", jsonMap["$defs"])
		require.Equal(t, "main", jsonMap["$anchor"])
	})

	t.Run("Complex Reference Structure", func(t *testing.T) {
		// Create a schema with complex reference structure
		s, err := schema.NewBuilder().
			ID("https://example.com/complex-refs").
			Type(schema.ObjectType).
			Property("base", schema.NewBuilder().Reference("#/definitions/base").MustBuild()).
			Property("dynamic", schema.NewBuilder().DynamicReference("#/definitions/dynamic").MustBuild()).
			Property("external", schema.NewBuilder().Reference("https://external.com/schema").MustBuild()).
			Build()
		require.NoError(t, err)
		
		// Serialize to JSON
		jsonData, err := json.Marshal(s)
		require.NoError(t, err)
		
		// Parse back to verify
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)
		
		// Verify complex structure
		require.Equal(t, "https://example.com/complex-refs", jsonMap["$id"])
		require.Equal(t, "object", jsonMap["type"])
		
		props, ok := jsonMap["properties"].(map[string]interface{})
		require.True(t, ok)
		
		// Check base property reference
		baseProp, ok := props["base"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "#/definitions/base", baseProp["$ref"])
		
		// Check dynamic property reference
		dynamicProp, ok := props["dynamic"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "#/definitions/dynamic", dynamicProp["$dynamicRef"])
		
		// Check external property reference
		externalProp, ok := props["external"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "https://external.com/schema", externalProp["$ref"])
	})
}

// TestSchemaIdentificationAndAnchoring tests schema identification mechanisms
func TestSchemaIdentificationAndAnchoring(t *testing.T) {
	t.Run("Schema Identification", func(t *testing.T) {
		testCases := []struct {
			name string
			id   string
		}{
			{"Simple ID", "https://example.com/schema"},
			{"ID with Path", "https://example.com/schemas/person.json"},
			{"ID with Fragment", "https://example.com/schema#person"},
			{"ID with Query", "https://example.com/schema?version=1"},
			{"Relative ID", "/schemas/person"},
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

	t.Run("Schema Anchoring", func(t *testing.T) {
		// Test various anchor patterns
		testCases := []string{
			"main",
			"person-schema",
			"address_schema",
			"item.schema",
			"123-schema",
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

	t.Run("Combined ID and Anchor", func(t *testing.T) {
		// Test schema with both ID and anchor
		s, err := schema.NewBuilder().
			ID("https://example.com/schemas/main.json").
			Anchor("root").
			Build()
		require.NoError(t, err)
		
		require.Equal(t, "https://example.com/schemas/main.json", s.ID())
		require.Equal(t, "root", s.Anchor())
	})
}