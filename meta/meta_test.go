package meta_test

import (
	"testing"

	"github.com/lestrrat-go/json-schema/meta"
	"github.com/stretchr/testify/require"
)

// TestValidate exercises the precompiled 2020-12 meta-schema validator,
// including the "$dynamicRef": "#meta" recursion the meta-schema applies to
// every subschema. That recursion resolves back into the meta validator itself
// (it declares "$dynamicAnchor": "meta" at its root), so nested subschemas are
// validated, not just the top level.
//
// Regression: validating an ordinary object-schema document used to panic with
// a nil-pointer dereference because the "#meta" $dynamicRef had no schema
// document to resolve against at runtime and a nil base schema was dereferenced
// during anchor lookup.
func TestValidate(t *testing.T) {
	t.Run("ObjectSchemaDocument", func(t *testing.T) {
		doc := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"age":  map[string]any{"type": "integer", "minimum": 0},
			},
			"required": []any{"name"},
		}

		require.NotPanics(t, func() {
			require.NoError(t, meta.Validate(t.Context(), doc),
				"an ordinary object schema is a valid JSON Schema document")
		})
	})

	t.Run("NestedSubschemaValidated", func(t *testing.T) {
		// The inner {"type": 123} is not a valid schema ("type" must be a string
		// or array of strings). It is only reachable through "$dynamicRef":
		// "#meta", so rejecting it proves the dynamic reference recurses into the
		// meta-schema rather than degrading to accept-everything.
		doc := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"x": map[string]any{"type": float64(123)},
			},
		}

		require.Error(t, meta.Validate(t.Context(), doc),
			"a schema with a malformed nested subschema must be rejected")
	})

	t.Run("DeeplyNestedValidSchema", func(t *testing.T) {
		doc := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"items": map[string]any{
					"type":  "array",
					"items": map[string]any{"type": "string", "minLength": float64(1)},
				},
			},
		}

		require.NoError(t, meta.Validate(t.Context(), doc))
	})

	t.Run("NonSchemaRejected", func(t *testing.T) {
		require.Error(t, meta.Validate(t.Context(), "not a schema"))
	})
}
