# How to use `github.com/lestrrat-go/json-schema`

Task-oriented guides for building, validating against, and generating code from JSON Schema 2020-12 documents in Go.

If you would rather read code, every file under [`examples/`](../examples) is a runnable `go test` example.

* [Getting Started](./00-getting-started.md) — install, and the build → compile → validate flow
* [Building Schemas](./01-building-schemas.md) — the fluent builder, convenience constructors, loading from JSON
* [Validating Data](./02-validating.md) — compiling validators, reading errors, tracing
* [References](./03-references.md) — `$ref`, `$id`, `$anchor`, `$dynamicRef`, and the resolver
* [Vocabularies & the Meta-Schema](./04-vocabularies-and-meta-schema.md) — turning `format` into an assertion, validating that a document *is* a schema
* [Code Generation](./05-code-generation.md) — emitting pre-compiled validator code
* [Command Line Tool](./06-command-line-tool.md) — `lint` and `gen-validator`
* [FAQ](./99-faq.md)

## The two-package model

```
schema package          validator package
─────────────           ─────────────────
*schema.Schema   ──────▶  validator.Compile(ctx, schema)  ──────▶  validator.Interface
(inert data)              (compiles once)                          (run many times)
```

A schema is pure data — you build it or unmarshal it from JSON. To validate, you **compile** it into a `validator.Interface` and call `Validate`. Schemas never validate themselves; there is no `schema.Validate`.

## API reference

Full API docs: <https://pkg.go.dev/github.com/lestrrat-go/json-schema>.
