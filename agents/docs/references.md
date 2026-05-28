<!-- Agent-consumed file. Reference resolution is subtle; verify against source before editing. -->

# Reference Resolution

Covers `$ref`, `$id`, `$anchor`, `$dynamicRef`, `$dynamicAnchor`, and the `schema.Resolver`. The entire required JSON Schema 2020-12 suite passes; treat the behaviors below as load-bearing, not incidental.

## Files

- `uri.go` — `ResolveURI` (RFC 3986 base+ref join).
- `registry.go` — `resourceIndex` (absolute URI → schema, plus anchors), `FindDynamicAnchor`, child-schema enumeration.
- `resolver.go` — `Resolver`: a `registryResolver` stacked ahead of any caller-supplied resolvers, then a final object resolver (via `lestrrat-go/jsref/v2`). `NewResolver(...ResolverOption)`, `RegisterRoot`, `RegisterDocument`, `RegisterFS`, `ResourceFor`, `ResolveReference`.
- `resolver_options.go` — `ResolverOption`, `WithResolver`, and the opt-in resolver factories `HTTPResolver`, `FSResolver(fs.FS)`, `DirResolver(dir)` plus the `fs.FS`-backed `fsResolver`.
- `validator/compiler.go` — the eager `$ref` resolution block.
- `validator/reference.go` — `ReferenceValidator`, `DynamicReferenceValidator`, `plainAnchorFragment`.

## In-document `$id` registry

A `resourceIndex` maps absolute URIs → schemas and is consulted by the `registryResolver` before any opt-in network/FS resolver. Built at the root `Compile` via `Resolver.RegisterRoot`.

`RegisterRoot` is **deduped per root** (a `registered` map guarded by `Resolver.mu`). This is required: validate-time recompiles must NOT re-index, or a data race results. Do not remove the dedup.

## `$ref` resolves eagerly at compile time

The `$ref` block in `compiler.go` resolves and compiles the target during `Compile`, with one exception for recursion:

- **Data-bounded recursive `$ref`** (e.g. a tree node referencing itself) is detected via a `dataDepth` counter carried in ctx, incremented at the top of `compileObjectValidator`/`compileArrayValidator` (the child-applying keywords). At the circular-reference guard: if depth *increased* since the ref was entered → data-bounded recursion → defer to a lazy `ReferenceValidator`. If depth is *unchanged* → a pure cycle → return a compile error (this preserves `TestCircularReferenceDetection`).
- Sibling keywords on a `$ref` schema are merged via `combineReferenceWithConstraints`.

When a `$ref` enters another resource, `compileSchema` sets the base URI to the reference's absolute retrieval URI and the base schema to the enclosing resource (`ResourceFor`). If the target is an already-registered resource, it sets a one-shot `skipIDRebase` flag (`withSkipIDRebase`) so the `$id` re-base is not applied twice (which would duplicate a path segment).

## `$dynamicRef` resolves at runtime

`$dynamicRef` uses the **runtime dynamic scope**, not compile-time resolution:

- A `dynamicScopeValidator` wraps every schema carrying `$id` or `$dynamicAnchor` and pushes it onto the ctx dynamic scope during `Validate`. Following a `$ref` into a resource also pushes that resource.
- `DynamicReferenceValidator.Validate` resolves per-call (caches compiled validators by target pointer under a mutex; does not memoize the resolution itself).
- `resolveDynamicRef` does **bookending**: first resolve the fragment lexically the way `$ref` would; only if that lexical target declares a `$dynamicAnchor` of the same name does it walk the scope outermost-first via `schema.FindDynamicAnchor` (which stops at a nested `$id`).
- Sibling keywords (e.g. `unevaluatedProperties` next to `$dynamicRef`) are combined with `combineReferenceWithConstraints`, as for `$ref`.

## Remote / preloaded documents

- `Resolver.RegisterDocument(uri, root)` preloads a document under an explicit *retrieval* URI; it becomes addressable both by that URI and by its own canonical `$id`. The conformance suite loads its `remotes/` tree this way (`loadRemotes`/`newSuiteResolver` in `schema_compliance_test.go`).
- `Resolver.RegisterFS(baseURI, fsys)` walks an `fs.FS` and registers every `.json` file under `baseURI` joined with its path. Works with `embed.FS`, `os.DirFS`, `fstest.MapFS`.
- **External access is opt-in.** A bare `NewResolver()` resolves only from memory (registry + preloaded docs); an external `$ref` that is not preloaded fails rather than being fetched. To allow it, pass a resolver explicitly: `NewResolver(WithResolver(HTTPResolver()))` for HTTP/HTTPS, `NewResolver(WithResolver(DirResolver(".")))` or `WithResolver(FSResolver(fsys))` for files. The FS resolvers are backed by `io/fs` and accept JSON or YAML documents.
- Offline preloading (RegisterDocument/RegisterFS) remains preferred for tests and reproducible builds.

## Pre-compiled meta validator + `$dynamicRef` (a real footgun, already fixed)

The generated `meta` validator (`meta/meta_gen.go`) is a flattened builder tree with **no** schema document / resolver / dynamic-scope wrappers at runtime. Its many `NewDynamicReferenceValidator("#meta")` nodes therefore have nothing to resolve against the normal way.

Fix in place: a **context-carried dynamic-anchor → validator registry** (`schema.WithDynamicAnchorValidator` / `DynamicAnchorValidatorFromContext`, values are `validator.Interface` stored as `any`). `DynamicReferenceValidator.Validate` checks this registry first for a plain-anchor fragment (`plainAnchorFragment`) and delegates if present. Hand-written `meta/meta.go` registers `metaValidator` under `"meta"` and delegates, so `"#meta"` recurses into the whole meta validator (nested subschemas ARE validated; an accept-all fallback was rejected because it would silently skip them). `resolveDynamicRef` and `ResolveAnchor` were also hardened against a nil base schema. See `meta/meta_test.go`.

If you change anything about how validators without a live resolver handle `$dynamicRef`, re-run `go test ./meta/...` — this is exactly where it breaks.
