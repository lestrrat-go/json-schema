package validator_test

import (
	"testing"
	"testing/fstest"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// TestRegisterFS verifies that a tree of schemas exposed as an fs.FS (here an
// in-memory fstest.MapFS, but equally an embed.FS or os.DirFS) can be preloaded
// in one call and referenced by URL, with relative refs resolving within the FS.
func TestRegisterFS(t *testing.T) {
	fsys := fstest.MapFS{
		"integer.json":       {Data: []byte(`{"type":"integer"}`)},
		"nested/foo.json":    {Data: []byte(`{"type":"object","properties":{"name":{"$ref":"string.json"}}}`)},
		"nested/string.json": {Data: []byte(`{"type":"string"}`)},
		"notschema.txt":      {Data: []byte(`ignored`)},
	}

	r := schema.NewResolver()
	require.NoError(t, r.RegisterFS("http://localhost:1234", fsys))

	var top schema.Schema
	require.NoError(t, top.UnmarshalJSON([]byte(`{"$ref":"http://localhost:1234/integer.json"}`)))
	ctx := schema.WithResolver(t.Context(), r)
	v, err := validator.Compile(ctx, &top)
	require.NoError(t, err)
	_, err = v.Validate(t.Context(), 7)
	require.NoError(t, err)
	_, err = v.Validate(t.Context(), "nope")
	require.Error(t, err)

	// Relative ref inside a registered FS document resolves within the FS.
	var nested schema.Schema
	require.NoError(t, nested.UnmarshalJSON([]byte(`{"$ref":"http://localhost:1234/nested/foo.json"}`)))
	v, err = validator.Compile(ctx, &nested)
	require.NoError(t, err)
	_, err = v.Validate(t.Context(), map[string]any{"name": "ok"})
	require.NoError(t, err)
	_, err = v.Validate(t.Context(), map[string]any{"name": 1})
	require.Error(t, err)
}
