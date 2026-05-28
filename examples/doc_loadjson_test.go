package examples_test

import (
	"context"
	"encoding/json"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// Example_docLoadJSON loads a schema authored as JSON. *schema.Schema implements
// json.Unmarshaler, so json.Unmarshal is all it takes; the result compiles and
// validates exactly like a schema built with the fluent builder.
func Example_docLoadJSON() {
	const doc = `{
		"type": "object",
		"properties": { "city": { "type": "string", "minLength": 1 } },
		"required": ["city"]
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(doc), &s); err != nil {
		fmt.Println("parse failed:", err)
		return
	}

	ctx := context.Background()
	v, err := validator.Compile(ctx, &s)
	if err != nil {
		fmt.Println("compile failed:", err)
		return
	}

	_, err = v.Validate(ctx, map[string]any{"city": "Kyoto"})
	fmt.Println("with city:   ", err == nil)
	_, err = v.Validate(ctx, map[string]any{})
	fmt.Println("without city:", err == nil)
	// Output:
	// with city:    true
	// without city: false
}
