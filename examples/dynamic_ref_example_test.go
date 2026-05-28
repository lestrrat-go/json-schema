package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_dynamic_ref validates a recursive structure. The schema marks itself
// with $dynamicAnchor "node" and, for each child, refers back to "#node" via
// $dynamicRef — so the same schema applies at every level of the tree.
func Example_dynamic_ref() {
	built := schema.NewBuilder().
		ID("https://example.com/tree").
		DynamicAnchor("node").
		Types(schema.ObjectType).
		Property("value", schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
		Property("children", schema.NewBuilder().
			Types(schema.ArrayType).
			Items(schema.NewBuilder().DynamicReference("#node").MustBuild()).
			MustBuild()).
		Required("value").
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"$id": "https://example.com/tree",
		"$dynamicAnchor": "node",
		"type": "object",
		"properties": {
			"value": { "type": "integer" },
			"children": {
				"type": "array",
				"items": { "$dynamicRef": "#node" }
			}
		},
		"required": ["value"]
	}`)
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	good := map[string]any{
		"value": 1,
		"children": []any{
			map[string]any{"value": 2},
			map[string]any{"value": 3, "children": []any{map[string]any{"value": 4}}},
		},
	}
	bad := map[string]any{
		"value":    1,
		"children": []any{map[string]any{"value": "not-an-integer"}},
	}

	fmt.Println("# every node has an integer value")
	report(schemas, good)
	fmt.Println("# a nested node has a non-integer value")
	report(schemas, bad)
	// Output:
	// # every node has an integer value
	// programmatic valid=true
	// from-json    valid=true
	// # a nested node has a non-integer value
	// programmatic valid=false
	// from-json    valid=false
}
