# Validating Data

Validation is two steps: **compile** a schema into a validator, then **validate** data with it.

## Compile once, validate many

`validator.Compile(ctx, s)` returns a `validator.Interface`. Compiling walks the schema once and is the expensive step; the returned validator is safe to **reuse across goroutines** and across many `Validate` calls. Do the compile at startup, keep the `Interface` around, and call `Validate(ctx, data)` per request.

`data` is any decoded JSON value — `map[string]any`, `[]any`, `string`, `float64`, `bool`, `nil`, etc. (the shapes `encoding/json` produces into an `any`). To validate raw JSON text without decoding it yourself first, see [Validating raw JSON text](#validating-raw-json-text).

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

## Validating raw JSON text

If you start from JSON bytes, `validator.ValidateJSON(ctx, v, data)` decodes and validates in one call — no manual `json.Unmarshal` step. It uses the same compiled validator, so the compile-once / validate-many guidance above still applies; only the decoding differs.

Two things to know:

- **Numbers keep their precision.** `ValidateJSON` decodes with `json.Decoder.UseNumber()`, so a 64-bit identifier larger than 2^53 is validated exactly instead of being rounded by `float64`. (Integer values outside the `int64` range cannot be validated as integers and are reported as an error.)
- **Exactly one value.** The input must contain a single top-level JSON value; trailing content after it (other than whitespace) is rejected. Empty or whitespace-only input is an error.

<!-- INCLUDE(examples/validate_json_example_test.go) -->
```go
package examples_test

import (
  "context"
  "fmt"

  schema "github.com/lestrrat-go/json-schema"
  "github.com/lestrrat-go/json-schema/validator"
)

// Example_validateJSON validates raw JSON text directly with
// validator.ValidateJSON, skipping a manual json.Unmarshal step. Numbers are
// decoded as json.Number, so a 64-bit identifier larger than 2^53 is validated
// exactly rather than being rounded by float64.
func Example_validateJSON() {
  s := schema.NewBuilder().
    Types(schema.ObjectType).
    Property("id", schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
    Property("role", schema.Enum("admin", "user").MustBuild()).
    Required("id", "role").
    MustBuild()

  ctx := context.Background()
  v, err := validator.Compile(ctx, s)
  if err != nil {
    fmt.Println("compile failed:", err)
    return
  }

  for _, data := range [][]byte{
    []byte(`{"id": 9007199254740993, "role": "admin"}`), // large id, valid
    []byte(`{"id": 1, "role": "root"}`),                 // role not in enum
    []byte(`{"role": "admin"}`),                         // missing required id
  } {
    _, err := validator.ValidateJSON(ctx, v, data)
    fmt.Printf("valid=%t\n", err == nil)
  }
  // Output:
  // valid=true
  // valid=false
  // valid=false
}
```
source: [examples/validate_json_example_test.go](https://github.com/lestrrat-go/json-schema/blob/main/examples/validate_json_example_test.go)
<!-- END INCLUDE -->

## Configuring Compile and Validate

Optional behavior is configured two ways:

- **Compile options** passed to `validator.Compile(ctx, schema, opts...)`:
  - A custom [reference resolver](./03-references.md) — `validator.WithResolver(r)`. Note that external (`network`/`filesystem`) access is **opt-in** on the resolver itself; see [References](./03-references.md).
  - A different [vocabulary set](./04-vocabularies-and-meta-schema.md) — `validator.WithVocabularySet(vs)`.
- **A trace logger** carried on the `context.Context` — `validator.WithTraceSlog(ctx, logger)` — and read while validating (see [Tracing](#tracing)).

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
