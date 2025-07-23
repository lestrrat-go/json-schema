package schemactx

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/lestrrat-go/blackmagic"
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

// Consolidated context functions using ValidationContext

// WithResolver adds a resolver to the consolidated validation context
func WithResolver(ctx context.Context, resolver any) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.Resolver = resolver
	return WithValidationContext(ctx, &newVctx)
}

// ResolverFromContext retrieves the resolver from context, returns error if not present or incompatible
func ResolverFromContext(ctx context.Context, dst any) error {
	vctx := ValidationContextFrom(ctx)
	if vctx.Resolver == nil {
		return fmt.Errorf("resolver not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, vctx.Resolver)
}

// WithRootSchema adds a root schema to the context
func WithRootSchema(ctx context.Context, rootSchema any) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.RootSchema = rootSchema
	return WithValidationContext(ctx, &newVctx)
}

// RootSchemaFromContext retrieves the root schema from context, returns error if not present or incompatible
func RootSchemaFromContext(ctx context.Context, dst any) error {
	vctx := ValidationContextFrom(ctx)
	if vctx.RootSchema == nil {
		return fmt.Errorf("root schema not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, vctx.RootSchema)
}

// WithBaseSchema adds a base schema to the context for reference resolution
func WithBaseSchema(ctx context.Context, baseSchema any) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.BaseSchema = baseSchema
	return WithValidationContext(ctx, &newVctx)
}

// BaseSchemaFromContext retrieves the base schema from context, returns error if not present or incompatible
func BaseSchemaFromContext(ctx context.Context, dst any) error {
	vctx := ValidationContextFrom(ctx)
	if vctx.BaseSchema == nil {
		return fmt.Errorf("base schema not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, vctx.BaseSchema)
}

// WithBaseURI adds a base URI to the context for reference resolution
func WithBaseURI(ctx context.Context, baseURI string) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.BaseURI = baseURI
	return WithValidationContext(ctx, &newVctx)
}

// BaseURIFromContext extracts the base URI from context, returns error if not present or incompatible
func BaseURIFromContext(ctx context.Context, dst any) error {
	vctx := ValidationContextFrom(ctx)
	if vctx.BaseURI == "" {
		return fmt.Errorf("base URI not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, vctx.BaseURI)
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

// DynamicScopeFromContext retrieves the dynamic scope chain from context, returns error if not present or incompatible
func DynamicScopeFromContext(ctx context.Context, dst any) error {
	vctx := ValidationContextFrom(ctx)
	if len(vctx.DynamicScope) == 0 {
		return fmt.Errorf("dynamic scope not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, vctx.DynamicScope)
}

// WithVocabularySet adds a vocabulary set to the context
func WithVocabularySet(ctx context.Context, vocabSet any) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.VocabularySet = vocabSet
	return WithValidationContext(ctx, &newVctx)
}

// VocabularySetFromContext retrieves the vocabulary set from context, returns error if not present or incompatible
func VocabularySetFromContext(ctx context.Context, dst any) error {
	vctx := ValidationContextFrom(ctx)
	if vctx.VocabularySet == nil {
		return fmt.Errorf("vocabulary set not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, vctx.VocabularySet)
}

// WithReferenceStack adds a reference stack to the context for circular reference detection
func WithReferenceStack(ctx context.Context, stack []string) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.ReferenceStack = stack
	return WithValidationContext(ctx, &newVctx)
}

// ReferenceStackFromContext retrieves the reference stack from context, returns error if not present or incompatible
func ReferenceStackFromContext(ctx context.Context, dst any) error {
	vctx := ValidationContextFrom(ctx)
	if len(vctx.ReferenceStack) == 0 {
		return fmt.Errorf("reference stack not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, vctx.ReferenceStack)
}

// WithEvaluationContext adds evaluation context to the context
func WithEvaluationContext(ctx context.Context, ec *EvaluationContext) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.Evaluation = ec
	return WithValidationContext(ctx, &newVctx)
}

func EvaluationContextFromContext(ctx context.Context, dst any) error {
	vctx := ValidationContextFrom(ctx)
	if vctx.Evaluation == nil {
		return fmt.Errorf("evaluated items not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, vctx.Evaluation)
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
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
