package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_objectProperties validates an object with typed properties and a
// required field. NonEmptyString() is type:string,minLength:1 and
// PositiveInteger() is type:integer,minimum:0.
func Example_objectProperties() {
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

// Example_objectAdditionalProperties rejects any property that is not declared,
// using additionalProperties:false.
func Example_objectAdditionalProperties() {
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

// Example_objectPatternProperties constrains properties whose names match a
// regular expression. Combined with additionalProperties:false, only names
// matching the pattern are permitted, and their values must be strings.
func Example_objectPatternProperties() {
	built := schema.NewBuilder().
		Types(schema.ObjectType).
		PatternProperty("^S_", schema.NewBuilder().Types(schema.StringType).MustBuild()).
		AdditionalProperties(schema.FalseSchema()).
		MustBuild()

	loaded := loadSchemaJSON(`{
		"type": "object",
		"patternProperties": {
			"^S_": { "type": "string" }
		},
		"additionalProperties": false
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# matching name, string value")
	report(schemas, map[string]any{"S_name": "widget"})
	fmt.Println("# matching name, non-string value")
	report(schemas, map[string]any{"S_name": 5})
	// Output:
	// # matching name, string value
	// programmatic valid=true
	// from-json    valid=true
	// # matching name, non-string value
	// programmatic valid=false
	// from-json    valid=false
}
