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

```go
package main

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

func main() {
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
		panic(err)
	}

	// 3. Validate data.
	_, err = v.Validate(ctx, map[string]any{
		"name":  "Ada Lovelace",
		"email": "ada@example.com",
		"age":   36,
	})
	fmt.Println("valid:", err == nil)

	_, err = v.Validate(ctx, map[string]any{"name": "", "email": "nope"})
	fmt.Println("valid:", err == nil) // false
}
```

## Why compile?

Compiling walks the schema once and produces an optimized validator tree. You compile a given schema **once** and reuse the resulting `validator.Interface` for every piece of data — it is safe to share across goroutines. For deployments where you want to skip even the one-time compile, you can [generate validator code](./05-code-generation.md) ahead of time.

## A note about `context.Context`

Both `Compile` and `Validate` take a `context.Context`. The context is how you pass in optional behavior — a custom [reference resolver](./03-references.md), a different set of [vocabularies](./04-vocabularies-and-meta-schema.md), or a [validation trace](./02-validating.md#tracing). Pass the *same* configured context to both `Compile` and `Validate`; configuring one but not the other can change results.

## Next

- [Building Schemas](./01-building-schemas.md)
- [Validating Data](./02-validating.md)
