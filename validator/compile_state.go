package validator

import (
	"context"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/vocabulary"
	"github.com/lestrrat-go/option/v3"
)

// compileConfig holds the immutable, whole-compilation inputs. It is shared by
// pointer across every compileState derived during a single Compile call: these
// values never change as recursion descends.
type compileConfig struct {
	resolver *schema.Resolver
	vocab    *vocabulary.VocabularySet
}

// compileState is the explicit per-recursion-edge carrier for compilation. It
// replaces the values that used to be smuggled through context.Context
// (resolver, root/base schema, base URI, reference stack, ref depths, data
// depth). It is passed BY VALUE so a callee deriving a modified state (entering
// an $id resource, crossing a data boundary, pushing a reference) cannot leak
// that change back to its caller or siblings — exactly the fork semantics the
// old `ctx = schema.WithX(ctx, …)` calls provided.
//
// cfg is a shared pointer (immutable for the whole compile); the remaining
// fields are copied on each value-copy of the struct. Slice/map fields
// (referenceStack, refDepths) are replaced with fresh copies when pushed, so the
// value-copy is safe.
type compileState struct {
	cfg *compileConfig

	rootSchema *schema.Schema // the document root (constant after entry)
	baseSchema *schema.Schema // enclosing resource; changes on $id boundaries
	baseURI    string         // enclosing resource's base URI

	referenceStack []string       // active $ref chain for cycle detection
	refDepths      map[string]int // data depth at which each active $ref was entered
	dataDepth      int            // child-applying keyword boundaries crossed

	// skipIDRebase marks that the caller already set the base URI to the target
	// resource's canonical URI (from the registry), so compileSchema must not
	// re-base the target's $id again (which would double a path segment). It
	// applies only to the immediate schema, never its nested subschemas.
	skipIDRebase bool
}

// newCompileState builds the initial compileState for a top-level Compile call.
// Explicit CompileOptions take precedence; for backward compatibility, any
// values the caller placed on ctx via the public schema.With* / vocabulary.WithSet
// helpers seed anything an option did not set. Defaults (fresh resolver, default
// vocabulary) are applied so the rest of the compiler never sees a nil config.
func newCompileState(ctx context.Context, s *schema.Schema, options []CompileOption) compileState {
	var optResolver *schema.Resolver
	var optVocab *vocabulary.VocabularySet
	var optBaseURI string
	haveBaseURI := false
	for _, o := range options {
		switch o.Ident() {
		case identResolver{}:
			optResolver = option.MustGet[*schema.Resolver](o)
		case identVocabularySet{}:
			optVocab = option.MustGet[*vocabulary.VocabularySet](o)
		case identBaseURI{}:
			optBaseURI = option.MustGet[string](o)
			haveBaseURI = true
		}
	}

	resolver := optResolver
	if resolver == nil {
		resolver = schema.ResolverFromContext(ctx)
	}
	if resolver == nil {
		resolver = schema.NewResolver()
	}

	vocab := optVocab
	if vocab == nil {
		vocab = vocabulary.SetFromContext(ctx)
	}
	if vocab == nil {
		vocab = vocabulary.DefaultSet()
	}

	baseURI := optBaseURI
	if !haveBaseURI {
		baseURI = schema.BaseURIFromContext(ctx)
	}

	rootSchema := schema.RootSchemaFromContext(ctx)
	if rootSchema == nil {
		rootSchema = s
	}
	// Eager resolution requires the $id/anchor index to exist before the first
	// $ref is compiled; register the root up front. RegisterRoot is deduped per
	// root inside the resolver, so this is safe to call repeatedly.
	resolver.RegisterRoot(rootSchema)

	baseSchema := schema.BaseSchemaFromContext(ctx)
	if baseSchema == nil {
		baseSchema = s
	}

	return compileState{
		cfg:            &compileConfig{resolver: resolver, vocab: vocab},
		rootSchema:     rootSchema,
		baseSchema:     baseSchema,
		baseURI:        baseURI,
		referenceStack: schema.ReferenceStackFromContext(ctx),
		refDepths:      schema.RefDepthsFromContext(ctx),
		dataDepth:      schema.DataDepthFromContext(ctx),
	}
}

// withBase returns a copy of cs whose enclosing resource (base schema and base
// URI) has been replaced — used when compilation crosses into a schema that
// declares its own $id, or follows a $ref into another resource.
func (cs compileState) withBase(base *schema.Schema, baseURI string) compileState {
	cs.baseSchema = base
	cs.baseURI = baseURI
	return cs
}

// withBaseSchema returns a copy of cs with only the base schema replaced.
func (cs compileState) withBaseSchema(base *schema.Schema) compileState {
	cs.baseSchema = base
	return cs
}

// withBaseURI returns a copy of cs with only the base URI replaced.
func (cs compileState) withBaseURI(baseURI string) compileState {
	cs.baseURI = baseURI
	return cs
}

// incDataDepth returns a copy of cs with the data depth incremented by one,
// recording that compilation has crossed a child-applying keyword boundary
// (object/array applicators) — the signal that distinguishes data-bounded
// recursion from a pure reference cycle.
func (cs compileState) incDataDepth() compileState {
	cs.dataDepth++
	return cs
}

// pushReference returns a copy of cs with reference appended to the active
// reference chain and its entry data-depth recorded. The slice and map are
// copied so the parent's view is untouched.
func (cs compileState) pushReference(reference string) compileState {
	newStack := make([]string, len(cs.referenceStack)+1)
	copy(newStack, cs.referenceStack)
	newStack[len(cs.referenceStack)] = reference
	cs.referenceStack = newStack

	newDepths := make(map[string]int, len(cs.refDepths)+1)
	for k, v := range cs.refDepths {
		newDepths[k] = v
	}
	newDepths[reference] = cs.dataDepth
	cs.refDepths = newDepths

	return cs
}
