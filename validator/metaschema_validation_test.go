package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// TestMetaschemaValidation tests that raw JSON data is properly validated against
// the JSON Schema metaschema to catch semantic errors that should be rejected
// according to the JSON Schema 2020-12 specification.
//
// This test uses map[string]any (raw data) instead of attempting to unmarshal
// into a Schema struct because:
//
// 1. Our Schema.UnmarshalJSON correctly rejects malformed schemas like {"type": 1}
//    at the parsing level with errors like "cannot unmarshal number into Go value of type string"
//
// 2. However, metaschema validation should happen at the SEMANTIC level, where the
//    validator validates raw JSON against metaschema rules to determine if the JSON
//    represents a valid JSON Schema according to the specification
//
// 3. The failing defs.json test expects that when validating raw data like
//    {"$defs": {"foo": {"type": 1}}} against a schema that references the metaschema,
//    the validator should detect that {"type": 1} violates the metaschema rules
//    (type must be string or array of strings, not number)
//
// 4. This separation allows the validator to validate potentially malformed schemas
//    and provide proper metaschema validation errors, rather than just parsing errors
func TestMetaschemaValidation(t *testing.T) {
	t.Run("Validate invalid type definition against metaschema", func(t *testing.T) {
		// This schema references the JSON Schema metaschema - any data validated
		// against it should conform to the JSON Schema 2020-12 specification
		metaschemaRefJSON := `{
			"$schema": "https://json-schema.org/draft/2020-12/schema",
			"$ref": "https://json-schema.org/draft/2020-12/schema"
		}`

		var metaschemaRef schema.Schema
		require.NoError(t, metaschemaRef.UnmarshalJSON([]byte(metaschemaRefJSON)))

		v, err := validator.CompileSchema(&metaschemaRef)
		require.NoError(t, err)

		// IMPORTANT: We use map[string]any here instead of trying to unmarshal {"type": 1}
		// into a Schema struct because:
		// - Schema.UnmarshalJSON would reject this at parsing time with a type error
		// - But metaschema validation should happen at validation time against raw data
		// - The validator needs to detect that "type": 1 violates JSON Schema rules
		//   (type field must be string like "string" or array like ["string", "number"])
		invalidTypeData := map[string]any{
			"type": 1, // Invalid according to JSON Schema spec: type must be string or array of strings
		}

		// Expected: validation should fail with a metaschema validation error
		// Current issue: validation passes when it should detect the spec violation
		_, err = v.Validate(context.Background(), invalidTypeData)
		if err == nil {
			t.Errorf("Expected validation to FAIL for invalid type definition, but it PASSED")
			t.Logf("Data that should be invalid according to JSON Schema spec: %+v", invalidTypeData)
		} else {
			t.Logf("Validation correctly FAILED with metaschema error: %v", err)
		}
	})

	t.Run("Validate defs.json test case against metaschema", func(t *testing.T) {
		// This reproduces the exact failing test case from the JSON Schema test suite:
		// tests/draft2020-12/defs.json -> "validate definition against metaschema" -> "invalid definition schema"
		//
		// The test validates that schemas containing $defs are themselves valid according to the metaschema.
		// When a schema definition violates the JSON Schema spec, validation should fail.
		metaschemaRefJSON := `{
			"$schema": "https://json-schema.org/draft/2020-12/schema",
			"$ref": "https://json-schema.org/draft/2020-12/schema"
		}`

		var metaschemaRef schema.Schema
		require.NoError(t, metaschemaRef.UnmarshalJSON([]byte(metaschemaRefJSON)))

		v, err := validator.CompileSchema(&metaschemaRef)
		require.NoError(t, err)

		// IMPORTANT: This uses map[string]any to represent the raw JSON data from the test suite.
		// The failing test case has data: {"$defs": {"foo": {"type": 1}}} where "type": 1 is invalid.
		// We can't unmarshal this into a Schema struct because Schema.UnmarshalJSON correctly
		// rejects {"type": 1} at parsing time. But the metaschema validation should happen at
		// the semantic validation level to catch that the nested "foo" definition violates
		// the JSON Schema specification (type must be string or array, not number).
		invalidDefsData := map[string]any{
			"$defs": map[string]any{
				"foo": map[string]any{
					"type": 1, // Invalid according to JSON Schema spec: type must be string or array of strings
				},
			},
		}

		// Expected: validation should fail because the "foo" definition violates metaschema rules
		// Current issue: validation passes when it should detect the nested schema violation
		_, err = v.Validate(context.Background(), invalidDefsData)
		if err == nil {
			t.Errorf("Expected validation to FAIL for invalid definition schema, but it PASSED")
			t.Logf("This reproduces the exact defs.json test suite failure")
			t.Logf("Data that should be invalid according to metaschema: %+v", invalidDefsData)
		} else {
			t.Logf("Validation correctly FAILED with metaschema error: %v", err)
		}
	})

	t.Run("Validate valid schema against metaschema", func(t *testing.T) {
		// Positive control test - this should pass to ensure our metaschema validation
		// works correctly for valid schemas (not just rejecting invalid ones)
		metaschemaRefJSON := `{
			"$schema": "https://json-schema.org/draft/2020-12/schema",
			"$ref": "https://json-schema.org/draft/2020-12/schema"
		}`

		var metaschemaRef schema.Schema
		require.NoError(t, metaschemaRef.UnmarshalJSON([]byte(metaschemaRefJSON)))

		v, err := validator.CompileSchema(&metaschemaRef)
		require.NoError(t, err)

		// This uses map[string]any for a valid schema that should pass metaschema validation.
		// This serves as a positive control to ensure that when we fix the metaschema validation
		// to properly reject invalid schemas, we don't break validation of valid schemas.
		validSchemaData := map[string]any{
			"type":      "string", // Valid: type as string (correct according to JSON Schema spec)
			"minLength": 1,        // Valid: minLength constraint with proper type
		}

		_, err = v.Validate(context.Background(), validSchemaData)
		if err != nil {
			t.Errorf("Expected validation to PASS for valid schema, but it FAILED: %v", err)
		} else {
			t.Logf("Validation correctly PASSED for valid schema (positive control)")
		}
	})
}