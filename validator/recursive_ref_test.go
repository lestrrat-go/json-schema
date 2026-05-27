package validator_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// TestRecursiveReference covers data-bounded recursive $ref: a reference that
// cycles back through a child-applying keyword terminates on the instance and
// must compile and validate, whereas a pure $ref cycle (no data consumed) is a
// compile-time error.
func TestRecursiveReference(t *testing.T) {
	compile := func(t *testing.T, jsonSchema string) (validator.Interface, error) {
		t.Helper()
		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))
		ctx := schema.WithResolver(t.Context(), schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, &s)
		return validator.Compile(ctx, &s)
	}

	t.Run("self reference via $ref to document root", func(t *testing.T) {
		// {"$ref":"#"} under properties: recursion bounded by nesting depth.
		v, err := compile(t, `{
			"type": "object",
			"required": ["name"],
			"properties": {"name": {"type": "string"}, "child": {"$ref": "#"}}
		}`)
		require.NoError(t, err)

		_, err = v.Validate(t.Context(), map[string]any{"name": "root"})
		require.NoError(t, err)
		_, err = v.Validate(t.Context(), map[string]any{
			"name":  "root",
			"child": map[string]any{"name": "kid", "child": map[string]any{"name": "grandkid"}},
		})
		require.NoError(t, err)
		// A nested child missing the required "name" must fail.
		_, err = v.Validate(t.Context(), map[string]any{
			"name":  "root",
			"child": map[string]any{"child": map[string]any{}},
		})
		require.Error(t, err)
	})

	t.Run("mutually recursive schemas via $defs", func(t *testing.T) {
		v, err := compile(t, `{
			"type": "object",
			"properties": {
				"value": {"type": "number"},
				"next": {"$ref": "#/$defs/node"}
			},
			"$defs": {
				"node": {
					"type": "object",
					"properties": {"value": {"type": "number"}, "next": {"$ref": "#"}}
				}
			}
		}`)
		require.NoError(t, err)

		_, err = v.Validate(t.Context(), map[string]any{
			"value": 1.0,
			"next":  map[string]any{"value": 2.0, "next": map[string]any{"value": 3.0}},
		})
		require.NoError(t, err)
		_, err = v.Validate(t.Context(), map[string]any{
			"value": 1.0,
			"next":  map[string]any{"value": "not a number"},
		})
		require.Error(t, err)
	})

	t.Run("pure $ref cycle is a compile-time error", func(t *testing.T) {
		_, err := compile(t, `{
			"$ref": "#/$defs/a",
			"$defs": {
				"a": {"$ref": "#/$defs/b"},
				"b": {"$ref": "#/$defs/a"}
			}
		}`)
		require.Error(t, err)
		require.Contains(t, err.Error(), "circular reference")
	})
}
