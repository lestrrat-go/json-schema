package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_multipleTypes allows a value to be any of several primitive types by
// passing more than one type to Types (which serializes as a JSON "type" array).
func Example_multipleTypes() {
	built := schema.NewBuilder().
		Types(schema.StringType, schema.IntegerType).
		MustBuild()

	loaded := loadSchema("testdata/multiple_types.json")
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# a string")
	report(schemas, "hello")
	fmt.Println("# an integer")
	report(schemas, 42)
	fmt.Println("# a boolean (neither)")
	report(schemas, true)
	// Output:
	// # a string
	// programmatic valid=true
	// from-json    valid=true
	// # an integer
	// programmatic valid=true
	// from-json    valid=true
	// # a boolean (neither)
	// programmatic valid=false
	// from-json    valid=false
}
