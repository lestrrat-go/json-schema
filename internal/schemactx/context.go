package schemactx

import (
	"context"
	"fmt"

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

// Context key types for passing data through context
type resolverKey struct{}
type rootSchemaKey struct{}
type baseSchemaKey struct{}
type baseURIKey struct{}
type dynamicScopeKey struct{}
type vocabularyKey struct{}
type referenceStackKey struct{}
type evaluatedPropertiesKey struct{}
type evaluatedItemsKey struct{}

// WithResolver adds a resolver to the context
func WithResolver(ctx context.Context, resolver any) context.Context {
	return context.WithValue(ctx, resolverKey{}, resolver)
}

// ResolverFromContext retrieves the resolver from context, returns error if not present or incompatible
func ResolverFromContext(ctx context.Context, dst any) error {
	v := ctx.Value(resolverKey{})
	if v == nil {
		return fmt.Errorf("resolver not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, v)
}

// WithRootSchema adds a root schema to the context
func WithRootSchema(ctx context.Context, rootSchema any) context.Context {
	return context.WithValue(ctx, rootSchemaKey{}, rootSchema)
}

// RootSchemaFromContext retrieves the root schema from context, returns error if not present or incompatible
func RootSchemaFromContext(ctx context.Context, dst any) error {
	v := ctx.Value(rootSchemaKey{})
	if v == nil {
		return fmt.Errorf("root schema not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, v)
}

// WithBaseSchema adds a base schema to the context for reference resolution
func WithBaseSchema(ctx context.Context, baseSchema any) context.Context {
	return context.WithValue(ctx, baseSchemaKey{}, baseSchema)
}

// BaseSchemaFromContext retrieves the base schema from context, returns error if not present or incompatible
func BaseSchemaFromContext(ctx context.Context, dst any) error {
	v := ctx.Value(baseSchemaKey{})
	if v == nil {
		return fmt.Errorf("base schema not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, v)
}

// WithBaseURI adds a base URI to the context for reference resolution
func WithBaseURI(ctx context.Context, baseURI string) context.Context {
	return context.WithValue(ctx, baseURIKey{}, baseURI)
}

// BaseURIFromContext extracts the base URI from context, returns error if not present or incompatible
func BaseURIFromContext(ctx context.Context, dst any) error {
	v := ctx.Value(baseURIKey{})
	if v == nil {
		return fmt.Errorf("base URI not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, v)
}

// WithDynamicScope adds a schema to the dynamic scope chain in the context
func WithDynamicScope(ctx context.Context, s any) context.Context {
	// Get current scope chain from context
	var currentScope []any
	if existing := ctx.Value(dynamicScopeKey{}); existing != nil {
		if scope, ok := existing.([]any); ok {
			currentScope = scope
		}
	}

	// Add new schema to chain
	newScope := make([]any, len(currentScope)+1)
	copy(newScope, currentScope)
	newScope[len(currentScope)] = s

	return context.WithValue(ctx, dynamicScopeKey{}, newScope)
}

// DynamicScopeFromContext retrieves the dynamic scope chain from context, returns error if not present or incompatible
func DynamicScopeFromContext(ctx context.Context, dst any) error {
	v := ctx.Value(dynamicScopeKey{})
	if v == nil {
		return fmt.Errorf("dynamic scope not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, v)
}

// WithVocabularySet adds a vocabulary set to the context
func WithVocabularySet(ctx context.Context, vocabSet any) context.Context {
	return context.WithValue(ctx, vocabularyKey{}, vocabSet)
}

// VocabularySetFromContext retrieves the vocabulary set from context, returns error if not present or incompatible
func VocabularySetFromContext(ctx context.Context, dst any) error {
	v := ctx.Value(vocabularyKey{})
	if v == nil {
		return fmt.Errorf("vocabulary set not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, v)
}

// WithReferenceStack adds a reference stack to the context for circular reference detection
func WithReferenceStack(ctx context.Context, stack []string) context.Context {
	return context.WithValue(ctx, referenceStackKey{}, stack)
}

// ReferenceStackFromContext retrieves the reference stack from context, returns error if not present or incompatible
func ReferenceStackFromContext(ctx context.Context, dst any) error {
	v := ctx.Value(referenceStackKey{})
	if v == nil {
		return fmt.Errorf("reference stack not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, v)
}

func WithEvaluatedProperties(ctx context.Context, props EvaluatedProperties) context.Context {
	return context.WithValue(ctx, evaluatedPropertiesKey{}, props)
}

func EvaluatedPropertiesFromContext(ctx context.Context, dst any) error {
	v := ctx.Value(evaluatedPropertiesKey{})
	if v == nil {
		return fmt.Errorf("evaluated properties not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, v)
}

func WithEvaluatedItems(ctx context.Context, items EvaluatedItems) context.Context {
	return context.WithValue(ctx, evaluatedItemsKey{}, items)
}

func EvaluatedItemsFromContext(ctx context.Context, dst any) error {
	v := ctx.Value(evaluatedItemsKey{})
	if v == nil {
		return fmt.Errorf("evaluated items not found in context")
	}
	return blackmagic.AssignIfCompatible(dst, v)
}
