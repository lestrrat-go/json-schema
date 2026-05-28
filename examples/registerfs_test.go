package examples_test

import (
	"context"
	"fmt"
	"testing/fstest"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// Example_resolverRegisterFS preloads a whole tree of schema documents from a
// filesystem with Resolver.RegisterFS. Each ".json" file is registered under the
// base URI joined with its path, so a $ref to that URI resolves offline. The
// fs.FS here is an in-memory fstest.MapFS, but an embed.FS or os.DirFS works the
// same way.
func Example_resolverRegisterFS() {
	fsys := fstest.MapFS{
		"address.json": &fstest.MapFile{Data: []byte(
			`{"type":"object","properties":{"city":{"type":"string","minLength":1}},"required":["city"]}`,
		)},
	}

	built := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("home", schema.NewBuilder().Reference("https://example.com/schemas/address.json").MustBuild()).
		Required("home").
		MustBuild()
	loaded := loadSchema("testdata/registerfs_main.json")

	validateWith := func(mainSchema *schema.Schema, data any) bool {
		r := schema.NewResolver()
		if err := r.RegisterFS("https://example.com/schemas/", fsys); err != nil {
			return false
		}
		ctx := schema.WithResolver(context.Background(), r)
		v, err := validator.Compile(ctx, mainSchema)
		if err != nil {
			return false
		}
		_, err = v.Validate(ctx, data)
		return err == nil
	}

	good := map[string]any{"home": map[string]any{"city": "Kyoto"}}
	bad := map[string]any{"home": map[string]any{}} // missing required city

	fmt.Printf("programmatic good=%t bad=%t\n", validateWith(built, good), validateWith(built, bad))
	fmt.Printf("from-json    good=%t bad=%t\n", validateWith(loaded, good), validateWith(loaded, bad))
	// Output:
	// programmatic good=true bad=false
	// from-json    good=true bad=false
}
