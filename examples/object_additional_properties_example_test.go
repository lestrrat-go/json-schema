package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_object_additional_properties rejects any property that is not declared,
// using additionalProperties:false.
func Example_object_additional_properties() {
	built := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("id", schema.PositiveInteger().MustBuild()).
		AdditionalProperties(schema.FalseSchema()).
		MustBuild()

	loaded := loadSchemaJSON(`{
		"type": "object",
		"properties": {
			"id": { "type": "integer", "minimum": 0 }
		},
		"additionalProperties": false
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# only declared properties")
	report(schemas, map[string]any{"id": 1})
	fmt.Println("# undeclared property present")
	report(schemas, map[string]any{"id": 1, "extra": true})
	// Output:
	// # only declared properties
	// programmatic valid=true
	// from-json    valid=true
	// # undeclared property present
	// programmatic valid=false
	// from-json    valid=false
}
