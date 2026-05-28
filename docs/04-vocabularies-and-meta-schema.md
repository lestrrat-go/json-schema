# Vocabularies & the Meta-Schema

## Vocabularies in one minute

JSON Schema 2020-12 groups keywords into *vocabularies* (core, applicator, validation, format-annotation, format-assertion, content, unevaluated, meta-data). A schema's `$vocabulary` declares which are in effect. This library tracks the active set on the `context.Context`.

The default set (`vocabulary.DefaultSet()`) matches the spec default — notably, **format-assertion is disabled**, so `format` is an annotation only.

## Making `format` assert

By default, `"format": "email"` (and `uuid`, `date-time`, `uri`, …) is informational and will **not** reject malformed values. This is the JSON Schema default, and it surprises people. To enforce formats, enable every standard vocabulary — including format-assertion — with `vocabulary.WithSet(ctx, vocabulary.AllEnabled())` on the context you compile *and* validate with:

<!-- INCLUDE(examples/doc_format_test.go) -->
```go
package examples_test

import (
  "context"
  "fmt"

  schema "github.com/lestrrat-go/json-schema"
  "github.com/lestrrat-go/json-schema/validator"
  "github.com/lestrrat-go/json-schema/vocabulary"
)

// Example_docFormatAssertion shows that "format" is an annotation by default (it
// does not reject bad values) and only asserts when the format-assertion
// vocabulary is enabled via vocabulary.AllEnabled.
func Example_docFormatAssertion() {
  s := schema.NewBuilder().
    Types(schema.ObjectType).
    Property("email", schema.Email().MustBuild()).
    MustBuild()

  data := map[string]any{"email": "not-an-email"}

  // Default vocabularies: format is annotation-only, so this passes.
  def := context.Background()
  v, _ := validator.Compile(def, s)
  _, err := v.Validate(def, data)
  fmt.Println("default set: valid:", err == nil)

  // All vocabularies enabled: format-assertion is on, so this fails.
  strict := context.Background()
  vs, _ := validator.Compile(strict, s, validator.WithVocabularySet(vocabulary.AllEnabled()))
  _, err = vs.Validate(strict, data)
  fmt.Println("all enabled: valid:", err == nil)
  // Output:
  // default set: valid: true
  // all enabled: valid: false
}
```
source: [examples/doc_format_test.go](https://github.com/lestrrat-go/json-schema/blob/main/examples/doc_format_test.go)
<!-- END INCLUDE -->

With `vocabulary.AllEnabled()`, a `schema.Email()` property rejects `"not-an-email"`; with the default set, it does not.

| Set | `format` behavior |
|-----|-------------------|
| `vocabulary.DefaultSet()` (default) | annotation only — never rejects |
| `vocabulary.AllEnabled()` | assertion — rejects malformed values |

Use the same configured context for `Compile` and `Validate`.

## Selecting vocabularies explicitly

`vocabulary.NewVocabularySet()` plus `Enable`/`Disable` lets you build a custom set; `vocabulary.ExtractVocabularySet(schema)` derives the set declared by a schema's `$vocabulary`. The standard vocabulary URIs are available as constants (`vocabulary.FormatAssertionURL`, `vocabulary.ValidationURL`, …).

## Validating that a document *is* a schema (the meta-schema)

The `meta` package answers a different question: "is this JSON a valid JSON Schema 2020-12 document?" It ships a pre-compiled meta-schema validator, so you do not pay compilation cost.

<!-- INCLUDE(examples/doc_meta_test.go) -->
```go
package examples_test

import (
  "context"
  "fmt"

  "github.com/lestrrat-go/json-schema/meta"
)

// Example_docMeta validates that a document is itself a valid JSON Schema 2020-12
// document, using the pre-compiled meta-schema validator in the meta package.
func Example_docMeta() {
  ctx := context.Background()

  validSchema := map[string]any{"type": "string", "minLength": 1}
  fmt.Println("valid schema:  ", meta.Validate(ctx, validSchema) == nil)

  notASchema := "not a schema"
  fmt.Println("invalid schema:", meta.Validate(ctx, notASchema) == nil)
  // Output:
  // valid schema:   true
  // invalid schema: false
}
```
source: [examples/doc_meta_test.go](https://github.com/lestrrat-go/json-schema/blob/main/examples/doc_meta_test.go)
<!-- END INCLUDE -->

`meta.Validate(ctx, doc)` is a convenience wrapper; `meta.Validator()` returns the underlying reusable `validator.Interface` if you want to hold it directly. This is useful for linting user-supplied schemas before you try to compile them. (The CLI's [`lint`](./06-command-line-tool.md) command is the command-line counterpart.)

## Next

- [Code Generation](./05-code-generation.md)
- [Command Line Tool](./06-command-line-tool.md)
