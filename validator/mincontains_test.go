package validator

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

func TestMinContainsZeroBehavior(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		containsSchema *schema.Schema
		minContains    *uint
		maxContains    *uint
		input          []any
		shouldPass     bool
		description    string
	}{
		{
			name:           "minContains_0_empty_array",
			containsSchema: schema.NewBuilder().Types(schema.StringType).MustBuild(),
			minContains:    uintPtr(0),
			maxContains:    nil,
			input:          []any{},
			shouldPass:     true,
			description:    "Empty array should pass with minContains=0",
		},
		{
			name:           "minContains_0_no_matches",
			containsSchema: schema.NewBuilder().Types(schema.StringType).MustBuild(),
			minContains:    uintPtr(0),
			maxContains:    nil,
			input:          []any{1, 2, 3},
			shouldPass:     true,
			description:    "Array with no matches should pass with minContains=0",
		},
		{
			name:           "minContains_0_with_matches",
			containsSchema: schema.NewBuilder().Types(schema.StringType).MustBuild(),
			minContains:    uintPtr(0),
			maxContains:    nil,
			input:          []any{"hello", 1, 2},
			shouldPass:     true,
			description:    "Array with matches should pass with minContains=0",
		},
		{
			name:           "minContains_0_with_maxContains_2",
			containsSchema: schema.NewBuilder().Types(schema.StringType).MustBuild(),
			minContains:    uintPtr(0),
			maxContains:    uintPtr(2),
			input:          []any{"hello", "world", "extra"},
			shouldPass:     false,
			description:    "Array with 3 matches should fail with minContains=0, maxContains=2",
		},
		{
			name:           "minContains_1_empty_array",
			containsSchema: schema.NewBuilder().Types(schema.StringType).MustBuild(),
			minContains:    uintPtr(1),
			maxContains:    nil,
			input:          []any{},
			shouldPass:     false,
			description:    "Empty array should fail with minContains=1",
		},
		{
			name:           "minContains_1_no_matches",
			containsSchema: schema.NewBuilder().Types(schema.StringType).MustBuild(),
			minContains:    uintPtr(1),
			maxContains:    nil,
			input:          []any{1, 2, 3},
			shouldPass:     false,
			description:    "Array with no matches should fail with minContains=1",
		},
		{
			name:           "minContains_1_with_matches",
			containsSchema: schema.NewBuilder().Types(schema.StringType).MustBuild(),
			minContains:    uintPtr(1),
			maxContains:    nil,
			input:          []any{"hello", 1, 2},
			shouldPass:     true,
			description:    "Array with 1 match should pass with minContains=1",
		},
		{
			name:           "no_minContains_empty_array",
			containsSchema: schema.NewBuilder().Types(schema.StringType).MustBuild(),
			minContains:    nil,
			maxContains:    nil,
			input:          []any{},
			shouldPass:     false,
			description:    "Empty array should fail with contains (default behavior requires at least 1)",
		},
		{
			name:           "no_minContains_no_matches",
			containsSchema: schema.NewBuilder().Types(schema.StringType).MustBuild(),
			minContains:    nil,
			maxContains:    nil,
			input:          []any{1, 2, 3},
			shouldPass:     false,
			description:    "Array with no matches should fail with contains (default behavior requires at least 1)",
		},
		{
			name:           "no_minContains_with_matches",
			containsSchema: schema.NewBuilder().Types(schema.StringType).MustBuild(),
			minContains:    nil,
			maxContains:    nil,
			input:          []any{"hello", 1, 2},
			shouldPass:     true,
			description:    "Array with matches should pass with contains (default behavior requires at least 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the schema
			builder := schema.NewBuilder().Types(schema.ArrayType)
			builder.Contains(tt.containsSchema)
			
			if tt.minContains != nil {
				builder.MinContains(*tt.minContains)
			}
			if tt.maxContains != nil {
				builder.MaxContains(*tt.maxContains)
			}

			testSchema := builder.MustBuild()

			// Compile the validator
			validator, err := Compile(ctx, testSchema)
			require.NoError(t, err, "Failed to compile validator")

			// Test validation
			_, err = validator.Validate(ctx, tt.input)
			if tt.shouldPass {
				require.NoError(t, err, "Test case: %s - %s", tt.name, tt.description)
			} else {
				require.Error(t, err, "Test case: %s - %s", tt.name, tt.description)
			}
		})
	}
}

// Helper function to create uint pointer
func uintPtr(v uint) *uint {
	return &v
}