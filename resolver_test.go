package schema_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

func TestResolver(t *testing.T) {
	t.Run("NewResolver", func(t *testing.T) {
		resolver := schema.NewResolver()
		require.NotNil(t, resolver)
	})
}

func TestResolveLocalReference(t *testing.T) {
	// Since we can't add definitions directly (not in current schema),
	// let's test with a more realistic schema structure
	jsonSchema := `{
		"$id": "https://example.com/person",
		"type": "object",
		"properties": {
			"name": {"$ref": "#/$defs/stringType"},
			"age": {"$ref": "#/$defs/intType"}
		},
		"$defs": {
			"stringType": {"type": "string"},
			"intType": {"type": "integer", "minimum": 0}
		}
	}`

	var base schema.Schema
	require.NoError(t, base.UnmarshalJSON([]byte(jsonSchema)))

	resolver := schema.NewResolver()

	t.Run("resolve string definition", func(t *testing.T) {
		var resolved schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), &base)
		err := resolver.ResolveReference(ctx, &resolved, "#/$defs/stringType")
		require.NoError(t, err)
		require.True(t, resolved.ContainsType(schema.StringType))
	})

	t.Run("resolve integer definition", func(t *testing.T) {
		var resolved schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), &base)
		err := resolver.ResolveReference(ctx, &resolved, "#/$defs/intType")
		require.NoError(t, err)
		require.True(t, resolved.ContainsType(schema.IntegerType))
		require.Equal(t, float64(0), resolved.Minimum()) // JSON numbers are float64
	})

	t.Run("resolve non-existent reference", func(t *testing.T) {
		var resolved schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), &base)
		err := resolver.ResolveReference(ctx, &resolved, "#/$defs/nonexistent")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to resolve local reference")
	})
}

func TestResolveFileReference(t *testing.T) {
	// Create temporary schema files
	tmpDir := t.TempDir()

	// Create person.json
	personSchema := map[string]any{
		"$id":  "https://example.com/person",
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"age":  map[string]any{"type": "integer", "minimum": 0},
		},
		"required": []string{"name"},
	}

	personFile := filepath.Join(tmpDir, "person.json")
	personData, err := json.Marshal(personSchema)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(personFile, personData, 0644))

	// Create address.json that references person.json
	addressSchema := map[string]any{
		"$id":  "https://example.com/address",
		"type": "object",
		"properties": map[string]any{
			"street":   map[string]any{"type": "string"},
			"resident": map[string]any{"$ref": "person.json"},
		},
	}

	addressFile := filepath.Join(tmpDir, "address.json")
	addressData, err := json.Marshal(addressSchema)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(addressFile, addressData, 0644))

	// Change to tmpDir for relative path resolution
	t.Chdir(tmpDir)

	resolver := schema.NewResolver()

	t.Run("resolve file reference", func(t *testing.T) {
		var resolved schema.Schema
		ctx := context.Background()
		err := resolver.ResolveReference(ctx, &resolved, "person.json")
		require.NoError(t, err)
		require.Equal(t, "https://example.com/person", resolved.ID())
		require.True(t, resolved.ContainsType(schema.ObjectType))
	})

	t.Run("resolve file reference with fragment", func(t *testing.T) {
		var resolved schema.Schema
		ctx := context.Background()
		err := resolver.ResolveReference(ctx, &resolved, "person.json#/properties/name")
		require.NoError(t, err)
		require.True(t, resolved.ContainsType(schema.StringType))
	})

	t.Run("resolve non-existent file", func(t *testing.T) {
		var resolved schema.Schema
		ctx := context.Background()
		err := resolver.ResolveReference(ctx, &resolved, "nonexistent.json")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to resolve external reference")
	})
}

