package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/lestrrat-go/json-schema/vocabulary"
)

func main() {
	if err := _main(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func _main() error {
	// Get the root directory (where main go.mod is located)
	rootDir, err := findRootDir()
	if err != nil {
		return fmt.Errorf("failed to find root directory: %w", err)
	}

	// Load the main meta-schema
	metaSchemaPath := filepath.Join(rootDir, "metaschema-2020-12.json")
	metaSchemaData, err := os.ReadFile(metaSchemaPath)
	if err != nil {
		return fmt.Errorf("failed to read meta-schema file %q: %w", metaSchemaPath, err)
	}

	// Parse the meta-schema
	var metaSchema schema.Schema
	if err := json.Unmarshal(metaSchemaData, &metaSchema); err != nil {
		return fmt.Errorf("failed to unmarshal meta-schema: %w", err)
	}

	fmt.Printf("Successfully loaded meta-schema with ID: %s\n", metaSchema.ID())

	// Debug: Raw JSON check
	if bytes.Contains(metaSchemaData, []byte(`"type"`)) {
		fmt.Printf("Raw JSON contains 'type' field\n")
	} else {
		fmt.Printf("Raw JSON does NOT contain 'type' field\n")
	}

	// Debug: Check if meta-schema has types
	if metaSchema.Has(TypesField) {
		types := metaSchema.Types()
		fmt.Printf("Meta-schema has types: %v\n", types)

		// Debug: Test type compilation in isolation
		fmt.Printf("Testing if 'type' keyword is enabled in context...\n")
		testCtx := context.Background()
		testVocabSet := vocabulary.AllEnabled()
		testCtx = vocabulary.WithSet(testCtx, testVocabSet)
		isTypeEnabled := vocabulary.IsKeywordEnabledInContext(testCtx, "type")
		fmt.Printf("Type keyword enabled: %v\n", isTypeEnabled)

		// Debug: Test what base schema would be created
		fmt.Printf("Simulating createBaseSchema behavior...\n")
		baseBuilder := schema.NewBuilder()
		if len(metaSchema.Types()) > 0 {
			baseBuilder.Types(metaSchema.Types()...)
			fmt.Printf("Base schema would have types: %v\n", metaSchema.Types())
		}
		baseSchema := baseBuilder.MustBuild()

		// Test compiling the base schema in isolation
		baseValidator, err := validator.Compile(testCtx, baseSchema)
		if err != nil {
			fmt.Printf("Base schema compilation failed: %v\n", err)
		} else {
			fmt.Printf("Base schema compiled successfully to: %T\n", baseValidator)

			// Test the base validator
			testValues := []any{"string", 123, true, map[string]any{"key": "value"}}
			for _, value := range testValues {
				_, err := baseValidator.Validate(testCtx, value)
				fmt.Printf("Base validator - Value %T: %v\n", value, err == nil)
			}

			// Debug: Check which allOf compilation path will be taken
			fmt.Printf("Checking allOf compilation path...\n")
			fmt.Printf("metaSchema.Has(AllOfField): %v\n", metaSchema.Has(AllOfField))
			fmt.Printf("hasBaseConstraints(metaSchema): %v\n", len(metaSchema.Types()) > 0) // This is what hasBaseConstraints checks for types

			// Check if it has unevaluated fields that would trigger special handling
			hasUnevaluatedProperties := metaSchema.Has(UnevaluatedPropertiesField)
			hasUnevaluatedItems := metaSchema.Has(UnevaluatedItemsField)
			fmt.Printf("metaSchema.Has(UnevaluatedPropertiesField): %v\n", hasUnevaluatedProperties)
			fmt.Printf("metaSchema.Has(UnevaluatedItemsField): %v\n", hasUnevaluatedItems)
		}
	} else {
		fmt.Printf("Meta-schema has NO types field!\n")
	}

	// Compile the meta-schema to a validator
	ctx := context.Background()

	// Set up vocabulary context for JSON Schema 2020-12
	// Use AllEnabled to ensure all vocabularies are enabled for meta-schema compilation
	vocabSet := vocabulary.AllEnabled()
	ctx = vocabulary.WithSet(ctx, vocabSet)

	// Set up resolver for meta-schema references
	resolver := schema.NewResolver()
	ctx = schema.WithResolver(ctx, resolver)

	// Set up base schema context for reference resolution
	ctx = schema.WithBaseSchema(ctx, &metaSchema)

	// Set up base URI for relative reference resolution within metaschema
	// Use the directory base URI, not the specific schema file URI
	// This allows meta/core to resolve to https://json-schema.org/draft/2020-12/meta/core
	ctx = schema.WithBaseURI(ctx, "https://json-schema.org/draft/2020-12/")

	// Compile the meta-schema to a validator
	compiledValidator, err := validator.Compile(ctx, &metaSchema)
	if err != nil {
		return fmt.Errorf("failed to compile meta-schema: %w", err)
	}

	fmt.Printf("Successfully compiled meta-schema to validator of type: %T\n", compiledValidator)

	// Debug: Print the structure of the compiled validator
	debugValidator(compiledValidator, 0)

	// Use the code generator to generate the builder chain
	generator := validator.NewCodeGenerator()
	var builderBuf bytes.Buffer

	err = generator.Generate(&builderBuf, compiledValidator)
	if err != nil {
		return fmt.Errorf("failed to generate validator code: %w", err)
	}

	builderChain := builderBuf.String()
	fmt.Printf("Generated builder chain: %s\n", builderChain)

	// Create the package content using the generated builder chain
	packageContent := fmt.Sprintf(`// Code generated by internal/cmd/genmeta. DO NOT EDIT.

// Package meta provides a pre-compiled validator for JSON Schema 2020-12 meta-schema.
// This validator can be used to validate JSON Schema documents themselves.
package meta

import (
	"context"
	"github.com/lestrrat-go/json-schema/keywords"
	"github.com/lestrrat-go/json-schema/validator"
)

func init() {
	// Ensure keywords package is imported (referenced by generated code)
	_ = keywords.Schema
}

// metaValidator holds the pre-compiled validator for the JSON Schema 2020-12 meta-schema
var metaValidator validator.Interface

func init() {
	// Generated validator using the code generator from the actual meta-schema
	metaValidator = %s
}

// Validator returns a pre-compiled validator for the JSON Schema 2020-12 meta-schema.
// This validator can be used to validate JSON Schema documents themselves.
//
// Example usage:
//
//	validator := meta.Validator()
//	result, err := validator.Validate(ctx, jsonSchemaDocument)
func Validator() validator.Interface {
	return metaValidator
}

// Validate validates a JSON Schema document against the JSON Schema 2020-12 meta-schema.
// This is a convenience function that uses the pre-compiled validator.
//
// Example usage:
//
//	err := meta.Validate(ctx, jsonSchemaDocument)
//	if err != nil {
//	    // The document is not a valid JSON Schema
//	}
func Validate(ctx context.Context, jsonSchemaDocument any) error {
	_, err := metaValidator.Validate(ctx, jsonSchemaDocument)
	return err
}
`, builderChain)

	// Format the generated code
	formattedCode, err := format.Source([]byte(packageContent))
	if err != nil {
		scanner := bufio.NewScanner(strings.NewReader(packageContent))
		line := 1
		for scanner.Scan() {
			fmt.Printf("%03d: %s\n", line, scanner.Text())
			line++
		}
		return fmt.Errorf("failed to format generated code: %w", err)
	}

	// Ensure meta directory exists
	metaDir := filepath.Join(rootDir, "meta")
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		return fmt.Errorf("failed to create meta directory: %w", err)
	}

	// Write the generated code to meta/meta_gen.go
	outputPath := filepath.Join(metaDir, "meta_gen.go")
	if err := os.WriteFile(outputPath, formattedCode, 0644); err != nil {
		return fmt.Errorf("failed to write generated code to %q: %w", outputPath, err)
	}

	fmt.Printf("Successfully generated meta-schema validator at: %s\n", outputPath)
	return nil
}

// findRootDir finds the root directory containing the main go.mod file
func findRootDir() (string, error) {
	// Start from current directory and walk up until we find go.mod
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Check if this is the main go.mod (contains github.com/lestrrat-go/json-schema)
			content, err := os.ReadFile(goModPath)
			if err == nil && len(content) > 0 {
				// Simple check - if it doesn't contain "replace", it's likely the main module
				if !bytes.Contains(content, []byte("replace github.com/lestrrat-go/json-schema")) {
					return dir, nil
				}
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find root go.mod file")
}

// debugValidator recursively prints the structure of validators to help debug what's being compiled
func debugValidator(v validator.Interface, depth int) {
	if v == nil {
		return
	}

	if depth > 10 {
		fmt.Printf("%s... (max depth reached)\n", strings.Repeat("  ", depth))
		return
	}

	indent := strings.Repeat("  ", depth)
	vType := reflect.TypeOf(v)
	if vType.Kind() == reflect.Ptr {
		vType = vType.Elem()
	}

	fmt.Printf("%s%s (%T)\n", indent, vType.Name(), v)

	// Try to access the internal structure using reflection for known types
	vValue := reflect.ValueOf(v)
	if vValue.Kind() == reflect.Ptr {
		vValue = vValue.Elem()
	}

	// Handle MultiValidator specifically
	if vType.Name() == "MultiValidator" {
		// Look for validators field
		if vValue.IsValid() {
			for i := 0; i < vValue.NumField(); i++ {
				field := vValue.Field(i)
				fieldType := vValue.Type().Field(i)

				if fieldType.Name == "validators" && field.Kind() == reflect.Slice {
					fmt.Printf("%s  %d validators:\n", indent, field.Len())
					for j := 0; j < field.Len(); j++ {
						child := field.Index(j)
						if child.CanInterface() {
							if childValidator, ok := child.Interface().(validator.Interface); ok {
								debugValidator(childValidator, depth+2)
							}
						}
					}
				} else if fieldType.Name == "mode" {
					fmt.Printf("%s  mode: %v\n", indent, field.Interface())
				}
			}
		}
	}

	// Print all fields for other validator types to see their structure
	if vType.Name() != "MultiValidator" && vValue.IsValid() {
		for i := 0; i < vValue.NumField(); i++ {
			field := vValue.Field(i)
			fieldType := vValue.Type().Field(i)

			// Skip unexported fields that might cause issues
			if !fieldType.IsExported() {
				continue
			}

			// Print field values for debugging
			if field.CanInterface() {
				switch field.Kind() {
				case reflect.Slice:
					fmt.Printf("%s  %s: [%d items]\n", indent, fieldType.Name, field.Len())
					if field.Len() < 5 { // Only show details for small slices
						for j := 0; j < field.Len(); j++ {
							item := field.Index(j)
							if item.CanInterface() {
								fmt.Printf("%s    [%d]: %v\n", indent, j, item.Interface())
							}
						}
					}
				case reflect.Map:
					fmt.Printf("%s  %s: map[%d entries]\n", indent, fieldType.Name, field.Len())
					if field.Len() < 5 { // Only show details for small maps
						for _, key := range field.MapKeys() {
							value := field.MapIndex(key)
							if key.CanInterface() && value.CanInterface() {
								fmt.Printf("%s    %v: %v\n", indent, key.Interface(), value.Interface())
							}
						}
					}
				case reflect.Ptr:
					if !field.IsNil() {
						fmt.Printf("%s  %s: %v\n", indent, fieldType.Name, field.Interface())
					} else {
						fmt.Printf("%s  %s: nil\n", indent, fieldType.Name)
					}
				default:
					fmt.Printf("%s  %s: %v\n", indent, fieldType.Name, field.Interface())
				}
			}
		}
	}
}
