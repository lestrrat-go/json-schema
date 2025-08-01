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
		isStringSchema := schema.NewBuilder().
			Property("type", schema.NewBuilder().Const("string").MustBuild()).
			MustBuild()

		stringValueSchema := schema.NewBuilder().
			Types(schema.StringType).
			MinLength(1).
			MustBuild()

		numberValueSchema := schema.NewBuilder().
			Types(schema.NumberType).
			Minimum(0).
			MustBuild()

		thenSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("type", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("value", schema.NewBuilder().Reference("#/$defs/stringValue").MustBuild()).
			Required("type", "value").
			MustBuild()

		elseSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("type", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("value", schema.NewBuilder().Reference("#/$defs/numberValue").MustBuild()).
			Required("type", "value").
			MustBuild()

		s := schema.NewBuilder().
			IfSchema(schema.NewBuilder().Reference("#/$defs/isString").MustBuild()).
			ThenSchema(thenSchema).
			ElseSchema(elseSchema).
			Definitions("isString", isStringSchema).
			Definitions("stringValue", stringValueSchema).
			Definitions("numberValue", numberValueSchema).
			MustBuild()

		// Set up context with resolver and root schema
		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, s)

		v, err := validator.Compile(ctx, s)
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
		hasSimpleFlagSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("simple", schema.NewBuilder().Const(true).MustBuild()).
			Required("simple").
			MustBuild()

		basicConfigSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("simple", schema.NewBuilder().Types(schema.BooleanType).MustBuild()).
			Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Required("simple").
			MustBuild()

		advancedConfigSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("features", schema.NewBuilder().
				Types(schema.ArrayType).
				Items(schema.NewBuilder().Types(schema.StringType).MustBuild()).
				MustBuild()).
			Property("enabled", schema.NewBuilder().Types(schema.BooleanType).MustBuild()).
			Required("features", "enabled").
			MustBuild()

		s := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("category", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("config", schema.NewBuilder().
				IfSchema(schema.NewBuilder().Reference("#/$defs/hasSimpleFlag").MustBuild()).
				ThenSchema(schema.NewBuilder().Reference("#/$defs/basicConfig").MustBuild()).
				ElseSchema(schema.NewBuilder().Reference("#/$defs/advancedConfig").MustBuild()).
				MustBuild()).
			Required("category", "config").
			Definitions("hasSimpleFlag", hasSimpleFlagSchema).
			Definitions("basicConfig", basicConfigSchema).
			Definitions("advancedConfig", advancedConfigSchema).
			MustBuild()

		// Set up context with resolver and root schema
		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, s)

		v, err := validator.Compile(ctx, s)
		require.NoError(t, err)

		t.Run("basic config - then branch", func(t *testing.T) {
			obj := map[string]any{
				"category": "basic",
				"config": map[string]any{
					"simple": true,
					"name":   "test",
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
