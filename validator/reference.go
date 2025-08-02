package validator // ReferenceValidator handles schema references ($ref) with lazy resolution and circular detection
import (
	"context"
	"fmt"
	"strings"
	"sync"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/internal/pool"
)

var resolverPool = pool.New[*schema.Resolver](
	func() *schema.Resolver { return schema.NewResolver() },
	func(r *schema.Resolver) *schema.Resolver { return r }, // Resolvers are stateless, no cleanup needed
)

type ReferenceValidator struct {
	reference    string
	resolvedOnce sync.Once
	resolved     Interface
	resolveErr   error
	rootSchema   *schema.Schema
}

func (r *ReferenceValidator) Validate(ctx context.Context, v any) (Result, error) {
	// Lazy resolution - only resolve when actually needed for validation
	r.resolvedOnce.Do(func() {
		r.resolved, r.resolveErr = r.resolveReference(ctx)
	})

	if r.resolveErr != nil {
		return nil, fmt.Errorf("reference resolution failed for %s: %w", r.reference, r.resolveErr)
	}

	return r.resolved.Validate(ctx, v)
}

func (r *ReferenceValidator) resolveReference(ctx context.Context) (Interface, error) {
	// Get a resolver from the pool
	resolver := resolverPool.Get()
	defer resolverPool.Put(resolver)

	// Use the validator's stored root schema
	rootSchema := r.rootSchema
	if rootSchema == nil {
		return nil, fmt.Errorf("no root schema available for reference resolution: %s", r.reference)
	}

	// Check for circular references by looking at context
	if stack := schema.ReferenceStackFromContext(ctx); stack != nil {
		for _, ref := range stack {
			if ref == r.reference {
				return nil, fmt.Errorf("circular reference detected: %s", r.reference)
			}
		}
		// Add current reference to stack
		newStack := make([]string, len(stack)+1)
		copy(newStack, stack)
		newStack[len(stack)] = r.reference
		ctx = schema.WithReferenceStack(ctx, newStack)
	} else {
		// Start new reference stack
		ctx = schema.WithReferenceStack(ctx, []string{r.reference})
	}

	// Resolve the reference to get the target schema
	var targetSchema schema.Schema
	baseURI := schema.BaseURIFromContext(ctx)
	refCtx := ctx
	if baseURI != "" {
		refCtx = schema.WithBaseURI(ctx, baseURI)
	}
	// Add base schema context for reference resolution
	if rootSchema := schema.RootSchemaFromContext(ctx); rootSchema != nil {
		refCtx = schema.WithBaseSchema(refCtx, rootSchema)
	}
	if err := resolver.ResolveReference(refCtx, &targetSchema, r.reference); err != nil {
		return nil, fmt.Errorf("failed to resolve reference %s: %w", r.reference, err)
	}

	// Compile the resolved schema into a validator
	// IMPORTANT: Keep the original root schema context to ensure nested references can be resolved
	return Compile(ctx, &targetSchema)
}

// DynamicReferenceValidator handles $dynamicRef with proper dynamic scope resolution
type DynamicReferenceValidator struct {
	reference    string
	resolvedOnce sync.Once
	resolved     Interface
	resolveErr   error
	rootSchema   *schema.Schema
	dynamicScope []*schema.Schema // Store the dynamic scope chain from compilation
}

// NewDynamicReferenceValidator creates a new DynamicReferenceValidator for the given reference
func NewDynamicReferenceValidator(reference string) *DynamicReferenceValidator {
	return &DynamicReferenceValidator{
		reference: reference,
	}
}

func (dr *DynamicReferenceValidator) Validate(ctx context.Context, v any) (Result, error) {
	// Lazy resolution - only resolve when actually needed for validation
	dr.resolvedOnce.Do(func() {
		dr.resolved, dr.resolveErr = dr.resolveDynamicReference(ctx)
	})

	if dr.resolveErr != nil {
		return nil, fmt.Errorf("dynamic reference resolution failed for %s: %w", dr.reference, dr.resolveErr)
	}

	return dr.resolved.Validate(ctx, v)
}

func (dr *DynamicReferenceValidator) resolveDynamicReference(ctx context.Context) (Interface, error) {
	// Get a resolver from the pool
	resolver := resolverPool.Get()
	defer resolverPool.Put(resolver)

	// Use the validator's stored root schema
	rootSchema := dr.rootSchema
	if rootSchema == nil {
		return nil, fmt.Errorf("no root schema available for dynamic reference resolution: %s", dr.reference)
	}

	// Create context with stored dynamic scope chain from compilation time
	ctxWithScope := ctx
	if dr.dynamicScope != nil {
		// Build context with all scope elements at once to avoid nested context in loop
		//nolint:fatcontext // Intentional: building dynamic scope chain requires nested contexts
		for _, scope := range dr.dynamicScope {
			ctxWithScope = schema.WithDynamicScope(ctxWithScope, scope)
		}
	}

	// Check for circular references by looking at context
	if stack := schema.ReferenceStackFromContext(ctxWithScope); stack != nil {
		for _, ref := range stack {
			if ref == dr.reference {
				return nil, fmt.Errorf("circular reference detected: %s", dr.reference)
			}
		}
		// Add current reference to stack
		newStack := make([]string, len(stack)+1)
		copy(newStack, stack)
		newStack[len(stack)] = dr.reference
		ctxWithScope = schema.WithReferenceStack(ctxWithScope, newStack)
	} else {
		// Start new reference stack
		ctxWithScope = schema.WithReferenceStack(ctxWithScope, []string{dr.reference})
	}

	// Resolve the $dynamicRef using stored dynamic scope chain
	targetSchema, err := resolveDynamicRef(ctxWithScope, resolver, rootSchema, dr.reference)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dynamic reference %s: %w", dr.reference, err)
	}

	// If the target schema has relative references, we need to ensure they're resolved
	// against the correct base URI. For metaschema, this is crucial.
	if targetSchema.Has(schema.IDField) && targetSchema.ID() != "" {
		// Set the base URI from the target schema's $id
		if baseURI := extractBaseURI(targetSchema.ID()); baseURI != "" {
			ctxWithScope = schema.WithBaseURI(ctxWithScope, baseURI)
		}
	}

	// Compile the resolved target schema
	return Compile(ctxWithScope, targetSchema)
}

