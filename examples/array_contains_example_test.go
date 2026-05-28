package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_array_contains requires at least two elements matching a subschema, via
// contains plus minContains.
func Example_array_contains() {
	built := schema.NewBuilder().
		Types(schema.ArrayType).
		Contains(schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
		MinContains(2).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "array",
		"contains": { "type": "integer" },
		"minContains": 2
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# two integers present")
	report(schemas, []any{"a", 1, 2})
	fmt.Println("# only one integer present")
	report(schemas, []any{"a", 1})
	// Output:
	// # two integers present
	// programmatic valid=true
	// from-json    valid=true
	// # only one integer present
	// programmatic valid=false
	// from-json    valid=false
}
