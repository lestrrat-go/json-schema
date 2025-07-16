package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestDependentSchemasDebug(t *testing.T) {
	t.Run("incompatible root and dependent schema", func(t *testing.T) {
		// Test case from JSON Schema compliance test suite
		jsonSchema := `{
			"properties": {
				"foo": {}
			},
			"dependentSchemas": {
				"foo": {
					"properties": {
						"bar": {}
					},
					"additionalProperties": false
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		v, err := validator.Compile(context.Background(), &s)
		require.NoError(t, err)

		// This should FAIL because:
		// 1. Root allows "foo" property
		// 2. But dependent schema (triggered by "foo") only allows "bar" property with additionalProperties: false
		// 3. So "foo" violates the dependent schema's additionalProperties constraint
		data := map[string]any{"foo": 1}
		_, err = v.Validate(context.Background(), data)
		t.Logf("Validation result for {\"foo\": 1}: %v", err)
		require.Error(t, err, "Should fail because foo violates dependent schema's additionalProperties: false")

		// This should FAIL for the same reason
		dataBoth := map[string]any{"foo": 1, "bar": 2}
		_, err = v.Validate(context.Background(), dataBoth)
		t.Logf("Validation result for {\"foo\": 1, \"bar\": 2}: %v", err)
		require.Error(t, err, "Should fail because foo violates dependent schema's additionalProperties: false")

		// This should PASS because no "foo" property, so no dependent schema triggered
		dataBar := map[string]any{"bar": 1}
		_, err = v.Validate(context.Background(), dataBar)
		t.Logf("Validation result for {\"bar\": 1}: %v", err)
		require.NoError(t, err)

		// This should PASS because no dependency triggered
		dataBaz := map[string]any{"baz": 1}
		_, err = v.Validate(context.Background(), dataBaz)
		t.Logf("Validation result for {\"baz\": 1}: %v", err)
		require.NoError(t, err)
	})
}
