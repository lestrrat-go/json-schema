package examples_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

func Example() {
	// Build a JSON Schema using the fluent builder API
	userSchema := schema.NewBuilder().
		Schema(schema.Version).
		Types(schema.ObjectType).
		Property("name", schema.NonEmptyString().MustBuild()).
		Property("email", schema.Email().MustBuild()).
		Property("age", schema.PositiveInteger().MustBuild()).
		Required("name", "email").
		MustBuild()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(userSchema); err != nil {
		fmt.Printf("failed to encode schema: %s\n", err)
		return
	}

	// Compile the schema into an optimized validator
	v, err := validator.Compile(context.Background(), userSchema)
	if err != nil {
		fmt.Printf("failed to compile validator: %s\n", err)
		return
	}

	// Validate data
	validUser := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	_, err = v.Validate(context.Background(), validUser)
	if err != nil {
		fmt.Printf("validation failed: %s\n", err)
		return
	}

	fmt.Println("User data is valid!")

	// Test with invalid data
	invalidUser := map[string]any{
		"name":  "", // Empty name should fail
		"email": "not-an-email",
	}

	_, err = v.Validate(context.Background(), invalidUser)
	if err != nil {
		fmt.Printf("validation failed as expected: %s\n", err)
	}
	// OUTPUT:
	// {
	//   "$schema": "https://json-schema.org/draft/2020-12/schema",
	//   "properties": {
	//     "age": {
	//       "minimum": 0,
	//       "type": "integer"
	//     },
	//     "email": {
	//       "format": "email",
	//       "type": "string"
	//     },
	//     "name": {
	//       "minLength": 1,
	//       "type": "string"
	//     }
	//   },
	//   "required": [
	//     "name",
	//     "email"
	//   ],
	//   "type": "object"
	// }
	// User data is valid!
	// validation failed as expected: invalid value passed to ObjectValidator: property validation failed for name: invalid value passed to StringValidator: string length (0) shorter then minLength (1)
}
