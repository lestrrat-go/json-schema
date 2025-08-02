package examples_test

import (
	"context"
	"fmt"
	"log"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// ExampleValidator_basic demonstrates basic validation using the validator package
func ExampleValidator_basic() {
	// Create a simple string schema that requires minimum length of 3
	s := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(3).
		MustBuild()

	// Compile the schema into a validator
	v, err := validator.Compile(context.Background(), s)
	if err != nil {
		log.Fatal(err)
	}

	// Test valid data
	ctx := context.Background()
	if _, err := v.Validate(ctx, "hello"); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("Valid: 'hello' passes validation")
	}

	// Test invalid data
	if _, err := v.Validate(ctx, "hi"); err != nil {
		fmt.Printf("Invalid: %v\n", err)
	} else {
		fmt.Println("Unexpected: 'hi' should have failed")
	}

	// OUTPUT:
	// Valid: 'hello' passes validation
	// Invalid: string length validation failed: value has 2 characters, needs at least 3
}

// ExampleValidator_object demonstrates object validation with properties
func ExampleValidator_object() {
	// Create a schema for a user object
	userSchema := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("name", schema.NonEmptyString().MustBuild()).
		Property("age", schema.PositiveInteger().MustBuild()).
		Property("email", schema.Email().MustBuild()).
		Required("name", "email").
		AdditionalProperties(schema.FalseSchema()).
		MustBuild()

	v, err := validator.Compile(context.Background(), userSchema)
	if err != nil {
		log.Fatal(err)
	}

	// Valid user
	validUser := map[string]any{
		"name":  "Alice",
		"age":   30,
		"email": "alice@example.com",
	}

	ctx := context.Background()
	if _, err := v.Validate(ctx, validUser); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("Valid user object passed validation")
	}

	// Invalid user - missing required field
	invalidUser := map[string]any{
		"name": "Bob",
		// missing email
	}

	if _, err := v.Validate(ctx, invalidUser); err != nil {
		fmt.Printf("Invalid user: %v\n", err)
	}

	// OUTPUT:
	// Valid user object passed validation
	// Invalid user: required field "email" is missing
}

// ExampleValidator_composition demonstrates schema composition with anyOf
func ExampleValidator_composition() {
	// Create a schema that accepts either a string or a number
	stringOrNumber := schema.AnyOf(
		schema.NonEmptyString().MustBuild(),
		schema.PositiveNumber().MustBuild(),
	).MustBuild()

	v, err := validator.Compile(context.Background(), stringOrNumber)
	if err != nil {
		log.Fatal(err)
	}

	// Test string value
	ctx := context.Background()
	if _, err := v.Validate(ctx, "hello"); err != nil {
		fmt.Printf("String validation failed: %v\n", err)
	} else {
		fmt.Println("String value accepted")
	}

	// Test number value
	if _, err := v.Validate(ctx, 42.5); err != nil {
		fmt.Printf("Number validation failed: %v\n", err)
	} else {
		fmt.Println("Number value accepted")
	}

	// Test invalid value
	if _, err := v.Validate(ctx, true); err != nil {
		fmt.Printf("Boolean rejected: %v\n", err)
	}

	// OUTPUT:
	// String value accepted
	// Number value accepted
	// Boolean rejected: anyOf validation failed: none of the schemas matched
}

// ExampleValidator_array demonstrates array validation with items and constraints
func ExampleValidator_array() {
	// Create schema for array of positive integers with max 5 items
	arraySchema := schema.NewBuilder().
		Types(schema.ArrayType).
		Items(schema.PositiveInteger().MustBuild()).
		MaxItems(5).
		UniqueItems(true).
		MustBuild()

	v, err := validator.Compile(context.Background(), arraySchema)
	if err != nil {
		log.Fatal(err)
	}

	// Valid array
	ctx := context.Background()
	validArray := []any{1, 2, 3, 4, 5}
	if _, err := v.Validate(ctx, validArray); err != nil {
		fmt.Printf("Valid array failed: %v\n", err)
	} else {
		fmt.Println("Valid array passed")
	}

	// Invalid - too many items
	tooManyItems := []any{1, 2, 3, 4, 5, 6}
	if _, err := v.Validate(ctx, tooManyItems); err != nil {
		fmt.Printf("Too many items: %v\n", err)
	}

	// Invalid - duplicate items
	duplicateItems := []any{1, 2, 2, 3}
	if _, err := v.Validate(ctx, duplicateItems); err != nil {
		fmt.Printf("Duplicate items: %v\n", err)
	}

	// OUTPUT:
	// Valid array passed
	// Too many items: array has 6 items, maximum allowed is 5
	// Duplicate items: array contains duplicate items
}