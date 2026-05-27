package validator_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// TestReferenceIDResolution covers $ref resolution that depends on $id base-URI
// handling: URN base URIs, nested/bundled $id re-basing, and percent-encoded
// JSON Pointer fragments. These mirror cases from the JSON Schema Test Suite's
// ref.json that require an in-document $id registry rather than external
// retrieval.
func TestReferenceIDResolution(t *testing.T) {
	compile := func(t *testing.T, jsonSchema string) validator.Interface {
		t.Helper()
		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))
		ctx := schema.WithResolver(t.Context(), schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, &s)
		v, err := validator.Compile(ctx, &s)
		require.NoError(t, err)
		return v
	}

	t.Run("URN base URI with URN and JSON pointer ref", func(t *testing.T) {
		// $ref names the document by its URN $id plus a JSON pointer; it must
		// resolve to the in-document $defs/bar, not be fetched externally.
		v := compile(t, `{
			"$id": "urn:uuid:deadbeef-1234-00ff-ff00-4321feebdaed",
			"properties": {"foo": {"$ref": "urn:uuid:deadbeef-1234-00ff-ff00-4321feebdaed#/$defs/bar"}},
			"$defs": {"bar": {"type": "string"}}
		}`)

		_, err := v.Validate(t.Context(), map[string]any{"foo": "a string"})
		require.NoError(t, err)
		_, err = v.Validate(t.Context(), map[string]any{"foo": 12})
		require.Error(t, err)
	})

	t.Run("URN base URI with URN and anchor ref", func(t *testing.T) {
		v := compile(t, `{
			"$id": "urn:uuid:deadbeef-1234-ff00-00ff-4321feebdaed",
			"properties": {"foo": {"$ref": "urn:uuid:deadbeef-1234-ff00-00ff-4321feebdaed#something"}},
			"$defs": {"bar": {"$anchor": "something", "type": "string"}}
		}`)

		_, err := v.Validate(t.Context(), map[string]any{"foo": "a string"})
		require.NoError(t, err)
		_, err = v.Validate(t.Context(), map[string]any{"foo": 12})
		require.Error(t, err)
	})

	t.Run("nested $id re-bases relative local pointer", func(t *testing.T) {
		// foo is a nested resource ($id schema-relative-uri-defs2.json). Its
		// "#/$defs/inner" must resolve within foo, not the document root.
		v := compile(t, `{
			"$id": "http://example.com/schema-relative-uri-defs1.json",
			"properties": {
				"foo": {
					"$id": "schema-relative-uri-defs2.json",
					"$defs": {"inner": {"properties": {"bar": {"type": "string"}}}},
					"$ref": "#/$defs/inner"
				}
			},
			"$ref": "schema-relative-uri-defs2.json"
		}`)

		_, err := v.Validate(t.Context(), map[string]any{"foo": map[string]any{"bar": "ok"}})
		require.NoError(t, err)
		_, err = v.Validate(t.Context(), map[string]any{"foo": map[string]any{"bar": 12}})
		require.Error(t, err)
	})

	t.Run("escaped pointer ref percent-decodes the fragment", func(t *testing.T) {
		v := compile(t, `{
			"$defs": {
				"tilde~field": {"type": "integer"},
				"slash/field": {"type": "integer"},
				"percent%field": {"type": "integer"}
			},
			"properties": {
				"tilde": {"$ref": "#/$defs/tilde~0field"},
				"slash": {"$ref": "#/$defs/slash~1field"},
				"percent": {"$ref": "#/$defs/percent%25field"}
			}
		}`)

		_, err := v.Validate(t.Context(), map[string]any{"percent": 1})
		require.NoError(t, err)
		_, err = v.Validate(t.Context(), map[string]any{"percent": "not an integer"})
		require.Error(t, err)
	})
}
