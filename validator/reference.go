package validator // ReferenceValidator handles schema references ($ref) with lazy resolution and circular detection
import (
	"context"
	"fmt"
	"strings"
	"sync"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/vocabulary"
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

func (r *ReferenceValidator) Validate(ctx context.Context, v any, options ...ValidateOption) (Result, error) {
	return r.evaluate(ctx, v, newEvalState(ctx, options))
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
	// All resolution inputs were captured into the validator at compile time, so
	// this lazy (validate-time) resolution is self-contained.
	resolver := r.resolver
	if resolver == nil {
		resolver = schema.NewResolver()
	}

	rootSchema := r.rootSchema
	if rootSchema == nil {
		return nil, fmt.Errorf("no root schema available for reference resolution: %s", r.reference)
	}

	baseSchema := r.baseSchema
	if baseSchema == nil {
		baseSchema = rootSchema
	}
	baseURI := r.baseURI

	// Resolve the reference against the enclosing resource captured at compile
	// time, so deferred recursive references in a nested $id resource still
	// resolve within that resource.
	var targetSchema schema.Schema
	if err := resolver.ResolveReference(ctx, &targetSchema, r.reference, baseSchema, baseURI); err != nil {
		return nil, fmt.Errorf("failed to resolve reference %s: %w", r.reference, err)
	}

	// Recompile the resolved schema. The reference is seeded onto the recompile's
	// reference stack so any cycle within the target is classified the same way
	// the original compile would have classified it.
	cs := compileState{
		cfg:            &compileConfig{resolver: resolver, vocab: vocabulary.DefaultSet()},
		rootSchema:     rootSchema,
		baseSchema:     baseSchema,
		baseURI:        baseURI,
		referenceStack: []string{r.reference},
		refDepths:      map[string]int{r.reference: 0},
	}
	compiled, err := compile(ctx, &targetSchema, cs)
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

func (d *dynamicScopeValidator) Validate(ctx context.Context, v any, options ...ValidateOption) (Result, error) {
	return d.evaluate(ctx, v, newEvalState(ctx, options))
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

func (dr *DynamicReferenceValidator) Validate(ctx context.Context, v any, options ...ValidateOption) (Result, error) {
	return dr.evaluate(ctx, v, newEvalState(ctx, options))
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
		if rv := st.dynamicAnchorValidators[name]; rv != nil {
			// The registered validator stands in for an outermost resource, so it
			// re-enters with fresh dynamic scope; the anchor registry is carried
			// forward so nested $dynamicRefs to the same anchor still resolve.
			return evalChild(ctx, rv, v, &evalState{dynamicAnchorValidators: st.dynamicAnchorValidators})
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
		resolver = schema.NewResolver()
	}
	baseSchema := dr.baseSchema
	if baseSchema == nil {
		baseSchema = dr.rootSchema
	}
	return resolveDynamicRef(ctx, resolver, baseSchema, dr.baseURI, dr.reference, st.dynamicScope)
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
		resolver = schema.NewResolver()
	}
	cs := compileState{
		cfg:        &compileConfig{resolver: resolver, vocab: vocabulary.DefaultSet()},
		rootSchema: dr.rootSchema,
		baseSchema: target,
	}
	if target.HasID() && target.ID() != "" {
		// Resolve the target's (possibly relative) $id against the base URI under
		// which the $dynamicRef itself was resolved, so the target's own relative
		// references resolve within its resource.
		if base := schema.ResolveURI(dr.baseURI, target.ID()); base != "" {
			cs.baseURI = base
		}
	}
	if cs.rootSchema == nil {
		cs.rootSchema = target
	}

	v, err := compile(ctx, target, cs)
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
func resolveDynamicRef(ctx context.Context, resolver *schema.Resolver, baseSchema *schema.Schema, baseURI string, dynamicRef string, scopeChain []*schema.Schema) (*schema.Schema, error) {
	// Determine whether the fragment is a plain anchor name (eligible for
	// dynamic-scope bookending) versus a JSON pointer or no fragment at all.
	anchorName := plainAnchorFragment(dynamicRef)

	// Resolve the lexical target as $ref would.
	var lexical schema.Schema
	var lexErr error
	if dynamicRef == "#"+anchorName && anchorName != "" {
		lexErr = resolver.ResolveAnchor(ctx, &lexical, anchorName, baseSchema)
	} else {
		lexErr = resolver.ResolveReference(ctx, &lexical, dynamicRef, baseSchema, baseURI)
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
