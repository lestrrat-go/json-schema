# github.com/lestrrat-go/json-schema [![CI](https://github.com/lestrrat-go/json-schema/actions/workflows/ci.yml/badge.svg)](https://github.com/lestrrat-go/json-schema/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/lestrrat-go/json-schema.svg)](https://pkg.go.dev/github.com/lestrrat-go/json-schema) [![codecov.io](https://codecov.io/github/lestrrat-go/json-schema/coverage.svg?branch=main)](https://codecov.io/github/lestrrat-go/json-schema?branch=main)

Go module implementing JSON Schema validation following the [JSON Schema 2020-12](https://json-schema.org/draft/2020-12/schema) specification.

# Features

* Complete JSON Schema 2020-12 specification support
  * All validation keywords (type, properties, items, etc.)
  * Schema composition (allOf, anyOf, oneOf)
  * Conditional schemas (if/then/else)
  * Reference resolution ($ref, $anchor, $dynamicRef)
  * Format validation (email, uri, date-time, etc.)
  * Content validation (base64, json-pointer, etc.)
* Clean separation between schema construction and validation
  * Build schemas using fluent builder API
  * Compile schemas into optimized validators for high-performance validation
* Code generation support
  * Generate pre-compiled validator code for deployment optimization
  * Skip runtime schema compilation overhead
* Command-line tool for schema validation and code generation
* Comprehensive test coverage including JSON Schema Test Suite compliance

# SYNOPSIS

<!-- INCLUDE(examples/json_schema_readme_example_test.go) -->
```go
package examples_test

import (
  "context"
  "fmt"

  schema "github.com/lestrrat-go/json-schema"
  "github.com/lestrrat-go/json-schema/validator"
)

func Example() {
  // Build a JSON Schema using the fluent builder API
  userSchema := schema.NewBuilder().
    Types(schema.ObjectType).
    Property("name", schema.NonEmptyString().MustBuild()).
    Property("email", schema.Email().MustBuild()).
    Property("age", schema.PositiveInteger().MustBuild()).
    Required("name", "email").
    MustBuild()

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
  // User data is valid!
  // validation failed as expected: invalid value passed to ObjectValidator: property validation failed for name: invalid value passed to StringValidator: string length (0) shorter then minLength (1)
}```
source: [examples/json_schema_readme_example_test.go](https://github.com/lestrrat-go/json-schema/blob/main/examples/json_schema_readme_example_test.go)
<!-- END INCLUDE -->

# Command Line Tool

The `json-schema` command line tool provides schema validation and code generation:

```bash
# Install the tool
go install github.com/lestrrat-go/json-schema/cmd/json-schema@latest

# Validate a schema file
json-schema lint schema.json

# Generate validator code
json-schema gen-validator --name UserValidator schema.json

# Read from stdin
echo '{"type": "string"}' | json-schema lint -
```

# How-to Documentation

* [API documentation](https://pkg.go.dev/github.com/lestrrat-go/json-schema)
* [Examples](./examples)

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

# Contributions

## Issues

For bug reports and feature requests, please try to follow the issue templates as much as possible.
For either bug reports or feature requests, failing tests are even better.

## Pull Requests

Please make sure to include tests that exercise the changes you made.

If you are editing auto-generated files (those files with the `_gen.go` suffix, please make sure that you do the following:

1. Edit the generator, not the generated files (e.g. internal/cmd/genobjects/main.go)
2. Run `./gen.sh` to generate the new code
3. Commit _both_ the generator _and_ the generated files

## Discussions / Usage

Please try GitHub Discussions for usage questions and general discussion.