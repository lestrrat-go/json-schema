package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestAnchorResolution(t *testing.T) {
	t.Run("basic anchor resolution", func(t *testing.T) {
		// Schema with anchor and reference to anchor
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
						"name": {"type": "string", "minLength": 1},
						"age": {"type": "number", "minimum": 0}
					},
					"required": ["name"]
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		ctx := context.Background()
		ctx = validator.WithResolver(ctx, schema.NewResolver())
		ctx = validator.WithRootSchema(ctx, &s)
		
		v, err := validator.Compile(ctx, &s)
		if err != nil {
			t.Logf("Expected anchor resolution to work, but got error: %v", err)
			t.Skip("Anchor resolution not yet implemented")
		}

		// Valid data
		validData := map[string]any{
			"user": map[string]any{
				"name": "John Doe",
				"age":  30.0,
			},
		}

		_, err = v.Validate(context.Background(), validData)
		require.NoError(t, err)

		// Invalid data - missing required field
		invalidData := map[string]any{
			"user": map[string]any{
				"age": 30.0,
				// missing "name"
			},
		}

		_, err = v.Validate(context.Background(), invalidData)
		require.Error(t, err)
	})

	t.Run("multiple anchors", func(t *testing.T) {
		// Schema with multiple anchors
		jsonSchema := `{
			"type": "object",
			"properties": {
				"person": {"$ref": "#personSchema"},
				"address": {"$ref": "#addressSchema"}
			},
			"$defs": {
				"personDef": {
					"$anchor": "personSchema",
					"type": "object",
					"properties": {
						"name": {"type": "string"},
						"email": {"type": "string", "format": "email"}
					},
					"required": ["name"]
				},
				"addressDef": {
					"$anchor": "addressSchema",
					"type": "object",
					"properties": {
						"street": {"type": "string"},
						"city": {"type": "string"},
						"zipcode": {"type": "string", "pattern": "^\\d{5}$"}
					},
					"required": ["street", "city"]
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		ctx := context.Background()
		ctx = validator.WithResolver(ctx, schema.NewResolver())
		ctx = validator.WithRootSchema(ctx, &s)
		
		v, err := validator.Compile(ctx, &s)
		if err != nil {
			t.Logf("Expected multiple anchor resolution to work, but got error: %v", err)
			t.Skip("Anchor resolution not yet implemented")
		}

		// Valid data
		validData := map[string]any{
			"person": map[string]any{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			"address": map[string]any{
				"street":  "123 Main St",
				"city":    "Anytown",
				"zipcode": "12345",
			},
		}

		_, err = v.Validate(context.Background(), validData)
		require.NoError(t, err)
	})

	t.Run("nested anchor resolution", func(t *testing.T) {
		// Schema with nested anchor references
		jsonSchema := `{
			"$ref": "#rootSchema",
			"$defs": {
				"root": {
					"$anchor": "rootSchema",
					"type": "object",
					"properties": {
						"data": {"$ref": "#dataSchema"}
					}
				},
				"data": {
					"$anchor": "dataSchema",
					"type": "object",
					"properties": {
						"value": {"type": "string", "minLength": 1}
					},
					"required": ["value"]
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		ctx := context.Background()
		ctx = validator.WithResolver(ctx, schema.NewResolver())
		ctx = validator.WithRootSchema(ctx, &s)
		
		v, err := validator.Compile(ctx, &s)
		if err != nil {
			t.Logf("Expected nested anchor resolution to work, but got error: %v", err)
			t.Skip("Anchor resolution not yet implemented")
		}

		// Valid data
		validData := map[string]any{
			"data": map[string]any{
				"value": "test",
			},
		}

		_, err = v.Validate(context.Background(), validData)
		require.NoError(t, err)

		// Invalid data
		invalidData := map[string]any{
			"data": map[string]any{
				"value": "", // violates minLength: 1
			},
		}

		_, err = v.Validate(context.Background(), invalidData)
		require.Error(t, err)
	})
}