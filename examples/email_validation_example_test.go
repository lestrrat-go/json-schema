package examples_test

import (
	"context"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

func ExampleEmail() {
	// Create a user schema with email validation
	userSchema := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("name", schema.NonEmptyString().MustBuild()).
		Property("email", schema.Email().MustBuild()).
		Property("age", schema.PositiveInteger().MustBuild()).
		Required("name", "email").
		MustBuild()

	// Validate user data
	validUser := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	v, _ := validator.Compile(context.Background(), userSchema)
	_, err := v.Validate(context.Background(), validUser)
	if err != nil {
		panic("This should not happen")
	}
	// Output:
}