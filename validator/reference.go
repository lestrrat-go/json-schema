package validator // ReferenceValidator handles schema references ($ref) with lazy resolution and circular detection
import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	schema "github.com/lestrrat-go/json-schema"
)

type ReferenceValidator struct {
	reference    string
	resolvedOnce sync.Once
	resolved     Interface
	resolveErr   error
	resolver     *schema.Resolver
	rootSchema   *schema.Schema
	baseSchema   *schema.Schema // Enclosing resource captured at compile time (nil = use root)
	baseURI      string         // Enclosing resource's base URI captured at compile time
}

func (r *ReferenceValidator) Validate(ctx context.Context, v any) (Result, error) {
	return r.evaluate(ctx, v, newEvalState(ctx))
}

func (r *ReferenceValidator) evaluate(ctx context.Context, v any, st *evalState) (Result, error) {
	// Lazy resolution - only resolve when actually needed for validation
	r.resolvedOnce.Do(func() {
		r.resolved, r.resolveErr = r.resolveReference(ctx)
	})

	if r.resolveErr != nil {
		return nil, fmt.Errorf("reference resolution failed for %s: %w", r.reference, r.resolveErr)
	}

	return evalChild(ctx, r.resolved, v, st)
}

func (r *ReferenceValidator) resolveReference(ctx context.Context) (Interface, error) {
	// Use stored resolver and root schema, fall back to context if not available
	resolver := r.resolver
	if resolver == nil {
		resolver = schema.ResolverFromContext(ctx)
		if resolver == nil {
			resolver = schema.NewResolver()
		}
	}

	rootSchema := r.rootSchema
	if rootSchema == nil {
		rootSchema = schema.RootSchemaFromContext(ctx)
		if rootSchema == nil {
			return nil, fmt.Errorf("no root schema available in context for reference resolution: %s", r.reference)
		}
	}

	// Check for circular references by looking at context
	if stack := schema.ReferenceStackFromContext(ctx); stack != nil {
		if slices.Contains(stack, r.reference) {
			return nil, fmt.Errorf("circular reference detected: %s", r.reference)
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

	// Resolve the reference against the enclosing resource captured at compile
	// time (falling back to context/root), so deferred recursive references in
	// a nested $id resource still resolve within that resource.
	var targetSchema schema.Schema
	baseURI := r.baseURI
	if baseURI == "" {
		baseURI = schema.BaseURIFromContext(ctx)
	}
	refCtx := ctx
	if baseURI != "" {
		refCtx = schema.WithBaseURI(ctx, baseURI)
	}
	baseSchema := r.baseSchema
	if baseSchema == nil {
		baseSchema = schema.BaseSchemaFromContext(ctx)
	}
	if baseSchema == nil {
		baseSchema = rootSchema
	}
	if baseSchema != nil {
		refCtx = schema.WithBaseSchema(refCtx, baseSchema)
	}
	if err := resolver.ResolveReference(refCtx, &targetSchema, r.reference); err != nil {
		return nil, fmt.Errorf("failed to resolve reference %s: %w", r.reference, err)
	}

	// Compile the resolved schema. Seed the resolver (carrying the in-document
	// $id registry), root schema, and base schema/URI: this runs at validation
	// time, when the incoming context generally lacks them, and without the
	// resolver nested references would fall back to external retrieval.
	compileCtx := schema.WithResolver(ctx, resolver)
	compileCtx = schema.WithRootSchema(compileCtx, rootSchema)
	if baseURI != "" {
		compileCtx = schema.WithBaseURI(compileCtx, baseURI)
	}
	if baseSchema != nil {
		compileCtx = schema.WithBaseSchema(compileCtx, baseSchema)
	}
	compiled, err := Compile(compileCtx, &targetSchema)
	if err != nil {
		return nil, err
	}

	// Following a $ref into another resource enters that resource's dynamic
	// scope, even when the reference targets a subschema within it. Push the
	// enclosing resource so $dynamicRef bookending sees it.
	if resource := resolver.ResourceFor(schema.ResolveURI(baseURI, r.reference)); resource != nil && resource != &targetSchema {
		compiled = &dynamicScopeValidator{schema: resource, inner: compiled}
	}
	return compiled, nil
}

// dynamicScopeValidator pushes its schema resource onto the runtime dynamic
// scope before delegating to the wrapped validator. Validation thus accumulates
// the chain of $dynamicAnchor-bearing resources actually entered along the
// instance's evaluation path, which is what $dynamicRef bookending consults.
type dynamicScopeValidator struct {
	schema *schema.Schema
	inner  Interface
}

func (d *dynamicScopeValidator) Validate(ctx context.Context, v any) (Result, error) {
	return d.evaluate(ctx, v, newEvalState(ctx))
}

func (d *dynamicScopeValidator) evaluate(ctx context.Context, v any, st *evalState) (Result, error) {
	return evalChild(ctx, d.inner, v, st.pushDynamicScope(d.schema))
}

// DynamicReferenceValidator handles $dynamicRef. Unlike $ref, a $dynamicRef can
// resolve to different targets on different validations depending on the runtime
// dynamic scope, so resolution happens per-Validate (not memoized once).
type DynamicReferenceValidator struct {
	reference  string
	resolver   *schema.Resolver
	rootSchema *schema.Schema
	baseSchema *schema.Schema // Enclosing resource for non-dynamic fallback resolution
	baseURI    string         // Enclosing resource's base URI

	mu    sync.Mutex
	cache map[*schema.Schema]Interface // compiled validators keyed by resolved target
}

// NewDynamicReferenceValidator creates a new DynamicReferenceValidator for the given reference
func NewDynamicReferenceValidator(reference string) *DynamicReferenceValidator {
	return &DynamicReferenceValidator{
		reference: reference,
	}
}

func (dr *DynamicReferenceValidator) Validate(ctx context.Context, v any) (Result, error) {
	return dr.evaluate(ctx, v, newEvalState(ctx))
}

func (dr *DynamicReferenceValidator) evaluate(ctx context.Context, v any, st *evalState) (Result, error) {
	// When the fragment is a plain $dynamicAnchor name and a validator has been
	// registered for it in the context, the registered validator stands in for
	// the outermost dynamic-scope resource declaring that anchor. This is how the
	// precompiled meta-schema validator satisfies "$dynamicRef": "#meta" — it
	// registers itself under "meta" and recurses, since no schema document is
	// available to resolve against at validation time. The registered validator
	// represents an outermost resource, so it re-enters with fresh state.
	if name := plainAnchorFragment(dr.reference); name != "" {
		if rv, ok := schema.DynamicAnchorValidatorFromContext(ctx, name).(Interface); ok && rv != nil {
			return rv.Validate(ctx, v)
		}
	}

	target, err := dr.resolveTarget(ctx, st)
	if err != nil {
		return nil, fmt.Errorf("dynamic reference resolution failed for %s: %w", dr.reference, err)
	}
	validator, err := dr.validatorFor(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("dynamic reference resolution failed for %s: %w", dr.reference, err)
	}
	return evalChild(ctx, validator, v, st)
}

// resolveTarget resolves the $dynamicRef against the current runtime dynamic
// scope carried in st.
func (dr *DynamicReferenceValidator) resolveTarget(ctx context.Context, st *evalState) (*schema.Schema, error) {
	resolver := dr.resolver
	if resolver == nil {
		if resolver = schema.ResolverFromContext(ctx); resolver == nil {
			resolver = schema.NewResolver()
		}
	}
	baseSchema := dr.baseSchema
	if baseSchema == nil {
		baseSchema = schema.BaseSchemaFromContext(ctx)
	}
	if baseSchema == nil {
		baseSchema = dr.rootSchema
	}
	if baseSchema == nil {
		baseSchema = schema.RootSchemaFromContext(ctx)
	}

	refCtx := ctx
	if dr.baseURI != "" {
		refCtx = schema.WithBaseURI(refCtx, dr.baseURI)
	}
	if baseSchema != nil {
		refCtx = schema.WithBaseSchema(refCtx, baseSchema)
	}
	return resolveDynamicRef(refCtx, resolver, baseSchema, dr.reference, st.dynamicScope)
}

// validatorFor compiles (and caches) the validator for a resolved target schema.
func (dr *DynamicReferenceValidator) validatorFor(ctx context.Context, target *schema.Schema) (Interface, error) {
	dr.mu.Lock()
	if dr.cache == nil {
		dr.cache = make(map[*schema.Schema]Interface)
	}
	if v, ok := dr.cache[target]; ok {
		dr.mu.Unlock()
		return v, nil
	}
	dr.mu.Unlock()

	resolver := dr.resolver
	if resolver == nil {
		resolver = schema.ResolverFromContext(ctx)
	}
	root := dr.rootSchema
	if root == nil {
		root = schema.RootSchemaFromContext(ctx)
	}
	compileCtx := ctx
	if resolver != nil {
		compileCtx = schema.WithResolver(compileCtx, resolver)
	}
	if root != nil {
		compileCtx = schema.WithRootSchema(compileCtx, root)
	}
	if target.HasID() && target.ID() != "" {
		// Resolve the target's (possibly relative) $id against the base URI under
		// which the $dynamicRef itself was resolved, so the target's own relative
		// references resolve within its resource.
		parentBase := dr.baseURI
		if parentBase == "" {
			parentBase = schema.BaseURIFromContext(ctx)
		}
		if base := schema.ResolveURI(parentBase, target.ID()); base != "" {
			compileCtx = schema.WithBaseURI(compileCtx, base)
		}
		compileCtx = schema.WithBaseSchema(compileCtx, target)
	}

	v, err := Compile(compileCtx, target)
	if err != nil {
		return nil, fmt.Errorf("failed to compile dynamic reference target %s: %w", dr.reference, err)
	}

	dr.mu.Lock()
	dr.cache[target] = v
	dr.mu.Unlock()
	return v, nil
}

// plainAnchorFragment returns the fragment of ref when it is a plain anchor name
// (e.g. "meta" for "#meta" or "extended#meta"), or "" when the reference has no
// fragment or the fragment is a JSON pointer.
func plainAnchorFragment(ref string) string {
	_, frag, found := strings.Cut(ref, "#")
	if !found || frag == "" || strings.HasPrefix(frag, "/") {
		return ""
	}
	return frag
}

// resolveDynamicRef resolves a $dynamicRef. It first resolves the reference the
// way $ref would (the "lexical" target). When the reference's fragment is a
// plain anchor (e.g. "#meta" or "extended#meta") and that lexical target itself
// declares a $dynamicAnchor of the same name — the bookending requirement — the
// reference instead resolves to the same $dynamicAnchor in the FIRST (outermost)
// resource of the runtime dynamic scope. Otherwise it behaves exactly like $ref.
func resolveDynamicRef(ctx context.Context, resolver *schema.Resolver, baseSchema *schema.Schema, dynamicRef string, scopeChain []*schema.Schema) (*schema.Schema, error) {
	baseURI := schema.BaseURIFromContext(ctx)
	// Only seed the base schema when one is actually available. A nil *Schema
	// boxed into the context's any-typed slot would defeat the presence check in
	// BaseSchemaFromContext and get dereferenced during anchor lookup.
	refCtx := ctx
	if baseSchema != nil {
		refCtx = schema.WithBaseSchema(refCtx, baseSchema)
	}
	if baseURI != "" {
		refCtx = schema.WithBaseURI(refCtx, baseURI)
	}

	// Determine whether the fragment is a plain anchor name (eligible for
	// dynamic-scope bookending) versus a JSON pointer or no fragment at all.
	anchorName := plainAnchorFragment(dynamicRef)

	// Resolve the lexical target as $ref would.
	var lexical schema.Schema
	var lexErr error
	if dynamicRef == "#"+anchorName && anchorName != "" {
		lexErr = resolver.ResolveAnchor(refCtx, &lexical, anchorName)
	} else {
		lexErr = resolver.ResolveReference(refCtx, &lexical, dynamicRef)
	}

	// Bookending: only consult the dynamic scope when the lexical target declares
	// a $dynamicAnchor matching the fragment.
	if anchorName != "" && lexErr == nil && lexical.HasDynamicAnchor() && lexical.DynamicAnchor() == anchorName {
		for i := range scopeChain {
			if found := schema.FindDynamicAnchor(scopeChain[i], anchorName); found != nil {
				return found, nil
			}
		}
	}

	if lexErr != nil {
		return nil, fmt.Errorf("failed to resolve dynamic reference %s: %w", dynamicRef, lexErr)
	}
	return &lexical, nil
}
