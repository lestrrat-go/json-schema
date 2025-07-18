package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// TestCousinUnevaluatedProperties tests that cousin schemas in allOf cannot see each other's evaluated properties
func TestCousinUnevaluatedProperties(t *testing.T) {
	t.Run("cousin unevaluatedProperties can't see inside cousins", func(t *testing.T) {
		// This test comes from the JSON Schema Test Suite
		// Schema: cousin schemas in allOf where one has unevaluatedProperties: false
		// should not be able to see properties defined in the other cousin
		
		// Create the first cousin schema that defines a "foo" property
		cousin1 := schema.NewBuilder().
			Property("foo", schema.NewBuilder().MustBuild()).
			MustBuild()
		
		// Create the second cousin schema with unevaluatedProperties: false
		cousin2 := schema.NewBuilder().
			UnevaluatedProperties(schema.SchemaFalse()).
			MustBuild()
		
		// Create the main schema with allOf containing both cousins
		mainSchema := schema.NewBuilder().
			AllOf(cousin1, cousin2).
			MustBuild()
		
		ctx := context.Background()
		validator, err := validator.Compile(ctx, mainSchema)
		require.NoError(t, err)
		
		// This should fail because the second cousin (with unevaluatedProperties: false)
		// cannot see the "foo" property defined in the first cousin
		testData := map[string]any{
			"foo": "bar",
		}
		
		_, err = validator.Validate(ctx, testData)
		require.Error(t, err, "cousin unevaluatedProperties should fail - cousins cannot see each other's properties")
	})
	
	t.Run("cousin unevaluatedProperties can't see inside cousins (reverse order)", func(t *testing.T) {
		// Same test but with the order reversed
		
		// Create the first cousin schema with unevaluatedProperties: false
		cousin1 := schema.NewBuilder().
			UnevaluatedProperties(schema.SchemaFalse()).
			MustBuild()
		
		// Create the second cousin schema that defines a "foo" property
		cousin2 := schema.NewBuilder().
			Property("foo", schema.NewBuilder().MustBuild()).
			MustBuild()
		
		// Create the main schema with allOf containing both cousins
		mainSchema := schema.NewBuilder().
			AllOf(cousin1, cousin2).
			MustBuild()
		
		ctx := context.Background()
		validator, err := validator.Compile(ctx, mainSchema)
		require.NoError(t, err)
		
		// This should fail because the first cousin (with unevaluatedProperties: false)
		// cannot see the "foo" property defined in the second cousin
		testData := map[string]any{
			"foo": "bar",
		}
		
		_, err = validator.Validate(ctx, testData)
		require.Error(t, err, "cousin unevaluatedProperties should fail - cousins cannot see each other's properties")
	})
}