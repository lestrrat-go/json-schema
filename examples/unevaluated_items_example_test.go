package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_unevaluated_items is the array counterpart of unevaluatedProperties:
// the prefixItems branch evaluates index 0, so any further element is
// "unevaluated" and rejected.
func Example_unevaluated_items() {
	base := schema.NewBuilder().
		Types(schema.ArrayType).
		PrefixItems(schema.NewBuilder().Types(schema.StringType).MustBuild()).
		MustBuild()

	built := schema.NewBuilder().
		AllOf(base).
		UnevaluatedItems(schema.FalseSchema()).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"allOf": [
			{
				"type": "array",
				"prefixItems": [{ "type": "string" }]
			}
		],
		"unevaluatedItems": false
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# only the evaluated element")
	report(schemas, []any{"first"})
	fmt.Println("# an unevaluated element is present")
	report(schemas, []any{"first", 2})
	// Output:
	// # only the evaluated element
	// programmatic valid=true
	// from-json    valid=true
	// # an unevaluated element is present
	// programmatic valid=false
	// from-json    valid=false
}
