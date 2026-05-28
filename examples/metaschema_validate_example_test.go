package examples_test

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/json-schema/meta"
)

func Example_metaschema_validate() {
	ctx := context.Background()

	// Example: validate an object schema using the convenience function
	objectSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"age":  map[string]any{"type": "integer", "minimum": 0},
		},
		"required": []string{"name"},
	}

	err := meta.Validate(ctx, objectSchema)
	if err != nil {
		// Schema is not valid according to JSON Schema meta-schema
		panic(err)
	}

	// Schema is valid!
	fmt.Println("Object schema is valid")
	// Output:
	// Object schema is valid
}
