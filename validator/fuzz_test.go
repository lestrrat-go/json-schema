package validator_test

import (
	"encoding/json"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// FuzzCompile feeds arbitrary bytes through the full unmarshal -> Compile path. Any
// schema that parses must compile without panicking; a compile error is acceptable
// (e.g. an unresolvable $ref), a panic is not.
func FuzzCompile(f *testing.F) {
	f.Add([]byte(`true`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"type":"object","properties":{"id":{"type":"integer"}}}`))
	f.Add([]byte(`{"type":["string","null"],"minLength":1}`))
	f.Add([]byte(`{"allOf":[{"minimum":0},{"maximum":10}]}`))
	f.Add([]byte(`{"$ref":"#/$defs/foo","$defs":{"foo":{"type":"string"}}}`))
	f.Add([]byte(`{"pattern":"^a.*z$"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var s schema.Schema
		if err := json.Unmarshal(data, &s); err != nil {
			return
		}
		_, _ = validator.Compile(t.Context(), &s)
	})
}

// FuzzValidateJSON fuzzes both halves of the pipeline at once: an arbitrary schema is
// compiled, then arbitrary data is validated against it. Neither Compile nor
// ValidateJSON may panic regardless of how malformed either input is.
func FuzzValidateJSON(f *testing.F) {
	f.Add([]byte(`{"type":"object","required":["id"]}`), []byte(`{"id":1}`))
	f.Add([]byte(`{"type":"integer","minimum":0}`), []byte(`42`))
	f.Add([]byte(`{"type":"string","pattern":"^x"}`), []byte(`"xyz"`))
	f.Add([]byte(`{"enum":["a","b"]}`), []byte(`"c"`))
	f.Add([]byte(`true`), []byte(``))
	f.Add([]byte(`{}`), []byte(`{"anything":true}`))

	f.Fuzz(func(t *testing.T, schemaData, valueData []byte) {
		var s schema.Schema
		if err := json.Unmarshal(schemaData, &s); err != nil {
			return
		}
		v, err := validator.Compile(t.Context(), &s)
		if err != nil {
			return
		}
		// Validation may pass or fail; we only require that it never panics.
		_, _ = validator.ValidateJSON(t.Context(), v, valueData)
	})
}
