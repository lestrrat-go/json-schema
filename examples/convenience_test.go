package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_convenienceBuilders composes several of the package's one-line helper
// constructors — PositiveInteger, Optional, NonEmptyString and Enum — into a
// single object schema. Optional(s) accepts either s or null.
func Example_convenienceBuilders() {
	built := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("id", schema.PositiveInteger().MustBuild()).
		Property("nickname", schema.Optional(schema.NonEmptyString().MustBuild()).MustBuild()).
		Property("role", schema.Enum("admin", "user").MustBuild()).
		Required("id", "role").
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "object",
		"properties": {
			"id": { "type": "integer", "minimum": 0 },
			"nickname": {
				"anyOf": [
					{ "type": "string", "minLength": 1 },
					{ "type": "null" }
				]
			},
			"role": { "enum": ["admin", "user"] }
		},
		"required": ["id", "role"]
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# required fields present, optional nickname omitted")
	report(schemas, map[string]any{"id": 1, "role": "admin"})
	fmt.Println("# optional nickname explicitly null")
	report(schemas, map[string]any{"id": 1, "role": "user", "nickname": nil})
	fmt.Println("# role outside the enum")
	report(schemas, map[string]any{"id": 1, "role": "root"})
	// Output:
	// # required fields present, optional nickname omitted
	// programmatic valid=true
	// from-json    valid=true
	// # optional nickname explicitly null
	// programmatic valid=true
	// from-json    valid=true
	// # role outside the enum
	// programmatic valid=false
	// from-json    valid=false
}
