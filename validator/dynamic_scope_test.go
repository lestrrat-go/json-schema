package validator_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// TestDynamicRefRuntimeScope covers $dynamicRef resolution that depends on the
// runtime dynamic scope (the resources actually entered while validating the
// instance), including the bookending requirement.
func TestDynamicRefRuntimeScope(t *testing.T) {
	compile := func(t *testing.T, jsonSchema string) validator.Interface {
		t.Helper()
		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))
		v, err := validator.Compile(t.Context(), &s, validator.WithResolver(schema.NewResolver()))
		require.NoError(t, err)
		return v
	}

	t.Run("resolves to outermost $dynamicAnchor in scope", func(t *testing.T) {
		// items' $dynamicRef "#items" must resolve to the root resource's
		// $dynamicAnchor (string), not list's placeholder, so the array must
		// contain strings.
		v := compile(t, `{
			"$id": "https://example.com/dynscope/root",
			"$ref": "list",
			"$defs": {
				"foo": {"$dynamicAnchor": "items", "type": "string"},
				"list": {
					"$id": "list",
					"type": "array",
					"items": {"$dynamicRef": "#items"},
					"$defs": {"items": {"$dynamicAnchor": "items"}}
				}
			}
		}`)

		_, err := v.Validate(t.Context(), []any{"a", "b"})
		require.NoError(t, err)
		_, err = v.Validate(t.Context(), []any{"a", 1})
		require.Error(t, err, "non-string item must fail the root's string $dynamicAnchor")
	})

	t.Run("no bookend behaves like a normal $ref", func(t *testing.T) {
		// extended has a static $anchor (not $dynamicAnchor) named meta, so the
		// $dynamicRef must NOT walk the dynamic scope to the root; baz validates
		// against extended (permissive), not root (foo const "pass").
		v := compile(t, `{
			"$id": "https://example.com/nobookend/derived",
			"$dynamicAnchor": "meta",
			"type": "object",
			"properties": {"foo": {"const": "pass"}},
			"$ref": "extended",
			"$defs": {
				"extended": {
					"$id": "extended",
					"$anchor": "meta",
					"type": "object",
					"properties": {"bar": {"$ref": "bar"}}
				},
				"bar": {
					"$id": "bar",
					"type": "object",
					"properties": {"baz": {"$dynamicRef": "extended#meta"}}
				}
			}
		}`)

		_, err := v.Validate(t.Context(), map[string]any{
			"foo": "pass",
			"bar": map[string]any{"baz": map[string]any{"foo": "anything"}},
		})
		require.NoError(t, err)
	})
}
