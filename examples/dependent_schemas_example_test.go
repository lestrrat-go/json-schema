package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_dependent_schemas applies an entire subschema when a property is
// present: if "credit_card" is present, the object must also satisfy a schema
// that requires "billing_address".
func Example_dependent_schemas() {
	dep := schema.NewBuilder().
		Required("billing_address").
		Property("billing_address", schema.NewBuilder().Types(schema.StringType).MustBuild()).
		MustBuild()

	built := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("credit_card", schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
		DependentSchemas(map[string]schema.SchemaOrBool{"credit_card": dep}).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "object",
		"properties": {
			"credit_card": { "type": "integer" }
		},
		"dependentSchemas": {
			"credit_card": {
				"required": ["billing_address"],
				"properties": {
					"billing_address": { "type": "string" }
				}
			}
		}
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# no credit_card, dependent schema not applied")
	report(schemas, map[string]any{})
	fmt.Println("# credit_card present, dependent schema unsatisfied")
	report(schemas, map[string]any{"credit_card": 1234})
	// Output:
	// # no credit_card, dependent schema not applied
	// programmatic valid=true
	// from-json    valid=true
	// # credit_card present, dependent schema unsatisfied
	// programmatic valid=false
	// from-json    valid=false
}
