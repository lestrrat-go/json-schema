package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_enum restricts a value to a fixed set, using the schema.Enum helper.
func Example_enum() {
	built := schema.Enum("red", "green", "blue").MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"enum": ["red", "green", "blue"]
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# member of the set")
	report(schemas, "green")
	fmt.Println("# not a member")
	report(schemas, "purple")
	// Output:
	// # member of the set
	// programmatic valid=true
	// from-json    valid=true
	// # not a member
	// programmatic valid=false
	// from-json    valid=false
}
