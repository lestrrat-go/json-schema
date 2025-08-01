package validator

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

func TestCompositeValidatorWithRefAndOtherConstraints(t *testing.T) {
	// Create a definition schema
	definitionSchema := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(3).
		MustBuild()

	// Create a root schema with the definition
	rootSchema := schema.NewBuilder().
		Definitions("StringType", definitionSchema).
		MustBuild()

	// Create a schema with both $ref and additional constraints
	schemaWithRef := schema.NewBuilder().
		Reference("#/$defs/StringType").
		MaxLength(10).
		Pattern("^[a-zA-Z]+$").
		MustBuild()

	// Test the hasOtherConstraints function
	require.True(t, hasOtherConstraints(schemaWithRef))

	// Test the createSchemaWithoutRef function
	withoutRef := createSchemaWithoutRef(schemaWithRef)
	require.False(t, withoutRef.Has(schema.ReferenceField))
	require.True(t, withoutRef.Has(schema.MaxLengthField))
	require.True(t, withoutRef.Has(schema.PatternField))
	require.Equal(t, schemaWithRef.MaxLength(), withoutRef.MaxLength())
	require.Equal(t, schemaWithRef.Pattern(), withoutRef.Pattern())

	// Test compilation and validation
	ctx := context.Background()
	ctx = schema.WithResolver(ctx, schema.NewResolver())
	ctx = schema.WithRootSchema(ctx, rootSchema)
	ctx = schema.WithBaseSchema(ctx, rootSchema)

	validator, err := Compile(ctx, schemaWithRef)
	require.NoError(t, err)

	// Test validation with various inputs
	testCases := []struct {
		name        string
		input       string
		shouldPass  bool
		description string
	}{
		{
			name:        "valid_string",
			input:       "hello",
			shouldPass:  true,
			description: "Should pass: length 5 (3-10), all letters",
		},
		{
			name:        "too_short",
			input:       "hi",
			shouldPass:  false,
			description: "Should fail: length 2 (below minLength 3 from $ref)",
		},
		{
			name:        "too_long",
			input:       "verylongstring",
			shouldPass:  false,
			description: "Should fail: length 14 (above maxLength 10)",
		},
		{
			name:        "contains_numbers",
			input:       "hello123",
			shouldPass:  false,
			description: "Should fail: contains numbers (violates pattern)",
		},
		{
			name:        "valid_long",
			input:       "abcdefghij",
			shouldPass:  true,
			description: "Should pass: length 10 (exactly maxLength), all letters",
		},
		{
			name:        "valid_short",
			input:       "abc",
			shouldPass:  true,
			description: "Should pass: length 3 (exactly minLength), all letters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validator.Validate(ctx, tc.input)
			if tc.shouldPass {
				require.NoError(t, err, "Test case: %s - %s", tc.name, tc.description)
			} else {
				require.Error(t, err, "Test case: %s - %s", tc.name, tc.description)
			}
		})
	}
}

func TestCompositeValidatorWithComplexRef(t *testing.T) {
	// Create a complex definition schema with object constraints
	definitionSchema := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
		Required("name").
		MustBuild()

	// Create a root schema with the definition
	rootSchema := schema.NewBuilder().
		Definitions("Person", definitionSchema).
		MustBuild()

	// Create a schema with both $ref and additional object constraints
	schemaWithRef := schema.NewBuilder().
		Reference("#/$defs/Person").
		Property("age", schema.NewBuilder().Types(schema.IntegerType).Minimum(0).MustBuild()).
		MaxProperties(5).
		MustBuild()

	// Test compilation and validation
	ctx := context.Background()
	ctx = schema.WithResolver(ctx, schema.NewResolver())
	ctx = schema.WithRootSchema(ctx, rootSchema)
	ctx = schema.WithBaseSchema(ctx, rootSchema)

	validator, err := Compile(ctx, schemaWithRef)
	require.NoError(t, err)

	// Test validation with various inputs
	testCases := []struct {
		name        string
		input       map[string]any
		shouldPass  bool
		description string
	}{
		{
			name:        "valid_person_with_age",
			input:       map[string]any{"name": "John", "age": 30},
			shouldPass:  true,
			description: "Should pass: has required name, age is valid integer",
		},
		{
			name:        "missing_required_name",
			input:       map[string]any{"age": 30},
			shouldPass:  false,
			description: "Should fail: missing required name field from $ref",
		},
		{
			name:        "negative_age",
			input:       map[string]any{"name": "John", "age": -5},
			shouldPass:  false,
			description: "Should fail: negative age violates minimum constraint",
		},
		{
			name:        "valid_person_no_age",
			input:       map[string]any{"name": "John"},
			shouldPass:  true,
			description: "Should pass: has required name, age is optional",
		},
		{
			name:        "too_many_properties",
			input:       map[string]any{"name": "John", "age": 30, "email": "john@example.com", "phone": "123-456-7890", "address": "123 Main St", "country": "US"},
			shouldPass:  false,
			description: "Should fail: 6 properties exceeds maxProperties 5",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validator.Validate(ctx, tc.input)
			if tc.shouldPass {
				require.NoError(t, err, "Test case: %s - %s", tc.name, tc.description)
			} else {
				require.Error(t, err, "Test case: %s - %s", tc.name, tc.description)
			}
		})
	}
}
