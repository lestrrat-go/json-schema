package schema_test

import (
	"encoding/json"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// In JSON Schema a "schema" is either an object (e.g. {"type":"string"}) or a
// boolean: `true` is the schema that accepts every value, and `false` is the
// schema that rejects every value. prefixItems applies a schema to each array
// position by index, and since each entry is itself a schema, an entry may be a
// boolean as well as an object.
//
// So prefixItems: [true, {"type":"string"}, false] reads as:
//   - position 0: `true`  -> any value is allowed here
//   - position 1: object  -> the value must be a string
//   - position 2: `false` -> no value is allowed at this position
//
// These tests pin down both that such a schema parses and what it means when
// validating data.
func TestPrefixItemsBooleanSchema(t *testing.T) {
	const src = `{"prefixItems": [true, {"type": "string"}, false]}`

	t.Run("boolean entries are preserved as boolean schemas", func(t *testing.T) {
		var s schema.Schema
		require.NoError(t, json.Unmarshal([]byte(src), &s))

		items := s.PrefixItems()
		require.Len(t, items, 3)

		// A bare `true`/`false` round-trips as a BoolSchema, not a *Schema.
		require.Equal(t, schema.BoolSchema(true), items[0])
		require.IsType(t, (*schema.Schema)(nil), items[1])
		require.Equal(t, schema.BoolSchema(false), items[2])

		_, err := json.Marshal(&s)
		require.NoError(t, err)
	})

	t.Run("the boolean entries control validation per position", func(t *testing.T) {
		var s schema.Schema
		require.NoError(t, json.Unmarshal([]byte(src), &s))
		v, err := validator.Compile(t.Context(), &s)
		require.NoError(t, err)

		// position 0 is `true`, so even a non-string (42) is accepted there;
		// position 1 is a string and "hello" satisfies it; there is no third
		// element, so the `false` at position 2 never applies.
		_, err = v.Validate(t.Context(), []any{42, "hello"})
		require.NoError(t, err, "true accepts any value at position 0")

		// position 1 must be a string; 99 is not.
		_, err = v.Validate(t.Context(), []any{42, 99})
		require.Error(t, err, "the object schema at position 1 rejects a non-string")

		// position 2 is `false`: any value present there is rejected.
		_, err = v.Validate(t.Context(), []any{42, "hello", "anything"})
		require.Error(t, err, "false rejects whatever value appears at position 2")
	})
}
