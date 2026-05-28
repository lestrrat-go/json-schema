# Command Line Tool

The `json-schema` CLI does two things: check that a schema is valid (`lint`), and emit pre-compiled validator code (`gen-validator`).

## Install

```bash
go install github.com/lestrrat-go/json-schema/cmd/json-schema@latest
```

## `lint` — validate a schema document

Parses a schema and compiles it, reporting any structural or semantic problem. Reads a file, or `-` for stdin.

```bash
json-schema lint schema.json
# Schema schema.json is valid

echo '{"type": "string"}' | json-schema lint -
# Schema stdin is valid
```

If the document is not a valid schema or cannot be compiled, `lint` prints the error and exits non-zero — handy as a pre-commit or CI check on schema files.

## `gen-validator` — emit validator code

Compiles a schema and prints Go source that rebuilds the validator directly, so production code can skip compilation. Reads a file or `-` for stdin.

```bash
json-schema gen-validator --name UserValidator user-schema.json
```

`--name` sets the generated variable name (default `val`). See [Code Generation](./05-code-generation.md) for a full example of the output and how to wire it into `go:generate`.

```bash
# from stdin, default variable name "val"
echo '{"type":"string","minLength":1}' | json-schema gen-validator -
# val := validator.String().MinLength(1).MustBuild()
```

## Summary

| Command | Purpose | Key flag |
|---------|---------|----------|
| `lint [file\|-]` | Verify a schema is valid | — |
| `gen-validator [file\|-]` | Print Go validator code | `--name <var>` (default `val`) |
