# Getting Started

## Install

As a library:

```bash
go get github.com/lestrrat-go/json-schema
```

As a command-line tool:

```bash
go install github.com/lestrrat-go/json-schema/cmd/json-schema@latest
```

The library requires the Go version pinned in `go.mod`. No build tags or `GOEXPERIMENT` flags are needed.

## The core flow

Working with this library is always three steps:

1. **Get a schema** — build one with the fluent API, or unmarshal one from JSON.
2. **Compile** it into a validator with `validator.Compile`.
3. **Validate** data with `Interface.Validate`.

<!-- INCLUDE(examples/doc_quickstart_test.go) -->
```go
package examples_test

import (
  "context"
  "fmt"

  schema "github.com/lestrrat-go/json-schema"
  "github.com/lestrrat-go/json-schema/validator"
)

// Example_docQuickStart shows the three core steps: build a schema, compile it
// into a validator, and validate data. A compiled validator is reusable across
// many Validate calls and safe for concurrent use.
func Example_docQuickStart() {
  ctx := context.Background()

  // 1. Build a schema.
  s := schema.NewBuilder().
    Schema(schema.Version).
    Types(schema.ObjectType).
    Property("name", schema.NonEmptyString().MustBuild()).
    Property("email", schema.Email().MustBuild()).
    Property("age", schema.PositiveInteger().MustBuild()).
    Required("name", "email").
    MustBuild()

  // 2. Compile it into a validator.
  v, err := validator.Compile(ctx, s)
  if err != nil {
    fmt.Println("compile failed:", err)
    return
  }

  // 3. Validate data.
  _, err = v.Validate(ctx, map[string]any{
    "name":  "Ada Lovelace",
    "email": "ada@example.com",
    "age":   36,
  })
  fmt.Println("valid record:", err == nil)

  _, err = v.Validate(ctx, map[string]any{"name": "", "email": "x"})
  fmt.Println("empty name:  ", err == nil)
  // Output:
  // valid record: true
  // empty name:   false
}
```
source: [examples/doc_quickstart_test.go](https://github.com/lestrrat-go/json-schema/blob/main/examples/doc_quickstart_test.go)
<!-- END INCLUDE -->

## Why compile?

Compiling walks the schema once and produces an optimized validator tree. You compile a given schema **once** and reuse the resulting `validator.Interface` for every piece of data — it is safe to share across goroutines. For deployments where you want to skip even the one-time compile, you can [generate validator code](./05-code-generation.md) ahead of time.

## A note about `context.Context`

Both `Compile` and `Validate` take a `context.Context`. The context is how you pass in optional behavior — a custom [reference resolver](./03-references.md), a different set of [vocabularies](./04-vocabularies-and-meta-schema.md), or a [validation trace](./02-validating.md#tracing). Pass the *same* configured context to both `Compile` and `Validate`; configuring one but not the other can change results.

## Next

- [Building Schemas](./01-building-schemas.md)
- [Validating Data](./02-validating.md)
