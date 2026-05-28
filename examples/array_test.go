package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_arrayItems validates a list whose elements share one schema, with a
// minimum length and a uniqueness constraint.
func Example_arrayItems() {
	built := schema.NewBuilder().
		Types(schema.ArrayType).
		Items(schema.NewBuilder().Types(schema.StringType).MustBuild()).
		MinItems(1).
		UniqueItems(true).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "array",
		"items": { "type": "string" },
		"minItems": 1,
		"uniqueItems": true
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# distinct strings")
	report(schemas, []any{"a", "b"})
	fmt.Println("# duplicate elements")
	report(schemas, []any{"a", "a"})
	// Output:
	// # distinct strings
	// programmatic valid=true
	// from-json    valid=true
	// # duplicate elements
	// programmatic valid=false
	// from-json    valid=false
}

// Example_arrayTuple validates a fixed-shape tuple with prefixItems, and forbids
// extra elements with items:false.
func Example_arrayTuple() {
	built := schema.NewBuilder().
		Types(schema.ArrayType).
		PrefixItems(
			schema.NewBuilder().Types(schema.StringType).MustBuild(),
			schema.NewBuilder().Types(schema.IntegerType).MustBuild(),
		).
		Items(schema.FalseSchema()).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "array",
		"prefixItems": [
			{ "type": "string" },
			{ "type": "integer" }
		],
		"items": false
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# [string, integer]")
	report(schemas, []any{"x", 1})
	fmt.Println("# extra element beyond the tuple")
	report(schemas, []any{"x", 1, "extra"})
	// Output:
	// # [string, integer]
	// programmatic valid=true
	// from-json    valid=true
	// # extra element beyond the tuple
	// programmatic valid=false
	// from-json    valid=false
}

// Example_arrayContains requires at least two elements matching a subschema, via
// contains plus minContains.
func Example_arrayContains() {
	built := schema.NewBuilder().
		Types(schema.ArrayType).
		Contains(schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
		MinContains(2).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"type": "array",
		"contains": { "type": "integer" },
		"minContains": 2
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# two integers present")
	report(schemas, []any{"a", 1, 2})
	fmt.Println("# only one integer present")
	report(schemas, []any{"a", 1})
	// Output:
	// # two integers present
	// programmatic valid=true
	// from-json    valid=true
	// # only one integer present
	// programmatic valid=false
	// from-json    valid=false
}
