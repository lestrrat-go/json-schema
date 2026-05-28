<!-- Agent-consumed file. Describes current architecture; verify against source before editing code. -->

# Architecture

## Separation of concerns: Schema vs. Validator

The central design decision is that **schema representation** and **validation execution** are different things in different packages.

| | `schema.Schema` (root pkg) | `validator.Interface` (validator pkg) |
|---|---|---|
| Role | Inert data: a JSON Schema document | Executable validation logic |
| Created by | `NewBuilder()...Build()` or `json.Unmarshal` | `validator.Compile(ctx, schema)` |
| Mutability | Effectively immutable after build | Built once, run many times |
| Concurrency | Safe to share | Safe to share + reuse |

Why it matters when editing:

- Adding a keyword touches **both** sides: the schema field/builder (generated from `objects.yml`) *and* the compile path in `validator/` that turns that field into a validator.
- Schemas never validate themselves. There is no `schema.Validate`. Validation always goes through a compiled `validator.Interface`.

## The compile pipeline

`validator.Compile(ctx, s)` (validator/compiler.go) is the only public entry into validation. It:

1. Seeds the `context.Context` with a default `*schema.Resolver`, the root schema, the current base schema/URI, and the active `*vocabulary.VocabularySet`.
2. Indexes the root's `$id`/`$anchor`/`$dynamicAnchor` via `Resolver.RegisterRoot` (deduped per root — see references.md).
3. Walks the schema, dispatching per keyword group into the unexported `compileXxxValidator` functions:
   - `compileObjectValidator`, `compileArrayValidator`, `compileStringValidator`, `compileIntegerValidator`, `compileNumberValidator`, `compileBooleanValidator`
   - `compileCompositeValidators` (`allOf`/`anyOf`/`oneOf`/`not`), `compileConditionalValidators` (`if`/`then`/`else`)
   - `compileContentValidator`, `compileValueConstraintsValidator` (`enum`/`const`), `compileUntypedValidator` (no `type:`)
4. Combines the per-group validators into one tree implementing `Interface`.
5. Wraps schemas carrying `$id` or `$dynamicAnchor` in a `dynamicScopeValidator` so the dynamic scope is pushed at *validation* time.

The output is a tree of small validators. `Interface.Validate(ctx, value)` runs the tree, returning `(Result, error)`; a non-nil error is a validation failure with a descriptive message.

## Context, not globals

State that must flow through compilation and validation is carried on `context.Context`, never in package globals. This keeps `Compile`/`Validate` reentrant and concurrency-safe. Key carriers:

| Concern | Set with | Read with | Package |
|---------|----------|-----------|---------|
| Active vocabularies | `vocabulary.WithSet` | `vocabulary.SetFromContext` | vocabulary |
| Reference resolver | `schema.WithResolver` | `schema.ResolverFromContext` | schema |
| Root / base schema | `schema.WithRootSchema` / `WithBaseSchema` | `...FromContext` | schema |
| Base URI for `$id` rebasing | `schema.WithBaseURI` | `schema.BaseURIFromContext` | schema |
| Dynamic scope chain (`$dynamicRef`) | `schema.WithDynamicScope` | `schema.DynamicScopeFromContext` | schema |
| Anchor→validator overrides | `schema.WithDynamicAnchorValidator` | `schema.DynamicAnchorValidatorFromContext` | schema |
| Validation trace | `validator.WithTraceSlog` | (internal) | validator |
| `dependentSchemas` map | `validator.WithDependentSchemas` | `validator.DependentSchemasFromContext` | validator |

Practical consequence: to change validation behavior (custom resolver, all vocabularies enabled, tracing) you pass options *through the same ctx* into both `Compile` and `Validate`. Compiling with one ctx and validating with another that lacks the resolver/vocabularies can change results.

## `unevaluatedProperties` / `unevaluatedItems`

These keywords need to know which properties/items *sibling and applicator* validators already evaluated. That information flows back up as the `Result` value: `*ObjectResult` carries `EvaluatedProperties()`, `*ArrayResult` carries `EvaluatedItems()`. Composite validators merge child results so an `unevaluated*` validator can subtract what was covered.

## Code generation has two unrelated meanings

Don't confuse them (see codegen.md):

1. **Object/builder generation** — `internal/cmd/genobjects` turns `objects.yml` into `schema_gen.go` + `builder_gen.go`. Build-time tooling for the library.
2. **Validator code generation** — `validator.CodeGenerator` turns a *compiled validator* back into Go builder source (the `gen-validator` CLI). A runtime feature for library users who want to skip compilation in production.

## File-naming conventions

| Pattern | Meaning |
|---------|---------|
| `*_gen.go` | Generated — DO NOT EDIT (edit the generator) |
| `*_test.go` | Tests |
| `objects.yml` | Input to `genobjects` |
