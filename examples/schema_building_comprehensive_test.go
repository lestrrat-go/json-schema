package examples_test

import (
	"encoding/json"
	"fmt"
	"log"

	schema "github.com/lestrrat-go/json-schema"
)

// ExampleBuilder_basic demonstrates basic schema construction using the Builder
func ExampleBuilder_basic() {
	// Build a simple string schema with constraints
	s, err := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(3).
		MaxLength(20).
		Pattern("^[a-zA-Z]+$").
		Build()
	if err != nil {
		log.Fatal(err)
	}

	// Print schema as JSON
	data, _ := json.MarshalIndent(s, "", "  ")
	fmt.Printf("String schema:\n%s\n", data)

	// OUTPUT:
	// String schema:
	// {
	//   "maxLength": 20,
	//   "minLength": 3,
	//   "pattern": "^[a-zA-Z]+$",
	//   "type": "string"
	// }
}

// ExampleBuilder_object demonstrates building object schemas with properties
func ExampleBuilder_object() {
	// Build an object schema for a person
	personSchema, err := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("firstName", schema.NonEmptyString().MustBuild()).
		Property("lastName", schema.NonEmptyString().MustBuild()).
		Property("age", schema.NewBuilder().
			Types(schema.IntegerType).
			Minimum(0).
			Maximum(150).
			MustBuild()).
		Property("email", schema.Email().MustBuild()).
		Required("firstName", "lastName").
		AdditionalProperties(schema.FalseSchema()).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	data, _ := json.MarshalIndent(personSchema, "", "  ")
	fmt.Printf("Person schema:\n%s\n", data)

	// OUTPUT:
	// Person schema:
	// {
	//   "additionalProperties": false,
	//   "properties": {
	//     "age": {
	//       "maximum": 150,
	//       "minimum": 0,
	//       "type": "integer"
	//     },
	//     "email": {
	//       "format": "email",
	//       "type": "string"
	//     },
	//     "firstName": {
	//       "minLength": 1,
	//       "type": "string"
	//     },
	//     "lastName": {
	//       "minLength": 1,
	//       "type": "string"
	//     }
	//   },
	//   "required": [
	//     "firstName",
	//     "lastName"
	//   ],
	//   "type": "object"
	// }
}

// ExampleBuilder_composition demonstrates schema composition with allOf, anyOf, oneOf
func ExampleBuilder_composition() {
	// Create a schema that must be both a string and have specific constraints
	stringWithLength := schema.AllOf(
		schema.NewBuilder().Types(schema.StringType).MustBuild(),
		schema.NewBuilder().MinLength(5).MustBuild(),
		schema.NewBuilder().MaxLength(50).MustBuild(),
	).MustBuild()

	data, _ := json.MarshalIndent(stringWithLength, "", "  ")
	fmt.Printf("AllOf composition:\n%s\n\n", data)

	// Create a schema that accepts multiple types
	stringOrNumber := schema.OneOf(
		schema.NonEmptyString().MustBuild(),
		schema.PositiveNumber().MustBuild(),
	).MustBuild()

	data, _ = json.MarshalIndent(stringOrNumber, "", "  ")
	fmt.Printf("OneOf composition:\n%s\n", data)

	// OUTPUT:
	// AllOf composition:
	// {
	//   "allOf": [
	//     {
	//       "type": "string"
	//     },
	//     {
	//       "minLength": 5
	//     },
	//     {
	//       "maxLength": 50
	//     }
	//   ]
	// }
	// 
	// OneOf composition:
	// {
	//   "oneOf": [
	//     {
	//       "minLength": 1,
	//       "type": "string"
	//     },
	//     {
	//       "minimum": 0,
	//       "type": "number"
	//     }
	//   ]
	// }
}

// ExampleBuilder_array demonstrates array schema construction
func ExampleBuilder_array() {
	// Build an array of unique positive integers with size constraints
	arraySchema, err := schema.NewBuilder().
		Types(schema.ArrayType).
		Items(schema.PositiveInteger().MustBuild()).
		MinItems(1).
		MaxItems(10).
		UniqueItems(true).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	data, _ := json.MarshalIndent(arraySchema, "", "  ")
	fmt.Printf("Array schema:\n%s\n", data)

	// OUTPUT:
	// Array schema:
	// {
	//   "items": {
	//     "minimum": 0,
	//     "type": "integer"
	//   },
	//   "maxItems": 10,
	//   "minItems": 1,
	//   "type": "array",
	//   "uniqueItems": true
	// }
}

// ExampleBuilder_clone demonstrates cloning existing schemas
func ExampleBuilder_clone() {
	// Create a base schema
	baseSchema := schema.Email().MustBuild()

	// Clone and extend the schema
	extendedSchema, err := schema.NewBuilder().
		Clone(baseSchema).
		MinLength(5).  // Add additional constraint
		MaxLength(100). // Add additional constraint
		Build()
	if err != nil {
		log.Fatal(err)
	}

	data, _ := json.MarshalIndent(extendedSchema, "", "  ")
	fmt.Printf("Extended email schema:\n%s\n", data)

	// OUTPUT:
	// Extended email schema:
	// {
	//   "format": "email",
	//   "maxLength": 100,
	//   "minLength": 5,
	//   "type": "string"
	// }
}

// ExampleBuilder_definitions demonstrates using definitions ($defs)
func ExampleBuilder_definitions() {
	// Create a person definition
	personDef := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("name", schema.NonEmptyString().MustBuild()).
		Property("age", schema.PositiveInteger().MustBuild()).
		Required("name").
		MustBuild()

	// Create a schema that uses definitions
	teamSchema, err := schema.NewBuilder().
		Schema(schema.Version). // Set the meta-schema
		Definitions("person", personDef).
		Types(schema.ObjectType).
		Property("members", schema.NewBuilder().
			Types(schema.ArrayType).
			Items(schema.NewBuilder().Reference("#/$defs/person").MustBuild()).
			MustBuild()).
		Required("members").
		Build()
	if err != nil {
		log.Fatal(err)
	}

	data, _ := json.MarshalIndent(teamSchema, "", "  ")
	fmt.Printf("Team schema with definitions:\n%s\n", data)

	// OUTPUT:
	// Team schema with definitions:
	// {
	//   "$defs": {
	//     "person": {
	//       "properties": {
	//         "age": {
	//           "minimum": 0,
	//           "type": "integer"
	//         },
	//         "name": {
	//           "minLength": 1,
	//           "type": "string"
	//         }
	//       },
	//       "required": [
	//         "name"
	//       ],
	//       "type": "object"
	//     }
	//   },
	//   "$schema": "https://json-schema.org/draft/2020-12/schema",
	//   "properties": {
	//     "members": {
	//       "items": {
	//         "$ref": "#/$defs/person"
	//       },
	//       "type": "array"
	//     }
	//   },
	//   "required": [
	//     "members"
	//   ],
	//   "type": "object"
	// }
}