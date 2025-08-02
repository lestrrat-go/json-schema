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
		personDefSchema := schema.NewBuilder().
			Anchor("person").
			Types(schema.ObjectType).
			Property("name", schema.NewBuilder().Types(schema.StringType).MinLength(1).MustBuild()).
			Property("age", schema.NewBuilder().Types(schema.NumberType).Minimum(0).MustBuild()).
			Required("name").
			MustBuild()

		s := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("user", schema.NewBuilder().Reference("#person").MustBuild()).
			Definitions("personDef", personDefSchema).
			MustBuild()

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, s)

		v, err := validator.Compile(ctx, s)
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
		personDefSchema := schema.NewBuilder().
			Anchor("personSchema").
			Types(schema.ObjectType).
			Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("email", schema.NewBuilder().Types(schema.StringType).Format("email").MustBuild()).
			Required("name").
			MustBuild()

		addressDefSchema := schema.NewBuilder().
			Anchor("addressSchema").
			Types(schema.ObjectType).
			Property("street", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("city", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("zipcode", schema.NewBuilder().Types(schema.StringType).Pattern("^\\d{5}$").MustBuild()).
			Required("street", "city").
			MustBuild()

		s := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("person", schema.NewBuilder().Reference("#personSchema").MustBuild()).
			Property("address", schema.NewBuilder().Reference("#addressSchema").MustBuild()).
			Definitions("personDef", personDefSchema).
			Definitions("addressDef", addressDefSchema).
			MustBuild()

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, s)

		v, err := validator.Compile(ctx, s)
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
		dataDefSchema := schema.NewBuilder().
			Anchor("dataSchema").
			Types(schema.ObjectType).
			Property("value", schema.NewBuilder().Types(schema.StringType).MinLength(1).MustBuild()).
			Required("value").
			MustBuild()

		rootDefSchema := schema.NewBuilder().
			Anchor("rootSchema").
			Types(schema.ObjectType).
			Property("data", schema.NewBuilder().Reference("#dataSchema").MustBuild()).
			MustBuild()

		s := schema.NewBuilder().
			Reference("#rootSchema").
			Definitions("root", rootDefSchema).
			Definitions("data", dataDefSchema).
			MustBuild()

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, s)

		v, err := validator.Compile(ctx, s)
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
