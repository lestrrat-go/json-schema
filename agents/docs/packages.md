<!-- Agent-consumed file. Keep terse, unambiguous, machine-parseable. -->

# Package Map

Module: `github.com/lestrrat-go/json-schema`. Single module, sub-packages below. Root imported as `schema`.

## schema (root)

JSON Schema data type + fluent builder + reference resolver + context helpers. Pure data; no validation logic.

- **Version** — const `"https://json-schema.org/draft/2020-12/schema"` (schema.go)
- **Schema** struct — generated in `schema_gen.go`. Inert; built or unmarshaled. Implements `json.Marshaler`/`json.Unmarshaler`.
- **New() \*Schema** — empty schema (`schema_gen.go`)
- **NewBuilder() \*Builder** / **(\*Builder) Build() (\*Schema, error)** / **MustBuild() \*Schema** / **Clone(\*Schema) \*Builder** / **Reset(FieldFlag) \*Builder** (`builder_gen.go`)
- Builder is chainable; one method per keyword. Notable: `Schema()`, `ID()`, `Anchor()`, `DynamicAnchor()`, `Reference()`, `DynamicReference()`, `Types(...PrimitiveType)`, `Property(name, *Schema)`, `Properties(...)`, `PatternProperty()`, `Required(...)`, `AdditionalProperties()`, `Items()`, `PrefixItems()`, `Contains()`, `AllOf()/AnyOf()/OneOf()/Not()`, `IfSchema()/ThenSchema()/ElseSchema()`, `Definitions(name, *Schema)`, `Minimum()/Maximum()/MultipleOf()`, `MinLength()/MaxLength()/Pattern()/Format()`, `Enum()/Const()/Default()`, `ContentEncoding()/ContentMediaType()/ContentSchema()`, `Vocabulary()`. Each has a `ResetXxx()`.
- **BoolSchema** type + **TrueSchema() BoolSchema** / **FalseSchema() BoolSchema** + **SchemaOrBool** interface — JSON Schema's `true`/`false` schemas (schema.go)
- **PrimitiveType** enum (primitives.go): `NullType`, `BooleanType`, `IntegerType`, `NumberType`, `StringType`, `ArrayType`, `ObjectType` (+ `InvalidType`). `NewPrimitiveType(string)`, `IsScalarPrimitiveType()`, `PrimitiveTypes` slice type.
- Convenience constructors → `*Builder` (patterns.go): `Email()`, `URI()`, `UUID()`, `Date()`, `DateTime()`, `NonEmptyString()`, `AlphanumericString()`, `PositiveNumber()`, `PositiveInteger()`, `Enum(...any)`, `OneOf(...*Schema)`, `AnyOf(...*Schema)`, `AllOf(...*Schema)`, `Optional(*Schema)` (schema-or-null).
- Field bitfield: `FieldFlag` constants `XxxField` (one per keyword) + grouped sets `StringConstraintFields`, `NumericConstraintFields`, `ObjectConstraintFields`, `ArrayConstraintFields`, `CompositionFields`, `ConditionalFields`, `ContentFields`, etc. `(*Schema).Has(FieldFlag) bool`, `HasAny(FieldFlag) bool`, plus `HasXxx()` per keyword.
- URI: **ResolveURI(base, ref string) string** (uri.go, RFC 3986).
- Resolver: **NewResolver(...ResolverOption) \*Resolver** (external access is opt-in; bare resolver is in-memory only); methods `RegisterRoot(*Schema)`, `RegisterDocument(uri string, root *Schema)`, `RegisterFS(baseURI string, fs.FS) error`, `ResourceFor(uri string) *Schema`, `ResolveReference(ctx, dst any, ref string) error` (resolver.go). **FindDynamicAnchor(resource *Schema, name string) *Schema** (registry.go).
- Resolver options (resolver_options.go): **ResolverOption**, **WithResolver(jsref.Resolver)**, and opt-in resolver factories **HTTPResolver() jsref.Resolver**, **FSResolver(fs.FS) jsref.Resolver**, **DirResolver(dir string) jsref.Resolver**.

## validator/

Compile schemas to validators; validate; generate validator code.

