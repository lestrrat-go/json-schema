package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_ref_defs reuses a subschema defined under $defs via $ref. Both
// firstName and lastName point at the same "name" definition.
func Example_ref_defs() {
	nameDef := schema.NonEmptyString().MustBuild()
	built := schema.NewBuilder().
		Types(schema.ObjectType).
		Definitions("name", nameDef).
		Property("firstName", schema.NewBuilder().Reference("#/$defs/name").MustBuild()).
		Property("lastName", schema.NewBuilder().Reference("#/$defs/name").MustBuild()).
		Required("firstName", "lastName").
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "object",
		"$defs": {
			"name": { "type": "string", "minLength": 1 }
		},
		"properties": {
			"firstName": { "$ref": "#/$defs/name" },
			"lastName": { "$ref": "#/$defs/name" }
		},
		"required": ["firstName", "lastName"]
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# both names present and non-empty")
	report(schemas, map[string]any{"firstName": "Ada", "lastName": "Lovelace"})
	fmt.Println("# empty firstName violates the referenced schema")
	report(schemas, map[string]any{"firstName": "", "lastName": "Lovelace"})
	// Output:
	// # both names present and non-empty
	// programmatic valid=true
	// from-json    valid=true
	// # empty firstName violates the referenced schema
	// programmatic valid=false
	// from-json    valid=false
}
