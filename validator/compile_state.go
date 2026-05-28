package validator

import (
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

// newCompileState builds the initial compileState for a top-level Compile call
// from the supplied options. Defaults (fresh resolver, default vocabulary) are
// applied so the rest of the compiler never sees a nil config. The root schema
// is both the document root and the initial base resource.
func newCompileState(s *schema.Schema, options []CompileOption) compileState {
	resolver := schema.NewResolver()
	vocab := vocabulary.DefaultSet()
	var baseURI string
	// By default the schema being compiled is its own document root and base
	// resource; WithBaseSchema overrides this for fragment compilation.
	doc := s
	for _, o := range options {
		switch o.Ident() {
		case identResolver{}:
			if r := option.MustGet[*schema.Resolver](o); r != nil {
				resolver = r
			}
		case identVocabularySet{}:
			if vs := option.MustGet[*vocabulary.VocabularySet](o); vs != nil {
				vocab = vs
			}
		case identBaseURI{}:
			baseURI = option.MustGet[string](o)
		case identBaseSchema{}:
			if bs := option.MustGet[*schema.Schema](o); bs != nil {
				doc = bs
			}
		}
	}

	// Eager resolution requires the $id/anchor index to exist before the first
	// $ref is compiled; register the document root up front. RegisterRoot is
	// deduped per root inside the resolver, so this is safe to call repeatedly.
	resolver.RegisterRoot(doc)

	return compileState{
		cfg:        &compileConfig{resolver: resolver, vocab: vocab},
		rootSchema: doc,
		baseSchema: doc,
		baseURI:    baseURI,
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
