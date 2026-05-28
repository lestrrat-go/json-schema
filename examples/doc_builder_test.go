package examples_test

import (
	"encoding/json"
	"fmt"
	"os"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_docBuilder builds an object schema with the fluent builder and marshals
// it to JSON. Object keys are emitted in a stable, sorted order, so the result is
// deterministic and round-trips cleanly.
func Example_docBuilder() {
	s := schema.NewBuilder().
		Schema(schema.Version).
		ID("https://example.com/user").
		Types(schema.ObjectType).
		Property("name", schema.NewBuilder().Types(schema.StringType).MinLength(1).MustBuild()).
		Property("age", schema.NewBuilder().Types(schema.IntegerType).Minimum(0).MustBuild()).
		Required("name").
		AdditionalProperties(schema.FalseSchema()).
		MustBuild()

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		fmt.Println("encode failed:", err)
	}
	// Output:
	// {
	//   "$id": "https://example.com/user",
	//   "$schema": "https://json-schema.org/draft/2020-12/schema",
	//   "additionalProperties": false,
	//   "properties": {
	//     "age": {
	//       "minimum": 0,
	//       "type": "integer"
	//     },
	//     "name": {
	//       "minLength": 1,
	//       "type": "string"
	//     }
	//   },
	//   "required": [
	//     "name"
	//   ],
	//   "type": "object"
	// }
}
