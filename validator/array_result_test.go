package validator_test

import (
	"testing"

	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// ArrayResult tracks which array indices were evaluated, for unevaluatedItems
// accounting. SetEvaluatedItem marks one index; SetEvaluatedItems replaces the
// whole annotation slice in one call (e.g. when a caller has already computed
// the full evaluated-items vector).
func TestArrayResultEvaluatedItems(t *testing.T) {
	t.Run("SetEvaluatedItem marks individual indices and grows the slice", func(t *testing.T) {
		r := validator.NewArrayResult()
		r.SetEvaluatedItem(2)
		require.Equal(t, []bool{false, false, true}, r.EvaluatedItems())
	})

	t.Run("SetEvaluatedItems replaces the entire slice", func(t *testing.T) {
		r := validator.NewArrayResult()
		r.SetEvaluatedItem(0)
		r.SetEvaluatedItems([]bool{false, true, true})
		require.Equal(t, []bool{false, true, true}, r.EvaluatedItems())
	})

	t.Run("EvaluatedItems returns a copy, not the backing slice", func(t *testing.T) {
		r := validator.NewArrayResult()
		r.SetEvaluatedItems([]bool{true, true})
		got := r.EvaluatedItems()
		got[0] = false
		require.Equal(t, []bool{true, true}, r.EvaluatedItems(), "mutating the returned slice must not affect the result")
	})
}
