package schema

import (
	"context"

	"github.com/lestrrat-go/json-schema/internal/schemactx"
)

// Context helper functions - these delegate to internal schemactx package

// WithReferenceBase adds a base schema to the context for reference resolution
func WithReferenceBase(ctx context.Context, baseSchema *Schema) context.Context {
	return schemactx.WithReferenceBase(ctx, baseSchema)
}

// ReferenceBaseFromContext retrieves the base schema from context, returns nil if not found
func ReferenceBaseFromContext(ctx context.Context) *Schema {
	var baseSchema *Schema
	if err := schemactx.ReferenceBaseFromContext(ctx, &baseSchema); err != nil {
		return nil
	}
	return baseSchema
}

// WithResolver adds a resolver to the context
func WithResolver(ctx context.Context, resolver *Resolver) context.Context {
	return schemactx.WithResolver(ctx, resolver)
}

// ResolverFromContext retrieves the resolver from context, returns nil if not found
func ResolverFromContext(ctx context.Context) *Resolver {
	var resolver *Resolver
	if err := schemactx.ResolverFromContext(ctx, &resolver); err != nil {
		return nil
	}
	return resolver
}

// WithRootSchema adds a root schema to the context
func WithRootSchema(ctx context.Context, rootSchema *Schema) context.Context {
	return schemactx.WithRootSchema(ctx, rootSchema)
}

// RootSchemaFromContext retrieves the root schema from context, returns nil if not found
func RootSchemaFromContext(ctx context.Context) *Schema {
	var rootSchema *Schema
	if err := schemactx.RootSchemaFromContext(ctx, &rootSchema); err != nil {
		return nil
	}
	return rootSchema
}

// WithBaseURI adds a base URI to the context for reference resolution
func WithBaseURI(ctx context.Context, baseURI string) context.Context {
	return schemactx.WithBaseURI(ctx, baseURI)
}

// BaseURIFromContext extracts the base URI from context, returns empty string if not present
func BaseURIFromContext(ctx context.Context) string {
	var baseURI string
	if err := schemactx.BaseURIFromContext(ctx, &baseURI); err != nil {
		return ""
	}
	return baseURI
}

// WithDynamicScope adds a schema to the dynamic scope chain in the context
func WithDynamicScope(ctx context.Context, s *Schema) context.Context {
	return schemactx.WithDynamicScope(ctx, s)
}

// DynamicScopeFromContext retrieves the dynamic scope chain from context, returns nil if not present
func DynamicScopeFromContext(ctx context.Context) []*Schema {
	var scope []any
	if err := schemactx.DynamicScopeFromContext(ctx, &scope); err != nil {
		return nil
	}

	// Convert []any to []*Schema
	result := make([]*Schema, 0, len(scope))
	for _, s := range scope {
		if schema, ok := s.(*Schema); ok {
			result = append(result, schema)
		}
	}
	return result
}

// WithReferenceStack adds a reference stack to the context for circular reference detection
func WithReferenceStack(ctx context.Context, stack []string) context.Context {
	return schemactx.WithReferenceStack(ctx, stack)
}

// ReferenceStackFromContext retrieves the reference stack from context, returns nil if not present
func ReferenceStackFromContext(ctx context.Context) []string {
	var stack []string
	if err := schemactx.ReferenceStackFromContext(ctx, &stack); err != nil {
		return nil
	}
	return stack
}

// Context keys for validator-specific data
type dependentSchemasKey struct{}

// WithDependentSchemas adds compiled dependent schema validators to the context
func WithDependentSchemas(ctx context.Context, dependentSchemas map[string]any) context.Context {
	return context.WithValue(ctx, dependentSchemasKey{}, dependentSchemas)
}

// DependentSchemasFromContext extracts compiled dependent schema validators from context, returns nil if none are associated with ctx
func DependentSchemasFromContext(ctx context.Context) map[string]any {
	if deps, ok := ctx.Value(dependentSchemasKey{}).(map[string]any); ok {
		return deps
	}
	return nil
}
