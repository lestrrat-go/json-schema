package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_stringConstraints validates a string against length bounds and a
// regular-expression pattern.
func Example_stringConstraints() {
	built := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(3).
		MaxLength(10).
		Pattern("^[a-z]+$").
		MustBuild()

	loaded := loadSchema("testdata/string_constraints.json")
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# lowercase, within length")
	report(schemas, "hello")
	fmt.Println("# too short and has uppercase")
	report(schemas, "Hi")
	// Output:
	// # lowercase, within length
	// programmatic valid=true
	// from-json    valid=true
	// # too short and has uppercase
	// programmatic valid=false
	// from-json    valid=false
}

// Example_stringFormat shows that "format" is an annotation by default and only
// becomes an assertion when the format-assertion vocabulary is enabled. The same
// schema (built with schema.Email(), or loaded from JSON) accepts a malformed
// address under the default vocabulary but rejects it under the strict one.
func Example_stringFormat() {
	built := schema.Email().MustBuild() // type:string, format:email
	loaded := loadSchema("testdata/string_email.json")

	const malformed = "not-an-email"
	fmt.Printf("default (annotation) built=%t json=%t\n", valid(built, malformed), valid(loaded, malformed))
	fmt.Printf("strict  (assertion)  built=%t json=%t\n", validStrict(built, malformed), validStrict(loaded, malformed))
	// Output:
	// default (annotation) built=true json=true
	// strict  (assertion)  built=false json=false
}
