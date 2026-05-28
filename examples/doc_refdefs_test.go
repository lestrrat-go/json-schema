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
