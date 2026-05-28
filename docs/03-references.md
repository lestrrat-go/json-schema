# References

JSON Schema lets you reuse and cross-link sub-schemas with `$ref` and `$dynamicRef`, anchored by `$id`, `$anchor`, and `$dynamicAnchor`. This library resolves all of them; this guide covers the common cases and how to point the resolver at external documents.

> [!IMPORTANT]
> **External access is opt-in.** A `schema.Resolver` resolves references **only from memory** by default — it never touches the network or the filesystem unless you explicitly enable it. An external `$ref` you have not preloaded (or whose access you have not opted into) fails to resolve rather than being silently fetched. To allow live access, pass `WithResolver(HTTPResolver())` (HTTP/HTTPS), `WithResolver(DirResolver("."))`, or `WithResolver(FSResolver(fsys))`. See [External references and the Resolver](#external-references-and-the-resolver) below.

## Reusing a sub-schema with `$ref` and `$defs`

Define a schema once under `$defs` and reference it with `$ref`:

<!-- INCLUDE(examples/doc_refdefs_test.go) -->
```go
package examples_test

import (
  "context"
  "fmt"

  schema "github.com/lestrrat-go/json-schema"
  "github.com/lestrrat-go/json-schema/validator"
)

// Example_docRefDefs defines a subschema under $defs once and references it from
// two properties with $ref. References within a document resolve automatically at
// Compile time — no resolver setup needed.
func Example_docRefDefs() {
  nameDef := schema.NonEmptyString().MustBuild()
  s := schema.NewBuilder().
    Types(schema.ObjectType).
    Definitions("name", nameDef).
    Property("firstName", schema.NewBuilder().Reference("#/$defs/name").MustBuild()).
    Property("lastName", schema.NewBuilder().Reference("#/$defs/name").MustBuild()).
    Required("firstName", "lastName").
    MustBuild()

  ctx := context.Background()
  v, _ := validator.Compile(ctx, s)
  check := func(data map[string]any) bool {
    _, err := v.Validate(ctx, data)
    return err == nil
  }

  fmt.Println("both names:  ", check(map[string]any{"firstName": "Ada", "lastName": "Lovelace"}))
  fmt.Println("empty first: ", check(map[string]any{"firstName": "", "lastName": "Lovelace"}))
  // Output:
  // both names:   true
  // empty first:  false
}
```
source: [examples/doc_refdefs_test.go](https://github.com/lestrrat-go/json-schema/blob/main/examples/doc_refdefs_test.go)
<!-- END INCLUDE -->

`Definitions(name, schema)` adds an entry under `$defs`; `Reference("#/$defs/name")` is a JSON-Pointer `$ref` to it. Both `firstName` and `lastName` validate against the same `name` definition. The equivalent JSON is:

```json
{
  "type": "object",
  "$defs": { "name": { "type": "string", "minLength": 1 } },
  "properties": {
    "firstName": { "$ref": "#/$defs/name" },
    "lastName":  { "$ref": "#/$defs/name" }
  },
  "required": ["firstName", "lastName"]
}
```

References within a document resolve automatically when you `Compile` — no extra setup.

## External references and the Resolver

When a `$ref` points at *another document* (an absolute URI, not a `#`-fragment of the current schema), the validator needs to find that document. That is the job of `schema.Resolver`: create one with `schema.NewResolver()`, register the documents it should know about, and pass it to `Compile` with `validator.WithResolver(r)`.

> **External access is opt-in.** A bare `schema.NewResolver()` resolves references **only from memory** — in-document `$id` resources and documents you preload (below). It never reaches the network or the filesystem on its own, so an external `$ref` you have not preloaded fails to resolve instead of being silently fetched. To allow live access, hand the resolver explicit resolvers:
>
> ```go
> r := schema.NewResolver(schema.WithResolver(schema.HTTPResolver()))      // allow HTTP/HTTPS
> r := schema.NewResolver(schema.WithResolver(schema.DirResolver(".")))   // allow local files under "."
> r := schema.NewResolver(schema.WithResolver(schema.FSResolver(fsys)))   // allow files from any io/fs
> ```
>
> Preloading (below) is preferred over live access for tests and reproducible builds.

### Preloading a tree of files: `RegisterFS`

`RegisterFS(baseURI, fsys)` walks any `fs.FS` and registers every `.json` file under `baseURI` joined with its path. This works with `embed.FS`, `os.DirFS`, or an in-memory `fstest.MapFS`:

<!-- INCLUDE(examples/doc_registerfs_test.go) -->
```go
package examples_test

import (
  "context"
  "fmt"
  "testing/fstest"

  schema "github.com/lestrrat-go/json-schema"
  "github.com/lestrrat-go/json-schema/validator"
)

// Example_docRegisterFS preloads schema documents from a filesystem with
// Resolver.RegisterFS, so a $ref to an external URI resolves offline. Each ".json"
// file is registered under the base URI joined with its path. The fs.FS here is an
// in-memory fstest.MapFS, but an embed.FS or os.DirFS works the same way.
func Example_docRegisterFS() {
  fsys := fstest.MapFS{
    "address.json": &fstest.MapFile{Data: []byte(
      `{"type":"object","properties":{"city":{"type":"string","minLength":1}},"required":["city"]}`,
    )},
  }

  main := schema.NewBuilder().
    Types(schema.ObjectType).
    Property("home", schema.NewBuilder().Reference("https://example.com/schemas/address.json").MustBuild()).
    Required("home").
    MustBuild()

  r := schema.NewResolver()
  if err := r.RegisterFS("https://example.com/schemas/", fsys); err != nil {
    fmt.Println("register failed:", err)
    return
  }

  ctx := context.Background()
  v, err := validator.Compile(ctx, main, validator.WithResolver(r))
  if err != nil {
    fmt.Println("compile failed:", err)
    return
  }

  _, err = v.Validate(ctx, map[string]any{"home": map[string]any{"city": "Kyoto"}})
  fmt.Println("with city:   ", err == nil)
  _, err = v.Validate(ctx, map[string]any{"home": map[string]any{}})
  fmt.Println("without city:", err == nil)
  // Output:
  // with city:    true
  // without city: false
}
```
source: [examples/doc_registerfs_test.go](https://github.com/lestrrat-go/json-schema/blob/main/examples/doc_registerfs_test.go)
<!-- END INCLUDE -->

The `$ref` to `https://example.com/schemas/address.json` resolves offline against the registered file.

### Registering a single document: `RegisterDocument` / `RegisterRoot`

- `RegisterDocument(uri, root)` preloads one document under an explicit retrieval URI. The document becomes addressable both by that URI **and** by its own canonical `$id`.
- `RegisterRoot(root)` indexes a schema's own `$id`/anchors (the root `Compile` does this for you automatically).

Preloading documents is preferred over live HTTP fetching (which is opt-in; see above) for tests and reproducible builds.

## `$id`, `$anchor`, `$dynamicAnchor`

- **`$id`** establishes a base URI; references inside that schema resolve relative to it. Set with `ID(...)`.
- **`$anchor`** names a location for plain `#name` references. Set with `Anchor(...)`.
- **`$dynamicAnchor`** / **`$dynamicRef`** implement *runtime* extension points: a `$dynamicRef` resolves against the outermost matching `$dynamicAnchor` in the current dynamic scope, which lets a base schema defer part of its definition to whatever schema referenced it. Set with `DynamicAnchor(...)` and `DynamicReference(...)`.

`$dynamicRef` is the mechanism behind recursive, extensible schemas (it is how the JSON Schema meta-schema references itself). For most application schemas, plain `$ref` + `$defs` is all you need.

## Recursive schemas

A schema may reference itself (e.g. a tree node whose children are the same node). Data-bounded recursion through `$ref` is supported — validation recurses as deep as the data goes. A reference cycle that is *not* bounded by data (a schema that would recurse forever with no input to consume) is reported as an error at compile time.

## Next

- [Vocabularies & the Meta-Schema](./04-vocabularies-and-meta-schema.md)
