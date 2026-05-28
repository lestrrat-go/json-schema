package examples_test

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// Example_option_base_schema shows how validator.WithBaseSchema compiles a schema
// fragment whose local "#/..." references resolve against a separate document.
// The fragment carries no $defs of its own; WithBaseSchema names the document it
// belongs to so the reference resolves there.
func Example_option_base_schema() {
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
