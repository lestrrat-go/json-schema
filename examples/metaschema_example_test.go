package examples_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/lestrrat-go/json-schema/meta"
	"github.com/stretchr/testify/require"
)

func TestMetaSchemaValidation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		schema      string
		shouldPass  bool
		description string
	}{
		{
			name: "ValidSimpleSchema",
			schema: `{
				"type": "string",
				"minLength": 1
			}`,
			shouldPass:  true,
			description: "A simple valid JSON Schema",
		},
		{
			name: "ValidComplexSchema",
			schema: `{
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "integer", "minimum": 0}
				},
				"required": ["name"]
			}`,
			shouldPass:  true,
			description: "A complex valid JSON Schema with object properties",
		},
		{
			name:        "ValidBooleanSchema",
			schema:      `true`,
			shouldPass:  true,
			description: "A boolean schema (true means accept everything)",
		},
		{
			name:        "ValidFalseBooleanSchema",
			schema:      `false`,
			shouldPass:  true,
			description: "A boolean schema (false means reject everything)",
		},
		{
			name: "ValidArraySchema",
			schema: `{
				"type": "array",
				"items": {"type": "string"},
				"minItems": 1,
				"maxItems": 10
			}`,
			shouldPass:  true,
			description: "A valid array schema",
		},
		{
			name:        "InvalidNonObjectNonBoolean",
			schema:      `"not a schema"`,
			shouldPass:  false,
			description: "A string is not a valid JSON Schema",
		},
		{
			name:        "InvalidNumber",
			schema:      `123`,
			shouldPass:  false,
			description: "A number is not a valid JSON Schema",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the test schema as JSON to get the actual value
			var schemaValue any
			err := json.Unmarshal([]byte(tt.schema), &schemaValue)
			require.NoError(t, err, "Test schema should be valid JSON")

			// Use the pre-compiled meta-schema validator
			validator := meta.Validator()
			require.NotNil(t, validator, "Meta validator should not be nil")

			// Validate the schema against the meta-schema
			_, err = validator.Validate(ctx, schemaValue)

			if tt.shouldPass {
				require.NoError(t, err, "Schema should be valid according to meta-schema: %s", tt.description)
				t.Logf("✓ %s: Schema is valid", tt.description)
			} else {
				require.Error(t, err, "Schema should be invalid according to meta-schema: %s", tt.description)
				t.Logf("✓ %s: Schema is correctly rejected - %v", tt.description, err)
			}
		})
	}
}

func TestMetaSchemaConvenienceFunction(t *testing.T) {
	ctx := context.Background()

	// Test the convenience Validate function
	validSchema := map[string]any{
		"type":      "string",
		"minLength": 1,
	}

	err := meta.Validate(ctx, validSchema)
	require.NoError(t, err, "Valid schema should pass validation")

	// Test with invalid schema
	invalidSchema := "not a schema"
	err = meta.Validate(ctx, invalidSchema)
	require.Error(t, err, "Invalid schema should fail validation")
}

func BenchmarkMetaSchemaValidation(b *testing.B) {
	ctx := context.Background()
	validator := meta.Validator()

	// Test schema
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"age":  map[string]any{"type": "integer", "minimum": 0},
		},
		"required": []string{"name"},
	}

	b.ResetTimer()
	for range b.N {
		_, err := validator.Validate(ctx, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func ExampleValidator() {
	ctx := context.Background()

	// Get the pre-compiled meta-schema validator
	validator := meta.Validator()

	// Example: validate a simple string schema
	stringSchema := map[string]any{
		"type":      "string",
		"minLength": 1,
		"maxLength": 100,
	}

	_, err := validator.Validate(ctx, stringSchema)
	if err != nil {
		// Schema is not valid according to JSON Schema meta-schema
		panic(err)
	}

	// Schema is valid!
	fmt.Println("Schema is valid")
	// Output:
	// Schema is valid
}

func ExampleValidate() {
	ctx := context.Background()

	// Example: validate an object schema using the convenience function
	objectSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"age":  map[string]any{"type": "integer", "minimum": 0},
		},
		"required": []string{"name"},
	}

	err := meta.Validate(ctx, objectSchema)
	if err != nil {
		// Schema is not valid according to JSON Schema meta-schema
		panic(err)
	}

	// Schema is valid!
	fmt.Println("Object schema is valid")
	// Output:
	// Object schema is valid
}
