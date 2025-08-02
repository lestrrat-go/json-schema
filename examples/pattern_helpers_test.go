package examples_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// Example_patterns_basic demonstrates the built-in pattern helpers
func Example_patterns_basic() {
	// Test various pattern helpers (in specific order for consistent output)
	patterns := []struct {
		name   string
		schema *schema.Schema
	}{
		{"Email", schema.Email().MustBuild()},
		{"URL", schema.URL().MustBuild()},
		{"UUID", schema.UUID().MustBuild()},
		{"Date", schema.Date().MustBuild()},
		{"DateTime", schema.DateTime().MustBuild()},
		{"NonEmptyString", schema.NonEmptyString().MustBuild()},
		{"PositiveNumber", schema.PositiveNumber().MustBuild()},
		{"PositiveInteger", schema.PositiveInteger().MustBuild()},
		{"AlphanumericString", schema.AlphanumericString().MustBuild()},
	}

	for _, p := range patterns {
		data, _ := json.Marshal(p.schema)
		fmt.Printf("%s: %s\n", p.name, data)
	}

	// OUTPUT:
	// Email: {"format":"email","type":"string"}
	// URL: {"format":"uri","type":"string"}
	// UUID: {"format":"uuid","type":"string"}
	// Date: {"format":"date","type":"string"}
	// DateTime: {"format":"date-time","type":"string"}
	// NonEmptyString: {"minLength":1,"type":"string"}
	// PositiveNumber: {"minimum":0,"type":"number"}
	// PositiveInteger: {"minimum":0,"type":"integer"}
	// AlphanumericString: {"pattern":"^[a-zA-Z0-9]+$","type":"string"}
}

// Example_patterns_enum demonstrates enum pattern helpers
func Example_patterns_enum() {
	// Create enum schemas using pattern helpers
	statusEnum := schema.Enum("active", "inactive", "pending").MustBuild()
	_ = schema.Enum(1, 2, 3, 4, 5).MustBuild()  // priorityEnum example
	_ = schema.Enum("low", 1, true, nil).MustBuild()  // mixedEnum example

	// Test the enum schemas
	ctx := context.Background()

	// Test status enum
	statusValidator, err := validator.Compile(ctx, statusEnum)
	if err != nil {
		log.Fatal(err)
	}

	testValues := []any{"active", "invalid", "pending"}
	fmt.Println("Status enum validation:")
	for _, val := range testValues {
		if _, err := statusValidator.Validate(ctx, val); err != nil {
			fmt.Printf("  %v: INVALID (%v)\n", val, err)
		} else {
			fmt.Printf("  %v: VALID\n", val)
		}
	}

	// Show the generated schema
	data, _ := json.MarshalIndent(statusEnum, "", "  ")
	fmt.Printf("\nStatus enum schema:\n%s\n", data)

	// OUTPUT:
	// Status enum validation:
	//   active: VALID
	//   invalid: INVALID (invalid value: invalid not found in enum [active inactive pending])
	//   pending: VALID
	// 
	// Status enum schema:
	// {
	//   "enum": [
	//     "active",
	//     "inactive",
	//     "pending"
	//   ]
	// }
}

// Example_patterns_composition demonstrates composition pattern helpers
func Example_patterns_composition() {
	// Create schemas using composition helpers
	stringSchema := schema.NonEmptyString().MustBuild()
	numberSchema := schema.PositiveNumber().MustBuild()
	boolSchema := schema.NewBuilder().Types(schema.BooleanType).MustBuild()

	// OneOf - exactly one must match
	oneOfSchema := schema.OneOf(stringSchema, numberSchema, boolSchema).MustBuild()
	
	// AnyOf - at least one must match  
	_ = schema.AnyOf(stringSchema, numberSchema).MustBuild()  // anyOfSchema example
	
	// AllOf - all must match
	constrainedString := schema.AllOf(
		schema.NewBuilder().Types(schema.StringType).MustBuild(),
		schema.NewBuilder().MinLength(5).MustBuild(),
		schema.NewBuilder().MaxLength(20).MustBuild(),
	).MustBuild()

	// Test OneOf schema
	ctx := context.Background()
	oneOfValidator, err := validator.Compile(ctx, oneOfSchema)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("OneOf validation results:")
	testValues := []any{"hello", 42, true, nil}
	for _, val := range testValues {
		if _, err := oneOfValidator.Validate(ctx, val); err != nil {
			fmt.Printf("  %v (%T): INVALID\n", val, val)
		} else {
			fmt.Printf("  %v (%T): VALID\n", val, val)
		}
	}

	// Show schemas
	data, _ := json.MarshalIndent(constrainedString, "", "  ")
	fmt.Printf("\nConstrained string (AllOf):\n%s\n", data)

	// OUTPUT:
	// OneOf validation results:
	//   hello (string): VALID
	//   42 (int): VALID
	//   true (bool): VALID
	//   <nil> (<nil>): INVALID
	// 
	// Constrained string (AllOf):
	// {
	//   "allOf": [
	//     {
	//       "type": "string"
	//     },
	//     {
	//       "minLength": 5
	//     },
	//     {
	//       "maxLength": 20
	//     }
	//   ]
	// }
}

