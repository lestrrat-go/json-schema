package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_number_exclusive_bounds validates a number against exclusive bounds: the
// value must be strictly between 0 and 1.
func Example_number_exclusive_bounds() {
	built := schema.NewBuilder().
		Types(schema.NumberType).
		ExclusiveMinimum(0).
		ExclusiveMaximum(1).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "number",
		"exclusiveMinimum": 0,
		"exclusiveMaximum": 1
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# 0.5 (strictly between 0 and 1)")
	report(schemas, 0.5)
	fmt.Println("# 0 (not strictly greater than 0)")
	report(schemas, 0)
	// Output:
	// # 0.5 (strictly between 0 and 1)
	// programmatic valid=true
	// from-json    valid=true
	// # 0 (not strictly greater than 0)
	// programmatic valid=false
	// from-json    valid=false
}
