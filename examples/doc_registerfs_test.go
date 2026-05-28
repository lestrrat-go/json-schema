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

	ctx := schema.WithResolver(context.Background(), r)
	v, err := validator.Compile(ctx, main)
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
