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
		personSchema, err := schema.NewBuilder().Types(schema.StringType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Definitions("person", personSchema).
			Build()
		require.NoError(t, err)
		require.True(t, s.Has(schema.DefinitionsField))
		defs := s.Definitions()
		require.Contains(t, defs, "person")
		require.Equal(t, personSchema, defs["person"])
	})

	t.Run("Schema with ID and Definitions", func(t *testing.T) {
		// Test schema with both ID and definitions
		personSchema, err := schema.NewBuilder().Types(schema.ObjectType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			ID("https://example.com/schemas/main").
			Definitions("person", personSchema).
			Build()
		require.NoError(t, err)
		require.Equal(t, "https://example.com/schemas/main", s.ID())
		defs := s.Definitions()
		require.Contains(t, defs, "person")
		require.Equal(t, personSchema, defs["person"])
	})

	t.Run("Complex Schema with Multiple References", func(t *testing.T) {
		// Test schema with multiple reference types
		baseSchema, err := schema.NewBuilder().Types(schema.ObjectType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			ID("https://example.com/schemas/complex").
			Reference("#/$defs/base").
			DynamicReference("#/$defs/dynamic").
			Definitions("base", baseSchema).
			Anchor("main").
			Build()
		require.NoError(t, err)

		require.Equal(t, "https://example.com/schemas/complex", s.ID())
		require.Equal(t, "#/$defs/base", s.Reference())
		require.Equal(t, "#/$defs/dynamic", s.DynamicReference())
		defs := s.Definitions()
		require.Contains(t, defs, "base")
		require.Equal(t, baseSchema, defs["base"])
		require.Equal(t, "main", s.Anchor())
	})
}

// TestSchemaReferenceResolution tests reference resolution behavior
func TestSchemaReferenceResolution(t *testing.T) {
	t.Run("Self-Reference", func(t *testing.T) {
		// Test schema that references itself
		s, err := schema.NewBuilder().
			ID("https://example.com/recursive").
			Types(schema.ObjectType).
			Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
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
			Types(schema.ObjectType).
			Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Build()
		require.NoError(t, err)

		addressSchema, err := schema.NewBuilder().
			ID("https://example.com/address").
			Types(schema.ObjectType).
			Property("street", schema.NewBuilder().Types(schema.StringType).MustBuild()).
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
			Types(schema.ObjectType).
			Property("items", schema.NewBuilder().
				Types(schema.ArrayType).
				Items(schema.NewBuilder().Reference("#/definitions/item").MustBuild()).
				MustBuild()).
			Build()
		require.NoError(t, err)

		// Verify nested reference structure
		require.Equal(t, "https://example.com/nested", s.ID())
		itemsProp := s.Properties()["items"]
		require.NotNil(t, itemsProp)
		require.NotNil(t, itemsProp.Items())
		itemsSchema, ok := itemsProp.Items().(*schema.Schema)
		require.True(t, ok, "Items should be a *Schema, not a boolean")
		require.Equal(t, "#/definitions/item", itemsSchema.Reference())
	})
}

// TestSchemaReferencesSerialization tests that references serialize correctly
func TestSchemaReferencesSerialization(t *testing.T) {
	t.Run("Reference Serialization", func(t *testing.T) {
		personSchema, err := schema.NewBuilder().Types(schema.StringType).Build()
		require.NoError(t, err)

		original, err := schema.NewBuilder().
			ID("https://example.com/test").
			Reference("#/$defs/person").
			DynamicReference("#/$defs/dynamic").
			Definitions("person", personSchema).
			Anchor("main").
			Build()
		require.NoError(t, err)

		// Serialize to JSON
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		// Verify JSON structure by unmarshaling back to Schema
		var s schema.Schema
		require.NoError(t, json.Unmarshal(jsonData, &s), `json.Unmarshal should succeed`)

		// Verify reference fields
		require.Equal(t, "https://example.com/test", s.ID())
		require.Equal(t, "#/$defs/person", s.Reference())
		require.Equal(t, "#/$defs/dynamic", s.DynamicReference())
		defs := s.Definitions()
		require.Contains(t, defs, "person")
		require.True(t, defs["person"].ContainsType(schema.StringType))
		require.Equal(t, "main", s.Anchor())
	})

	t.Run("Complex Reference Structure", func(t *testing.T) {
		// Create a schema with complex reference structure
		original, err := schema.NewBuilder().
			ID("https://example.com/complex-refs").
			Types(schema.ObjectType).
			Property("base", schema.NewBuilder().Reference("#/$defs/base").MustBuild()).
			Property("dynamic", schema.NewBuilder().DynamicReference("#/$defs/dynamic").MustBuild()).
			Property("external", schema.NewBuilder().Reference("https://external.com/schema").MustBuild()).
			Build()
		require.NoError(t, err)

		// Serialize to JSON
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		// Verify JSON structure by unmarshaling back to Schema
		var s schema.Schema
		require.NoError(t, json.Unmarshal(jsonData, &s), `json.Unmarshal should succeed`)

		// Verify complex structure
		require.Equal(t, "https://example.com/complex-refs", s.ID())
		require.True(t, s.ContainsType(schema.ObjectType))

		props := s.Properties()
		require.NotNil(t, props)

		// Check base property reference
		baseProp := props["base"]
		require.NotNil(t, baseProp)
		require.Equal(t, "#/$defs/base", baseProp.Reference())

		// Check dynamic property reference
		dynamicProp := props["dynamic"]
		require.NotNil(t, dynamicProp)
		require.Equal(t, "#/$defs/dynamic", dynamicProp.DynamicReference())

		// Check external property reference
		externalProp := props["external"]
		require.NotNil(t, externalProp)
		require.Equal(t, "https://external.com/schema", externalProp.Reference())
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
