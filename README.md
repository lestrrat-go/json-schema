# github.com/lestrrat-go/json-schema [![CI](https://github.com/lestrrat-go/json-schema/actions/workflows/ci.yml/badge.svg)](https://github.com/lestrrat-go/json-schema/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/lestrrat-go/json-schema.svg)](https://pkg.go.dev/github.com/lestrrat-go/json-schema) [![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/lestrrat-go/json-schema)

Go module implementing JSON Schema validation following the [JSON Schema 2020-12](https://json-schema.org/draft/2020-12/schema) specification — with a fluent schema builder, reusable compiled validators, and ahead-of-time validator code generation.

# Features

* Complete JSON Schema 2020-12 specification support
  * All validation keywords (type, properties, items, etc.)
  * Schema composition (allOf, anyOf, oneOf)
  * Conditional schemas (if/then/else)
  * Reference resolution ($ref, $anchor, $dynamicRef) — in-memory by default; network/filesystem access is opt-in
  * Format validation (email, uri, date-time, etc.)
  * Content validation (base64, json-pointer, etc.)
* Clean separation between schema construction and validation
  * Build schemas using fluent builder API
  * Compile schemas into optimized validators for high-performance validation
* Ahead-of-time code generation
  * Emit a schema as standalone Go validator source, compiled straight into your binary
  * No schema files to embed and no runtime schema compilation in production
* Command-line tool for schema validation and code generation
* Comprehensive test coverage including JSON Schema Test Suite compliance

> [!IMPORTANT]
> **Reference resolution is in-memory by default — external access is opt-in.**
> A `$ref`/`$dynamicRef` to another document resolves **only** from schemas you
> have registered in memory. The resolver does **not** reach out to the network
> or read the local filesystem unless you explicitly enable it. This prevents
> surprise outbound fetches (SSRF) and unexpected disk reads.
>
> **If you relied on the old behavior where external `$ref`s were fetched
> automatically, you must now opt in:**
>
> ```go
> // Default: in-memory only. Preload externals with RegisterFS / RegisterDocument.
> r := schema.NewResolver()
>
> // Opt in to live access when you want it:
> r = schema.NewResolver(schema.WithResolver(schema.HTTPResolver()))    // HTTP/HTTPS
> r = schema.NewResolver(schema.WithResolver(schema.DirResolver(".")))  // local files under "."
> r = schema.NewResolver(schema.WithResolver(schema.FSResolver(fsys)))  // any io/fs (embed.FS, …)
> ```
>
> See [References](./docs/03-references.md) for details.

# SYNOPSIS

<!-- INCLUDE(examples/json_schema_readme_example_test.go) -->
```go
package examples_test

import (
  "context"
  "encoding/json"
  "fmt"
  "os"

  schema "github.com/lestrrat-go/json-schema"
  "github.com/lestrrat-go/json-schema/validator"
)

func Example() {
  // Build a JSON Schema using the fluent builder API
  userSchema := schema.NewBuilder().
    Schema(schema.Version).
    Types(schema.ObjectType).
    Property("name", schema.NonEmptyString().MustBuild()).
    Property("email", schema.Email().MustBuild()).
    Property("age", schema.PositiveInteger().MustBuild()).
    Required("name", "email").
    MustBuild()
  enc := json.NewEncoder(os.Stdout)
  enc.SetIndent("", "  ")
  if err := enc.Encode(userSchema); err != nil {
    fmt.Printf("failed to encode schema: %s\n", err)
    return
  }

  // Compile the schema into an optimized validator
  v, err := validator.Compile(context.Background(), userSchema)
  if err != nil {
    fmt.Printf("failed to compile validator: %s\n", err)
    return
  }

  // Validate data
  validUser := map[string]any{
    "name":  "John Doe",
    "email": "john@example.com",
    "age":   30,
  }

  _, err = v.Validate(context.Background(), validUser)
  if err != nil {
    fmt.Printf("validation failed: %s\n", err)
    return
  }

  fmt.Println("User data is valid!")

  // Test with invalid data
  invalidUser := map[string]any{
    "name":  "", // Empty name should fail
    "email": "not-an-email",
  }

  _, err = v.Validate(context.Background(), invalidUser)
  if err != nil {
    fmt.Printf("validation failed as expected: %s\n", err)
  }
  // OUTPUT:
  // {
  //   "$schema": "https://json-schema.org/draft/2020-12/schema",
  //   "properties": {
  //     "age": {
  //       "minimum": 0,
  //       "type": "integer"
  //     },
  //     "email": {
  //       "format": "email",
  //       "type": "string"
  //     },
  //     "name": {
  //       "minLength": 1,
  //       "type": "string"
  //     }
  //   },
  //   "required": [
  //     "name",
  //     "email"
  //   ],
  //   "type": "object"
  // }
  // User data is valid!
  // validation failed as expected: invalid value passed to ObjectValidator: property validation failed for name: invalid value passed to StringValidator: string length (0) shorter then minLength (1)
}
```
source: [examples/json_schema_readme_example_test.go](https://github.com/lestrrat-go/json-schema/blob/main/examples/json_schema_readme_example_test.go)
<!-- END INCLUDE -->

# Install

As a library:

```bash
go get github.com/lestrrat-go/json-schema
```

As a command-line tool:

```bash
go install github.com/lestrrat-go/json-schema/cmd/json-schema@latest
```

# Command Line Tool

The `json-schema` command line tool provides schema validation and code generation.

## Basic Usage

```bash
# Validate a schema file
json-schema lint schema.json

# Generate validator code
json-schema gen-validator --name UserValidator schema.json

# Read from stdin
echo '{"type": "string"}' | json-schema lint -
```

## Complete Example

Given a JSON schema file `user-schema.json`:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "User",
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1
    },
    "email": {
      "type": "string",
      "format": "email"
    },
    "age": {
      "type": "integer",
      "minimum": 0,
      "maximum": 150
    }
  },
  "required": ["name", "email"],
  "additionalProperties": false
}
```

Generate a Go validator:

```bash
json-schema gen-validator --name UserValidator user-schema.json
```

Output:
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

This generated validator can be directly used in your Go code for high-performance validation.

# Documentation

* [How-to Guides](./docs) — building schemas, validating, references, vocabularies, code generation, and the CLI
* [API Reference](https://pkg.go.dev/github.com/lestrrat-go/json-schema)
* [Runnable Examples](./examples)
* Working on the library itself? See [`AGENTS.md`](./AGENTS.md)

# Description

This Go module implements JSON Schema validation according to the 2020-12 specification.
The library provides a clean separation between schema construction and validation:

1. **Schema Construction**: Build schemas using a fluent builder API or unmarshal from JSON
2. **Validator Compilation**: Convert schemas into optimized validator objects for high-performance validation
3. **Validation**: Use compiled validators to validate data with detailed error reporting
4. **Code Generation**: Generate Go code that creates pre-compiled validators

## Core Packages

| Package | Description |
|---------|-------------|
| [schema](https://pkg.go.dev/github.com/lestrrat-go/json-schema) | Schema construction and builder API |
| [validator](https://pkg.go.dev/github.com/lestrrat-go/json-schema/validator) | Validator compilation and validation logic |
| [cmd/json-schema](./cmd/json-schema) | Command-line tool for validation and code generation |

## Design Philosophy

This library follows a clear separation of concerns:

* **Schemas** are pure data structures for representing JSON Schema documents
* **Validators** are compiled, optimized objects for performing validation
* **Code Generation** allows pre-compilation of validators for deployment scenarios

This design provides:
- High performance through compiled validators
- Clean, intuitive API
- Flexibility for different use cases
- Strong type safety

## Why this library?

Most Go JSON Schema packages specialize in one direction — validating data, generating Go types *from* a schema, or inferring a schema *from* Go structs. This module instead covers the whole path from authoring to deployment:

* **Author** schemas with a fluent builder and ready-made constructors (`Email()`, `PositiveInteger()`, `Optional()`, …), or unmarshal them from JSON — the two are interchangeable.
* **Compile** a schema once into an optimized validator that is safe to reuse across goroutines and many requests.
* **Generate** that validator as plain Go source, so production binaries skip schema parsing and compilation entirely — a step most validators leave for runtime.

Schemas are fully encapsulated rather than open structs, so you cannot accidentally build a half-initialized or internally inconsistent schema, and the compiled form is free to optimize. The aim is one cohesive 2020-12 implementation that scales from a quick `lint` of a schema file to high-throughput services with no per-request setup.

# Contributions

## Issues

For bug reports and feature requests, please try to follow the issue templates as much as possible.
For either bug reports or feature requests, failing tests are even better.

## Pull Requests

Please make sure to include tests that exercise the changes you made.

See [`AGENTS.md`](./AGENTS.md) for a map of the codebase, the build/test/codegen workflow, and architectural notes — it is written for both human contributors and AI coding agents.

If you are editing auto-generated files (those with the `_gen.go` suffix), please do the following:

1. Edit the generator, not the generated files (e.g. `internal/cmd/genobjects/`)
2. Run `./gen.sh` to regenerate the code
3. Commit _both_ the generator _and_ the generated files

## Discussions / Usage

Please try GitHub Discussions for usage questions and general discussion.