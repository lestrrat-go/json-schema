package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_const pins a value to a single constant.
func Example_const() {
	built := schema.NewBuilder().Const("v1").MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"const": "v1"
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# the constant value")
	report(schemas, "v1")
	fmt.Println("# any other value")
	report(schemas, "v2")
	// Output:
	// # the constant value
	// programmatic valid=true
	// from-json    valid=true
	// # any other value
	// programmatic valid=false
	// from-json    valid=false
}
