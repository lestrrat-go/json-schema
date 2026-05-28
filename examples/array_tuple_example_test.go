package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_array_tuple validates a fixed-shape tuple with prefixItems, and forbids
// extra elements with items:false.
func Example_array_tuple() {
	built := schema.NewBuilder().
		Types(schema.ArrayType).
		PrefixItems(
			schema.NewBuilder().Types(schema.StringType).MustBuild(),
			schema.NewBuilder().Types(schema.IntegerType).MustBuild(),
		).
		Items(schema.FalseSchema()).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "array",
		"prefixItems": [
			{ "type": "string" },
			{ "type": "integer" }
		],
		"items": false
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# [string, integer]")
	report(schemas, []any{"x", 1})
	fmt.Println("# extra element beyond the tuple")
	report(schemas, []any{"x", 1, "extra"})
	// Output:
	// # [string, integer]
	// programmatic valid=true
	// from-json    valid=true
	// # extra element beyond the tuple
	// programmatic valid=false
	// from-json    valid=false
}
