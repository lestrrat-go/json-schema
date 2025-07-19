package validator

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
)

func TestCommonPatterns(t *testing.T) {
	testCases := []struct {
		name    string
		builder *schema.Builder
		valid   []any
		invalid []any
	}{
		{
			name:    "Email",
			builder: schema.Email(),
			valid:   []any{"user@example.com", "test.email+tag@domain.co.uk"},
			invalid: []any{"invalid-email", 123, "user@", "@domain.com"},
		},
		{
			name:    "URL",
			builder: schema.URL(),
			valid:   []any{"https://example.com", "http://localhost:8080/path"},
			invalid: []any{"not-a-url", 123, "://invalid"},
		},
		{
			name:    "NonEmptyString",
			builder: schema.NonEmptyString(),
			valid:   []any{"hello", "a", "non-empty"},
			invalid: []any{"", 123, nil},
		},
		{
			name:    "PositiveNumber",
			builder: schema.PositiveNumber(),
			valid:   []any{0, 1, 42.5, 100},
			invalid: []any{-1, -0.1, "not-a-number"},
		},
		{
			name:    "PositiveInteger",
			builder: schema.PositiveInteger(),
			valid:   []any{0, 1, 42, 100},
			invalid: []any{-1, 3.14, "not-an-integer"},
		},
		{
			name:    "UUID",
			builder: schema.UUID(),
			valid:   []any{"550e8400-e29b-41d4-a716-446655440000", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
			invalid: []any{"not-a-uuid", "550e8400-e29b-41d4-a716", 123},
		},
		{
			name:    "AlphanumericString",
			builder: schema.AlphanumericString(),
			valid:   []any{"abc123", "ABC", "123"},
			invalid: []any{"hello-world", "with spaces", "special!", 123},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schemaObj := tc.builder.MustBuild()
			// For pattern helpers, we want format validation to be enforced
			// Set up context with format-assertion enabled
			ctx := context.Background()
			ctx = WithVocabularySet(ctx, AllEnabled())
			v, err := Compile(ctx, schemaObj)
			if err != nil {
				t.Fatalf("Failed to compile validator: %v", err)
			}

			// Test valid cases
			for _, validValue := range tc.valid {
				_, err := v.Validate(context.Background(), validValue)
				if err != nil {
					t.Errorf("Expected %v to be valid, got error: %v", validValue, err)
				}
			}

			// Test invalid cases
			for _, invalidValue := range tc.invalid {
				_, err := v.Validate(context.Background(), invalidValue)
				if err == nil {
					t.Errorf("Expected %v to be invalid, but validation passed", invalidValue)
				}
			}
		})
	}
}

func ExampleEmail() {
	// Create a user schema with email validation
	userSchema := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("name", schema.NonEmptyString().MustBuild()).
		Property("email", schema.Email().MustBuild()).
		Property("age", schema.PositiveInteger().MustBuild()).
		Required("name", "email").
		MustBuild()

	// Validate user data
	validUser := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	v, _ := Compile(context.Background(), userSchema)
	_, err := v.Validate(context.Background(), validUser)
	if err != nil {
		panic("This should not happen")
	}
	// Output:
}