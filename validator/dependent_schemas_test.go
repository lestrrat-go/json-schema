package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestDependentSchemas(t *testing.T) {
	t.Run("single dependency", func(t *testing.T) {
		// Schema from the JSON Schema test suite
		jsonSchema := `{
			"dependentSchemas": {
				"bar": {
					"properties": {
						"foo": {"type": "integer"},
						"bar": {"type": "integer"}
					}
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		v, err := validator.Compile(context.Background(), &s)
		require.NoError(t, err)

		// Valid case - both properties satisfy the dependent schema
		validData := map[string]any{"foo": 1, "bar": 2}
		_, err = v.Validate(context.Background(), validData)
		require.NoError(t, err)

		// Valid case - no dependency (bar property not present)
		noDependencyData := map[string]any{"foo": "quux"}
		_, err = v.Validate(context.Background(), noDependencyData)
		require.NoError(t, err)

		// Invalid case - wrong type for foo when bar is present
		wrongTypeFoo := map[string]any{"foo": "quux", "bar": 2}
		_, err = v.Validate(context.Background(), wrongTypeFoo)
		require.Error(t, err)

		// Invalid case - wrong type for bar when bar is present
		wrongTypeBar := map[string]any{"foo": 2, "bar": "quux"}
		_, err = v.Validate(context.Background(), wrongTypeBar)
		require.Error(t, err)

		// Invalid case - wrong types for both
		wrongTypeBoth := map[string]any{"foo": "quux", "bar": "quux"}
		_, err = v.Validate(context.Background(), wrongTypeBoth)
		require.Error(t, err)

		// Valid case - ignores arrays
		arrayData := []any{"bar"}
		_, err = v.Validate(context.Background(), arrayData)
		require.NoError(t, err)

		// Valid case - ignores strings
		stringData := "foobar"
		_, err = v.Validate(context.Background(), stringData)
		require.NoError(t, err)
	})

	t.Run("multiple dependencies", func(t *testing.T) {
		jsonSchema := `{
			"dependentSchemas": {
				"quux": {
					"properties": {
						"foo": {"type": "integer"},
						"bar": {"type": "integer"}
					}
				},
				"foo": {
					"properties": {
						"bar": {"type": "string"}
					}
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		v, err := validator.Compile(context.Background(), &s)
		require.NoError(t, err)

		// Invalid case - foo dependency requires bar to be string, but quux dependency requires bar to be integer
		// This should fail because both dependencies apply and they conflict
		conflictData := map[string]any{"foo": 1, "bar": 2, "quux": "baz"}
		_, err = v.Validate(context.Background(), conflictData)
		require.Error(t, err)

		// Valid case - only foo dependency applies
		onlyFooData := map[string]any{"foo": 1, "bar": "string"}
		_, err = v.Validate(context.Background(), onlyFooData)
		require.NoError(t, err)

		// Valid case - only quux dependency applies
		onlyQuuxData := map[string]any{"bar": 2, "quux": "baz"}
		_, err = v.Validate(context.Background(), onlyQuuxData)
		require.NoError(t, err)
	})

	t.Run("empty dependent schemas", func(t *testing.T) {
		jsonSchema := `{
			"dependentSchemas": {}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		v, err := validator.Compile(context.Background(), &s)
		require.NoError(t, err)

		// Any data should be valid
		data := map[string]any{"foo": "bar"}
		_, err = v.Validate(context.Background(), data)
		require.NoError(t, err)
	})

	t.Run("dependent schema with complex validation", func(t *testing.T) {
		jsonSchema := `{
			"type": "object",
			"dependentSchemas": {
				"credit_card": {
					"properties": {
						"billing_address": {"type": "string"}
					},
					"required": ["billing_address"]
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		v, err := validator.Compile(context.Background(), &s)
		require.NoError(t, err)

		// Valid case - credit_card present with required billing_address
		validData := map[string]any{
			"credit_card":     "1234-5678-9012-3456",
			"billing_address": "123 Main St",
		}
		_, err = v.Validate(context.Background(), validData)
		require.NoError(t, err)

		// Invalid case - credit_card present but missing billing_address
		invalidData := map[string]any{
			"credit_card": "1234-5678-9012-3456",
		}
		_, err = v.Validate(context.Background(), invalidData)
		require.Error(t, err)

		// Valid case - no credit_card, so no dependency
		noDependencyData := map[string]any{
			"payment_method": "cash",
		}
		_, err = v.Validate(context.Background(), noDependencyData)
		require.NoError(t, err)
	})

	t.Run("dependent schemas with references", func(t *testing.T) {
		jsonSchema := `{
			"type": "object",
			"dependentSchemas": {
				"name": {"$ref": "#/$defs/person"}
			},
			"$defs": {
				"person": {
					"type": "object",
					"properties": {
						"name": {"type": "string"},
						"age": {"type": "integer", "minimum": 0}
					},
					"required": ["name", "age"]
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, &s)

		v, err := validator.Compile(ctx, &s)
		require.NoError(t, err)

		// Valid case - name present with required age
		validData := map[string]any{
			"name": "John Doe",
			"age":  30,
		}
		_, err = v.Validate(context.Background(), validData)
		require.NoError(t, err)

		// Invalid case - name present but missing age
		invalidData := map[string]any{
			"name": "John Doe",
		}
		_, err = v.Validate(context.Background(), invalidData)
		require.Error(t, err)

		// Invalid case - name present but invalid age
		invalidAgeData := map[string]any{
			"name": "John Doe",
			"age":  -5,
		}
		_, err = v.Validate(context.Background(), invalidAgeData)
		require.Error(t, err)

		// Valid case - no name, so no dependency
		noDependencyData := map[string]any{
			"title": "Dr.",
		}
		_, err = v.Validate(context.Background(), noDependencyData)
		require.NoError(t, err)
	})
}
