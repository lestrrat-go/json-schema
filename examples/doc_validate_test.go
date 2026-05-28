package examples_test

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// Example_docValidate compiles a schema once and reuses the validator for several
// inputs. Validate returns a non-nil error when the data is invalid.
func Example_docValidate() {
	s := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("id", schema.PositiveInteger().MustBuild()).
		Property("role", schema.Enum("admin", "user").MustBuild()).
		Required("id", "role").
		MustBuild()

	ctx := context.Background()
	v, err := validator.Compile(ctx, s)
	if err != nil {
		fmt.Println("compile failed:", err)
		return
	}

	for _, data := range []map[string]any{
		{"id": 1, "role": "admin"}, // valid
		{"id": 1, "role": "root"},  // role not in enum
	} {
		_, err := v.Validate(ctx, data)
		fmt.Printf("valid=%t\n", err == nil)
	}
	// Output:
	// valid=true
	// valid=false
}
