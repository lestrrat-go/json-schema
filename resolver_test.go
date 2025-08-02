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
	// Create schema with $defs using our builder
	base := schema.NewBuilder().
		ID("https://example.com/person").
		Types(schema.ObjectType).
		Property("name", schema.NewBuilder().Reference("#/$defs/stringType").MustBuild()).
		Property("age", schema.NewBuilder().Reference("#/$defs/intType").MustBuild()).
		Definitions("stringType", schema.NewBuilder().Types(schema.StringType).MustBuild()).
		Definitions("intType", schema.NewBuilder().Types(schema.IntegerType).Minimum(0).MustBuild()).
		MustBuild()

	resolver := schema.NewResolver()

	t.Run("resolve string definition", func(t *testing.T) {
		var resolved schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), base)
		err := resolver.ResolveReference(ctx, &resolved, "#/$defs/stringType")
		require.NoError(t, err)
		require.True(t, resolved.ContainsType(schema.StringType))
	})

	t.Run("resolve integer definition", func(t *testing.T) {
		var resolved schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), base)
		err := resolver.ResolveReference(ctx, &resolved, "#/$defs/intType")
		require.NoError(t, err)
		require.True(t, resolved.ContainsType(schema.IntegerType))
		require.Equal(t, float64(0), resolved.Minimum()) // JSON numbers are float64
	})

	t.Run("resolve non-existent reference", func(t *testing.T) {
		var resolved schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), base)
		err := resolver.ResolveReference(ctx, &resolved, "#/$defs/nonexistent")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to resolve local reference")
	})
}

