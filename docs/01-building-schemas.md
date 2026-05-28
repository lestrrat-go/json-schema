# Building Schemas

A `*schema.Schema` is inert data. You can produce one two ways — programmatically with the builder, or by unmarshaling JSON. Both yield the same thing, and both round-trip to identical, stably-sorted JSON.

## The fluent builder

`schema.NewBuilder()` returns a `*Builder` with one chainable method per JSON Schema keyword. Finish with `Build() (*Schema, error)` or `MustBuild() *Schema` (panics on error).

```go
s := schema.NewBuilder().
	Schema(schema.Version).                 // $schema: the 2020-12 dialect URI
	ID("https://example.com/user").         // $id
	Types(schema.ObjectType).               // type: object
	Property("name", schema.NewBuilder().Types(schema.StringType).MinLength(1).MustBuild()).
	Property("age", schema.NewBuilder().Types(schema.IntegerType).Minimum(0).MustBuild()).
	Required("name").
	AdditionalProperties(schema.FalseSchema()).
	MustBuild()
```

`schema.Version` is the constant `"https://json-schema.org/draft/2020-12/schema"`.

### Types

`Types` is variadic and takes `PrimitiveType` constants — a schema may permit more than one type:

```go
schema.NewBuilder().Types(schema.StringType, schema.NullType).MustBuild()
```

The constants are `NullType`, `BooleanType`, `IntegerType`, `NumberType`, `StringType`, `ArrayType`, `ObjectType`.

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

JSON Schema allows `true` and `false` as whole schemas (accept-anything / reject-everything). Use `schema.TrueSchema()` and `schema.FalseSchema()` wherever a sub-schema is accepted:

```go
schema.NewBuilder().
	Types(schema.ObjectType).
	AdditionalProperties(schema.FalseSchema()). // forbid unlisted properties
	MustBuild()
```

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

```go
s := schema.NewBuilder().
	Types(schema.ObjectType).
	Property("id", schema.PositiveInteger().MustBuild()).
	Property("nickname", schema.Optional(schema.NonEmptyString().MustBuild()).MustBuild()).
	Property("role", schema.Enum("admin", "user").MustBuild()).
	Required("id", "role").
	MustBuild()
```

> Note: `format` keywords (from `Email()`, `UUID()`, etc.) are **annotations by default** and do not reject bad values until you enable format-assertion. See [Vocabularies](./04-vocabularies-and-meta-schema.md).

## Loading a schema from JSON

`*schema.Schema` implements `json.Unmarshaler`, so loading is just `json.Unmarshal`:

```go
func loadSchema(path string) *schema.Schema {
	buf, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var s schema.Schema
	if err := json.Unmarshal(buf, &s); err != nil {
		panic(err)
	}
	return &s
}
```

A schema built programmatically and the equivalent schema loaded from JSON are interchangeable — they compile and validate identically.

## Serializing a schema

`*schema.Schema` also implements `json.Marshaler`. Object keys are emitted in a stable, sorted order, so marshaling is deterministic and round-trips cleanly:

```go
buf, _ := json.Marshal(s)
// {"$id":"...","$schema":"...","properties":{...},"required":[...],"type":"object"}
```

## Next

- [Validating Data](./02-validating.md)
- [References](./03-references.md) — reuse sub-schemas with `$ref` and `$defs`
