# References

JSON Schema lets you reuse and cross-link sub-schemas with `$ref` and `$dynamicRef`, anchored by `$id`, `$anchor`, and `$dynamicAnchor`. This library resolves all of them; this guide covers the common cases and how to point the resolver at external documents.

## Reusing a sub-schema with `$ref` and `$defs`

Define a schema once under `$defs` and reference it with `$ref`:

```go
nameDef := schema.NonEmptyString().MustBuild()

s := schema.NewBuilder().
	Types(schema.ObjectType).
	Definitions("name", nameDef).
	Property("firstName", schema.NewBuilder().Reference("#/$defs/name").MustBuild()).
	Property("lastName", schema.NewBuilder().Reference("#/$defs/name").MustBuild()).
	Required("firstName", "lastName").
	MustBuild()
```

`Definitions(name, schema)` adds an entry under `$defs`; `Reference("#/$defs/name")` is a JSON-Pointer `$ref` to it. Both `firstName` and `lastName` now validate against the same `name` definition. The equivalent JSON is:

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

When a `$ref` points at *another document* (an absolute URI, not a `#`-fragment of the current schema), the validator needs to find that document. That is the job of `schema.Resolver`. Create one, register the documents it should know about, and put it on the context:

```go
import "context"

r := schema.NewResolver()
// ... register documents (below) ...

ctx := schema.WithResolver(context.Background(), r)
v, err := validator.Compile(ctx, mainSchema)
// validate with the SAME ctx
_, err = v.Validate(ctx, data)
```

### Preloading a tree of files: `RegisterFS`

`RegisterFS(baseURI, fsys)` walks any `fs.FS` and registers every `.json` file under `baseURI` joined with its path. This works with `embed.FS`, `os.DirFS`, or an in-memory `fstest.MapFS`:

```go
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
	panic(err)
}

ctx := schema.WithResolver(context.Background(), r)
v, _ := validator.Compile(ctx, main)
_, err := v.Validate(ctx, map[string]any{"home": map[string]any{"city": "Kyoto"}})
```

Now a `$ref` to `https://example.com/schemas/address.json` resolves offline against the registered file.

### Registering a single document: `RegisterDocument` / `RegisterRoot`

- `RegisterDocument(uri, root)` preloads one document under an explicit retrieval URI. The document becomes addressable both by that URI **and** by its own canonical `$id`.
- `RegisterRoot(root)` indexes a schema's own `$id`/anchors (the root `Compile` does this for you automatically).

Preloading documents is preferred over live HTTP fetching for tests and reproducible builds.

## `$id`, `$anchor`, `$dynamicAnchor`

- **`$id`** establishes a base URI; references inside that schema resolve relative to it. Set with `ID(...)`.
- **`$anchor`** names a location for plain `#name` references. Set with `Anchor(...)`.
- **`$dynamicAnchor`** / **`$dynamicRef`** implement *runtime* extension points: a `$dynamicRef` resolves against the outermost matching `$dynamicAnchor` in the current dynamic scope, which lets a base schema defer part of its definition to whatever schema referenced it. Set with `DynamicAnchor(...)` and `DynamicReference(...)`.

`$dynamicRef` is the mechanism behind recursive, extensible schemas (it is how the JSON Schema meta-schema references itself). For most application schemas, plain `$ref` + `$defs` is all you need.

## Recursive schemas

A schema may reference itself (e.g. a tree node whose children are the same node). Data-bounded recursion through `$ref` is supported — validation recurses as deep as the data goes. A reference cycle that is *not* bounded by data (a schema that would recurse forever with no input to consume) is reported as an error at compile time.

## Next

- [Vocabularies & the Meta-Schema](./04-vocabularies-and-meta-schema.md)
