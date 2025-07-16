package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestDependentSchemasIsolation(t *testing.T) {
	t.Run("dependent schema alone", func(t *testing.T) {
		// Test just the dependent schema part
		dependentSchemaJSON := `{
			"properties": {
				"bar": {}
			},
			"additionalProperties": false
		}`

		var depSchema schema.Schema
		require.NoError(t, depSchema.UnmarshalJSON([]byte(dependentSchemaJSON)))

		v, err := validator.Compile(context.Background(), &depSchema)
		require.NoError(t, err)

		// This should FAIL - "foo" is not allowed because additionalProperties: false
		data := map[string]any{"foo": 1}
		_, err = v.Validate(context.Background(), data)
		t.Logf("Dependent schema validation result for {\"foo\": 1}: %v", err)
		require.Error(t, err, "Dependent schema should reject foo")

		// This should PASS - "bar" is allowed
		dataBar := map[string]any{"bar": 1}
		_, err = v.Validate(context.Background(), dataBar)
		t.Logf("Dependent schema validation result for {\"bar\": 1}: %v", err)
		require.NoError(t, err)
	})

	t.Run("direct dependent schemas validator", func(t *testing.T) {
		// Test the DependentSchemasValidator directly
		dependentSchemaJSON := `{
			"properties": {
				"bar": {}
			},
			"additionalProperties": false
		}`

		var depSchema schema.Schema
		require.NoError(t, depSchema.UnmarshalJSON([]byte(dependentSchemaJSON)))

		ctx := context.Background()
		depSchemas := map[string]*schema.Schema{"foo": &depSchema}

		depValidator, err := validator.DependentSchemasValidator(ctx, depSchemas)
		require.NoError(t, err)

		// This should FAIL - when "foo" is present, dependent schema rejects it
		data := map[string]any{"foo": 1}
		_, err = depValidator.Validate(context.Background(), data)
		t.Logf("DependentSchemasValidator result for {\"foo\": 1}: %v", err)
		require.Error(t, err, "Should fail because dependent schema is triggered and rejects foo")

		// This should PASS - no "foo" property, so no dependent schema triggered
		dataBar := map[string]any{"bar": 1}
		_, err = depValidator.Validate(context.Background(), dataBar)
		t.Logf("DependentSchemasValidator result for {\"bar\": 1}: %v", err)
		require.NoError(t, err)
	})
}
