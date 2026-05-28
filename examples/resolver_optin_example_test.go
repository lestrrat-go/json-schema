package examples_test

import (
	"context"
	"fmt"
	"testing/fstest"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// Example_resolverOptIn shows that external reference resolution is opt-in. A
// bare schema.NewResolver() resolves only from memory; to let a $ref reach the
// filesystem you must hand it an explicit resolver. Here FSResolver reads from
// an in-memory fs.FS on demand (use DirResolver(".") for the local filesystem or
// HTTPResolver() for the network the same way). Contrast with Example_docRegisterFS,
// which preloads the documents instead.
func Example_resolverOptIn() {
	fsys := fstest.MapFS{
		"address.json": &fstest.MapFile{Data: []byte(
			`{"type":"object","properties":{"city":{"type":"string","minLength":1}},"required":["city"]}`,
		)},
	}

	main := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("home", schema.NewBuilder().Reference("address.json").MustBuild()).
		Required("home").
		MustBuild()

	ctx := context.Background()

	// Default resolver: no filesystem access, so the external $ref cannot resolve.
	if _, err := validator.Compile(ctx, main, validator.WithResolver(schema.NewResolver())); err != nil {
		fmt.Println("default (in-memory only):", err != nil)
	}

	// Opt in to filesystem access with FSResolver.
	r := schema.NewResolver(schema.WithResolver(schema.FSResolver(fsys)))
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
	// default (in-memory only): true
	// with city:    true
	// without city: false
}