// Example_patterns_optional demonstrates optional pattern helper
func Example_patterns_optional() {
	// Create an optional string schema (string or null)
	optionalString := schema.Optional(schema.NonEmptyString().MustBuild()).MustBuild()

	// Test with validator
	ctx := context.Background()
	v, err := validator.Compile(ctx, optionalString)
	if err != nil {
		log.Fatal(err)
	}

	// Test various values
	testValues := []any{"hello", nil, "", 42}
	fmt.Println("Optional string validation:")
	for _, val := range testValues {
		if _, err := v.Validate(ctx, val); err != nil {
			fmt.Printf("  %v: INVALID (%v)\n", val, err)
		} else {
			fmt.Printf("  %v: VALID\n", val)
		}
	}

	// Show the schema structure
	data, _ := json.MarshalIndent(optionalString, "", "  ")
	fmt.Printf("\nOptional string schema:\n%s\n", data)

	// OUTPUT:
	// Optional string validation:
	//   hello: VALID
	//   <nil>: VALID
	//   : INVALID (anyOf validation failed: none of the validators passed)
	//   42: INVALID (anyOf validation failed: none of the validators passed)
	// 
	// Optional string schema:
	// {
	//   "anyOf": [
	//     {
	//       "minLength": 1,
	//       "type": "string"
	//     },
	//     {
	//       "type": "null"
	//     }
	//   ]
	// }
}

// Example_patterns_combined demonstrates combining pattern helpers
func Example_patterns_combined() {
	// Create a complex schema combining multiple patterns
	userSchema := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("id", schema.UUID().MustBuild()).
		Property("email", schema.Email().MustBuild()).
		Property("website", schema.Optional(schema.URL().MustBuild()).MustBuild()).
		Property("age", schema.Optional(schema.PositiveInteger().MustBuild()).MustBuild()).
		Property("status", schema.Enum("active", "inactive", "suspended").MustBuild()).
		Property("tags", schema.NewBuilder().
			Types(schema.ArrayType).
			Items(schema.AlphanumericString().MustBuild()).
			UniqueItems(true).
			MustBuild()).
		Required("id", "email", "status").
		AdditionalProperties(schema.FalseSchema()).
		MustBuild()

	data, _ := json.MarshalIndent(userSchema, "", "  ")
	fmt.Printf("Combined pattern user schema:\n%s\n", data)

	// OUTPUT:
	// Combined pattern user schema:
	// {
	//   "additionalProperties": false,
	//   "properties": {
	//     "age": {
	//       "anyOf": [
	//         {
	//           "minimum": 0,
	//           "type": "integer"
	//         },
	//         {
	//           "type": "null"
	//         }
	//       ]
	//     },
	//     "email": {
	//       "format": "email",
	//       "type": "string"
	//     },
	//     "id": {
	//       "format": "uuid",
	//       "type": "string"
	//     },
	//     "status": {
	//       "enum": [
	//         "active",
	//         "inactive",
	//         "suspended"
	//       ]
	//     },
	//     "tags": {
	//       "items": {
	//         "pattern": "^[a-zA-Z0-9]+$",
	//         "type": "string"
	//       },
	//       "type": "array",
	//       "uniqueItems": true
	//     },
	//     "website": {
	//       "anyOf": [
	//         {
	//           "format": "uri",
	//           "type": "string"
	//         },
	//         {
	//           "type": "null"
	//         }
	//       ]
	//     }
	//   },
	//   "required": [
	//     "id",
	//     "email",
	//     "status"
	//   ],
	//   "type": "object"
	// }
}