package examples_test

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// Example_docQuickStart shows the three core steps: build a schema, compile it
// into a validator, and validate data. A compiled validator is reusable across
// many Validate calls and safe for concurrent use.
func Example_docQuickStart() {
	ctx := context.Background()

	// 1. Build a schema.
	s := schema.NewBuilder().
		Schema(schema.Version).
		Types(schema.ObjectType).
		Property("name", schema.NonEmptyString().MustBuild()).
		Property("email", schema.Email().MustBuild()).
		Property("age", schema.PositiveInteger().MustBuild()).
		Required("name", "email").
		MustBuild()

	// 2. Compile it into a validator.
	v, err := validator.Compile(ctx, s)
	if err != nil {
		fmt.Println("compile failed:", err)
		return
	}

	// 3. Validate data.
	_, err = v.Validate(ctx, map[string]any{
		"name":  "Ada Lovelace",
		"email": "ada@example.com",
		"age":   36,
	})
	fmt.Println("valid record:", err == nil)

	_, err = v.Validate(ctx, map[string]any{"name": "", "email": "x"})
	fmt.Println("empty name:  ", err == nil)
	// Output:
	// valid record: true
	// empty name:   false
}
