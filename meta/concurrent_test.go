package meta_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/lestrrat-go/json-schema/meta"
	"github.com/stretchr/testify/require"
)

// TestConcurrentMetaValidate fires the single shared meta-schema validator from
// many goroutines. This is the highest-value race check: the first concurrent
// uses race to initialize ReferenceValidator.resolvedOnce and the
// DynamicReferenceValidator target cache, while the "#meta" $dynamicRefs exercise
// the per-call $dynamicAnchor registry. Run with -race.
func TestConcurrentMetaValidate(t *testing.T) {
	v := meta.Validator()
	doc := map[string]any{
		"type":       "object",
		"properties": map[string]any{"name": map[string]any{"type": "string"}},
	}

	var failures atomic.Int64
	var wg sync.WaitGroup
	for range 64 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := v.Validate(t.Context(), doc); err != nil {
				failures.Add(1)
			}
		}()
	}
	wg.Wait()

	require.Zero(t, failures.Load(), "concurrent meta validations failed")
}
