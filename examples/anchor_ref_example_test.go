package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_anchor_ref labels a subschema with $anchor and references it by name
// with "#name" (as opposed to a JSON pointer like "#/$defs/name").
func Example_anchor_ref() {
	nameDef := schema.NewBuilder().
		Anchor("name").
		Types(schema.StringType).
		MinLength(1).
		MustBuild()

	built := schema.NewBuilder().
		Types(schema.ObjectType).
		Definitions("name", nameDef).
		Property("first", schema.NewBuilder().Reference("#name").MustBuild()).
		Required("first").
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "object",
		"$defs": {
			"name": { "$anchor": "name", "type": "string", "minLength": 1 }
		},
		"properties": {
			"first": { "$ref": "#name" }
		},
		"required": ["first"]
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# non-empty first name")
	report(schemas, map[string]any{"first": "Ada"})
	fmt.Println("# empty first name violates the anchored schema")
	report(schemas, map[string]any{"first": ""})
	// Output:
	// # non-empty first name
	// programmatic valid=true
	// from-json    valid=true
	// # empty first name violates the anchored schema
	// programmatic valid=false
	// from-json    valid=false
}
