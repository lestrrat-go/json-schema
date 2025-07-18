package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestUnevaluatedItems(t *testing.T) {
	t.Run("unevaluatedItems: false - with contains", func(t *testing.T) {
		// Schema with contains that evaluates items matching a pattern, and unevaluatedItems: false
		containsSchema, err := schema.NewBuilder().
			Types(schema.StringType).
			Pattern("^hello").
			Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Types(schema.ArrayType).
			Contains(containsSchema).
			UnevaluatedItems(schema.SchemaFalse()).
			Build()
		require.NoError(t, err)

		v, err := validator.Compile(context.Background(), s)
		require.NoError(t, err)

		t.Run("all items evaluated by contains - should pass", func(t *testing.T) {
			// Array where all items match the contains pattern
			_, err := v.Validate(context.Background(), []any{"hello world", "hello there"})
			require.NoError(t, err)
		})

		t.Run("unevaluated items present - should fail", func(t *testing.T) {
			// Array where some items don't match contains pattern - they become unevaluated
			_, err := v.Validate(context.Background(), []any{"hello world", "goodbye"})
			require.Error(t, err)
			require.Contains(t, err.Error(), "unevaluated item")
		})
	})

	t.Run("unevaluatedItems: true - with contains", func(t *testing.T) {
		// Schema with contains that evaluates items matching a pattern, and unevaluatedItems: true
		containsSchema, err := schema.NewBuilder().
			Types(schema.StringType).
			Pattern("^hello").
			Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Types(schema.ArrayType).
			Contains(containsSchema).
			UnevaluatedItems(schema.SchemaTrue()).
			Build()
		require.NoError(t, err)

		v, err := validator.Compile(context.Background(), s)
		require.NoError(t, err)

		t.Run("all items evaluated by contains - should pass", func(t *testing.T) {
			// Array where all items match the contains pattern
			_, err := v.Validate(context.Background(), []any{"hello world", "hello there"})
			require.NoError(t, err)
		})

		t.Run("unevaluated items present - should still pass", func(t *testing.T) {
			// Array where some items don't match contains pattern - unevaluatedItems: true allows them
			_, err := v.Validate(context.Background(), []any{"hello world", "goodbye"})
			require.NoError(t, err)
		})
	})

	t.Run("unevaluatedItems with schema - with contains", func(t *testing.T) {
		// Schema with contains that evaluates items matching a pattern, and unevaluatedItems with number schema
		containsSchema, err := schema.NewBuilder().
			Types(schema.StringType).
			Pattern("^hello").
			Build()
		require.NoError(t, err)

		unevaluatedItemsSchema, err := schema.NewBuilder().
			Types(schema.NumberType).
			Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Types(schema.ArrayType).
			Contains(containsSchema).
			UnevaluatedItems(unevaluatedItemsSchema).
			Build()
		require.NoError(t, err)

		v, err := validator.Compile(context.Background(), s)
		require.NoError(t, err)

		t.Run("all items evaluated by contains - should pass", func(t *testing.T) {
			// Array where all items match the contains pattern
			_, err := v.Validate(context.Background(), []any{"hello world", "hello there"})
			require.NoError(t, err)
		})

		t.Run("unevaluated items match unevaluatedItems schema - should pass", func(t *testing.T) {
			// Array where some items don't match contains but do match unevaluatedItems schema
			_, err := v.Validate(context.Background(), []any{"hello world", 123})
			require.NoError(t, err)
		})

		t.Run("unevaluated items don't match unevaluatedItems schema - should fail", func(t *testing.T) {
			// Array where some items don't match contains and don't match unevaluatedItems schema
			_, err := v.Validate(context.Background(), []any{"hello world", true})
			require.Error(t, err)
			require.Contains(t, err.Error(), "unevaluated item validation failed")
		})
	})
}
