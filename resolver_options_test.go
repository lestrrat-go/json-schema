package schema_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

// TestResolverExternalAccessOptIn documents the core contract: NewResolver
// resolves only from memory by default, and network/filesystem access must be
// opted into explicitly via WithResolver.
func TestResolverExternalAccessOptIn(t *testing.T) {
	t.Run("network", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"$id":"https://example.com/person","type":"object"}`))
		}))
		defer server.Close()
		ref := server.URL + "/person.json"

		t.Run("declined by default", func(t *testing.T) {
			var resolved schema.Schema
			err := schema.NewResolver().ResolveReference(t.Context(), &resolved, ref, nil, "")
			require.Error(t, err, "bare resolver must not fetch over the network")
		})

		t.Run("allowed with WithResolver(HTTPResolver())", func(t *testing.T) {
			r := schema.NewResolver(schema.WithResolver(schema.HTTPResolver()))
			var resolved schema.Schema
			require.NoError(t, r.ResolveReference(t.Context(), &resolved, ref, nil, ""))
			require.True(t, resolved.ContainsType(schema.ObjectType))
		})
	})

	t.Run("filesystem", func(t *testing.T) {
		fsys := fstest.MapFS{
			"schemas/person.json": &fstest.MapFile{Data: []byte(`{"$id":"https://example.com/person","type":"string"}`)},
		}
		const ref = "schemas/person.json"

		t.Run("declined by default", func(t *testing.T) {
			var resolved schema.Schema
			err := schema.NewResolver().ResolveReference(t.Context(), &resolved, ref, nil, "")
			require.Error(t, err, "bare resolver must not read the filesystem")
		})

		t.Run("allowed with WithResolver(FSResolver(fsys))", func(t *testing.T) {
			r := schema.NewResolver(schema.WithResolver(schema.FSResolver(fsys)))
			var resolved schema.Schema
			require.NoError(t, r.ResolveReference(t.Context(), &resolved, ref, nil, ""))
			require.True(t, resolved.ContainsType(schema.StringType))
		})
	})
}
