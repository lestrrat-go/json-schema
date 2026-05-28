<!-- Agent-consumed file. Verify test names/paths before relying on them. -->

# Testing

Assertions use `github.com/stretchr/testify/require` (not `assert`).

## Conformance suite

The authority for correctness is the upstream [JSON Schema Test Suite](https://github.com/json-schema-org/JSON-Schema-Test-Suite). It is **not committed**; a script vendors it into `tests/`:

```bash
./init-test-suite.sh           # clone into tests/ (idempotent)
./init-test-suite.sh <commit>  # pin/update to a specific suite commit
```

If `tests/` is missing, `TestSpecificationCompliance` finds no cases. Run `init-test-suite.sh` before concluding a compliance test "can't run".

`schema_compliance_test.go`:

- `TestSpecificationCompliance` walks the suite's draft2020-12 cases. It **skips in `-short` mode** and **skips any path containing `optional/`** (formats, ECMAScript regex, etc. — the only part not yet fully supported).
- Remote refs: the suite's `tests/remotes/` tree is preloaded via the resolver (`loadRemotes` / `newSuiteResolver`) and served logically at `http://localhost:1234/...`, so `$ref`s to remotes resolve offline. This uses `Resolver.RegisterDocument` (see references.md).
- Status: the entire **required** 2020-12 suite passes (1723 pass / 0 fail / 0 skip at last count). A new failure in this test is a real regression, not a flaky case.

Other top-level tests are hand-written feature tests (`schema_references_test.go`, `resolver_test.go`, `schema_metaschema_test.go`, `boolean_schema_test.go`, the `validator/` package tests, the `meta/` tests, etc.). Runnable usage examples live in `examples/` as Go `Example` functions and table tests.

## Running

```bash
go test ./...                 # everything (needs tests/ for compliance)
go test -short ./...          # skip the full suite walk
go test ./validator/...       # validator package only
go test ./meta/...            # meta-schema + $dynamicRef edge cases — run after touching reference resolution
go test -run TestSpecificationCompliance .   # the conformance walk
```

No build tags, no `GOEXPERIMENT`.

## Debugging validation with a trace

`validator.WithTraceSlog(ctx, logger)` attaches a structured `*slog.Logger` that records the validation walk. Wire it before `Compile`/`Validate` and gate it on `testing.Verbose()` so normal runs stay quiet:

```go
logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
ctx := validator.WithTraceSlog(context.Background(), logger)
v, _ := validator.Compile(ctx, s)
_, err := v.Validate(ctx, data)
```

This is the fastest way to see *which* keyword/branch rejected an input when an error message alone is ambiguous.

## Meta-schema tests are the canary for dynamic refs

The generated `meta` validator has no live resolver at runtime, so it exercises the context-carried dynamic-anchor registry path (`meta/meta_test.go`). Any change to how validators handle `$dynamicRef` without a resolver should be checked with `go test ./meta/...`.
