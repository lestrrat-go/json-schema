package examples_test

import (
	"fmt"
	"log"

	schema "github.com/lestrrat-go/json-schema"
)

// ExampleSchemaMap_basic demonstrates basic SchemaMap usage
func ExampleSchemaMap_basic() {
	// Create a schema with properties
	personSchema, err := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("name", schema.NonEmptyString().MustBuild()).
		Property("age", schema.PositiveInteger().MustBuild()).
		Property("email", schema.Email().MustBuild()).
		Required("name").
		Build()
	if err != nil {
		log.Fatal(err)
	}

	// Access properties using SchemaMap
	properties := personSchema.Properties()

	// Check the number of properties
	fmt.Printf("Number of properties: %d\n", properties.Len())

	// Get all property names
	propertyNames := properties.Keys()
	fmt.Printf("Property names: %v\n", propertyNames)

	// Access individual properties
	var nameSchema schema.Schema
	if err := properties.Get("name", &nameSchema); err != nil {
		fmt.Printf("Error getting name property: %v\n", err)
	} else {
		fmt.Printf("Name property type: %v\n", nameSchema.Types())
		fmt.Printf("Name property min length: %d\n", nameSchema.MinLength())
	}

	// Try to access non-existent property
	var nonExistentSchema schema.Schema
	if err := properties.Get("nonexistent", &nonExistentSchema); err != nil {
		fmt.Printf("Expected error for non-existent property: %v\n", err)
	}

	// OUTPUT:
	// Number of properties: 3
	// Property names: [age email name]
	// Name property type: [string]
	// Name property min length: 1
	// Expected error for non-existent property: schema "nonexistent" not found
}

// ExampleSchemaMap_definitions demonstrates SchemaMap with definitions
func ExampleSchemaMap_definitions() {
	// Create a schema with definitions
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
			Required("street", "city").
			MustBuild()).
		Definitions("person", schema.NewBuilder().
			Types(schema.ObjectType).
			Property("name", schema.NonEmptyString().MustBuild()).
			Property("homeAddress", schema.NewBuilder().
				Reference("#/$defs/address").
				MustBuild()).
			Required("name").
			MustBuild()).
		Types(schema.ObjectType).
		Property("owner", schema.NewBuilder().
			Reference("#/$defs/person").
			MustBuild()).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	// Access definitions using SchemaMap
	definitions := mainSchema.Definitions()

	fmt.Printf("Number of definitions: %d\n", definitions.Len())
	fmt.Printf("Definition names: %v\n", definitions.Keys())

	// Get a specific definition
	var addressSchema schema.Schema
	if err := definitions.Get("address", &addressSchema); err != nil {
		fmt.Printf("Error getting address definition: %v\n", err)
	} else {
		fmt.Printf("Address definition type: %v\n", addressSchema.Types())
		fmt.Printf("Address has properties: %t\n", addressSchema.Has(schema.PropertiesField))

		// Access nested properties within the definition
		addressProps := addressSchema.Properties()
		fmt.Printf("Address properties: %v\n", addressProps.Keys())
	}

	// OUTPUT:
	// Number of definitions: 2
	// Definition names: [address person]
	// Address definition type: [object]
	// Address has properties: true
	// Address properties: [city street zipCode]
}

