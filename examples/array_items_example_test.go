package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_array_items validates a list whose elements share one schema, with a
// minimum length and a uniqueness constraint.
func Example_array_items() {
	built := schema.NewBuilder().
		Types(schema.ArrayType).
		Items(schema.NewBuilder().Types(schema.StringType).MustBuild()).
		MinItems(1).
		UniqueItems(true).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "array",
		"items": { "type": "string" },
		"minItems": 1,
		"uniqueItems": true
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# distinct strings")
	report(schemas, []any{"a", "b"})
	fmt.Println("# duplicate elements")
	report(schemas, []any{"a", "a"})
	// Output:
	// # distinct strings
	// programmatic valid=true
	// from-json    valid=true
	// # duplicate elements
	// programmatic valid=false
	// from-json    valid=false
}
