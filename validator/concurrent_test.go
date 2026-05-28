package validator_test

import (
	"sync"
	"sync/atomic"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// TestConcurrentValidate exercises the design law that a compiled validator is
// immutable config and all per-call state is decoupled from it: one validator
// must be safe to fire as `go v.Validate(...)` from many goroutines. Run with
// -race to catch any state leaking onto the receiver. The schema uses $ref and
// unevaluatedItems so the per-call dynamic scope / evaluation paths are exercised
// concurrently.
func TestConcurrentValidate(t *testing.T) {
	const src = `{
		"$id": "https://example.com/concurrent",
		"type": "object",
		"properties": {
			"name": {"$ref": "#/$defs/nonEmpty"},
			"tags": {"type": "array", "items": {"$ref": "#/$defs/nonEmpty"}, "unevaluatedItems": false}
		},
		"required": ["name"],
		"$defs": {"nonEmpty": {"type": "string", "minLength": 1}}
	}`
	var s schema.Schema
	require.NoError(t, s.UnmarshalJSON([]byte(src)))

	v, err := validator.Compile(t.Context(), &s, validator.WithResolver(schema.NewResolver()))
	require.NoError(t, err)

	valid := map[string]any{"name": "ok", "tags": []any{"a", "b"}}
	invalid := map[string]any{"tags": []any{""}} // missing required name, empty tag

	var mismatches atomic.Int64
	var wg sync.WaitGroup
	for i := range 64 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				if _, err := v.Validate(t.Context(), valid); err != nil {
					mismatches.Add(1)
				}
				return
			}
			if _, err := v.Validate(t.Context(), invalid); err == nil {
				mismatches.Add(1)
			}
		}(i)
	}
	wg.Wait()

	require.Zero(t, mismatches.Load(), "concurrent validations produced inconsistent results")
}
