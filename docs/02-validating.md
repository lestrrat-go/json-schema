# Validating Data

Validation is two steps: **compile** a schema into a validator, then **validate** data with it.

## Compile once, validate many

```go
import "github.com/lestrrat-go/json-schema/validator"

v, err := validator.Compile(ctx, s) // s is a *schema.Schema
if err != nil {
	// the schema itself is invalid / cannot be compiled
}

result, err := v.Validate(ctx, data)
```

`Compile` returns a `validator.Interface`. Compiling walks the schema once and is the expensive step; the returned validator is safe to **reuse across goroutines** and across many `Validate` calls. Do the compile at startup, keep the `Interface` around, and call `Validate` per request.

`data` is any decoded JSON value — `map[string]any`, `[]any`, `string`, `float64`, `bool`, `nil`, etc. (the shapes `encoding/json` produces into an `any`).

## Reading the result

`Validate` returns `(Result, error)`:

- **`error == nil`** → the data is valid.
- **`error != nil`** → validation failed; the error describes what and where.

```go
if _, err := v.Validate(ctx, data); err != nil {
	fmt.Println("invalid:", err)
	// invalid: ... property validation failed for name: ... string length (0) shorter then minLength (1)
}
```

The `Result` value carries validation annotations (chiefly which properties/items were evaluated, used internally for `unevaluatedProperties`/`unevaluatedItems`). Most callers only need the error.

## A complete example

```go
ctx := context.Background()

s := schema.NewBuilder().
	Types(schema.ObjectType).
	Property("id", schema.PositiveInteger().MustBuild()).
	Property("role", schema.Enum("admin", "user").MustBuild()).
	Required("id", "role").
	MustBuild()

v, err := validator.Compile(ctx, s)
if err != nil {
	panic(err)
}

for _, data := range []map[string]any{
	{"id": 1, "role": "admin"}, // valid
	{"id": 1, "role": "root"},  // invalid: role not in enum
} {
	_, err := v.Validate(ctx, data)
	fmt.Printf("valid=%t\n", err == nil)
}
```

## Pass the same context to Compile and Validate

Optional behavior is configured on the `context.Context` and read at both compile and validate time:

- A custom [reference resolver](./03-references.md) (`schema.WithResolver`)
- A different [vocabulary set](./04-vocabularies-and-meta-schema.md) (`vocabulary.WithSet`)
- A [trace logger](#tracing) (`validator.WithTraceSlog`)

Build the context once and pass it to **both** calls. Compiling with a resolver but validating without it (or vice-versa) can lead to surprising results.

## `format` does not assert by default

By default the validator follows the JSON Schema 2020-12 default: `format` is an **annotation**, not an assertion, so `"format": "email"` will not reject `"not-an-email"`. To make formats enforce, enable the format-assertion vocabulary — see [Vocabularies & the Meta-Schema](./04-vocabularies-and-meta-schema.md).

## Tracing

When an error message alone does not make it obvious *why* an input was rejected, attach a structured trace logger before compiling and validating:

```go
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
ctx := validator.WithTraceSlog(context.Background(), logger)

v, _ := validator.Compile(ctx, s)
_, err := v.Validate(ctx, data) // emits a step-by-step trace of the validation walk
```

The trace shows which keyword and branch each value hit, which is the fastest way to debug a failing `anyOf`, `if/then/else`, or a deep nested property.

## Next

- [References](./03-references.md)
- [Vocabularies & the Meta-Schema](./04-vocabularies-and-meta-schema.md)