// extractBaseURI extracts the base URI from a reference for context resolution
func extractBaseURI(reference string) string {
	// Handle absolute URIs
	if strings.HasPrefix(reference, "http://") || strings.HasPrefix(reference, "https://") {
		// Split on '#' to get the URI part without fragment
		parts := strings.Split(reference, "#")
		uri := parts[0]

		// Find the last '/' to get the directory path
		lastSlash := strings.LastIndex(uri, "/")
		if lastSlash != -1 {
			return uri[:lastSlash+1] // Include the trailing slash
		}
		return uri + "/" // Add trailing slash if not present
	}

	// For relative references, we can't determine base URI without context
	return ""
}

// resolveDynamicRef resolves a $dynamicRef by looking up the dynamic scope chain
// for the nearest schema with a matching $dynamicAnchor
func resolveDynamicRef(ctx context.Context, resolver *schema.Resolver, rootSchema *schema.Schema, dynamicRef string) (*schema.Schema, error) {
	// Parse the dynamic reference - it should be in the form "#anchorName"
	if !strings.HasPrefix(dynamicRef, "#") {
		// For non-anchor dynamic refs, treat as normal reference
		var targetSchema schema.Schema
		baseURI := schema.BaseURIFromContext(ctx)
		refCtx := schema.WithBaseSchema(ctx, rootSchema)
		if baseURI != "" {
			refCtx = schema.WithBaseURI(refCtx, baseURI)
		}
		if err := resolver.ResolveReference(refCtx, &targetSchema, dynamicRef); err != nil {
			return nil, fmt.Errorf("failed to resolve dynamic reference %s: %w", dynamicRef, err)
		}
		return &targetSchema, nil
	}

	// Check if this is a JSON pointer reference (starts with #/)
	if strings.HasPrefix(dynamicRef, "#/") {
		// For JSON pointer references, try dynamic anchor lookup first, then fall back to normal reference
		// Get the dynamic scope chain from context
		scopeChain := schema.DynamicScopeFromContext(ctx)

		// Search the dynamic scope chain from oldest to most recent for matching $dynamicAnchor
		anchorName := dynamicRef[1:] // Remove the '#' prefix
		for i := range scopeChain {
			currentSchema := scopeChain[i]

			// Check if this schema has a matching $dynamicAnchor
			if currentSchema.Has(schema.DynamicAnchorField) && currentSchema.DynamicAnchor() == anchorName {
				return currentSchema, nil
			}
		}

		// No matching $dynamicAnchor found, fall back to normal JSON pointer resolution
		var targetSchema schema.Schema
		refCtx := schema.WithBaseSchema(ctx, rootSchema)
		if err := resolver.ResolveReference(refCtx, &targetSchema, dynamicRef); err != nil {
			return nil, fmt.Errorf("failed to resolve dynamic reference %s: %w", dynamicRef, err)
		}
		return &targetSchema, nil
	}

	// This is a plain anchor reference (e.g., "#anchorName")
	anchorName := dynamicRef[1:] // Remove the '#' prefix

	// Get the dynamic scope chain from context
	scopeChain := schema.DynamicScopeFromContext(ctx)

	// Search the dynamic scope chain from oldest to most recent
	// $dynamicRef should resolve to the nearest enclosing schema with matching $dynamicAnchor
	// "Nearest enclosing" means closest to the root, not most recently added
	for i := range scopeChain {
		currentSchema := scopeChain[i]

		// Check if this schema has a matching $dynamicAnchor
		if currentSchema.Has(schema.DynamicAnchorField) && currentSchema.DynamicAnchor() == anchorName {
			return currentSchema, nil
		}
	}

	// If no matching $dynamicAnchor found in dynamic scope, fall back to normal anchor resolution
	// This is the correct behavior according to JSON Schema spec
	var targetSchema schema.Schema
	refCtx := schema.WithBaseSchema(ctx, rootSchema)
	if err := resolver.ResolveAnchor(refCtx, &targetSchema, anchorName); err != nil {
		return nil, fmt.Errorf("failed to resolve dynamic reference %s (no matching $dynamicAnchor in scope): %w", dynamicRef, err)
	}

	return &targetSchema, nil
}
