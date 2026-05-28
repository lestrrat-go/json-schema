package examples_test

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/json-schema/meta"
)

// Example_docMeta validates that a document is itself a valid JSON Schema 2020-12
// document, using the pre-compiled meta-schema validator in the meta package.
func Example_docMeta() {
	ctx := context.Background()

	validSchema := map[string]any{"type": "string", "minLength": 1}
	fmt.Println("valid schema:  ", meta.Validate(ctx, validSchema) == nil)

	notASchema := "not a schema"
	fmt.Println("invalid schema:", meta.Validate(ctx, notASchema) == nil)
	// Output:
	// valid schema:   true
	// invalid schema: false
}