func TestResolveFileReference(t *testing.T) {
	// Create temporary schema files
	tmpDir := t.TempDir()

	// Create person.json using builder
	personSchema := schema.NewBuilder().
		ID("https://example.com/person").
		Types(schema.ObjectType).
		Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
		Property("age", schema.NewBuilder().Types(schema.IntegerType).Minimum(0).MustBuild()).
		Required("name").
		MustBuild()

	personFile := filepath.Join(tmpDir, "person.json")
	personData, err := personSchema.MarshalJSON()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(personFile, personData, 0644))

	// Create address.json that references person.json using builder
	addressSchema := schema.NewBuilder().
		ID("https://example.com/address").
		Types(schema.ObjectType).
		Property("street", schema.NewBuilder().Types(schema.StringType).MustBuild()).
		Property("resident", schema.NewBuilder().Reference("person.json").MustBuild()).
		MustBuild()

	addressFile := filepath.Join(tmpDir, "address.json")
	addressData, err := addressSchema.MarshalJSON()
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
			// Create schema using builder instead of map
			personSchema := schema.NewBuilder().
				ID("https://example.com/person").
				Types(schema.ObjectType).
				Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
				Property("age", schema.NewBuilder().Types(schema.IntegerType).Minimum(0).MustBuild()).
				Definitions("nameType", schema.NewBuilder().Types(schema.StringType).MinLength(1).MustBuild()).
				MustBuild()

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

	// Create schema using builder and convert to YAML-formatted JSON
	yamlSchema := schema.NewBuilder().
		ID("https://example.com/yaml-schema").
		Types(schema.ObjectType).
		Property("title", schema.NewBuilder().Types(schema.StringType).MustBuild()).
		Property("count", schema.NewBuilder().Types(schema.IntegerType).Minimum(0).MustBuild()).
		Definitions("positiveInt", schema.NewBuilder().Types(schema.IntegerType).Minimum(1).MustBuild()).
		MustBuild()

	// Marshal to JSON and then format as YAML-style for file content
	schemaJSON, err := yamlSchema.MarshalJSON()
	require.NoError(t, err)

	// Write as JSON since resolver supports JSON in .yaml files
	yamlFile := filepath.Join(tmpDir, "schema.yaml")
	require.NoError(t, os.WriteFile(yamlFile, schemaJSON, 0644))

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
		// Schema with JSON pointer reference using builder
		baseSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("user", schema.NewBuilder().Reference("#/$defs/person").MustBuild()).
			Definitions("person", schema.NewBuilder().
				Types(schema.ObjectType).
				Property("name", schema.NewBuilder().Types(schema.StringType).MinLength(1).MustBuild()).
				Required("name").
				MustBuild()).
			MustBuild()

		resolver := schema.NewResolver()

		// Test ResolveJSONReference directly
		var resolved schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), baseSchema)
		err := resolver.ResolveJSONReference(ctx, &resolved, "#/$defs/person")
		require.NoError(t, err)

		// Verify resolved schema
		require.True(t, resolved.ContainsType(schema.ObjectType))
		require.True(t, resolved.Has(schema.PropertiesField))
		resolvedProps := resolved.Properties()
		var nameSchema schema.Schema
		err = resolvedProps.Get("name", &nameSchema)
		require.NoError(t, err)
		require.True(t, nameSchema.ContainsType(schema.StringType))
	})

	t.Run("ResolveAnchor", func(t *testing.T) {
		// Schema with anchor using builder
		baseSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("user", schema.NewBuilder().Reference("#person").MustBuild()).
			Definitions("personDef", schema.NewBuilder().
				Anchor("person").
				Types(schema.ObjectType).
				Property("name", schema.NewBuilder().Types(schema.StringType).MinLength(1).MustBuild()).
				Required("name").
				MustBuild()).
			MustBuild()

		resolver := schema.NewResolver()

		// Test ResolveAnchor directly
		var resolved schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), baseSchema)
		err := resolver.ResolveAnchor(ctx, &resolved, "person")
		require.NoError(t, err)

		// Verify resolved schema
		require.True(t, resolved.ContainsType(schema.ObjectType))
		require.True(t, resolved.Has(schema.PropertiesField))
		resolvedProps := resolved.Properties()
		var nameSchema schema.Schema
		err = resolvedProps.Get("name", &nameSchema)
		require.NoError(t, err)
		require.True(t, nameSchema.ContainsType(schema.StringType))
		require.Equal(t, "person", resolved.Anchor())
	})

	t.Run("ResolveReference unified API dispatching", func(t *testing.T) {
		// Schema with both JSON pointer and anchor references using builder
		baseSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("userByPointer", schema.NewBuilder().Reference("#/$defs/person").MustBuild()).
			Property("userByAnchor", schema.NewBuilder().Reference("#person").MustBuild()).
			Definitions("person", schema.NewBuilder().
				Anchor("person").
				Types(schema.ObjectType).
				Property("name", schema.NewBuilder().Types(schema.StringType).MinLength(1).MustBuild()).
				Required("name").
				MustBuild()).
			MustBuild()

		resolver := schema.NewResolver()

		// Test unified API with JSON pointer reference
		var resolvedPointer schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), baseSchema)
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
		// Schema without the requested anchor using builder
		baseSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Definitions("person", schema.NewBuilder().
				Anchor("person").
				Types(schema.StringType).
				MustBuild()).
			MustBuild()

		resolver := schema.NewResolver()

		// Try to resolve non-existent anchor
		var resolved schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), baseSchema)
		err := resolver.ResolveAnchor(ctx, &resolved, "nonexistent")
		require.Error(t, err)
		require.Contains(t, err.Error(), "anchor nonexistent not found")
	})

	t.Run("ResolveJSONReference error handling", func(t *testing.T) {
		// Schema without the requested definition using builder
		baseSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Definitions("person", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			MustBuild()

		resolver := schema.NewResolver()

		// Try to resolve non-existent JSON pointer
		var resolved schema.Schema
		ctx := schema.WithBaseSchema(context.Background(), baseSchema)
		err := resolver.ResolveJSONReference(ctx, &resolved, "#/$defs/nonexistent")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to resolve local JSON pointer reference")
	})

	t.Run("ResolveJSONReference with external URLs", func(t *testing.T) {
		// Create test HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/person.json":
				// Create schema using builder instead of map
				personSchema := schema.NewBuilder().
					ID("https://example.com/person").
					Types(schema.ObjectType).
					Definitions("nameType", schema.NewBuilder().
						Anchor("personName").
						Types(schema.StringType).
						MinLength(1).
						MustBuild()).
					MustBuild()

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
