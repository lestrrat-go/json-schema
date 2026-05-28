# Validating Data

Validation is two steps: **compile** a schema into a validator, then **validate** data with it.

## Compile once, validate many

`validator.Compile(ctx, s)` returns a `validator.Interface`. Compiling walks the schema once and is the expensive step; the returned validator is safe to **reuse across goroutines** and across many `Validate` calls. Do the compile at startup, keep the `Interface` around, and call `Validate(ctx, data)` per request.

`data` is any decoded JSON value — `map[string]any`, `[]any`, `string`, `float64`, `bool`, `nil`, etc. (the shapes `encoding/json` produces into an `any`).

## Reading the result

`Validate` returns `(Result, error)`:

- **`error == nil`** → the data is valid.
- **`error != nil`** → validation failed; the error describes what and where (e.g. `property validation failed for name: ... string length (0) shorter then minLength (1)`).

The `Result` value carries validation annotations (chiefly which properties/items were evaluated, used internally for `unevaluatedProperties`/`unevaluatedItems`). Most callers only need the error.

## A complete example

Compile once, then validate several inputs against the reused validator:

<!-- INCLUDE(examples/doc_validate_test.go) -->
```go
package examples_test

import (
  "context"
  "fmt"

  schema "github.com/lestrrat-go/json-schema"
  "github.com/lestrrat-go/json-schema/validator"
)

// Example_docValidate compiles a schema once and reuses the validator for several
// inputs. Validate returns a non-nil error when the data is invalid.
func Example_docValidate() {
  s := schema.NewBuilder().
    Types(schema.ObjectType).
    Property("id", schema.PositiveInteger().MustBuild()).
    Property("role", schema.Enum("admin", "user").MustBuild()).
    Required("id", "role").
    MustBuild()

  ctx := context.Background()
  v, err := validator.Compile(ctx, s)
  if err != nil {
    fmt.Println("compile failed:", err)
    return
  }

  for _, data := range []map[string]any{
    {"id": 1, "role": "admin"}, // valid
    {"id": 1, "role": "root"},  // role not in enum
  } {
    _, err := v.Validate(ctx, data)
    fmt.Printf("valid=%t\n", err == nil)
  }
  // Output:
  // valid=true
  // valid=false
}
```
source: [examples/doc_validate_test.go](https://github.com/lestrrat-go/json-schema/blob/main/examples/doc_validate_test.go)
<!-- END INCLUDE -->

## Pass the same context to Compile and Validate

Optional behavior is configured on the `context.Context` and read at both compile and validate time:

- A custom [reference resolver](./03-references.md) (`schema.WithResolver`)
- A different [vocabulary set](./04-vocabularies-and-meta-schema.md) (`vocabulary.WithSet`)
- A [trace logger](#tracing) (`validator.WithTraceSlog`)

Build the context once and pass it to **both** calls. Compiling with a resolver but validating without it (or vice-versa) can lead to surprising results.

## `format` does not assert by default

By default the validator follows the JSON Schema 2020-12 default: `format` is an **annotation**, not an assertion, so `"format": "email"` will not reject `"not-an-email"`. To make formats enforce, enable the format-assertion vocabulary — see [Vocabularies & the Meta-Schema](./04-vocabularies-and-meta-schema.md).

## Tracing

When an error message alone does not make it obvious *why* an input was rejected, attach a structured trace logger with `validator.WithTraceSlog` before compiling and validating. The trace shows which keyword and branch each value hit — the fastest way to debug a failing `anyOf`, `if/then/else`, or a deep nested property. (Point the handler at `os.Stderr` in real use; the example discards it for deterministic output.)

<!-- INCLUDE(examples/doc_tracing_test.go) -->
```go
package examples_test

import (
  "bytes"
  "context"
  "fmt"
  "log/slog"

  schema "github.com/lestrrat-go/json-schema"
  "github.com/lestrrat-go/json-schema/validator"
)

// Example_docTracing attaches a structured trace logger with WithTraceSlog. The
// logger records the validation walk keyword by keyword, which is the fastest way
// to see why an input was rejected. In real use point the handler at os.Stderr;
// here it writes to a buffer that is never printed, so the example output stays
// deterministic.
func Example_docTracing() {
  s := schema.NewBuilder().
    Types(schema.ObjectType).
    Property("id", schema.PositiveInteger().MustBuild()).
    Required("id").
    MustBuild()

  var traceOut bytes.Buffer
  logger := slog.New(slog.NewTextHandler(&traceOut, &slog.HandlerOptions{Level: slog.LevelDebug}))
  ctx := validator.WithTraceSlog(context.Background(), logger)

  v, err := validator.Compile(ctx, s)
  if err != nil {
    fmt.Println("compile failed:", err)
    return
  }

  _, err = v.Validate(ctx, map[string]any{"id": 1})
  fmt.Println("valid:", err == nil)
  // Output:
  // valid: true
}
```
source: [examples/doc_tracing_test.go](https://github.com/lestrrat-go/json-schema/blob/main/examples/doc_tracing_test.go)
<!-- END INCLUDE -->

## Next

- [References](./03-references.md)
- [Vocabularies & the Meta-Schema](./04-vocabularies-and-meta-schema.md)
