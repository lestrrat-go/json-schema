package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestDependentSchemasDebug2(t *testing.T) {
	t.Run("debug schema type inference", func(t *testing.T) {
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

		// Check what types are defined
		t.Logf("Schema types: %v", s.Types())
		t.Logf("Has explicit types: %v", len(s.Types()) > 0)
		t.Logf("Has properties: %v", s.HasProperties())
		t.Logf("Has dependent schemas: %v", s.HasDependentSchemas())

		_, err := validator.CompileSchema(&s)
		require.NoError(t, err)

		// Try to test without the root properties constraint
		jsonSchemaNoProps := `{
			"type": "object",
			"dependentSchemas": {
				"foo": {
					"properties": {
						"bar": {}
					},
					"additionalProperties": false
				}
			}
		}`

		var s2 schema.Schema
		require.NoError(t, s2.UnmarshalJSON([]byte(jsonSchemaNoProps)))

		v2, err := validator.CompileSchema(&s2)
		require.NoError(t, err)

		// This should FAIL because dependent schema rejects "foo"
		data := map[string]any{"foo": 1}
		_, err = v2.Validate(context.Background(), data)
		t.Logf("Without root properties - validation result for {\"foo\": 1}: %v", err)
		require.Error(t, err, "Should fail because dependent schema rejects foo")
	})
}