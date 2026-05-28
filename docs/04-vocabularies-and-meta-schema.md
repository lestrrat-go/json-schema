# Vocabularies & the Meta-Schema

## Vocabularies in one minute

JSON Schema 2020-12 groups keywords into *vocabularies* (core, applicator, validation, format-annotation, format-assertion, content, unevaluated, meta-data). A schema's `$vocabulary` declares which are in effect. This library tracks the active set on the `context.Context`.

The default set (`vocabulary.DefaultSet()`) matches the spec default — notably, **format-assertion is disabled**, so `format` is an annotation only.

## Making `format` assert

By default, `"format": "email"` (and `uuid`, `date-time`, `uri`, …) is informational and will **not** reject malformed values. This is the JSON Schema default, and it surprises people. To enforce formats, enable every standard vocabulary — including format-assertion — on the context you compile and validate with:

```go
import "github.com/lestrrat-go/json-schema/vocabulary"

ctx := vocabulary.WithSet(context.Background(), vocabulary.AllEnabled())

v, err := validator.Compile(ctx, s)   // same ctx
_, err = v.Validate(ctx, data)        // same ctx
```

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

```go
import "github.com/lestrrat-go/json-schema/meta"

// Convenience wrapper:
err := meta.Validate(ctx, map[string]any{
	"type":      "string",
	"minLength": 1,
})
// err == nil → it's a valid schema

err = meta.Validate(ctx, "not a schema") // err != nil
```

Or hold the validator directly to reuse it:

```go
v := meta.Validator() // a validator.Interface
_, err := v.Validate(ctx, schemaDocument)
```

This is useful for linting user-supplied schemas before you try to compile them. (The CLI's [`lint`](./06-command-line-tool.md) command is the command-line counterpart.)

## Next

- [Code Generation](./05-code-generation.md)
- [Command Line Tool](./06-command-line-tool.md)
