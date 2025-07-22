package schemactx

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/lestrrat-go/blackmagic"
)

// EvaluatedProperties is a map that tracks properties that have been evaluated
// by previous validators. The key is the property name, and the value is
// a struct{} to indicate that the property has been evaluated.
type EvaluatedProperties map[string]struct{} // NOT map[string]bool, because the presence of the key should indicate evaluation

// EvaluatedItems is a slice of booleans indicating whether each item in an array
// has been evaluated by previous validators. The index corresponds to the item
// position in the array, and the boolean indicates whether that item has been evaluated.
type EvaluatedItems []bool // A slice of booleans indicating whether each item has been evaluated

// ValidationContext consolidates all validation-related context data into a single struct
// to reduce the proliferation of individual context helper functions
type ValidationContext struct {
	Resolver            any
	RootSchema          any
	BaseSchema          any
	BaseURI             string
	DynamicScope        []any
	VocabularySet       any
	ReferenceStack      []string
	EvaluatedProperties EvaluatedProperties
	EvaluatedItems      EvaluatedItems
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

// WithEvaluatedProperties adds evaluated properties to the context
func WithEvaluatedProperties(ctx context.Context, props EvaluatedProperties) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.EvaluatedProperties = props
	return WithValidationContext(ctx, &newVctx)
}

// EvaluatedPropertiesFromContext retrieves evaluated properties from context
func EvaluatedPropertiesFromContext(ctx context.Context, dst any) error {
	vctx := ValidationContextFrom(ctx)
	if vctx.EvaluatedProperties == nil {
		return fmt.Errorf("evaluated properties not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, vctx.EvaluatedProperties)
}

// WithEvaluatedItems adds evaluated items to the context
func WithEvaluatedItems(ctx context.Context, items EvaluatedItems) context.Context {
	vctx := ValidationContextFrom(ctx)
	newVctx := *vctx // copy
	newVctx.EvaluatedItems = items
	return WithValidationContext(ctx, &newVctx)
}

// EvaluatedItemsFromContext retrieves evaluated items from context
func EvaluatedItemsFromContext(ctx context.Context, dst any) error {
	vctx := ValidationContextFrom(ctx)
	if vctx.EvaluatedItems == nil {
		return fmt.Errorf("evaluated items not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, vctx.EvaluatedItems)
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
