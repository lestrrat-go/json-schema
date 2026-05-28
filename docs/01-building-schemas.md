# Building Schemas

A `*schema.Schema` is inert data. You can produce one two ways — programmatically with the builder, or by unmarshaling JSON. Both yield the same thing, and both round-trip to identical, stably-sorted JSON.

## The fluent builder

`schema.NewBuilder()` returns a `*Builder` with one chainable method per JSON Schema keyword. Finish with `Build() (*Schema, error)` or `MustBuild() *Schema` (panics on error). The example below builds an object schema and marshals it back to JSON:

<!-- INCLUDE(examples/doc_builder_test.go) -->
```go
package examples_test

import (
  "encoding/json"
  "fmt"
  "os"

  schema "github.com/lestrrat-go/json-schema"
)

// Example_docBuilder builds an object schema with the fluent builder and marshals
// it to JSON. Object keys are emitted in a stable, sorted order, so the result is
// deterministic and round-trips cleanly.
func Example_docBuilder() {
  s := schema.NewBuilder().
    Schema(schema.Version).
    ID("https://example.com/user").
    Types(schema.ObjectType).
    Property("name", schema.NewBuilder().Types(schema.StringType).MinLength(1).MustBuild()).
    Property("age", schema.NewBuilder().Types(schema.IntegerType).Minimum(0).MustBuild()).
    Required("name").
    AdditionalProperties(schema.FalseSchema()).
    MustBuild()

  enc := json.NewEncoder(os.Stdout)
  enc.SetIndent("", "  ")
  if err := enc.Encode(s); err != nil {
    fmt.Println("encode failed:", err)
  }
  // Output:
  // {
  //   "$id": "https://example.com/user",
  //   "$schema": "https://json-schema.org/draft/2020-12/schema",
  //   "additionalProperties": false,
  //   "properties": {
  //     "age": {
  //       "minimum": 0,
  //       "type": "integer"
  //     },
  //     "name": {
  //       "minLength": 1,
  //       "type": "string"
  //     }
  //   },
  //   "required": [
  //     "name"
  //   ],
  //   "type": "object"
  // }
}
```
source: [examples/doc_builder_test.go](https://github.com/lestrrat-go/json-schema/blob/main/examples/doc_builder_test.go)
<!-- END INCLUDE -->

`schema.Version` is the constant `"https://json-schema.org/draft/2020-12/schema"`.

### Types

`Types` is variadic and takes `PrimitiveType` constants — a schema may permit more than one type, e.g. `Types(schema.StringType, schema.NullType)` for a string-or-null. The constants are `NullType`, `BooleanType`, `IntegerType`, `NumberType`, `StringType`, `ArrayType`, `ObjectType`.

### Keyword coverage

The builder has a method for every 2020-12 keyword. A non-exhaustive map:

| Area | Methods |
|------|---------|
| Identity | `Schema`, `ID`, `Anchor`, `DynamicAnchor`, `Comment`, `Vocabulary` |
| References | `Reference` (`$ref`), `DynamicReference` (`$dynamicRef`), `Definitions` |
| Strings | `MinLength`, `MaxLength`, `Pattern`, `Format` |
| Numbers | `Minimum`, `Maximum`, `ExclusiveMinimum`, `ExclusiveMaximum`, `MultipleOf` |
| Objects | `Property`, `Properties`, `PatternProperty`, `AdditionalProperties`, `PropertyNames`, `Required`, `MinProperties`, `MaxProperties`, `DependentRequired`, `DependentSchemas`, `UnevaluatedProperties` |
| Arrays | `Items`, `PrefixItems`, `Contains`, `MinItems`, `MaxItems`, `UniqueItems`, `MinContains`, `MaxContains`, `UnevaluatedItems` |
| Composition | `AllOf`, `AnyOf`, `OneOf`, `Not` |
| Conditionals | `IfSchema`, `ThenSchema`, `ElseSchema` |
| Values | `Enum`, `Const`, `Default` |
| Content | `ContentEncoding`, `ContentMediaType`, `ContentSchema` |

Every keyword method has a matching `ResetXxx()` that clears it.

### Boolean schemas

JSON Schema allows `true` and `false` as whole schemas (accept-anything / reject-everything). Use `schema.TrueSchema()` and `schema.FalseSchema()` wherever a sub-schema is accepted — for example `AdditionalProperties(schema.FalseSchema())` forbids unlisted properties (as in the builder example above).

## Convenience constructors

For common shapes, the package ships one-line constructors that each return a pre-seeded `*Builder` (so you still call `MustBuild()`):

| Constructor | Produces |
|-------------|----------|
| `schema.NonEmptyString()` | string, `minLength: 1` |
| `schema.AlphanumericString()` | string with an alphanumeric pattern |
| `schema.Email()` | string, `format: email` |
| `schema.URL()` | string, `format: uri` |
| `schema.UUID()` | string, `format: uuid` |
| `schema.Date()` / `schema.DateTime()` | string, `format: date` / `date-time` |
| `schema.PositiveInteger()` | integer, `minimum: 0` |
| `schema.PositiveNumber()` | number, `minimum: 0` |
| `schema.Enum(vals...)` | `enum` of the given values |
| `schema.OneOf(...)` / `AnyOf(...)` / `AllOf(...)` | composition over the given `*Schema`s |
| `schema.Optional(s)` | accepts `s` **or** `null` |

<!-- INCLUDE(examples/doc_convenience_test.go) -->
```go
package examples_test

import (
  "context"
  "fmt"

  schema "github.com/lestrrat-go/json-schema"
  "github.com/lestrrat-go/json-schema/validator"
)

// Example_docConvenience composes several one-line convenience constructors —
// PositiveInteger, Optional, NonEmptyString and Enum — into one object schema.
// Optional(s) accepts either s or null.
func Example_docConvenience() {
  s := schema.NewBuilder().
    Types(schema.ObjectType).
    Property("id", schema.PositiveInteger().MustBuild()).
    Property("nickname", schema.Optional(schema.NonEmptyString().MustBuild()).MustBuild()).
    Property("role", schema.Enum("admin", "user").MustBuild()).
    Required("id", "role").
    MustBuild()

  ctx := context.Background()
  v, _ := validator.Compile(ctx, s)
  check := func(data map[string]any) bool {
    _, err := v.Validate(ctx, data)
    return err == nil
  }

  fmt.Println("admin:        ", check(map[string]any{"id": 1, "role": "admin"}))
  fmt.Println("null nickname:", check(map[string]any{"id": 1, "role": "user", "nickname": nil}))
  fmt.Println("bad role:     ", check(map[string]any{"id": 1, "role": "root"}))
  // Output:
  // admin:         true
  // null nickname: true
  // bad role:      false
}
```
source: [examples/doc_convenience_test.go](https://github.com/lestrrat-go/json-schema/blob/main/examples/doc_convenience_test.go)
<!-- END INCLUDE -->

> Note: `format` keywords (from `Email()`, `UUID()`, etc.) are **annotations by default** and do not reject bad values until you enable format-assertion. See [Vocabularies](./04-vocabularies-and-meta-schema.md).

## Loading a schema from JSON

`*schema.Schema` implements `json.Unmarshaler`, so loading is just `json.Unmarshal`. A schema built programmatically and the equivalent schema loaded from JSON are interchangeable — they compile and validate identically:

<!-- INCLUDE(examples/doc_loadjson_test.go) -->
```go
package examples_test

import (
  "context"
  "encoding/json"
  "fmt"

  schema "github.com/lestrrat-go/json-schema"
  "github.com/lestrrat-go/json-schema/validator"
)

// Example_docLoadJSON loads a schema authored as JSON. *schema.Schema implements
// json.Unmarshaler, so json.Unmarshal is all it takes; the result compiles and
// validates exactly like a schema built with the fluent builder.
func Example_docLoadJSON() {
  const doc = `{
    "type": "object",
    "properties": { "city": { "type": "string", "minLength": 1 } },
    "required": ["city"]
  }`

  var s schema.Schema
  if err := json.Unmarshal([]byte(doc), &s); err != nil {
    fmt.Println("parse failed:", err)
    return
  }

  ctx := context.Background()
  v, err := validator.Compile(ctx, &s)
  if err != nil {
    fmt.Println("compile failed:", err)
    return
  }

  _, err = v.Validate(ctx, map[string]any{"city": "Kyoto"})
  fmt.Println("with city:   ", err == nil)
  _, err = v.Validate(ctx, map[string]any{})
  fmt.Println("without city:", err == nil)
  // Output:
  // with city:    true
  // without city: false
}
```
source: [examples/doc_loadjson_test.go](https://github.com/lestrrat-go/json-schema/blob/main/examples/doc_loadjson_test.go)
<!-- END INCLUDE -->

## Serializing a schema

`*schema.Schema` also implements `json.Marshaler`. Object keys are emitted in a stable, sorted order, so marshaling is deterministic and round-trips cleanly — the [fluent builder example](#the-fluent-builder) above marshals a schema and shows the resulting JSON.

## Next

- [Validating Data](./02-validating.md)
- [References](./03-references.md) — reuse sub-schemas with `$ref` and `$defs`