// ExampleSchemaMap_patternProperties demonstrates SchemaMap with pattern properties
func ExampleSchemaMap_patternProperties() {
	// Create a schema with pattern properties
	flexibleSchema, err := schema.NewBuilder().
		Types(schema.ObjectType).
		PatternProperty("^str_", schema.NonEmptyString().MustBuild()).
		PatternProperty("^num_", schema.PositiveNumber().MustBuild()).
		PatternProperty("^bool_", schema.NewBuilder().Types(schema.BooleanType).MustBuild()).
		AdditionalProperties(schema.FalseSchema()).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	// Access pattern properties using SchemaMap
	patternProperties := flexibleSchema.PatternProperties()

	fmt.Printf("Number of pattern properties: %d\n", patternProperties.Len())
	fmt.Printf("Pattern property patterns: %v\n", patternProperties.Keys())

	// Get specific pattern property
	var stringPatternSchema schema.Schema
	if err := patternProperties.Get("^str_", &stringPatternSchema); err != nil {
		fmt.Printf("Error getting string pattern: %v\n", err)
	} else {
		fmt.Printf("String pattern type: %v\n", stringPatternSchema.Types())
		fmt.Printf("String pattern min length: %d\n", stringPatternSchema.MinLength())
	}

	// OUTPUT:
	// Number of pattern properties: 3
	// Pattern property patterns: [^bool_ ^num_ ^str_]
	// String pattern type: [string]
	// String pattern min length: 1
}

// ExampleSchemaMap_iteration demonstrates iterating over SchemaMap
func ExampleSchemaMap_iteration() {
	// Create a schema with multiple properties
	configSchema, err := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("database_url", schema.URL().MustBuild()).
		Property("api_key", schema.UUID().MustBuild()).
		Property("debug_mode", schema.NewBuilder().Types(schema.BooleanType).MustBuild()).
		Property("max_connections", schema.PositiveInteger().MustBuild()).
		Property("timeout_seconds", schema.PositiveNumber().MustBuild()).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	// Iterate over all properties
	properties := configSchema.Properties()
	fmt.Printf("Configuration schema properties (%d total):\n", properties.Len())

	for _, propName := range properties.Keys() {
		var propSchema schema.Schema
		if err := properties.Get(propName, &propSchema); err != nil {
			fmt.Printf("  %s: ERROR - %v\n", propName, err)
			continue
		}

		// Show property details
		types := propSchema.Types()
		var details []string

		if len(types) > 0 {
			details = append(details, fmt.Sprintf("type=%v", types))
		}

		if propSchema.Has(schema.FormatField) {
			details = append(details, fmt.Sprintf("format=%s", propSchema.Format()))
		}

		if propSchema.Has(schema.MinimumField) {
			details = append(details, fmt.Sprintf("minimum=%g", propSchema.Minimum()))
		}

		fmt.Printf("  %s: %v\n", propName, details)
	}

	// OUTPUT:
	// Configuration schema properties (5 total):
	//   api_key: [type=[string] format=uuid]
	//   database_url: [type=[string] format=uri]
	//   debug_mode: [type=[boolean]]
	//   max_connections: [type=[integer] minimum=0]
	//   timeout_seconds: [type=[number] minimum=0]
}

// ExampleSchemaMap_empty demonstrates handling empty SchemaMaps
func ExampleSchemaMap_empty() {
	// Create a simple schema without properties or definitions
	simpleSchema := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(1).
		MustBuild()

	// Check empty SchemaMaps
	properties := simpleSchema.Properties()
	definitions := simpleSchema.Definitions()
	patternProperties := simpleSchema.PatternProperties()

	fmt.Printf("Properties length: %d\n", properties.Len())
	fmt.Printf("Properties keys: %v\n", properties.Keys())

	fmt.Printf("Definitions length: %d\n", definitions.Len())
	fmt.Printf("Definitions keys: %v\n", definitions.Keys())

	fmt.Printf("Pattern properties length: %d\n", patternProperties.Len())
	fmt.Printf("Pattern properties keys: %v\n", patternProperties.Keys())

	// Try to get from empty SchemaMap
	var emptySchema schema.Schema
	if err := properties.Get("anything", &emptySchema); err != nil {
		fmt.Printf("Expected error from empty properties: %v\n", err)
	}

	// OUTPUT:
	// Properties length: 0
	// Properties keys: []
	// Definitions length: 0
	// Definitions keys: []
	// Pattern properties length: 0
	// Pattern properties keys: []
	// Expected error from empty properties: schema "anything" not found
}
