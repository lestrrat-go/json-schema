package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_unevaluated_properties rejects properties that no other keyword has
// already evaluated. The "name" property is evaluated by the allOf branch, so it
// is allowed; any other property is "unevaluated" and rejected.
func Example_unevaluated_properties() {
	base := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
		MustBuild()

	built := schema.NewBuilder().
		AllOf(base).
		UnevaluatedProperties(schema.FalseSchema()).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"allOf": [
			{
				"type": "object",
				"properties": {
					"name": { "type": "string" }
				}
			}
		],
		"unevaluatedProperties": false
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# only the evaluated property")
	report(schemas, map[string]any{"name": "Ada"})
	fmt.Println("# an unevaluated property is present")
	report(schemas, map[string]any{"name": "Ada", "extra": true})
	// Output:
	// # only the evaluated property
	// programmatic valid=true
	// from-json    valid=true
	// # an unevaluated property is present
	// programmatic valid=false
	// from-json    valid=false
}