func TestResolveHTTPReference(t *testing.T) {
	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/person.json":
			personSchema := map[string]any{
				"$id":  "https://example.com/person",
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"age":  map[string]any{"type": "integer", "minimum": 0},
				},
				"$defs": map[string]any{
					"nameType": map[string]any{"type": "string", "minLength": 1},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(personSchema)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	resolver := schema.NewResolver()

	t.Run("resolve HTTP reference", func(t *testing.T) {
		var resolved schema.Schema
		ctx := context.Background()
		err := resolver.ResolveReference(ctx, &resolved, server.URL+"/person.json")
		require.NoError(t, err)
		require.Equal(t, "https://example.com/person", resolved.ID())
		require.True(t, resolved.ContainsType(schema.ObjectType))
	})

	t.Run("resolve HTTP reference with fragment", func(t *testing.T) {
		var resolved schema.Schema
		ctx := context.Background()
		err := resolver.ResolveReference(ctx, &resolved, server.URL+"/person.json#/$defs/nameType")
		require.NoError(t, err)
		require.True(t, resolved.ContainsType(schema.StringType))
		// Note: minLength would need to be checked if it was part of the schema struct
	})

	t.Run("resolve HTTP 404", func(t *testing.T) {
		var resolved schema.Schema
		ctx := context.Background()
		err := resolver.ResolveReference(ctx, &resolved, server.URL+"/nonexistent.json")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to resolve external reference")
	})
}

func TestValidateReference(t *testing.T) {
	testCases := []struct {
		name      string
		reference string
		expectErr bool
	}{
		{"empty reference", "", true},
		{"valid local reference", "#/definitions/person", false},
		{"valid root reference", "#", false},
		{"valid relative file reference", "person.json", false},
		{"valid relative with fragment", "person.json#/properties/name", false},
		{"valid absolute HTTP reference", "https://example.com/person.json", false},
		{"valid absolute with fragment", "https://example.com/person.json#/definitions/name", false},
		{"valid file URI", "file:///tmp/schema.json", false},
		{"invalid URI", "ht tp://invalid", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := schema.ValidateReference(tc.reference)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestResolverWithYAMLFiles(t *testing.T) {
	// Create temporary YAML schema file
	tmpDir := t.TempDir()

	yamlContent := `
$id: https://example.com/yaml-schema
type: object
properties:
  title:
    type: string
  count:
    type: integer
    minimum: 0
$defs:
  positiveInt:
    type: integer
    minimum: 1
`

	yamlFile := filepath.Join(tmpDir, "schema.yaml")
	require.NoError(t, os.WriteFile(yamlFile, []byte(yamlContent), 0644))

	// Change to tmpDir for relative path resolution
	t.Chdir(tmpDir)

	resolver := schema.NewResolver()

	t.Run("resolve YAML file", func(t *testing.T) {
		var resolved schema.Schema
		ctx := context.Background()
		err := resolver.ResolveReference(ctx, &resolved, "schema.yaml")
		require.NoError(t, err)
		require.Equal(t, "https://example.com/yaml-schema", resolved.ID())
		require.True(t, resolved.ContainsType(schema.ObjectType))
	})

	t.Run("resolve YAML file with fragment", func(t *testing.T) {
		var resolved schema.Schema
		ctx := context.Background()
		err := resolver.ResolveReference(ctx, &resolved, "schema.yaml#/$defs/positiveInt")
		require.NoError(t, err)
		require.True(t, resolved.ContainsType(schema.IntegerType))
		require.Equal(t, float64(1), resolved.Minimum()) // JSON numbers are float64
	})
}

func TestResolverSeparateAPIs(t *testing.T) {
	t.Run("ResolveJSONReference", func(t *testing.T) {
		// Schema with JSON pointer reference
		jsonSchema := `{
			"type": "object",
			"properties": {
				"user": {"$ref": "#/$defs/person"}
			},
			"$defs": {
				"person": {
					"type": "object",
					"properties": {
						"name": {"type": "string", "minLength": 1}
					},
					"required": ["name"]
				}
			}
		}`

		var baseSchema schema.Schema
		require.NoError(t, baseSchema.UnmarshalJSON([]byte(jsonSchema)))

		resolver := schema.NewResolver()

		// Test ResolveJSONReference directly
		var resolved schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), &baseSchema)
		err := resolver.ResolveJSONReference(ctx, &resolved, "#/$defs/person")
		require.NoError(t, err)

		// Verify resolved schema
		require.True(t, resolved.ContainsType(schema.ObjectType))
		require.True(t, resolved.HasProperties())
		nameSchema := resolved.Properties()["name"]
		require.NotNil(t, nameSchema)
		require.True(t, nameSchema.ContainsType(schema.StringType))
	})

	t.Run("ResolveAnchor", func(t *testing.T) {
		// Schema with anchor
		jsonSchema := `{
			"type": "object",
			"properties": {
				"user": {"$ref": "#person"}
			},
			"$defs": {
				"personDef": {
					"$anchor": "person",
					"type": "object",
					"properties": {
						"name": {"type": "string", "minLength": 1}
					},
					"required": ["name"]
				}
			}
		}`

		var baseSchema schema.Schema
		require.NoError(t, baseSchema.UnmarshalJSON([]byte(jsonSchema)))

		resolver := schema.NewResolver()

		// Test ResolveAnchor directly
		var resolved schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), &baseSchema)
		err := resolver.ResolveAnchor(ctx, &resolved, "person")
		require.NoError(t, err)

		// Verify resolved schema
		require.True(t, resolved.ContainsType(schema.ObjectType))
		require.True(t, resolved.HasProperties())
		nameSchema := resolved.Properties()["name"]
		require.NotNil(t, nameSchema)
		require.True(t, nameSchema.ContainsType(schema.StringType))
		require.Equal(t, "person", resolved.Anchor())
	})

	t.Run("ResolveReference unified API dispatching", func(t *testing.T) {
		// Schema with both JSON pointer and anchor references
		jsonSchema := `{
			"type": "object",
			"properties": {
				"userByPointer": {"$ref": "#/$defs/person"},
				"userByAnchor": {"$ref": "#person"}
			},
			"$defs": {
				"person": {
					"$anchor": "person",
					"type": "object",
					"properties": {
						"name": {"type": "string", "minLength": 1}
					},
					"required": ["name"]
				}
			}
		}`

		var baseSchema schema.Schema
		require.NoError(t, baseSchema.UnmarshalJSON([]byte(jsonSchema)))

		resolver := schema.NewResolver()

		// Test unified API with JSON pointer reference
		var resolvedPointer schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), &baseSchema)
		err := resolver.ResolveReference(ctx, &resolvedPointer, "#/$defs/person")
		require.NoError(t, err)
		require.True(t, resolvedPointer.ContainsType(schema.ObjectType))
		require.Equal(t, "person", resolvedPointer.Anchor())

		// Test unified API with anchor reference
		var resolvedAnchor schema.Schema
		err = resolver.ResolveReference(ctx, &resolvedAnchor, "#person")
		require.NoError(t, err)
		require.True(t, resolvedAnchor.ContainsType(schema.ObjectType))
		require.Equal(t, "person", resolvedAnchor.Anchor())

		// Both should resolve to the same schema
		require.Equal(t, resolvedPointer.Anchor(), resolvedAnchor.Anchor())
	})

	t.Run("ResolveAnchor error handling", func(t *testing.T) {
		// Schema without the requested anchor
		jsonSchema := `{
			"type": "object",
			"$defs": {
				"person": {
					"$anchor": "person",
					"type": "string"
				}
			}
		}`

		var baseSchema schema.Schema
		require.NoError(t, baseSchema.UnmarshalJSON([]byte(jsonSchema)))

		resolver := schema.NewResolver()

		// Try to resolve non-existent anchor
		var resolved schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), &baseSchema)
		err := resolver.ResolveAnchor(ctx, &resolved, "nonexistent")
		require.Error(t, err)
		require.Contains(t, err.Error(), "anchor nonexistent not found")
	})

	t.Run("ResolveJSONReference error handling", func(t *testing.T) {
		// Schema without the requested definition
		jsonSchema := `{
			"type": "object",
			"$defs": {
				"person": {
					"type": "string"
				}
			}
		}`

		var baseSchema schema.Schema
		require.NoError(t, baseSchema.UnmarshalJSON([]byte(jsonSchema)))

		resolver := schema.NewResolver()

		// Try to resolve non-existent JSON pointer
		var resolved schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), &baseSchema)
		err := resolver.ResolveJSONReference(ctx, &resolved, "#/$defs/nonexistent")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to resolve local JSON pointer reference")
	})

	t.Run("ResolveJSONReference with external URLs", func(t *testing.T) {
		// Create test HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/person.json":
				personSchema := map[string]any{
					"$id":  "https://example.com/person",
					"type": "object",
					"$defs": map[string]any{
						"nameType": map[string]any{
							"$anchor":   "personName",
							"type":      "string",
							"minLength": 1,
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(personSchema)
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		resolver := schema.NewResolver()

		// Test ResolveJSONReference with external URL and JSON pointer
		var resolved schema.Schema
		ctx := context.Background() // No base schema needed for external reference
		err := resolver.ResolveJSONReference(ctx, &resolved, server.URL+"/person.json#/$defs/nameType")
		require.NoError(t, err)
		require.True(t, resolved.ContainsType(schema.StringType))
		require.Equal(t, "personName", resolved.Anchor())
	})
}
