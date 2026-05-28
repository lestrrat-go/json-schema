package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_one_of accepts a value matching exactly one subschema. 9 is a multiple
// of 3 only, so it matches one branch; 15 is a multiple of both 3 and 5, so it
// matches two and is rejected.
func Example_one_of() {
	built := schema.NewBuilder().
		OneOf(
			schema.NewBuilder().Types(schema.IntegerType).MultipleOf(3).MustBuild(),
			schema.NewBuilder().Types(schema.IntegerType).MultipleOf(5).MustBuild(),
		).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"oneOf": [
			{ "type": "integer", "multipleOf": 3 },
			{ "type": "integer", "multipleOf": 5 }
		]
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# 9 (multiple of 3 only)")
	report(schemas, 9)
	fmt.Println("# 15 (multiple of both, matches two branches)")
	report(schemas, 15)
	// Output:
	// # 9 (multiple of 3 only)
	// programmatic valid=true
	// from-json    valid=true
	// # 15 (multiple of both, matches two branches)
	// programmatic valid=false
	// from-json    valid=false
}
