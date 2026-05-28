package examples_test

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// Example_externalResolver resolves a $ref that points at a separate schema
// document. A schema.Resolver holds the external document (keyed by its URI) and
// is placed on the context so the compiler can follow the reference. The same
// wiring is used whether the documents are built programmatically or loaded from
// JSON files.
func Example_externalResolver() {
	// Built programmatically.
	address := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("city", schema.NonEmptyString().MustBuild()).
		Required("city").
		MustBuild()
	main := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("home", schema.NewBuilder().Reference("https://example.com/address").MustBuild()).
		Required("home").
		MustBuild()

	// Loaded from JSON.
	loadedMain := loadSchema("testdata/resolver_main.json")
	loadedAddr := loadSchema("testdata/resolver_address.json")

	validateWith := func(mainSchema, addressSchema *schema.Schema, data any) bool {
		r := schema.NewResolver()
		r.RegisterDocument("https://example.com/address", addressSchema)
		ctx := schema.WithResolver(context.Background(), r)
		v, err := validator.Compile(ctx, mainSchema)
		if err != nil {
			return false
		}
		_, err = v.Validate(ctx, data)
		return err == nil
	}

	good := map[string]any{"home": map[string]any{"city": "Paris"}}
	bad := map[string]any{"home": map[string]any{}} // missing the required city

	fmt.Printf("programmatic good=%t bad=%t\n", validateWith(main, address, good), validateWith(main, address, bad))
	fmt.Printf("from-json    good=%t bad=%t\n", validateWith(loadedMain, loadedAddr, good), validateWith(loadedMain, loadedAddr, bad))
	// Output:
	// programmatic good=true bad=false
	// from-json    good=true bad=false
}
