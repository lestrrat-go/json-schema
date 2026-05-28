package examples_test

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// Example_external_resolver resolves a $ref that points at a separate schema
// document. A schema.Resolver holds the external document (keyed by its URI) and
// is passed to Compile so the compiler can follow the reference. The same
// wiring is used whether the documents are built programmatically or loaded from
// JSON files.
func Example_external_resolver() {
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

	// The equivalent schemas authored as JSON.
	loadedMain := loadSchemaJSON(`{
		"type": "object",
		"properties": {
			"home": { "$ref": "https://example.com/address" }
		},
		"required": ["home"]
	}`)
	loadedAddr := loadSchemaJSON(`{
		"$id": "https://example.com/address",
		"type": "object",
		"properties": {
			"city": { "type": "string", "minLength": 1 }
		},
		"required": ["city"]
	}`)

	validateWith := func(mainSchema, addressSchema *schema.Schema, data any) bool {
		r := schema.NewResolver()
		r.RegisterDocument("https://example.com/address", addressSchema)
		v, err := validator.Compile(context.Background(), mainSchema, validator.WithResolver(r))
		if err != nil {
			return false
		}
		_, err = v.Validate(context.Background(), data)
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
