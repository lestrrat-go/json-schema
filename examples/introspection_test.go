package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_schemaIntrospection reads a schema's constraints back out through its
// accessors. Every keyword has a Has<Keyword> presence check and a getter, and
// HasAny tests a group of fields at once using the package's field-flag
// constants. A built schema and the equivalent loaded-from-JSON schema report
// the same structure.
func Example_schemaIntrospection() {
	built := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(3).
		MaxLength(20).
		Pattern("^[a-z]+$").
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "string",
		"minLength": 3,
		"maxLength": 20,
		"pattern": "^[a-z]+$"
	}`)

	describe := func(s *schema.Schema) string {
		return fmt.Sprintf("types=%v string?=%t pattern=%q maxLength(set=%t)=%d stringConstraints?=%t",
			s.Types(),
			s.ContainsType(schema.StringType),
			s.Pattern(),
			s.HasMaxLength(), s.MaxLength(),
			s.HasAny(schema.StringConstraintFields),
		)
	}

	fmt.Println("programmatic:", describe(built))
	fmt.Println("from-json:   ", describe(loaded))
	// Output:
	// programmatic: types=[string] string?=true pattern="^[a-z]+$" maxLength(set=true)=20 stringConstraints?=true
	// from-json:    types=[string] string?=true pattern="^[a-z]+$" maxLength(set=true)=20 stringConstraints?=true
}
