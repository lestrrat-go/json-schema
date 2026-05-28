package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_string_constraints validates a string against length bounds and a
// regular-expression pattern.
func Example_string_constraints() {
	built := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(3).
		MaxLength(10).
		Pattern("^[a-z]+$").
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "string",
		"minLength": 3,
		"maxLength": 10,
		"pattern": "^[a-z]+$"
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# lowercase, within length")
	report(schemas, "hello")
	fmt.Println("# too short and has uppercase")
	report(schemas, "Hi")
	// Output:
	// # lowercase, within length
	// programmatic valid=true
	// from-json    valid=true
	// # too short and has uppercase
	// programmatic valid=false
	// from-json    valid=false
}
