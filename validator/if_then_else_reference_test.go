package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestIfThenElseWithReferences(t *testing.T) {
	t.Run("if/then/else with references", func(t *testing.T) {
		// Schema with if/then/else that uses references
		// The if condition checks the root object to determine the data type
		jsonSchema := `{
			"if": {"$ref": "#/$defs/isString"},
			"then": {
				"type": "object",
				"properties": {
					"type": {"type": "string"},
					"value": {"$ref": "#/$defs/stringValue"}
				},
				"required": ["type", "value"]
			},
			"else": {
				"type": "object",
				"properties": {
					"type": {"type": "string"}, 
					"value": {"$ref": "#/$defs/numberValue"}
				},
				"required": ["type", "value"]
			},
			"$defs": {
				"isString": {
					"properties": {
						"type": {"const": "string"}
					}
				},
				"stringValue": {
					"type": "string",
					"minLength": 1
				},
				"numberValue": {
					"type": "number",
					"minimum": 0
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		// Set up context with resolver and root schema
		ctx := context.Background()
		ctx = validator.WithResolver(ctx, schema.NewResolver())
		ctx = validator.WithRootSchema(ctx, &s)
		
		v, err := validator.Compile(ctx, &s)
		require.NoError(t, err)

		t.Run("then branch - string type", func(t *testing.T) {
			obj := map[string]any{
				"type":  "string",
				"value": "hello",
			}
			_, err := v.Validate(context.Background(), obj)
			require.NoError(t, err)
		})

		t.Run("then branch - invalid string", func(t *testing.T) {
			obj := map[string]any{
				"type":  "string",
				"value": "", // violates minLength: 1
			}
			_, err := v.Validate(context.Background(), obj)
			require.Error(t, err)
		})

		t.Run("else branch - number type", func(t *testing.T) {
			obj := map[string]any{
				"type":  "number",
				"value": 42.5,
			}
			_, err := v.Validate(context.Background(), obj)
			require.NoError(t, err)
		})

		t.Run("else branch - invalid number", func(t *testing.T) {
			obj := map[string]any{
				"type":  "number",
				"value": -1, // violates minimum: 0
			}
			_, err := v.Validate(context.Background(), obj)
			require.Error(t, err)
		})

		t.Run("else branch - wrong type", func(t *testing.T) {
			obj := map[string]any{
				"type":  "number",
				"value": "should be number", // violates type: number
			}
			_, err := v.Validate(context.Background(), obj)
			require.Error(t, err)
		})
	})

	t.Run("property-level if/then/else with references", func(t *testing.T) {
		// Schema that uses references in property-level if/then/else
		jsonSchema := `{
			"type": "object",
			"properties": {
				"category": {"type": "string"},
				"config": {
					"if": {"$ref": "#/$defs/hasSimpleFlag"},
					"then": {"$ref": "#/$defs/basicConfig"},
					"else": {"$ref": "#/$defs/advancedConfig"}
				}
			},
			"required": ["category", "config"],
			"$defs": {
				"hasSimpleFlag": {
					"type": "object",
					"properties": {
						"simple": {"const": true}
					},
					"required": ["simple"]
				},
				"basicConfig": {
					"type": "object",
					"properties": {
						"simple": {"type": "boolean"},
						"name": {"type": "string"}
					},
					"required": ["simple"]
				},
				"advancedConfig": {
					"type": "object",
					"properties": {
						"features": {"type": "array", "items": {"type": "string"}},
						"enabled": {"type": "boolean"}
					},
					"required": ["features", "enabled"]
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		// Set up context with resolver and root schema
		ctx := context.Background()
		ctx = validator.WithResolver(ctx, schema.NewResolver())
		ctx = validator.WithRootSchema(ctx, &s)
		
		v, err := validator.Compile(ctx, &s)
		require.NoError(t, err)

		t.Run("basic config - then branch", func(t *testing.T) {
			obj := map[string]any{
				"category": "basic",
				"config": map[string]any{
					"simple": true,
					"name": "test",
				},
			}
			_, err := v.Validate(context.Background(), obj)
			require.NoError(t, err)
		})

		t.Run("advanced config - else branch", func(t *testing.T) {
			obj := map[string]any{
				"category": "advanced",
				"config": map[string]any{
					"features": []any{"feature1", "feature2"},
					"enabled":  true,
				},
			}
			_, err := v.Validate(context.Background(), obj)
			require.NoError(t, err)
		})

		t.Run("invalid advanced config - missing required field", func(t *testing.T) {
			obj := map[string]any{
				"category": "advanced",
				"config": map[string]any{
					"features": []any{"feature1", "feature2"},
					// missing "enabled" field
				},
			}
			_, err := v.Validate(context.Background(), obj)
			require.Error(t, err)
		})
	})
}