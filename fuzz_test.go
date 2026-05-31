package schema_test

import (
	"encoding/json"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
)

// FuzzUnmarshalSchema feeds arbitrary bytes to Schema unmarshaling and, for any
// input that parses cleanly, marshals it back. The contract under test is "never
// panic": malformed input must surface as an error, and a schema that round-trips
// through MarshalJSON must not crash.
func FuzzUnmarshalSchema(f *testing.F) {
	f.Add([]byte(`true`))
	f.Add([]byte(`false`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"type":"object","properties":{"id":{"type":"integer"}},"required":["id"]}`))
	f.Add([]byte(`{"type":["string","null"]}`))
	f.Add([]byte(`{"$ref":"#/$defs/foo","$defs":{"foo":{"type":"string"}}}`))
	f.Add([]byte(`{"allOf":[{"minimum":0},{"maximum":10}]}`))
	f.Add([]byte(``))
	f.Add([]byte(`not-json`))

	f.Fuzz(func(_ *testing.T, data []byte) {
		var s schema.Schema
		if err := json.Unmarshal(data, &s); err != nil {
			return
		}
		// A schema that unmarshaled successfully must be re-marshalable without panicking.
		_, _ = json.Marshal(&s)
	})
}

// FuzzPrimitiveTypes targets PrimitiveTypes.UnmarshalJSON directly. It accepts a
// single type string or an array of them; the goal is to ensure no input — empty,
// scalar, or malformed — panics (see the empty-input panic fixed in #47).
func FuzzPrimitiveTypes(f *testing.F) {
	f.Add([]byte(`"string"`))
	f.Add([]byte(`["string","null"]`))
	f.Add([]byte(`[]`))
	f.Add([]byte(``))
	f.Add([]byte(`null`))
	f.Add([]byte(`123`))
	f.Add([]byte(`["string",123]`))

	f.Fuzz(func(_ *testing.T, data []byte) {
		var pt schema.PrimitiveTypes
		_ = json.Unmarshal(data, &pt)
	})
}
