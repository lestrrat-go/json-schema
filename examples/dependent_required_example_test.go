package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_dependent_required makes one property's presence require another's:
// if "credit_card" is present, "billing_address" must be too.
func Example_dependent_required() {
	built := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
		Property("credit_card", schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
		Property("billing_address", schema.NewBuilder().Types(schema.StringType).MustBuild()).
		DependentRequired(map[string][]string{"credit_card": {"billing_address"}}).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "object",
		"properties": {
			"name": { "type": "string" },
			"credit_card": { "type": "integer" },
			"billing_address": { "type": "string" }
		},
		"dependentRequired": {
			"credit_card": ["billing_address"]
		}
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# no credit_card, so no dependency")
	report(schemas, map[string]any{"name": "Ada"})
	fmt.Println("# credit_card without billing_address")
	report(schemas, map[string]any{"credit_card": 1234})
	// Output:
	// # no credit_card, so no dependency
	// programmatic valid=true
	// from-json    valid=true
	// # credit_card without billing_address
	// programmatic valid=false
	// from-json    valid=false
}
