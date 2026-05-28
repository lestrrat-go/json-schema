package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_object_pattern_properties constrains properties whose names match a
// regular expression. Combined with additionalProperties:false, only names
// matching the pattern are permitted, and their values must be strings.
func Example_object_pattern_properties() {
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
