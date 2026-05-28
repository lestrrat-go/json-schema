# Code Generation

For deployments where you want to skip even the one-time `Compile` step, the library can emit Go source that reconstructs a validator directly — no schema document, no compilation at startup. You generate the code once (at build time), commit it, and call the generated builder in production.

There are two ways to do this: the [command-line tool](./06-command-line-tool.md) and the programmatic API.

## With the CLI

```bash
json-schema gen-validator --name UserValidator user-schema.json
```

Given `user-schema.json`:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "name": { "type": "string", "minLength": 1 },
    "email": { "type": "string", "format": "email" },
    "age": { "type": "integer", "minimum": 0, "maximum": 150 }
  },
  "required": ["name", "email"],
  "additionalProperties": false
}
```

it prints a Go assignment you can paste into your code:

```go
UserValidator := validator.Object().
	Required("name", "email").
	Properties(
		validator.PropPair("age", validator.Integer().Minimum(0).Maximum(150).MustBuild()),
		validator.PropPair("email", validator.String().MustBuild()),
		validator.PropPair("name", validator.String().MinLength(1).MustBuild()),
	).
	AdditionalProperties(false).
	StrictObjectType(true).
	MustBuild()
```

`--name` sets the variable name (default `val`). Wire it into your build with `go:generate`:

```go
//go:generate json-schema gen-validator --name UserValidator user-schema.json
```

## Programmatically

`validator.NewCodeGenerator()` returns a `CodeGenerator`; its `Generate` method writes builder source for any compiled validator into an `io.Writer`:

<!-- INCLUDE(examples/doc_codegen_test.go) -->
```go
package examples_test

import (
  "bytes"
  "context"
  "fmt"

  schema "github.com/lestrrat-go/json-schema"
  "github.com/lestrrat-go/json-schema/validator"
)

// Example_docCodegen compiles a schema and emits Go source that reconstructs the
// validator directly, so production code can skip compilation. This is the
// programmatic form of the `json-schema gen-validator` CLI command.
func Example_docCodegen() {
  s := schema.NewBuilder().
    Types(schema.StringType).
    MinLength(1).
    MustBuild()

  v, err := validator.Compile(context.Background(), s)
  if err != nil {
    fmt.Println("compile failed:", err)
    return
  }

  var buf bytes.Buffer
  if err := validator.NewCodeGenerator().Generate(&buf, v); err != nil {
    fmt.Println("generate failed:", err)
    return
  }
  // Generate emits the raw builder calls (one per line). The gen-validator CLI
  // additionally assigns them to a variable and runs the result through gofmt.
  fmt.Print(buf.String())
  // Output:
  // validator.String().
  // MinLength(1).
  // MustBuild()
}
```
source: [examples/doc_codegen_test.go](https://github.com/lestrrat-go/json-schema/blob/main/examples/doc_codegen_test.go)
<!-- END INCLUDE -->

This is exactly what `gen-validator` does internally: compile, `Generate` into a buffer, prepend `name :=`, then run it through `go/format` for the indented layout shown above.

## The generated validator builders

Generated code uses the same `validator` builders you can write by hand: `validator.Object()`, `validator.String()`, `validator.Integer()`, `validator.Number()`, `validator.Array()`, `validator.Boolean()`, `validator.Null()`, the composition helpers `validator.AllOf/AnyOf/OneOf(...)`, and `validator.PropPair(name, v)` for object properties. So the output is readable, reviewable Go — not an opaque blob.

## Next

- [Command Line Tool](./06-command-line-tool.md)
