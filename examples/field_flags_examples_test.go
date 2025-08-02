package examples_test

import (
	"fmt"
	"log"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_fieldFlags_has demonstrates using Has() method for field checking
func Example_fieldFlags_has() {
	// Create a schema with various fields
	complexSchema, err := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(5).
		MaxLength(50).
		Pattern("^[a-zA-Z0-9]+$").
		Format("email").
		Build()
	if err != nil {
		log.Fatal(err)
	}

	// Check individual fields
	fmt.Printf("Has types: %t\n", complexSchema.Has(schema.TypesField))
	fmt.Printf("Has minLength: %t\n", complexSchema.Has(schema.MinLengthField))
	fmt.Printf("Has maxLength: %t\n", complexSchema.Has(schema.MaxLengthField))
	fmt.Printf("Has pattern: %t\n", complexSchema.Has(schema.PatternField))
	fmt.Printf("Has format: %t\n", complexSchema.Has(schema.FormatField))
	fmt.Printf("Has minimum: %t\n", complexSchema.Has(schema.MinimumField))

	// Check multiple fields at once (ALL must be present)
	fmt.Printf("Has minLength AND maxLength: %t\n", 
		complexSchema.Has(schema.MinLengthField | schema.MaxLengthField))
	fmt.Printf("Has pattern AND format: %t\n", 
		complexSchema.Has(schema.PatternField | schema.FormatField))
	fmt.Printf("Has minLength AND minimum: %t\n", 
		complexSchema.Has(schema.MinLengthField | schema.MinimumField))

	// OUTPUT:
	// Has types: true
	// Has minLength: true
	// Has maxLength: true
	// Has pattern: true
	// Has format: true
	// Has minimum: false
	// Has minLength AND maxLength: true
	// Has pattern AND format: true
	// Has minLength AND minimum: false
}

// Example_fieldFlags_hasAny demonstrates using HasAny() method for field checking
func Example_fieldFlags_hasAny() {
	// Create different types of schemas
	stringSchema := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(1).
		MustBuild()

	numberSchema := schema.NewBuilder().
		Types(schema.NumberType).
		Minimum(0).
		Maximum(100).
		MustBuild()

	objectSchema := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("name", schema.NonEmptyString().MustBuild()).
		Required("name").
		MustBuild()

	// Check if any string constraint is present
	stringConstraints := schema.MinLengthField | schema.MaxLengthField | schema.PatternField
	fmt.Printf("String schema has any string constraint: %t\n", 
		stringSchema.HasAny(stringConstraints))
	fmt.Printf("Number schema has any string constraint: %t\n", 
		numberSchema.HasAny(stringConstraints))

	// Check if any numeric constraint is present
	numericConstraints := schema.MinimumField | schema.MaximumField | schema.MultipleOfField
	fmt.Printf("String schema has any numeric constraint: %t\n", 
		stringSchema.HasAny(numericConstraints))
	fmt.Printf("Number schema has any numeric constraint: %t\n", 
		numberSchema.HasAny(numericConstraints))

	// Check if any object-related field is present
	objectConstraints := schema.PropertiesField | schema.RequiredField | schema.AdditionalPropertiesField
	fmt.Printf("Object schema has any object constraint: %t\n", 
		objectSchema.HasAny(objectConstraints))
	fmt.Printf("String schema has any object constraint: %t\n", 
		stringSchema.HasAny(objectConstraints))

	// OUTPUT:
	// String schema has any string constraint: true
	// Number schema has any string constraint: false
	// String schema has any numeric constraint: false
	// Number schema has any numeric constraint: true
	// Object schema has any object constraint: true
	// String schema has any object constraint: false
}

// Example_fieldFlags_conditionalLogic demonstrates conditional logic using field flags
func Example_fieldFlags_conditionalLogic() {
	// Create various schemas to test
	schemas := []*schema.Schema{
		schema.Email().MustBuild(),
		schema.PositiveInteger().MustBuild(),
		schema.NewBuilder().Types(schema.ArrayType).Items(schema.NonEmptyString().MustBuild()).MustBuild(),
		schema.NewBuilder().Types(schema.ObjectType).Property("test", schema.NonEmptyString().MustBuild()).MustBuild(),
	}

	for i, s := range schemas {
		fmt.Printf("Schema %d analysis:\n", i+1)
		
		// Determine schema category
		if s.Has(schema.TypesField) {
			types := s.Types()
			fmt.Printf("  Primary type: %v\n", types)
		}

		// Check validation capabilities
		var capabilities []string
		
		if s.HasAny(schema.MinLengthField | schema.MaxLengthField | schema.PatternField) {
			capabilities = append(capabilities, "string validation")
		}
		
		if s.HasAny(schema.MinimumField | schema.MaximumField | schema.MultipleOfField) {
			capabilities = append(capabilities, "numeric validation")
		}
		
		if s.HasAny(schema.PropertiesField | schema.PatternPropertiesField | schema.RequiredField) {
			capabilities = append(capabilities, "object validation")
		}
		
		if s.HasAny(schema.ItemsField | schema.MinItemsField | schema.MaxItemsField) {
			capabilities = append(capabilities, "array validation")
		}
		
		if s.HasAny(schema.FormatField) {
			capabilities = append(capabilities, "format validation")
		}

		if len(capabilities) > 0 {
			fmt.Printf("  Capabilities: %v\n", capabilities)
		} else {
			fmt.Printf("  Capabilities: basic type checking only\n")
		}

		// Check for references
		if s.Has(schema.ReferenceField) {
			fmt.Printf("  Uses reference: %s\n", s.Reference())
		}

		// Check for composition
		var composition []string
		if s.Has(schema.AllOfField) {
			composition = append(composition, "allOf")
		}
		if s.Has(schema.AnyOfField) {
			composition = append(composition, "anyOf")
		}
		if s.Has(schema.OneOfField) {
			composition = append(composition, "oneOf")
		}
		if len(composition) > 0 {
			fmt.Printf("  Uses composition: %v\n", composition)
		}

		fmt.Println()
	}

	// OUTPUT:
	// Schema 1 analysis:
	//   Primary type: [string]
	//   Capabilities: [format validation]
	// 
	// Schema 2 analysis:
	//   Primary type: [integer]
	//   Capabilities: [numeric validation]
	// 
	// Schema 3 analysis:
	//   Primary type: [array]
	//   Capabilities: [array validation]
	// 
	// Schema 4 analysis:
	//   Primary type: [object]
	//   Capabilities: [object validation]
}

// Example_fieldFlags_builderReset demonstrates using field flags with Builder.Reset()
func Example_fieldFlags_builderReset() {
	// Start with a complex schema
	builder := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(5).
		MaxLength(50).
		Pattern("^[a-zA-Z]+$").
		Format("email")

	// Build initial schema
	initialSchema := builder.MustBuild()
	fmt.Printf("Initial schema has minLength: %t\n", initialSchema.Has(schema.MinLengthField))
	fmt.Printf("Initial schema has format: %t\n", initialSchema.Has(schema.FormatField))

	// Reset specific fields and rebuild
	builder.Reset(schema.MinLengthField | schema.FormatField)
	resetSchema := builder.MustBuild()

	fmt.Printf("After reset - has minLength: %t\n", resetSchema.Has(schema.MinLengthField))
	fmt.Printf("After reset - has maxLength: %t\n", resetSchema.Has(schema.MaxLengthField))
	fmt.Printf("After reset - has pattern: %t\n", resetSchema.Has(schema.PatternField))
	fmt.Printf("After reset - has format: %t\n", resetSchema.Has(schema.FormatField))

	// OUTPUT:
	// Initial schema has minLength: true
	// Initial schema has format: true
	// After reset - has minLength: false
	// After reset - has maxLength: true
	// After reset - has pattern: true
	// After reset - has format: false
}

// Example_fieldFlags_composition demonstrates field flags with composition schemas
func Example_fieldFlags_composition() {
	// Create composition schemas
	baseString := schema.NewBuilder().Types(schema.StringType).MustBuild()
	lengthConstraint := schema.NewBuilder().MinLength(5).MaxLength(20).MustBuild()
	patternConstraint := schema.NewBuilder().Pattern("^[a-zA-Z0-9]+$").MustBuild()

	allOfSchema := schema.AllOf(baseString, lengthConstraint, patternConstraint).MustBuild()
	anyOfSchema := schema.AnyOf(
		schema.Email().MustBuild(),
		schema.UUID().MustBuild(),
		schema.URL().MustBuild(),
	).MustBuild()

	fmt.Printf("AllOf schema analysis:\n")
	fmt.Printf("  Has types: %t\n", allOfSchema.Has(schema.TypesField))
	fmt.Printf("  Has allOf: %t\n", allOfSchema.Has(schema.AllOfField))
	fmt.Printf("  Has minLength: %t\n", allOfSchema.Has(schema.MinLengthField))
	fmt.Printf("  Has composition: %t\n", 
		allOfSchema.HasAny(schema.AllOfField | schema.AnyOfField | schema.OneOfField))

	fmt.Printf("\nAnyOf schema analysis:\n")
	fmt.Printf("  Has types: %t\n", anyOfSchema.Has(schema.TypesField))
	fmt.Printf("  Has anyOf: %t\n", anyOfSchema.Has(schema.AnyOfField))
	fmt.Printf("  Has format: %t\n", anyOfSchema.Has(schema.FormatField))
	fmt.Printf("  Has composition: %t\n", 
		anyOfSchema.HasAny(schema.AllOfField | schema.AnyOfField | schema.OneOfField))

	// OUTPUT:
	// AllOf schema analysis:
	//   Has types: false
	//   Has allOf: true
	//   Has minLength: false
	//   Has composition: true
	// 
	// AnyOf schema analysis:
	//   Has types: false
	//   Has anyOf: true
	//   Has format: false
	//   Has composition: true
}