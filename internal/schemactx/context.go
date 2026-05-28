package schemactx

import (
	"context"
	"fmt"
	"log/slog"
)

type EvaluatedProperties struct {
	props map[string]struct{} // Evaluated properties
}

func (ep *EvaluatedProperties) Keys() []string {
	if ep == nil || ep.props == nil {
		return nil
	}
	keys := make([]string, 0, len(ep.props))
	for k := range ep.props {
		keys = append(keys, k)
	}
	return keys
}

func (ep *EvaluatedProperties) IsEvaluated(prop string) bool {
	if ep == nil || ep.props == nil {
		return false
	}
	_, exists := ep.props[prop]
	return exists
}

func (ep *EvaluatedProperties) MarkEvaluated(prop string) {
	if ep == nil {
		return
	}

	if ep.props == nil {
		ep.props = make(map[string]struct{})
	}
	ep.props[prop] = struct{}{}
}

type EvaluatedItems struct {
	items []bool // Evaluated items
}

func (ei *EvaluatedItems) IsEvaluated(index int) bool {
	if ei == nil || ei.items == nil || index < 0 || index >= len(ei.items) {
		return false
	}
	return ei.items[index]
}

func (ei *EvaluatedItems) Set(index int, value bool) {
	if cap(ei.items) < index+1 {
		allocated := make([]bool, index+1)
		copy(allocated, ei.items)
		ei.items = allocated
	}

	ei.items[index] = value
}

func (ei *EvaluatedItems) Copy(other *EvaluatedItems) {
	for i, v := range other.Values() {
		ei.Set(i, v)
	}
}

func (ei *EvaluatedItems) Values() []bool {
	if ei == nil || ei.items == nil {
		return nil
	}
	return ei.items
}

// EvaluationContext holds the context for validation evaluation, including
// evaluated properties and items. This context is used to track which properties
// and items have been evaluated by previous validators.
type EvaluationContext struct {
	// Properties is a map that tracks properties that have been evaluated
	// by previous validators. The key is the property name, and the value is
	// a struct{} to indicate that the property has been evaluated.
	Properties EvaluatedProperties
	// Items is a slice of booleans indicating whether each item in an array
	// has been evaluated by previous validators. The index corresponds to the item
	// position in the array, and the boolean indicates whether that item has been evaluated.
	Items EvaluatedItems // Evaluated items
}

// ValidationContext consolidates all validation-related context data into a single struct
// to reduce the proliferation of individual context helper functions
type ValidationContext struct {
	Resolver       any
	RootSchema     any
	BaseSchema     any
	BaseURI        string
	DynamicScope   []any
	VocabularySet  any
	ReferenceStack []string
	RefDepths      map[string]int // data depth at which each active reference was entered
	DataDepth      int            // number of child-applying keyword boundaries crossed during compilation
	Evaluation     *EvaluationContext
}

// Context key for the consolidated validation context
type validationContextKey struct{}

// WithValidationContext adds or updates the validation context
func WithValidationContext(ctx context.Context, vctx *ValidationContext) context.Context {
	return context.WithValue(ctx, validationContextKey{}, vctx)
}

// ValidationContextFrom retrieves the validation context from the context
// Returns a new empty ValidationContext if none exists
func ValidationContextFrom(ctx context.Context) *ValidationContext {
	if v := ctx.Value(validationContextKey{}); v != nil {
		if vctx, ok := v.(*ValidationContext); ok {
			return vctx
		}
	}
	return &ValidationContext{}
}

// valueFromContext is the shared core of every typed accessor below. It takes
// the raw field value, a flag reporting whether that field is actually present
// (each field carries its own emptiness rule — nil pointer, "" string, empty
// slice), and a human-readable name for error messages. It returns the value
// asserted to the caller-requested type T.
//
// PROGRESSION NOTE: when Go gains parameterized methods this helper, together
// with the per-field accessors, collapses into methods on *ValidationContext,
// e.g.
//
//	func (v *ValidationContext) Resolver[T any]() (T, error)
//	func (v *ValidationContext) BaseURI() (string, error)
//
// At that point a call like
//
//	schemactx.ResolverFromContext[*Resolver](ctx)
//
// becomes
//
//	schemactx.ValidationContextFrom(ctx).Resolver[*Resolver]()
//
// which is a mechanical, type-preserving change. The public schema.go wrappers
// (ResolverFromContext(ctx) *Resolver, ...) do not change in either step, so
// callers outside this package are insulated from the migration.
func valueFromContext[T any](raw any, present bool, what string) (T, error) {
	var zero T
	if !present {
		return zero, fmt.Errorf("%s not found in context", what)
	}
	v, ok := raw.(T)
	if !ok {
		return zero, fmt.Errorf("%s in context has incompatible type %T", what, raw)
	}
	return v, nil
}

// Consolidated context functions using ValidationContext

// WithResolver adds a resolver to the consolidated validation context
func WithResolver(ctx context.Context, resolver any) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.Resolver = resolver
	return WithValidationContext(ctx, &newVctx)
}

// ResolverFromContext retrieves the resolver from context, returns error if not present or incompatible.
// Generic because the concrete resolver type lives in the parent package and cannot be named here.
func ResolverFromContext[T any](ctx context.Context) (T, error) {
	vctx := ValidationContextFrom(ctx)
	return valueFromContext[T](vctx.Resolver, vctx.Resolver != nil, "resolver")
}

// WithRootSchema adds a root schema to the context
func WithRootSchema(ctx context.Context, rootSchema any) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.RootSchema = rootSchema
	return WithValidationContext(ctx, &newVctx)
}

