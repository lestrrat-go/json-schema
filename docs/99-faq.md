# Frequently Asked Questions

### My `format: email` (or `uuid`, `date-time`, …) isn't rejecting bad values. Why?

That's the JSON Schema 2020-12 default: `format` is an **annotation**, not an assertion, so it doesn't reject anything until you enable the format-assertion vocabulary. Compile and validate with `vocabulary.WithSet(ctx, vocabulary.AllEnabled())`. See [Vocabularies & the Meta-Schema](./04-vocabularies-and-meta-schema.md).

### How do I validate the same data many times efficiently?

`Compile` once, keep the returned `validator.Interface`, and call `Validate` as often as you like. Compilation is the expensive step; the compiled validator is safe to reuse across goroutines. See [Validating Data](./02-validating.md).

### Can I just unmarshal a schema from JSON instead of building it?

Yes. `*schema.Schema` implements `json.Unmarshaler`, so `json.Unmarshal(buf, &s)` is all it takes. A schema loaded from JSON and one built with `NewBuilder()` behave identically. See [Building Schemas](./01-building-schemas.md).

### Why is there no `schema.Validate(...)`?

Schemas are inert data; validation lives in the `validator` package. You compile a `*schema.Schema` into a `validator.Interface` and call `Validate` on that. This separation lets the validator be optimized and reused independently of the schema. See the [docs index](./README.md).

### How do I validate that a document is itself a valid JSON Schema?

Use the `meta` package: `meta.Validate(ctx, document)` or `meta.Validator()`. It runs the document against the pre-compiled 2020-12 meta-schema. See [Vocabularies & the Meta-Schema](./04-vocabularies-and-meta-schema.md).

### My `$ref` to an external URL won't resolve.

External references need a resolver that knows about the target document. Create a `schema.Resolver`, register the document(s) with `RegisterFS`/`RegisterDocument`, put it on the context with `schema.WithResolver`, and use that context for both `Compile` and `Validate`. See [References](./03-references.md).

### I configured the context but nothing changed.

Optional behavior (resolver, vocabularies, tracing) is read from the context at **both** compile and validate time. Pass the *same* configured context to `Compile` and to `Validate`. Configuring one but not the other is the most common cause of "it didn't take effect".

### Why did my recursive schema fail to compile with a circular-reference error?

A `$ref` cycle that consumes no data (it would recurse forever regardless of input) is rejected at compile time. Recursion that is bounded by the data being validated (e.g. a tree that's only as deep as the input) is fine. See [References](./03-references.md#recursive-schemas).

### How do I see why validation failed, beyond the error string?

Attach a trace logger with `validator.WithTraceSlog(ctx, logger)` before compiling/validating. It logs the validation walk keyword by keyword. See [Validating Data](./02-validating.md#tracing).

### Is there a way to avoid compilation cost entirely in production?

Yes — generate validator code ahead of time with `json-schema gen-validator` (or `validator.NewCodeGenerator()`), commit the output, and call the generated builder directly. See [Code Generation](./05-code-generation.md).
