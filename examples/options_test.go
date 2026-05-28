package examples_test

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/lestrrat-go/json-schema/vocabulary"
)

// Example_optionVocabularySet shows how validator.WithVocabularySet selects which
// vocabularies are enforced. By default "format" is annotation-only, so a
// malformed value still validates; enabling all vocabularies turns on
// format-assertion, making "format" an enforced constraint.
func Example_optionVocabularySet() {
	s := schema.Email().MustBuild() // type: string, format: email

	const malformed = "not-an-email"

	def, err := validator.Compile(context.Background(), s)
	if err != nil {
		panic(err)
	}
	strict, err := validator.Compile(context.Background(), s,
		validator.WithVocabularySet(vocabulary.AllEnabled()))
	if err != nil {
		panic(err)
	}

	_, defErr := def.Validate(context.Background(), malformed)
	_, strictErr := strict.Validate(context.Background(), malformed)

	fmt.Printf("default (annotation) valid=%t\n", defErr == nil)
	fmt.Printf("strict  (assertion)  valid=%t\n", strictErr == nil)
	// Output:
	// default (annotation) valid=true
	// strict  (assertion)  valid=false
}

// Example_optionBaseSchema shows how validator.WithBaseSchema compiles a schema
// fragment whose local "#/..." references resolve against a separate document.
// The fragment carries no $defs of its own; WithBaseSchema names the document it
// belongs to so the reference resolves there.
func Example_optionBaseSchema() {
	doc := schema.NewBuilder().
		Types(schema.ObjectType).
		Definitions("nonEmpty", schema.NewBuilder().Types(schema.StringType).MinLength(1).MustBuild()).
		MustBuild()

	fragment := schema.NewBuilder().Reference("#/$defs/nonEmpty").MustBuild()

	v, err := validator.Compile(context.Background(), fragment, validator.WithBaseSchema(doc))
	if err != nil {
		panic(err)
	}

	_, goodErr := v.Validate(context.Background(), "hello")
	_, badErr := v.Validate(context.Background(), "")

	fmt.Printf("good=%t bad=%t\n", goodErr == nil, badErr == nil)
	// Output:
	// good=true bad=false
}

// Example_optionBaseURI shows how validator.WithBaseURI resolves a relative $ref
// to an absolute URI. The address document is registered under its absolute URI;
// the main schema references it relatively, and WithBaseURI supplies the base the
// relative reference is resolved against.
func Example_optionBaseURI() {
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
