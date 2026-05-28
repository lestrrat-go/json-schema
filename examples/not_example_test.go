package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_not accepts any value that does NOT match the subschema.
func Example_not() {
	built := schema.NewBuilder().
		Not(schema.NewBuilder().Types(schema.StringType).MustBuild()).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"not": { "type": "string" }
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# a number (not a string)")
	report(schemas, 42)
	fmt.Println("# a string (matches the negated schema)")
	report(schemas, "nope")
	// Output:
	// # a number (not a string)
	// programmatic valid=true
	// from-json    valid=true
	// # a string (matches the negated schema)
	// programmatic valid=false
	// from-json    valid=false
}
