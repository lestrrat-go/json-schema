package schema_test

import (
	"encoding/json"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

// TestPrefixItemsBooleanSchema guards against a regression where prefixItems
// elements were typed as *Schema and could not represent boolean schemas
// (true/false), even though JSON Schema 2020-12 allows any schema — including a
// boolean — as a prefixItems element.
func TestPrefixItemsBooleanSchema(t *testing.T) {
	const src = `{
		"prefixItems": [true, {"const": "bar"}, false]
	}`

	var s schema.Schema
	require.NoError(t, json.Unmarshal([]byte(src), &s), "boolean prefixItems must parse")

	items := s.PrefixItems()
	require.Len(t, items, 3)

	first, ok := items[0].(schema.BoolSchema)
	require.True(t, ok, "items[0] should be a BoolSchema")
	require.True(t, bool(first))

	_, ok = items[1].(*schema.Schema)
	require.True(t, ok, "items[1] should be a *Schema")

	last, ok := items[2].(schema.BoolSchema)
	require.True(t, ok, "items[2] should be a BoolSchema")
	require.False(t, bool(last))

	// Round-trips back to JSON without error.
	_, err := json.Marshal(&s)
	require.NoError(t, err)
}
