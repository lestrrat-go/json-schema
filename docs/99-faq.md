# Frequently Asked Questions

### My `format: email` (or `uuid`, `date-time`, …) isn't rejecting bad values. Why?

That's the JSON Schema 2020-12 default: `format` is an **annotation**, not an assertion, so it doesn't reject anything until you enable the format-assertion vocabulary. Compile and validate with `vocabulary.WithSet(ctx, vocabulary.AllEnabled())`. See [Vocabularies & the Meta-Schema](./04-vocabularies-and-meta-schema.md).

### How do I validate the same data many times efficiently?

`Compile` once, keep the returned `validator.Interface`, and call `Validate` as often as you like. Compilation is the expensive step; the compiled validator is safe to reuse across goroutines. See [Validating Data](./02-validating.md).

### How do I validate a JSON byte slice, or keep large integers precise?

Call `validator.ValidateJSON(ctx, v, data)` with the raw `[]byte` instead of unmarshaling yourself. It decodes with `json.Decoder.UseNumber()`, so integers larger than 2^53 (e.g. 64-bit IDs) are validated exactly rather than being rounded by `float64`; integers outside the `int64` range are reported as an error. The input must be a single top-level JSON value with no trailing data. See [Validating raw JSON text](./02-validating.md#validating-raw-json-text).

### Can I just unmarshal a schema from JSON instead of building it?

Yes. `*schema.Schema` implements `json.Unmarshaler`, so `json.Unmarshal(buf, &s)` is all it takes. A schema loaded from JSON and one built with `NewBuilder()` behave identically. See [Building Schemas](./01-building-schemas.md).

### Why is there no `schema.Validate(...)`?

Schemas are inert data; validation lives in the `validator` package. You compile a `*schema.Schema` into a `validator.Interface` and call `Validate` on that. This separation lets the validator be optimized and reused independently of the schema. See the [docs index](./README.md).

### How do I validate that a document is itself a valid JSON Schema?

Use the `meta` package: `meta.Validate(ctx, document)` or `meta.Validator()`. It runs the document against the pre-compiled 2020-12 meta-schema. See [Vocabularies & the Meta-Schema](./04-vocabularies-and-meta-schema.md).

### My `$ref` to an external URL won't resolve.

By design, a resolver **resolves only from memory by default** — it will not fetch over the network or read the filesystem unless you opt in. You have two choices:

- **Preload (recommended):** create a `schema.Resolver`, register the target document(s) with `RegisterFS`/`RegisterDocument`, and pass it to `Compile` with `validator.WithResolver(r)`.
- **Opt into live access:** `schema.NewResolver(schema.WithResolver(schema.HTTPResolver()))` for HTTP/HTTPS, or `schema.WithResolver(schema.DirResolver("."))` / `schema.WithResolver(schema.FSResolver(fsys))` for files.

See [References](./03-references.md).

### I configured something but nothing changed.

The resolver and vocabulary set are **`Compile` options** — pass them to `validator.Compile(ctx, schema, validator.WithResolver(r), validator.WithVocabularySet(vs))`, not after the fact. The trace logger is carried on the `context.Context` (`validator.WithTraceSlog`) and read while validating. Configuring the wrong stage is the most common cause of "it didn't take effect". See [Validating Data](./02-validating.md#configuring-compile-and-validate).

### Why did my recursive schema fail to compile with a circular-reference error?

A `$ref` cycle that consumes no data (it would recurse forever regardless of input) is rejected at compile time. Recursion that is bounded by the data being validated (e.g. a tree that's only as deep as the input) is fine. See [References](./03-references.md#recursive-schemas).

### How do I see why validation failed, beyond the error string?

Attach a trace logger with `validator.WithTraceSlog(ctx, logger)` before compiling/validating. It logs the validation walk keyword by keyword. See [Validating Data](./02-validating.md#tracing).

### Is there a way to avoid compilation cost entirely in production?

Yes — generate validator code ahead of time with `json-schema gen-validator` (or `validator.NewCodeGenerator()`), commit the output, and call the generated builder directly. See [Code Generation](./05-code-generation.md).
