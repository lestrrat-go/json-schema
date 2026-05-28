package validator_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// TestRemoteRefViaRegisterDocument verifies that documents preloaded with
// Resolver.RegisterDocument resolve from memory (no network), including relative
// references inside a preloaded document resolving against its retrieval URI.
func TestRemoteRefViaRegisterDocument(t *testing.T) {
	mustParse := func(t *testing.T, src string) *schema.Schema {
		t.Helper()
		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(src)))
		return &s
	}

	t.Run("ref to a registered remote document", func(t *testing.T) {
		r := schema.NewResolver()
		r.RegisterDocument("http://localhost:1234/integer.json", mustParse(t, `{"type":"integer"}`))

		s := mustParse(t, `{"$ref":"http://localhost:1234/integer.json"}`)
		ctx := schema.WithResolver(t.Context(), r)
		v, err := validator.Compile(ctx, s)
		require.NoError(t, err)

		_, err = v.Validate(t.Context(), 42)
		require.NoError(t, err)
		_, err = v.Validate(t.Context(), "not an integer")
		require.Error(t, err)
	})

	t.Run("relative ref inside a remote resolves against its retrieval URI", func(t *testing.T) {
		r := schema.NewResolver()
		// foo.json lives in nested/ and references its sibling string.json.
		r.RegisterDocument("http://localhost:1234/nested/foo.json",
			mustParse(t, `{"type":"object","properties":{"name":{"$ref":"string.json"}}}`))
		r.RegisterDocument("http://localhost:1234/nested/string.json",
			mustParse(t, `{"type":"string"}`))

		s := mustParse(t, `{"$ref":"http://localhost:1234/nested/foo.json"}`)
		ctx := schema.WithResolver(t.Context(), r)
		v, err := validator.Compile(ctx, s)
		require.NoError(t, err)

		_, err = v.Validate(t.Context(), map[string]any{"name": "ok"})
		require.NoError(t, err)
		_, err = v.Validate(t.Context(), map[string]any{"name": 123})
		require.Error(t, err, "name must be a string per the sibling-resolved string.json")
	})
}
