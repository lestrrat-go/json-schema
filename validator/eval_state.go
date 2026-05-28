package validator

import (
	"context"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/option/v3"
)

// evalState is the explicit per-validation working set, threaded by pointer
// through the unexported evaluate method. It is allocated fresh at each
// top-level Validate entry and never stored on a validator, so a single compiled
// validator is safe to call concurrently from multiple goroutines (go
// v.Validate(...)): each call gets its own evalState.
//
// It currently carries the runtime dynamic scope — the chain of
// $dynamicAnchor-bearing resources entered along the instance's evaluation path,
// which $dynamicRef bookending consults. Other per-call state (the
// $dynamicAnchor validator registry, dependent-schema scope, trace logger)
// still travels on the context during this transitional stage and will move
// here as the context bag is dismantled.
type evalState struct {
	dynamicScope []*schema.Schema

	// dynamicAnchorValidators maps a $dynamicAnchor name to a validator that
	// stands in for the outermost resource declaring it, populated via
	// WithDynamicAnchorValidator. It lets a precompiled validator satisfy a
	// $dynamicRef when no schema document is available at validation time.
	dynamicAnchorValidators map[string]Interface
}

// evaluator is the internal recursion contract. Every in-package validator that
// recurses into child validators implements it so the per-call evalState is
// shared down the whole evaluation path; a gap (a recursing validator that only
// implemented the public Validate) would drop the accumulated dynamic scope.
type evaluator interface {
	evaluate(ctx context.Context, v any, st *evalState) (Result, error)
}

// newEvalState builds the fresh per-call state for a top-level Validate from the
// supplied options.
func newEvalState(_ context.Context, options []ValidateOption) *evalState {
	st := &evalState{}
	for _, o := range options {
		if o.Ident() == (identDynamicAnchorValidator{}) {
			reg := option.MustGet[dynamicAnchorRegistration](o)
			if st.dynamicAnchorValidators == nil {
				st.dynamicAnchorValidators = make(map[string]Interface)
			}
			st.dynamicAnchorValidators[reg.name] = reg.v
		}
	}
	return st
}

// pushDynamicScope returns a copy of st with s appended to the dynamic scope
// chain. The slice is copied so sibling evaluations never observe each other's
// scope — the same fork semantics the old ctx-based WithDynamicScope provided.
func (st *evalState) pushDynamicScope(s *schema.Schema) *evalState {
	newScope := make([]*schema.Schema, len(st.dynamicScope)+1)
	copy(newScope, st.dynamicScope)
	newScope[len(st.dynamicScope)] = s
	return &evalState{dynamicScope: newScope, dynamicAnchorValidators: st.dynamicAnchorValidators}
}

// evalChild dispatches into a child validator, sharing st when the child is an
// in-package validator (so the dynamic scope flows down) and falling back to the
// public Validate (a fresh state) for any foreign Interface — e.g. a validator
// registered under a $dynamicAnchor that deliberately stands in for an outermost
// resource.
func evalChild(ctx context.Context, child Interface, v any, st *evalState) (Result, error) {
	// All recursive validation flows through here, so a single cancellation check
	// bounds how long a deep or wide validation runs after ctx is cancelled.
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if e, ok := child.(evaluator); ok {
		return e.evaluate(ctx, v, st)
	}
	return child.Validate(ctx, v)
}
