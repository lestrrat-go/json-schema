package examples_test

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// Example_option_base_uri shows how validator.WithBaseURI resolves a relative $ref
// to an absolute URI. The address document is registered under its absolute URI;
// the main schema references it relatively, and WithBaseURI supplies the base the
// relative reference is resolved against.
func Example_option_base_uri() {
	address := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("city", schema.NonEmptyString().MustBuild()).
		Required("city").
		MustBuild()

	r := schema.NewResolver()
	r.RegisterDocument("https://example.com/schemas/address.json", address)

	main := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("home", schema.NewBuilder().Reference("address.json").MustBuild()).
		Required("home").
		MustBuild()

	v, err := validator.Compile(context.Background(), main,
		validator.WithResolver(r),
		validator.WithBaseURI("https://example.com/schemas/main.json"))
	if err != nil {
		panic(err)
	}

	_, goodErr := v.Validate(context.Background(), map[string]any{"home": map[string]any{"city": "Paris"}})
	_, badErr := v.Validate(context.Background(), map[string]any{"home": map[string]any{}})

	fmt.Printf("good=%t bad=%t\n", goodErr == nil, badErr == nil)
	// Output:
	// good=true bad=false
}
