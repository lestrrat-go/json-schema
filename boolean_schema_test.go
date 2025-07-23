package schema_test

import (
	"context"
	"encoding/json"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestBooleanSchemaUnmarshal(t *testing.T) {
	t.Run("oneOf with boolean values", func(t *testing.T) {
		jsonSchema := `{
			"oneOf": [
				true,
				false,
				{"type": "string"}
			]
		}`

		var s schema.Schema
		err := json.Unmarshal([]byte(jsonSchema), &s)
		require.NoError(t, err)

		require.True(t, s.HasOneOf())
		oneOf := s.OneOf()
		require.Len(t, oneOf, 3)

		// First element should be BoolSchema(true)
		schemaBool, ok := oneOf[0].(schema.BoolSchema)
		require.True(t, ok)
		require.True(t, bool(schemaBool))

		// Second element should be BoolSchema(false)
		schemaBool, ok = oneOf[1].(schema.BoolSchema)
		require.True(t, ok)
		require.False(t, bool(schemaBool))

		// Third element should be a *Schema
		schemaPtr, ok := oneOf[2].(*schema.Schema)
		require.True(t, ok)
		require.True(t, schemaPtr.ContainsType(schema.StringType))
	})

	t.Run("allOf with boolean values", func(t *testing.T) {
		jsonSchema := `{
			"allOf": [
				true,
				{"type": "number"}
			]
		}`

		var s schema.Schema
		err := json.Unmarshal([]byte(jsonSchema), &s)
		require.NoError(t, err)

		require.True(t, s.HasAllOf())
		allOf := s.AllOf()
		require.Len(t, allOf, 2)

		// First element should be BoolSchema(true)
		schemaBool, ok := allOf[0].(schema.BoolSchema)
		require.True(t, ok)
		require.True(t, bool(schemaBool))

		// Second element should be a *Schema
		schemaPtr, ok := allOf[1].(*schema.Schema)
		require.True(t, ok)
		require.True(t, schemaPtr.ContainsType(schema.NumberType))
	})

	t.Run("anyOf with boolean values", func(t *testing.T) {
		jsonSchema := `{
			"anyOf": [
				false,
				{"type": "integer"}
			]
		}`

		var s schema.Schema
		err := json.Unmarshal([]byte(jsonSchema), &s)
		require.NoError(t, err)

		require.True(t, s.HasAnyOf())
		anyOf := s.AnyOf()
		require.Len(t, anyOf, 2)

		// First element should be BoolSchema(false)
		schemaBool, ok := anyOf[0].(schema.BoolSchema)
		require.True(t, ok)
		require.False(t, bool(schemaBool))

		// Second element should be a *Schema
		schemaPtr, ok := anyOf[1].(*schema.Schema)
		require.True(t, ok)
		require.True(t, schemaPtr.ContainsType(schema.IntegerType))
	})
}

func TestBooleanSchemaValidation(t *testing.T) {
	t.Run("oneOf with true schema always passes", func(t *testing.T) {
		jsonSchema := `{
			"oneOf": [
				true
			]
		}`

		var s schema.Schema
		err := json.Unmarshal([]byte(jsonSchema), &s)
		require.NoError(t, err)

		v, err := validator.Compile(context.Background(), &s)
		require.NoError(t, err)

		// True schema should accept any value
		_, err = v.Validate(context.Background(), "hello")
		require.NoError(t, err)
		_, err = v.Validate(context.Background(), 42)
		require.NoError(t, err)
		_, err = v.Validate(context.Background(), true)
		require.NoError(t, err)
		_, err = v.Validate(context.Background(), nil)
		require.NoError(t, err)
	})

	t.Run("oneOf with false schema always fails", func(t *testing.T) {
		jsonSchema := `{
			"oneOf": [
				false
			]
		}`

		var s schema.Schema
		err := json.Unmarshal([]byte(jsonSchema), &s)
		require.NoError(t, err)

		v, err := validator.Compile(context.Background(), &s)
		require.NoError(t, err)

		// False schema should reject any value
		_, err = v.Validate(context.Background(), "hello")
		require.Error(t, err)
		_, err = v.Validate(context.Background(), 42)
		require.Error(t, err)
		_, err = v.Validate(context.Background(), true)
		require.Error(t, err)
		_, err = v.Validate(context.Background(), nil)
		require.Error(t, err)
	})

	t.Run("oneOf with mixed boolean and schema", func(t *testing.T) {
		jsonSchema := `{
			"oneOf": [
				false,
				{"type": "string"}
			]
		}`

		var s schema.Schema
		err := json.Unmarshal([]byte(jsonSchema), &s)
		require.NoError(t, err)

		v, err := validator.Compile(context.Background(), &s)
		require.NoError(t, err)

		// Should pass for strings (only the string schema passes)
		_, err = v.Validate(context.Background(), "hello")
		require.NoError(t, err)

		// Should fail for non-strings (false schema fails, string schema fails)
		_, err = v.Validate(context.Background(), 42)
		require.Error(t, err)
		_, err = v.Validate(context.Background(), true)
		require.Error(t, err)
	})

	t.Run("allOf with true schema", func(t *testing.T) {
		jsonSchema := `{
			"allOf": [
				true,
				{"type": "string"}
			]
		}`

		var s schema.Schema
		err := json.Unmarshal([]byte(jsonSchema), &s)
		require.NoError(t, err)

		v, err := validator.Compile(context.Background(), &s)
		require.NoError(t, err)

		// Should pass for strings (true schema passes, string schema passes)
		_, err = v.Validate(context.Background(), "hello")
		require.NoError(t, err)

		// Should fail for non-strings (true schema passes, but string schema fails)
		_, err = v.Validate(context.Background(), 42)
		require.Error(t, err)
	})

	t.Run("allOf with false schema", func(t *testing.T) {
		jsonSchema := `{
			"allOf": [
				false,
				{"type": "string"}
			]
		}`

		var s schema.Schema
		err := json.Unmarshal([]byte(jsonSchema), &s)
		require.NoError(t, err)

		v, err := validator.Compile(context.Background(), &s)
		require.NoError(t, err)

		// Should fail for everything (false schema always fails)
		_, err = v.Validate(context.Background(), "hello")
		require.Error(t, err)
		_, err = v.Validate(context.Background(), 42)
		require.Error(t, err)
	})
}
