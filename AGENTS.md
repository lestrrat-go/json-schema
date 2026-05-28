# AGENTS.md

## For Module Consumers

If you are writing code that *uses* this library (not developing the library itself):

- **How-to documentation**: See the [`docs/`](docs/) directory — task-oriented guides for building schemas, validating data, references, code generation, and the CLI.
- **Runnable examples**: See [`examples/`](examples/) — every `*_test.go` there is an executable `go test` example.
- **API reference**: Use `go doc` or <https://pkg.go.dev/github.com/lestrrat-go/json-schema>.

The rest of this document is for developing the library itself.

---

## Go Version

This project requires the Go version pinned in [`go.mod`](go.mod) (currently **Go 1.24.4**). No `GOEXPERIMENT` flags are required; a plain `go build`/`go test` works.

## Module Layout

Single Go module, `github.com/lestrrat-go/json-schema`, with sub-packages:

| Import path | Directory | Responsibility |
|-------------|-----------|----------------|
| `github.com/lestrrat-go/json-schema` | `/` (root) | Schema data type + fluent builder (the `schema` package) |
| `.../validator` | `validator/` | Compile schemas into validators; run validation; generate validator code |
| `.../vocabulary` | `vocabulary/` | 2020-12 vocabularies and per-context enable/disable |
| `.../keywords` | `keywords/` | String constants for every JSON Schema keyword |
| `.../meta` | `meta/` | Pre-compiled 2020-12 meta-schema validator |
| `.../cmd/json-schema` | `cmd/json-schema/` | CLI: `lint` and `gen-validator` |

The root package is imported as `schema` by convention: `import schema "github.com/lestrrat-go/json-schema"`.

See [`agents/docs/packages.md`](agents/docs/packages.md) for the full exported-API map.

## Code Generation

### Immutable Rule

**NEVER edit files ending in `_gen.go` directly.** They are generated. Edit the generator source, then regenerate.

### Generated Files

| Generated file | Generator source | Input |
|----------------|------------------|-------|
| `schema_gen.go`, `builder_gen.go` | `internal/cmd/genobjects/` | `internal/cmd/genobjects/objects.yml` |
| `meta/meta_gen.go` | `internal/cmd/genmeta/` | embedded 2020-12 meta-schema (no network) |
| `validator/int_gen.go`, `validator/number_gen.go` | numeric-validator generator | — |

### Regenerate

```bash
./gen.sh
```

`gen.sh` builds and runs `genobjects` (schema + builder) and `genmeta` (meta validator), then removes the temporary generator binaries. It must be run from anywhere — it `cd`s to its own directory. See [`agents/docs/codegen.md`](agents/docs/codegen.md) for what each generator emits and how validator code generation (the `gen-validator` CLI path) differs from object generation.

When you edit an `_gen.go` file's generator, commit **both** the generator change and the regenerated output.

## Testing

Use `github.com/stretchr/testify/require` for assertions.

The conformance suite is the upstream [JSON Schema Test Suite](https://github.com/json-schema-org/JSON-Schema-Test-Suite), vendored into `tests/` by a script (it is not committed):

```bash
./init-test-suite.sh           # clone the suite into tests/
./init-test-suite.sh <commit>  # pin to a specific suite commit
go test ./...
```

If `tests/` is absent, run `init-test-suite.sh` before claiming a compliance test is unrunnable. See [`agents/docs/testing.md`](agents/docs/testing.md) for the test layout, the `remotes/` loading path, and the `WithTraceSlog` debugging hook.

## Architecture in One Paragraph

A `*schema.Schema` is inert data (built fluently or unmarshaled from JSON). `validator.Compile(ctx, schema)` walks it once and produces a `validator.Interface` — a tree of small, type-specific validators. `Interface.Validate(ctx, data)` runs that tree. References (`$ref`/`$dynamicRef`) and vocabularies are resolved through the `context.Context`, not globals. See [`agents/docs/architecture.md`](agents/docs/architecture.md) and [`agents/docs/references.md`](agents/docs/references.md).

## Pre-Read Rules

Read the linked doc BEFORE working in that area.

| Trigger | Doc |
|---------|-----|
| Looking up exported types/functions across packages | [`agents/docs/packages.md`](agents/docs/packages.md) |
| Compilation pipeline, schema-vs-validator separation, context passing | [`agents/docs/architecture.md`](agents/docs/architecture.md) |
| Anything touching `$ref`, `$id`, `$anchor`, `$dynamicRef`, the `Resolver` | [`agents/docs/references.md`](agents/docs/references.md) |
| Object/builder code generation OR validator code generation | [`agents/docs/codegen.md`](agents/docs/codegen.md) |
| Running/writing tests, the conformance suite, validation tracing | [`agents/docs/testing.md`](agents/docs/testing.md) |

## Cache Maintenance

These docs cache repository state. Still read the source before modifying code.

1. When your change affects a doc below, update it in the same commit.
2. If you notice any doc is wrong or stale — even on an unrelated task — fix it immediately.

| Doc | Update trigger |
|-----|----------------|
| `agents/docs/packages.md` | New/renamed/removed exported functions, types, or packages |
| `agents/docs/architecture.md` | Changes to the compile pipeline, context keys, or separation of concerns |
| `agents/docs/references.md` | Changes to reference resolution, the registry, dynamic scope, or the `Resolver` |
| `agents/docs/codegen.md` | Changes to generators, `objects.yml` schema, or the `CodeGenerator` interface |
| `agents/docs/testing.md` | Changes to test infrastructure, the suite-loading scripts, or trace hooks |
