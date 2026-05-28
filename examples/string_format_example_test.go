package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_string_format shows that "format" is an annotation by default and only
// becomes an assertion when the format-assertion vocabulary is enabled. The same
// schema (built with schema.Email(), or loaded from JSON) accepts a malformed
// address under the default vocabulary but rejects it under the strict one.
func Example_string_format() {
	built := schema.Email().MustBuild() // type:string, format:email
	loaded := loadSchemaJSON(`{
		"type": "string",
		"format": "email"
	}`)

	const malformed = "not-an-email"
	fmt.Printf("default (annotation) built=%t json=%t\n", valid(built, malformed), valid(loaded, malformed))
	fmt.Printf("strict  (assertion)  built=%t json=%t\n", validStrict(built, malformed), validStrict(loaded, malformed))
	// Output:
	// default (annotation) built=true json=true
	// strict  (assertion)  built=false json=false
}