// RootSchemaFromContext retrieves the root schema from context, returns error if not present or incompatible.
// Generic because the concrete schema type lives in the parent package and cannot be named here.
func RootSchemaFromContext[T any](ctx context.Context) (T, error) {
	vctx := ValidationContextFrom(ctx)
	return valueFromContext[T](vctx.RootSchema, vctx.RootSchema != nil, "root schema")
}

// WithBaseSchema adds a base schema to the context for reference resolution
func WithBaseSchema(ctx context.Context, baseSchema any) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.BaseSchema = baseSchema
	return WithValidationContext(ctx, &newVctx)
}

// BaseSchemaFromContext retrieves the base schema from context, returns error if not present or incompatible.
// Generic because the concrete schema type lives in the parent package and cannot be named here.
func BaseSchemaFromContext[T any](ctx context.Context) (T, error) {
	vctx := ValidationContextFrom(ctx)
	return valueFromContext[T](vctx.BaseSchema, vctx.BaseSchema != nil, "base schema")
}

// WithBaseURI adds a base URI to the context for reference resolution
func WithBaseURI(ctx context.Context, baseURI string) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.BaseURI = baseURI
	return WithValidationContext(ctx, &newVctx)
}

// BaseURIFromContext extracts the base URI from context, returns error if not present.
// Concrete return type: string is nameable here, so no type parameter is needed.
func BaseURIFromContext(ctx context.Context) (string, error) {
	vctx := ValidationContextFrom(ctx)
	return valueFromContext[string](vctx.BaseURI, vctx.BaseURI != "", "base URI")
}

// WithDynamicScope adds a schema to the dynamic scope chain in the context
func WithDynamicScope(ctx context.Context, s any) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy

	// Add new schema to chain
	newScope := make([]any, len(vctx.DynamicScope)+1)
	copy(newScope, vctx.DynamicScope)
	newScope[len(vctx.DynamicScope)] = s
	newVctx.DynamicScope = newScope

	return WithValidationContext(ctx, &newVctx)
}

// DynamicScopeFromContext retrieves the dynamic scope chain from context, returns error if not present.
// The chain is stored as []any (its elements are schemas from the parent package); callers convert
// the elements to their concrete type.
func DynamicScopeFromContext(ctx context.Context) ([]any, error) {
	vctx := ValidationContextFrom(ctx)
	return valueFromContext[[]any](vctx.DynamicScope, len(vctx.DynamicScope) > 0, "dynamic scope")
}

// WithVocabularySet adds a vocabulary set to the context
func WithVocabularySet(ctx context.Context, vocabSet any) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.VocabularySet = vocabSet
	return WithValidationContext(ctx, &newVctx)
}

// VocabularySetFromContext retrieves the vocabulary set from context, returns error if not present or incompatible.
// Generic because the concrete vocabulary type lives in the vocabulary package and cannot be named here.
func VocabularySetFromContext[T any](ctx context.Context) (T, error) {
	vctx := ValidationContextFrom(ctx)
	return valueFromContext[T](vctx.VocabularySet, vctx.VocabularySet != nil, "vocabulary set")
}

// WithReferenceStack adds a reference stack to the context for circular reference detection
func WithReferenceStack(ctx context.Context, stack []string) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.ReferenceStack = stack
	return WithValidationContext(ctx, &newVctx)
}

// ReferenceStackFromContext retrieves the reference stack from context, returns error if not present.
// Concrete return type: []string is nameable here, so no type parameter is needed.
func ReferenceStackFromContext(ctx context.Context) ([]string, error) {
	vctx := ValidationContextFrom(ctx)
	return valueFromContext[[]string](vctx.ReferenceStack, len(vctx.ReferenceStack) > 0, "reference stack")
}

// WithRefDepths stores the per-reference data-depth map used to distinguish
// pure reference cycles from data-bounded recursion.
func WithRefDepths(ctx context.Context, depths map[string]int) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.RefDepths = depths
	return WithValidationContext(ctx, &newVctx)
}

// RefDepthsFromContext returns the per-reference data-depth map (nil if absent).
func RefDepthsFromContext(ctx context.Context) map[string]int {
	return ValidationContextFrom(ctx).RefDepths
}

// WithDataDepth sets the count of child-applying keyword boundaries crossed so far.
func WithDataDepth(ctx context.Context, depth int) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.DataDepth = depth
	return WithValidationContext(ctx, &newVctx)
}

// DataDepthFromContext returns the current data depth (0 if absent).
func DataDepthFromContext(ctx context.Context) int {
	return ValidationContextFrom(ctx).DataDepth
}

// WithEvaluationContext adds evaluation context to the context
func WithEvaluationContext(ctx context.Context, ec *EvaluationContext) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.Evaluation = ec
	return WithValidationContext(ctx, &newVctx)
}

// EvaluationContextFromContext retrieves the evaluation context, returns error if not present.
// Concrete return type: *EvaluationContext is defined in this package, so no type parameter is needed.
func EvaluationContextFromContext(ctx context.Context) (*EvaluationContext, error) {
	vctx := ValidationContextFrom(ctx)
	return valueFromContext[*EvaluationContext](vctx.Evaluation, vctx.Evaluation != nil, "evaluation context")
}

// Logging context functions

// Context key for trace slog logger
type traceSlogKey struct{}

// WithTraceSlog adds a trace slog logger to the context
func WithTraceSlog(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, traceSlogKey{}, logger)
}

// TraceSlogFromContext retrieves the trace slog logger from context
// Returns a no-op logger if no logger is associated with the context
func TraceSlogFromContext(ctx context.Context) *slog.Logger {
	if v := ctx.Value(traceSlogKey{}); v != nil {
		if logger, ok := v.(*slog.Logger); ok {
			return logger
		}
	}
	// Return no-op logger that discards all output
	return slog.New(slog.DiscardHandler)
}
