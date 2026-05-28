package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_clone_and_reset derives a new schema from an existing one. Builder.Clone
// copies all of a schema's fields into a builder, and Builder.Reset clears the
// named fields (identified by field-flag constants). Here a "relaxed" variant of
// a string schema is produced by cloning the base and dropping its maxLength —
// equivalent to building the relaxed schema directly.
func Example_clone_and_reset() {
	// The base schema authored as JSON: string, length 3..8.
	base := loadSchemaJSON(`{
		"type": "string",
		"minLength": 3,
		"maxLength": 8
	}`)

	// Two ways to reach the same goal: build the relaxed schema directly, or
	// clone the base and reset the maxLength field.
	relaxedDirect := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(3).
		MustBuild()
	relaxedClone := schema.NewBuilder().
		Clone(base).
		Reset(schema.MaxLengthField).
		MustBuild()

	const long = "abcdefghijklmnop" // 16 chars, exceeds the base maxLength of 8

	fmt.Printf("base accepts long string:    %t\n", valid(base, long))
	fmt.Printf("relaxed (direct) accepts it: %t\n", valid(relaxedDirect, long))
	fmt.Printf("relaxed (clone)  accepts it: %t\n", valid(relaxedClone, long))
	// Output:
	// base accepts long string:    false
	// relaxed (direct) accepts it: true
	// relaxed (clone)  accepts it: true
}
