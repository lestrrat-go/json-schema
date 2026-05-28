<!-- Agent-consumed file. Verify the CodeGenerator interface and generator sources before editing. -->

# Code Generation

"Code generation" means two unrelated things in this repo. Keep them straight.

## 1. Object + builder generation (build-time tooling)

Turns a YAML description of the schema keyword set into the `Schema` data type and its `Builder`.

| Output | Generator | Input |
|--------|-----------|-------|
| `schema_gen.go` (the `Schema` struct, accessors, field flags, `MarshalJSON`/`UnmarshalJSON`) | `internal/cmd/genobjects/` | `internal/cmd/genobjects/objects.yml` |
| `builder_gen.go` (the `Builder`, one chainable method + `ResetXxx` per keyword) | `internal/cmd/genobjects/` | same |
| `meta/meta_gen.go` (the `metaValidator` value) | `internal/cmd/genmeta/` | meta-schema embedded in the generator (no network) |
| `validator/int_gen.go`, `validator/number_gen.go` | numeric-validator generator | — |

Run everything with:

```bash
./gen.sh
```

`gen.sh` builds `genobjects`, runs it against `objects.yml`, builds `genmeta`, runs it, and deletes both temporary binaries. To add or change a schema keyword: edit `objects.yml` (and the generator if the shape is new), run `gen.sh`, commit generator + regenerated `_gen.go` together. **Never hand-edit `_gen.go`.**

`meta/meta.go` is hand-written and owns the public `Validator()` / `Validate()`; `genmeta` only emits the `metaValidator` value it wraps. This split exists so the meta validator can register itself under the `"meta"` dynamic anchor (see references.md) — logic that does not belong in generated output.

## 2. Validator code generation (a runtime library feature)

Turns a *compiled* `validator.Interface` back into Go builder source, so users can vendor a pre-built validator and skip `Compile` at startup.

```go
type CodeGenerator interface {
    Generate(dst io.Writer, v Interface) error
}

func NewCodeGenerator() CodeGenerator
```

(`validator/codegen_core.go`.) `Generate` walks the validator tree via a type switch over the concrete validator types (`*stringValidator`, `*objectValidator`, the composite validators, reference/content/dependent-schemas validators, …) and writes equivalent `validator.Xxx()....MustBuild()` calls. The CLI `gen-validator` wraps this: it compiles the input schema, calls `Generate` into a buffer, prepends `<name> :=`, and runs the result through `go/format`.

When you add a new validator type, you must extend the `Generate` type switch too, or `gen-validator` will fail with an unsupported-type error on schemas that use it. Round-trip coverage (compile → generate → the generated code validates identically) lives in the validator package tests.
