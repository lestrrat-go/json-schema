package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_all_of requires a value to satisfy every subschema simultaneously.
func Example_all_of() {
	built := schema.AllOf(
		schema.NewBuilder().Types(schema.IntegerType).Minimum(0).MustBuild(),
		schema.NewBuilder().Types(schema.IntegerType).MultipleOf(2).MustBuild(),
	).MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"allOf": [
			{ "type": "integer", "minimum": 0 },
			{ "type": "integer", "multipleOf": 2 }
		]
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# 4 (non-negative and even)")
	report(schemas, 4)
	fmt.Println("# 3 (odd)")
	report(schemas, 3)
	// Output:
	// # 4 (non-negative and even)
	// programmatic valid=true
	// from-json    valid=true
	// # 3 (odd)
	// programmatic valid=false
	// from-json    valid=false
}