- **Interface** — `Validate(ctx context.Context, value any) (Result, error)`. Non-nil error == validation failure (validator.go).
- **Result** — `type Result = any`. Annotation payload (e.g. `*ObjectResult`, `*ArrayResult`) used to track evaluated properties/items for `unevaluated*`.
- **Compile(ctx context.Context, s *schema.Schema, ...CompileOption) (Interface, error)** (compiler.go) — single entry point. Sets up resolver/root/base/vocabulary in ctx; rebases `$id`; wraps `$dynamicAnchor` schemas in a `dynamicScopeValidator`.
- **ValidateJSON(ctx, v Interface, data []byte, ...ValidateOption) (Result, error)** (json.go) — decode raw JSON (`UseNumber`, rejecting empty/trailing input) and validate via `v`. Numbers stay `json.Number` so large integers keep precision; helpers in `numeric.go` (`isNumeric`/`isJSONNumber`/`numericFloat`/`numericInt`) interpret them.
- Hand-written validator builders (each `Build()/MustBuild()`): **Object()** (`*ObjectValidatorBuilder`: `Properties()`, `PatternProperties()`, `AdditionalProperties()`, `PropertyNames()`, `Required()`, `DependentRequired()`, `DependentSchemas()`, `Min/MaxProperties()`, `UnevaluatedProperties()`, `StrictObjectType()`), **String()** (`MinLength/MaxLength/Pattern/Format`), **Array()** (`Items/PrefixItems/Contains/Min/MaxItems/UniqueItems/Min/MaxContains/AdditionalItems/UnevaluatedItems`), **Boolean()**, **Null() Interface**.
- Generated numeric builders: **Integer()** (`*IntegerValidatorBuilder`; methods take `int64`), **Number()** (`float64`) — `Minimum/Maximum/ExclusiveMinimum/ExclusiveMaximum/MultipleOf` (`int_gen.go`, `number_gen.go`).
- **PropPair(name string, v Interface) PropertyPair** — for `ObjectValidatorBuilder.Properties(...)` (object.go).
- Composition (multi.go): **AllOf/AnyOf/OneOf(...Interface) Interface**.
- Code generation: **CodeGenerator** interface — `Generate(dst io.Writer, v Interface) error`; **NewCodeGenerator() CodeGenerator** (codegen_core.go). Emits Go builder source from a compiled validator.
- Tracing: **WithTraceSlog(ctx, *slog.Logger) context.Context** (conditional.go) — structured validation trace.
- **WithDependentSchemas(ctx, map[string]Interface)** / **DependentSchemasFromContext(ctx)** (validator.go).
- Result helpers: `NewObjectResult()`, `NewArrayResult(size ...int)` and their `EvaluatedProperties/Items`/`SetEvaluatedProperty/Item` methods.

## vocabulary/

2020-12 vocabularies; per-context enable/disable. Drives whether keywords like `format` assert or merely annotate.

- **VocabularySet** — `Enable/Disable/IsEnabled/IsKeywordEnabled`.
- **DefaultSet() \*VocabularySet** — standard default (**format-assertion disabled**: `format` is annotation-only).
- **AllEnabled() \*VocabularySet** — every standard vocabulary incl. format-assertion (makes `format` assert).
- **NewVocabularySet()**, **DefaultRegistry() \*Registry**, **ExtractVocabularySet(*schema.Schema)**, **ResolveVocabularyFromMetaschema(ctx, uri)**, **ValidateVocabularyURI(uri)**.
- Context: **WithSet(ctx, *VocabularySet)** / **SetFromContext(ctx)** / **IsKeywordEnabledInContext(ctx, keyword)**.
- URI consts: `CoreURL`, `ApplicatorURL`, `UnevaluatedURL`, `ValidationURL`, `FormatAnnotationURL`, `FormatAssertionURL`, `ContentURL`, `MetaDataURL`.

## keywords/

String constants for every JSON Schema keyword (avoid hardcoded literals). Core/applicator/unevaluated/validation/format/content/metadata groups. E.g. `keywords.Type`, `keywords.Properties`, `keywords.Reference` (`"$ref"`), `keywords.DynamicReference`.

## meta/

Pre-compiled 2020-12 meta-schema validator (validate that a document *is* a valid schema).

- **Validator() validator.Interface** — the cached meta-schema validator (hand-written `meta.go`; value comes from generated `meta_gen.go`).
- **Validate(ctx, jsonSchemaDocument any) error** — convenience wrapper.
- Registers `metaValidator` under the `"meta"` dynamic anchor so `$dynamicRef: "#meta"` recurses (see references.md).

## cmd/json-schema/

CLI (`urfave/cli/v3`).

- `lint [filename|-]` — unmarshal + `validator.Compile`; prints `Schema <src> is valid` or the failure.
- `gen-validator [filename|-]` `--name <var>` (default `val`) — compile, then `NewCodeGenerator().Generate`; prints `<name> := <builder code>` formatted with `go/format`.

## internal/ (not public API)

- `internal/cmd/genobjects/` — generates `schema_gen.go` + `builder_gen.go` from `objects.yml`.
- `internal/cmd/genmeta/` — generates `meta/meta_gen.go` from the embedded meta-schema.
- `internal/field/` — `FieldFlag` bitfield definitions.

## External dependencies

`lestrrat-go/jsref/v2` (stacked reference resolvers: in-document, HTTP, FS, object), `lestrrat-go/codegen` (codegen helpers), `urfave/cli/v3` (CLI), `stretchr/testify` (tests).
