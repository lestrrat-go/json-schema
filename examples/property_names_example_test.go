package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_property_names constrains the names of an object's properties (rather
// than their values) with a subschema.
func Example_property_names() {
	built := schema.NewBuilder().
		Types(schema.ObjectType).
		PropertyNames(schema.NewBuilder().Pattern("^[a-z]+$").MustBuild()).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "object",
		"propertyNames": { "pattern": "^[a-z]+$" }
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# all names lowercase")
	report(schemas, map[string]any{"foo": 1, "bar": 2})
	fmt.Println("# a name with an uppercase letter")
	report(schemas, map[string]any{"Foo": 1})
	// Output:
	// # all names lowercase
	// programmatic valid=true
	// from-json    valid=true
	// # a name with an uppercase letter
	// programmatic valid=false
	// from-json    valid=false
}
