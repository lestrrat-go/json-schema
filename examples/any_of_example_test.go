package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_any_of accepts a value matching at least one of several subschemas.
func Example_any_of() {
	built := schema.AnyOf(
		schema.NewBuilder().Types(schema.StringType).MustBuild(),
		schema.NewBuilder().Types(schema.IntegerType).MustBuild(),
	).MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"anyOf": [
			{ "type": "string" },
			{ "type": "integer" }
		]
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# a string")
	report(schemas, "hello")
	fmt.Println("# a boolean (neither string nor integer)")
	report(schemas, true)
	// Output:
	// # a string
	// programmatic valid=true
	// from-json    valid=true
	// # a boolean (neither string nor integer)
	// programmatic valid=false
	// from-json    valid=false
}
