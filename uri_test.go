package schema_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

func TestResolveURI(t *testing.T) {
	testCases := []struct {
		name string
		base string
		ref  string
		want string
	}{
		{"relative against hierarchical base", "https://example.com/a/base.json", "int.json", "https://example.com/a/int.json"},
		{"dot-relative segment", "https://example.com/a/nested/foo.json", "./bar.json", "https://example.com/a/nested/bar.json"},
		{"fragment against hierarchical base", "https://example.com/a/base.json", "#/$defs/bar", "https://example.com/a/base.json#/$defs/bar"},
		{"fragment against URN (opaque)", "urn:uuid:deadbeef", "#/$defs/bar", "urn:uuid:deadbeef#/$defs/bar"},
		{"absolute ref ignores base", "urn:uuid:deadbeef", "urn:uuid:deadbeef#/$defs/bar", "urn:uuid:deadbeef#/$defs/bar"},
		{"empty ref returns base", "https://example.com/a", "", "https://example.com/a"},
		{"empty base returns ref", "", "int.json", "int.json"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, schema.ResolveURI(tc.base, tc.ref))
		})
	}
}
