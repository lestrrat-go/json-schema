package examples_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	schema "github.com/lestrrat-go/json-schema"
)

// ExampleResolver_basicReference demonstrates resolving local references
func ExampleResolver_basicReference() {
	// Create a schema with definitions and references
	mainSchema, err := schema.NewBuilder().
		Schema(schema.Version).
		Definitions("address", schema.NewBuilder().
			Types(schema.ObjectType).
			Property("street", schema.NonEmptyString().MustBuild()).
			Property("city", schema.NonEmptyString().MustBuild()).
			Property("zipCode", schema.NewBuilder().
				Types(schema.StringType).
				Pattern("^\\d{5}$").
				MustBuild()).
			Required("street", "city", "zipCode").
			MustBuild()).
		Types(schema.ObjectType).
		Property("homeAddress", schema.NewBuilder().
			Reference("#/$defs/address").
			MustBuild()).
		Property("workAddress", schema.NewBuilder().
			Reference("#/$defs/address").
			MustBuild()).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	// Create resolver and resolve a reference
	resolver := schema.NewResolver()
	ctx := schema.WithReferenceBase(context.Background(), mainSchema)

	var resolvedSchema schema.Schema
	err = resolver.ResolveReference(ctx, &resolvedSchema, "#/$defs/address")
	if err != nil {
		log.Fatal(err)
	}

	data, _ := json.MarshalIndent(&resolvedSchema, "", "  ")
	fmt.Printf("Resolved address schema:\n%s\n", data)

	// OUTPUT:
	// Resolved address schema:
	// {
	//   "properties": {
	//     "city": {
	//       "minLength": 1,
	//       "type": "string"
	//     },
	//     "street": {
	//       "minLength": 1,
	//       "type": "string"
	//     },
	//     "zipCode": {
	//       "pattern": "^\\d{5}$",
	//       "type": "string"
	//     }
	//   },
	//   "required": [
	//     "street",
	//     "city",
	//     "zipCode"
	//   ],
	//   "type": "object"
	// }
}

// ExampleResolver_anchorReference demonstrates resolving anchor references
func ExampleResolver_anchorReference() {
	// Create a schema with anchors
	schemaWithAnchors, err := schema.NewBuilder().
		Schema(schema.Version).
		Types(schema.ObjectType).
		Property("person", schema.NewBuilder().
			Anchor("personSchema").
			Types(schema.ObjectType).
			Property("name", schema.NonEmptyString().MustBuild()).
			Property("age", schema.PositiveInteger().MustBuild()).
			Required("name").
			MustBuild()).
		Property("spouse", schema.NewBuilder().
			Reference("#personSchema").
			MustBuild()).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	// Resolve the anchor reference
	resolver := schema.NewResolver()
	ctx := schema.WithReferenceBase(context.Background(), schemaWithAnchors)

	var resolvedSchema schema.Schema
	err = resolver.ResolveReference(ctx, &resolvedSchema, "#personSchema")
	if err != nil {
		log.Fatal(err)
	}

	data, _ := json.MarshalIndent(&resolvedSchema, "", "  ")
	fmt.Printf("Resolved person schema via anchor:\n%s\n", data)

	// OUTPUT:
	// Resolved person schema via anchor:
	// {
	//   "$anchor": "personSchema",
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

// ExampleResolver_context demonstrates using context for resolution
func ExampleResolver_context() {
	// Create a base schema for context
	baseSchema, err := schema.NewBuilder().
		Types(schema.ObjectType).
		Definitions("user", schema.NewBuilder().
			Types(schema.ObjectType).
			Property("username", schema.NonEmptyString().MustBuild()).
			Property("email", schema.Email().MustBuild()).
			Required("username", "email").
			MustBuild()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Set up context with base schema and resolver
	resolver := schema.NewResolver()
	ctx := context.Background()
	ctx = schema.WithReferenceBase(ctx, baseSchema)
	ctx = schema.WithResolver(ctx, resolver)

	// Retrieve resolver from context
	contextResolver := schema.ResolverFromContext(ctx)
	if contextResolver == nil {
		fmt.Println("No resolver found in context")
		return
	}

	fmt.Println("Successfully retrieved resolver from context")

	// Retrieve base schema from context
	contextBaseSchema := schema.ReferenceBaseFromContext(ctx)
	if contextBaseSchema == nil {
		fmt.Println("No base schema found in context")
		return
	}

	fmt.Println("Successfully retrieved base schema from context")
	fmt.Printf("Base schema has definitions: %t\n", contextBaseSchema.Has(schema.DefinitionsField))

	// OUTPUT:
	// Successfully retrieved resolver from context
	// Successfully retrieved base schema from context
	// Base schema has definitions: true
}

// ExampleValidateReference demonstrates reference validation
func ExampleValidateReference() {
	// Test various reference formats
	references := []string{
		"#/$defs/user",                                   // Local JSON pointer
		"#userAnchor",                                    // Local anchor
		"user.json#/$defs/person",                        // Relative with fragment
		"https://example.com/schema.json",                // Absolute URI
		"https://example.com/schema.json#/$defs/address", // Absolute with fragment
	}

	for _, ref := range references {
		if err := schema.ValidateReference(ref); err != nil {
			fmt.Printf("Invalid reference %q: %v\n", ref, err)
		} else {
			fmt.Printf("Valid reference: %q\n", ref)
		}
	}

	// Test invalid references
	invalidRefs := []string{
		"",          // Empty
		"not a uri", // Invalid format
	}

	for _, ref := range invalidRefs {
		if err := schema.ValidateReference(ref); err != nil {
			fmt.Printf("Expected invalid reference %q: %v\n", ref, err)
		}
	}

	// OUTPUT:
	// Valid reference: "#/$defs/user"
	// Valid reference: "#userAnchor"
	// Valid reference: "user.json#/$defs/person"
	// Valid reference: "https://example.com/schema.json"
	// Valid reference: "https://example.com/schema.json#/$defs/address"
	// Expected invalid reference "": reference cannot be empty
	// Expected invalid reference "not a uri": invalid reference format: fragment should be either a JSON Pointer starting with "/" or an anchor name
}
