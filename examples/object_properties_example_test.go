package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_object_properties validates an object with typed properties and a
// required field. NonEmptyString() is type:string,minLength:1 and
// PositiveInteger() is type:integer,minimum:0.
func Example_object_properties() {
	// Programmatic.
	built := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("name", schema.NonEmptyString().MustBuild()).
		Property("age", schema.PositiveInteger().MustBuild()).
		Required("name").
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "object",
		"properties": {
			"name": { "type": "string", "minLength": 1 },
			"age": { "type": "integer", "minimum": 0 }
		},
		"required": ["name"]
	}`)

	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# good input")
	report(schemas, map[string]any{"name": "Alice", "age": 30})
	fmt.Println("# bad input (missing name, negative age)")
	report(schemas, map[string]any{"age": -1})
	// Output:
	// # good input
	// programmatic valid=true
	// from-json    valid=true
	// # bad input (missing name, negative age)
	// programmatic valid=false
	// from-json    valid=false
}
