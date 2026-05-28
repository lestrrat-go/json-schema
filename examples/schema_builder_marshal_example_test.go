package examples_test

import (
	"encoding/json"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_schema_builder_marshal builds a schema with the fluent builder and shows
// that it marshals to the same canonical JSON as the equivalent schema authored
// as JSON. The library emits object keys in a stable, sorted order, so a
// round trip (build -> JSON, or JSON -> *Schema -> JSON) is deterministic.
func Example_schema_builder_marshal() {
	// Programmatic.
	built := schema.NewBuilder().
		Schema(schema.Version).
		ID("https://example.com/polygon").
		Types(schema.ObjectType).
		Property("validProp", schema.New()).
		AdditionalProperties(schema.TrueSchema()).
		MustBuild()

	// The equivalent schema authored as JSON.
	loaded := loadSchemaJSON(`{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"$id": "https://example.com/polygon",
		"type": "object",
		"properties": { "validProp": {} },
		"additionalProperties": true
	}`)

	for _, name := range []string{"programmatic", "from-json"} {
		s := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}[name]
		buf, err := json.Marshal(s)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%-12s %s\n", name, buf)
	}
	// Output:
	// programmatic {"$id":"https://example.com/polygon","$schema":"https://json-schema.org/draft/2020-12/schema","additionalProperties":true,"properties":{"validProp":{}},"type":"object"}
	// from-json    {"$id":"https://example.com/polygon","$schema":"https://json-schema.org/draft/2020-12/schema","additionalProperties":true,"properties":{"validProp":{}},"type":"object"}
}
