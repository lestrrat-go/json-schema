package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_anyOf accepts a value matching at least one of several subschemas.
func Example_anyOf() {
	built := schema.AnyOf(
		schema.NewBuilder().Types(schema.StringType).MustBuild(),
		schema.NewBuilder().Types(schema.IntegerType).MustBuild(),
	).MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"anyOf": [
			{ "type": "string" },
			{ "type": "integer" }
		]
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# a string")
	report(schemas, "hello")
	fmt.Println("# a boolean (neither string nor integer)")
	report(schemas, true)
	// Output:
	// # a string
	// programmatic valid=true
	// from-json    valid=true
	// # a boolean (neither string nor integer)
	// programmatic valid=false
	// from-json    valid=false
}

// Example_oneOf accepts a value matching exactly one subschema. 9 is a multiple
// of 3 only, so it matches one branch; 15 is a multiple of both 3 and 5, so it
// matches two and is rejected.
func Example_oneOf() {
	built := schema.NewBuilder().
		OneOf(
			schema.NewBuilder().Types(schema.IntegerType).MultipleOf(3).MustBuild(),
			schema.NewBuilder().Types(schema.IntegerType).MultipleOf(5).MustBuild(),
		).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"oneOf": [
			{ "type": "integer", "multipleOf": 3 },
			{ "type": "integer", "multipleOf": 5 }
		]
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# 9 (multiple of 3 only)")
	report(schemas, 9)
	fmt.Println("# 15 (multiple of both, matches two branches)")
	report(schemas, 15)
	// Output:
	// # 9 (multiple of 3 only)
	// programmatic valid=true
	// from-json    valid=true
	// # 15 (multiple of both, matches two branches)
	// programmatic valid=false
	// from-json    valid=false
}

// Example_allOf requires a value to satisfy every subschema simultaneously.
func Example_allOf() {
	built := schema.AllOf(
		schema.NewBuilder().Types(schema.IntegerType).Minimum(0).MustBuild(),
		schema.NewBuilder().Types(schema.IntegerType).MultipleOf(2).MustBuild(),
	).MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"allOf": [
			{ "type": "integer", "minimum": 0 },
			{ "type": "integer", "multipleOf": 2 }
		]
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# 4 (non-negative and even)")
	report(schemas, 4)
	fmt.Println("# 3 (odd)")
	report(schemas, 3)
	// Output:
	// # 4 (non-negative and even)
	// programmatic valid=true
	// from-json    valid=true
	// # 3 (odd)
	// programmatic valid=false
	// from-json    valid=false
}

// Example_not accepts any value that does NOT match the subschema.
func Example_not() {
	built := schema.NewBuilder().
		Not(schema.NewBuilder().Types(schema.StringType).MustBuild()).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"not": { "type": "string" }
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# a number (not a string)")
	report(schemas, 42)
	fmt.Println("# a string (matches the negated schema)")
	report(schemas, "nope")
	// Output:
	// # a number (not a string)
	// programmatic valid=true
	// from-json    valid=true
	// # a string (matches the negated schema)
	// programmatic valid=false
	// from-json    valid=false
}
