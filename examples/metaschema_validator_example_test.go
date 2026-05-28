package examples_test

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/json-schema/meta"
)

func Example_metaschema_validator() {
	ctx := context.Background()

	// Get the pre-compiled meta-schema validator
	validator := meta.Validator()

	// Example: validate a simple string schema
	stringSchema := map[string]any{
		"type":      "string",
		"minLength": 1,
		"maxLength": 100,
	}

	_, err := validator.Validate(ctx, stringSchema)
	if err != nil {
		// Schema is not valid according to JSON Schema meta-schema
		panic(err)
	}

	// Schema is valid!
	fmt.Println("Schema is valid")
	// Output:
	// Schema is valid
}
