package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestReferenceResolution(t *testing.T) {
	t.Run("local reference", func(t *testing.T) {
		// Schema with a local reference
		jsonSchema := `{
			"type": "object",
			"properties": {
				"name": {"$ref": "#/$defs/stringType"}
			},
			"$defs": {
				"stringType": {"type": "string", "minLength": 1}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		// Set up context with resolver and root schema
		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, &s)
		
		v, err := validator.Compile(ctx, &s)
		require.NoError(t, err)

		t.Run("valid object", func(t *testing.T) {
			obj := map[string]any{
				"name": "John",
			}
			_, err := v.Validate(context.Background(), obj)
			require.NoError(t, err)
		})

		t.Run("invalid object - empty string", func(t *testing.T) {
			obj := map[string]any{
				"name": "",
			}
			_, err := v.Validate(context.Background(), obj)
			require.Error(t, err)
		})

		t.Run("invalid object - non-string", func(t *testing.T) {
			obj := map[string]any{
				"name": 123,
			}
			_, err := v.Validate(context.Background(), obj)
			require.Error(t, err)
		})
	})

	t.Run("schema with $ref only", func(t *testing.T) {
		// A schema that is just a reference
		jsonSchema := `{
			"$ref": "#/$defs/personType",
			"$defs": {
				"personType": {
					"type": "object",
					"properties": {
						"name": {"type": "string"},
						"age": {"type": "integer", "minimum": 0}
					},
					"required": ["name"]
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		// Set up context with resolver and root schema
		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, &s)
		
		v, err := validator.Compile(ctx, &s)
		require.NoError(t, err)

		t.Run("valid person", func(t *testing.T) {
			person := map[string]any{
				"name": "Alice",
				"age":  30,
			}
			_, err := v.Validate(context.Background(), person)
			require.NoError(t, err)
		})

		t.Run("invalid person - missing name", func(t *testing.T) {
			person := map[string]any{
				"age": 30,
			}
			_, err := v.Validate(context.Background(), person)
			require.Error(t, err)
		})
	})
}