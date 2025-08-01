package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// TestArrayValidatorComprehensive tests all array validation features
func TestArrayValidatorComprehensive(t *testing.T) {
	t.Run("Basic Array Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			wantErr bool
			errMsg  string
		}{
			{
				name:    "valid empty array",
				value:   []any{},
				wantErr: false,
			},
			{
				name:    "valid array with elements",
				value:   []any{1, 2, 3},
				wantErr: false,
			},
			{
				name:    "valid string array",
				value:   []string{"a", "b", "c"},
				wantErr: false,
			},
			{
				name:    "valid int array",
				value:   []int{1, 2, 3},
				wantErr: false,
			},
			{
				name:    "valid mixed type array",
				value:   []any{1, "string", true, 3.14},
				wantErr: false,
			},
			{
				name:    "valid nested array",
				value:   []any{[]any{1, 2}, []any{3, 4}},
				wantErr: false,
			},
			{
				name:    "string should fail",
				value:   "not an array",
				wantErr: true,
				errMsg:  "expected array",
			},
			{
				name:    "object should fail",
				value:   map[string]any{"key": "value"},
				wantErr: true,
				errMsg:  "expected array",
			},
			{
				name:    "integer should fail",
				value:   123,
				wantErr: true,
				errMsg:  "expected array",
			},
			{
				name:    "nil should fail",
				value:   nil,
				wantErr: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().Types(schema.ArrayType).Build()
				require.NoError(t, err)

				v, err := validator.Compile(context.Background(), s)
				require.NoError(t, err)

				_, err = v.Validate(context.Background(), tc.value)
				if tc.wantErr {
					require.Error(t, err)
					if tc.errMsg != "" {
						require.Contains(t, err.Error(), tc.errMsg)
					}
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("Array Length Constraints", func(t *testing.T) {
		testCases := []struct {
			name     string
			value    []any
			minItems *uint
			maxItems *uint
			wantErr  bool
			errMsg   string
		}{
			// MinItems tests
			{
				name:     "valid minItems",
				value:    []any{1, 2, 3},
				minItems: uintPtr(2),
				wantErr:  false,
			},
			{
				name:     "exact minItems",
				value:    []any{1, 2},
				minItems: uintPtr(2),
				wantErr:  false,
			},
			{
				name:     "below minItems",
				value:    []any{1},
				minItems: uintPtr(2),
				wantErr:  true,
				errMsg:   "minimum items",
			},
			{
				name:     "empty array with minItems 1",
				value:    []any{},
				minItems: uintPtr(1),
				wantErr:  true,
			},
			{
				name:     "empty array with minItems 0",
				value:    []any{},
				minItems: uintPtr(0),
				wantErr:  false,
			},
			// MaxItems tests
			{
				name:     "valid maxItems",
				value:    []any{1, 2},
				maxItems: uintPtr(5),
				wantErr:  false,
			},
			{
				name:     "exact maxItems",
				value:    []any{1, 2, 3},
				maxItems: uintPtr(3),
				wantErr:  false,
			},
			{
				name:     "above maxItems",
				value:    []any{1, 2, 3, 4},
				maxItems: uintPtr(3),
				wantErr:  true,
				errMsg:   "maximum items",
			},
			// Combined tests
			{
				name:     "within min and max range",
				value:    []any{1, 2, 3},
				minItems: uintPtr(2),
				maxItems: uintPtr(5),
				wantErr:  false,
			},
			{
				name:     "below minItems with maxItems set",
				value:    []any{1},
				minItems: uintPtr(2),
				maxItems: uintPtr(5),
				wantErr:  true,
			},
			{
				name:     "above maxItems with minItems set",
				value:    []any{1, 2, 3, 4, 5, 6},
				minItems: uintPtr(2),
				maxItems: uintPtr(5),
				wantErr:  true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := schema.NewBuilder().Types(schema.ArrayType)
				if tc.minItems != nil {
					builder = builder.MinItems(*tc.minItems)
				}
				if tc.maxItems != nil {
					builder = builder.MaxItems(*tc.maxItems)
				}
				s, err := builder.Build()
				require.NoError(t, err)

				v, err := validator.Compile(context.Background(), s)
				require.NoError(t, err)

				_, err = v.Validate(context.Background(), tc.value)
				if tc.wantErr {
					require.Error(t, err)
					if tc.errMsg != "" {
						require.Contains(t, err.Error(), tc.errMsg)
					}
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("Items Schema Validation", func(t *testing.T) {
		testCases := []struct {
			name        string
			value       []any
			itemsSchema *schema.Schema
			wantErr     bool
			errMsg      string
		}{
			{
				name:        "all items match string schema",
				value:       []any{"a", "b", "c"},
				itemsSchema: schema.NewBuilder().Types(schema.StringType).MustBuild(),
				wantErr:     false,
			},
			{
				name:        "all items match integer schema",
				value:       []any{1, 2, 3},
				itemsSchema: schema.NewBuilder().Types(schema.IntegerType).MustBuild(),
				wantErr:     false,
			},
			{
				name:        "mixed types with string schema",
				value:       []any{"a", 1, "c"},
				itemsSchema: schema.NewBuilder().Types(schema.StringType).MustBuild(),
				wantErr:     true,
				errMsg:      "item validation failed",
			},
			{
				name:        "empty array with items schema",
				value:       []any{},
				itemsSchema: schema.NewBuilder().Types(schema.StringType).MustBuild(),
				wantErr:     false,
			},
			{
				name: "items with complex schema",
				value: []any{
					map[string]any{"name": "John", "age": 30},
					map[string]any{"name": "Jane", "age": 25},
				},
				itemsSchema: schema.NewBuilder().
					Types(schema.ObjectType).
					Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
					Property("age", schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
					MustBuild(),
				wantErr: false,
			},
			{
				name: "nested array items",
				value: []any{
					[]any{1, 2, 3},
					[]any{4, 5, 6},
				},
				itemsSchema: schema.NewBuilder().
					Types(schema.ArrayType).
					Items(schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
					MustBuild(),
				wantErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Types(schema.ArrayType).
					Items(tc.itemsSchema).
					Build()
				require.NoError(t, err)

				v, err := validator.Compile(context.Background(), s)
				require.NoError(t, err)

				_, err = v.Validate(context.Background(), tc.value)
				if tc.wantErr {
					require.Error(t, err)
					if tc.errMsg != "" {
						require.Contains(t, err.Error(), tc.errMsg)
					}
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("Tuple Validation (Positional Items)", func(t *testing.T) {
		testCases := []struct {
			name            string
			value           []any
			prefixItems     []*schema.Schema
			additionalItems any // bool or *schema.Schema
			wantErr         bool
			errMsg          string
		}{
			{
				name:  "exact tuple match",
				value: []any{"John", 30, true},
				prefixItems: []*schema.Schema{
					schema.NewBuilder().Types(schema.StringType).MustBuild(),
					schema.NewBuilder().Types(schema.IntegerType).MustBuild(),
					schema.NewBuilder().Types(schema.BooleanType).MustBuild(),
				},
				wantErr: false,
			},
			{
				name:  "tuple with fewer items",
				value: []any{"John", 30},
				prefixItems: []*schema.Schema{
					schema.NewBuilder().Types(schema.StringType).MustBuild(),
					schema.NewBuilder().Types(schema.IntegerType).MustBuild(),
					schema.NewBuilder().Types(schema.BooleanType).MustBuild(),
				},
				wantErr: false,
			},
			{
				name:  "tuple type mismatch",
				value: []any{123, 30, true}, // first should be string
				prefixItems: []*schema.Schema{
					schema.NewBuilder().Types(schema.StringType).MustBuild(),
					schema.NewBuilder().Types(schema.IntegerType).MustBuild(),
					schema.NewBuilder().Types(schema.BooleanType).MustBuild(),
				},
				wantErr: true,
				errMsg:  "prefixItems[0] validation failed",
			},
			{
				name:  "additional items allowed",
				value: []any{"John", 30, true, "extra", 42},
				prefixItems: []*schema.Schema{
					schema.NewBuilder().Types(schema.StringType).MustBuild(),
					schema.NewBuilder().Types(schema.IntegerType).MustBuild(),
					schema.NewBuilder().Types(schema.BooleanType).MustBuild(),
				},
				additionalItems: true,
				wantErr:         false,
			},
			{
				name:  "additional items forbidden",
				value: []any{"John", 30, true, "extra"},
				prefixItems: []*schema.Schema{
					schema.NewBuilder().Types(schema.StringType).MustBuild(),
					schema.NewBuilder().Types(schema.IntegerType).MustBuild(),
					schema.NewBuilder().Types(schema.BooleanType).MustBuild(),
				},
				additionalItems: false,
				wantErr:         true,
				errMsg:          "additionalItems validation failed",
			},
			{
				name:  "additional items with schema - valid strings",
				value: []any{"John", 30, true, "extra1", "extra2"},
				prefixItems: []*schema.Schema{
					schema.NewBuilder().Types(schema.StringType).MustBuild(),
					schema.NewBuilder().Types(schema.IntegerType).MustBuild(),
					schema.NewBuilder().Types(schema.BooleanType).MustBuild(),
				},
				additionalItems: schema.NewBuilder().Types(schema.StringType).MustBuild(),
				wantErr:         false,
			},
			{
				name:  "additional items violate schema",
				value: []any{"John", 30, true, 123}, // additional should be string
				prefixItems: []*schema.Schema{
					schema.NewBuilder().Types(schema.StringType).MustBuild(),
					schema.NewBuilder().Types(schema.IntegerType).MustBuild(),
					schema.NewBuilder().Types(schema.BooleanType).MustBuild(),
				},
				additionalItems: schema.NewBuilder().Types(schema.StringType).MustBuild(),
				wantErr:         true,
				errMsg:          "additionalItems validation failed",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := schema.NewBuilder().Types(schema.ArrayType)

				// Set prefix items (tuple validation) - use PrefixItems with all schemas at once
				builder = builder.PrefixItems(tc.prefixItems...)

				// Set additional items policy - AdditionalItems accepts SchemaOrBool
				switch ai := tc.additionalItems.(type) {
				case bool:
					builder = builder.AdditionalItems(schema.BoolSchema(ai))
				case *schema.Schema:
					builder = builder.AdditionalItems(ai)
				}

				s, err := builder.Build()
				require.NoError(t, err)

				v, err := validator.Compile(context.Background(), s)
				require.NoError(t, err)

				_, err = v.Validate(context.Background(), tc.value)
				if tc.wantErr {
					require.Error(t, err)
					if tc.errMsg != "" {
						require.Contains(t, err.Error(), tc.errMsg)
					}
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("Unique Items Validation", func(t *testing.T) {
		testCases := []struct {
			name        string
			value       []any
			uniqueItems bool
			wantErr     bool
			errMsg      string
		}{
			{
				name:        "unique items - all unique",
				value:       []any{1, 2, 3, "a", "b"},
				uniqueItems: true,
				wantErr:     false,
			},
			{
				name:        "unique items - duplicates",
				value:       []any{1, 2, 2, 3},
				uniqueItems: true,
				wantErr:     true,
				errMsg:      "duplicate items",
			},
			{
				name:        "unique items - string duplicates",
				value:       []any{"a", "b", "a"},
				uniqueItems: true,
				wantErr:     true,
				errMsg:      "duplicate items",
			},
			{
				name:        "unique items disabled - duplicates allowed",
				value:       []any{1, 1, 2, 2},
				uniqueItems: false,
				wantErr:     false,
			},
			{
				name:        "unique items - empty array",
				value:       []any{},
				uniqueItems: true,
				wantErr:     false,
			},
			{
				name:        "unique items - single item",
				value:       []any{42},
				uniqueItems: true,
				wantErr:     false,
			},
			{
				name:        "unique items - mixed types no duplicates",
				value:       []any{1, "1", true, 1.0},
				uniqueItems: true,
				wantErr:     false, // Different types are considered different
			},
			{
				name: "unique items - object duplicates",
				value: []any{
					map[string]any{"id": 1, "name": "John"},
					map[string]any{"id": 2, "name": "Jane"},
					map[string]any{"id": 1, "name": "John"}, // duplicate
				},
				uniqueItems: true,
				wantErr:     true,
			},
			{
				name: "unique items - array duplicates",
				value: []any{
					[]any{1, 2, 3},
					[]any{4, 5, 6},
					[]any{1, 2, 3}, // duplicate
				},
				uniqueItems: true,
				wantErr:     true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Types(schema.ArrayType).
					UniqueItems(tc.uniqueItems).
					Build()
				require.NoError(t, err)

				v, err := validator.Compile(context.Background(), s)
				require.NoError(t, err)

				_, err = v.Validate(context.Background(), tc.value)
				if tc.wantErr {
					require.Error(t, err)
					if tc.errMsg != "" {
						require.Contains(t, err.Error(), tc.errMsg)
					}
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("Contains Validation", func(t *testing.T) {
		testCases := []struct {
			name        string
			value       []any
			contains    *schema.Schema
			minContains *uint
			maxContains *uint
			wantErr     bool
			errMsg      string
		}{
			{
				name:     "contains string - found",
				value:    []any{1, "hello", 3, "world"},
				contains: schema.NewBuilder().Types(schema.StringType).MustBuild(),
				wantErr:  false,
			},
			{
				name:     "contains string - not found",
				value:    []any{1, 2, 3, 4},
				contains: schema.NewBuilder().Types(schema.StringType).MustBuild(),
				wantErr:  true,
				errMsg:   "does not contain required item",
			},
			{
				name:        "minContains satisfied",
				value:       []any{"a", "b", 1, "c"},
				contains:    schema.NewBuilder().Types(schema.StringType).MustBuild(),
				minContains: uintPtr(2),
				wantErr:     false,
			},
			{
				name:        "minContains not satisfied",
				value:       []any{"a", 1, 2, 3},
				contains:    schema.NewBuilder().Types(schema.StringType).MustBuild(),
				minContains: uintPtr(2),
				wantErr:     true,
				errMsg:      "minimum contains",
			},
			{
				name:        "maxContains satisfied",
				value:       []any{"a", "b", 1, 2},
				contains:    schema.NewBuilder().Types(schema.StringType).MustBuild(),
				maxContains: uintPtr(3),
				wantErr:     false,
			},
			{
				name:        "maxContains exceeded",
				value:       []any{"a", "b", "c", "d", 1},
				contains:    schema.NewBuilder().Types(schema.StringType).MustBuild(),
				maxContains: uintPtr(2),
				wantErr:     true,
				errMsg:      "maximum contains",
			},
			{
				name:        "contains range satisfied",
				value:       []any{"a", "b", "c", 1, 2},
				contains:    schema.NewBuilder().Types(schema.StringType).MustBuild(),
				minContains: uintPtr(2),
				maxContains: uintPtr(4),
				wantErr:     false,
			},
			{
				name: "contains complex schema",
				value: []any{
					map[string]any{"type": "user", "name": "John"},
					map[string]any{"type": "admin", "name": "Jane"},
					123,
				},
				contains: schema.NewBuilder().
					Types(schema.ObjectType).
					Property("type", schema.NewBuilder().
						Types(schema.StringType).
						Const("admin").
						MustBuild()).
					MustBuild(),
				wantErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := schema.NewBuilder().Types(schema.ArrayType)
				if tc.contains != nil {
					builder = builder.Contains(tc.contains)
				}
				if tc.minContains != nil {
					builder = builder.MinContains(*tc.minContains)
				}
				if tc.maxContains != nil {
					builder = builder.MaxContains(*tc.maxContains)
				}
				s, err := builder.Build()
				require.NoError(t, err)

				v, err := validator.Compile(context.Background(), s)
				require.NoError(t, err)

				_, err = v.Validate(context.Background(), tc.value)
				if tc.wantErr {
					require.Error(t, err)
					if tc.errMsg != "" {
						require.Contains(t, err.Error(), tc.errMsg)
					}
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("Combined Array Constraints", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   []any
			schema  func() *schema.Schema
			wantErr bool
			errMsg  string
		}{
			{
				name:  "all constraints satisfied",
				value: []any{"hello", "world", "unique"},
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Types(schema.ArrayType).
						Items(schema.NewBuilder().Types(schema.StringType).MustBuild()).
						MinItems(2).
						MaxItems(5).
						UniqueItems(true).
						Build()
					return s
				},
				wantErr: false,
			},
			{
				name:  "items valid but not unique",
				value: []any{"hello", "world", "hello"},
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Types(schema.ArrayType).
						Items(schema.NewBuilder().Types(schema.StringType).MustBuild()).
						MinItems(2).
						MaxItems(5).
						UniqueItems(true).
						Build()
					return s
				},
				wantErr: true,
				errMsg:  "duplicate items",
			},
			{
				name:  "unique but wrong item type",
				value: []any{"hello", 123, "world"},
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Types(schema.ArrayType).
						Items(schema.NewBuilder().Types(schema.StringType).MustBuild()).
						MinItems(2).
						MaxItems(5).
						UniqueItems(true).
						Build()
					return s
				},
				wantErr: true,
				errMsg:  "item validation failed",
			},
			{
				name:  "valid items and unique but too few",
				value: []any{"hello"},
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Types(schema.ArrayType).
						Items(schema.NewBuilder().Types(schema.StringType).MustBuild()).
						MinItems(2).
						MaxItems(5).
						UniqueItems(true).
						Build()
					return s
				},
				wantErr: true,
				errMsg:  "minimum items",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				v, err := validator.Compile(context.Background(), tc.schema())
				require.NoError(t, err)

				_, err = v.Validate(context.Background(), tc.value)
				if tc.wantErr {
					require.Error(t, err)
					if tc.errMsg != "" {
						require.Contains(t, err.Error(), tc.errMsg)
					}
				} else {
					require.NoError(t, err)
				}
			})
		}
	})
}
