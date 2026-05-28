package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_integer_constraints validates an integer against inclusive bounds and a
// multipleOf divisor.
func Example_integer_constraints() {
	built := schema.NewBuilder().
		Types(schema.IntegerType).
		Minimum(0).
		Maximum(100).
		MultipleOf(5).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "integer",
		"minimum": 0,
		"maximum": 100,
		"multipleOf": 5
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# 25 (in range, multiple of 5)")
	report(schemas, 25)
	fmt.Println("# 7 (not a multiple of 5)")
	report(schemas, 7)
	// Output:
	// # 25 (in range, multiple of 5)
	// programmatic valid=true
	// from-json    valid=true
	// # 7 (not a multiple of 5)
	// programmatic valid=false
	// from-json    valid=false
}
