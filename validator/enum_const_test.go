package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// TestEnumValidationComprehensive tests all enum validation features across types
func TestEnumValidationComprehensive(t *testing.T) {
	t.Run("String Enum Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			enum    []any
			wantErr bool
			errMsg  string
		}{
			{
				name:    "valid string in enum",
				value:   "red",
				enum:    []any{"red", "green", "blue"},
				wantErr: false,
			},
			{
				name:    "invalid string not in enum",
				value:   "yellow",
				enum:    []any{"red", "green", "blue"},
				wantErr: true,
				errMsg:  "enum",
			},
			{
				name:    "case sensitive enum",
				value:   "Red",
				enum:    []any{"red", "green", "blue"},
				wantErr: true,
			},
			{
				name:    "empty string in enum",
				value:   "",
				enum:    []any{"", "value"},
				wantErr: false,
			},
			{
				name:    "unicode strings in enum",
				value:   "café",
				enum:    []any{"café", "naïve", "résumé"},
				wantErr: false,
			},
			{
				name:    "single value enum",
				value:   "only",
				enum:    []any{"only"},
				wantErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					Enum(tc.enum...).
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

	t.Run("Numeric Enum Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			enum    []any
			wantErr bool
			errMsg  string
		}{
			{
				name:    "valid integer in enum",
				value:   2,
				enum:    []any{1, 2, 3, 5, 8},
				wantErr: false,
			},
			{
				name:    "invalid integer not in enum",
				value:   4,
				enum:    []any{1, 2, 3, 5, 8},
				wantErr: true,
				errMsg:  "enum",
			},
			{
				name:    "zero in integer enum",
				value:   0,
				enum:    []any{-1, 0, 1},
				wantErr: false,
			},
			{
				name:    "negative numbers in enum",
				value:   -5,
				enum:    []any{-10, -5, 0, 5, 10},
				wantErr: false,
			},
			{
				name:    "float in number enum",
				value:   3.14,
				enum:    []any{2.71, 3.14, 1.41},
				wantErr: false,
			},
			{
				name:    "integer matching float in enum",
				value:   5,
				enum:    []any{2.5, 5.0, 7.5},
				wantErr: false, // 5 should match 5.0
			},
			{
				name:    "float not in enum",
				value:   2.72,
				enum:    []any{2.71, 3.14, 1.41},
				wantErr: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Determine type based on value
				schemaType := schema.IntegerType
				if _, ok := tc.value.(float64); ok {
					schemaType = schema.NumberType
				}

				s, err := schema.NewBuilder().
					Types(schemaType).
					Enum(tc.enum...).
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

	t.Run("Boolean Enum Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   bool
			enum    []any
			wantErr bool
			errMsg  string
		}{
			{
				name:    "true in enum with both values",
				value:   true,
				enum:    []any{true, false},
				wantErr: false,
			},
			{
				name:    "false in enum with both values",
				value:   false,
				enum:    []any{true, false},
				wantErr: false,
			},
			{
				name:    "true in enum with only true",
				value:   true,
				enum:    []any{true},
				wantErr: false,
			},
			{
				name:    "false not in enum with only true",
				value:   false,
				enum:    []any{true},
				wantErr: true,
				errMsg:  "enum",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Types(schema.BooleanType).
					Enum(tc.enum...).
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

	t.Run("Mixed Type Enum Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			enum    []any
			wantErr bool
			errMsg  string
		}{
			{
				name:    "string value in mixed enum",
				value:   "hello",
				enum:    []any{"hello", 42, true, 3.14},
				wantErr: false,
			},
			{
				name:    "integer value in mixed enum",
				value:   42,
				enum:    []any{"hello", 42, true, 3.14},
				wantErr: false,
			},
			{
				name:    "boolean value in mixed enum",
				value:   true,
				enum:    []any{"hello", 42, true, 3.14},
				wantErr: false,
			},
			{
				name:    "float value in mixed enum",
				value:   3.14,
				enum:    []any{"hello", 42, true, 3.14},
				wantErr: false,
			},
			{
				name:    "value not in mixed enum",
				value:   "world",
				enum:    []any{"hello", 42, true, 3.14},
				wantErr: true,
				errMsg:  "enum",
			},
			{
				name:    "null value in mixed enum",
				value:   nil,
				enum:    []any{"hello", 42, nil, true},
				wantErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// For mixed type enums, don't specify a type constraint
				s, err := schema.NewBuilder().
					Enum(tc.enum...).
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

	t.Run("Complex Object Enum Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			enum    []any
			wantErr bool
			errMsg  string
		}{
			{
				name:  "object in enum",
				value: map[string]any{"name": "John", "age": 30},
				enum: []any{
					map[string]any{"name": "John", "age": 30},
					map[string]any{"name": "Jane", "age": 25},
				},
				wantErr: false,
			},
			{
				name:  "object not in enum - different value",
				value: map[string]any{"name": "John", "age": 31},
				enum: []any{
					map[string]any{"name": "John", "age": 30},
					map[string]any{"name": "Jane", "age": 25},
				},
				wantErr: true,
				errMsg:  "enum",
			},
			{
				name:  "object not in enum - extra property",
				value: map[string]any{"name": "John", "age": 30, "city": "NYC"},
				enum: []any{
					map[string]any{"name": "John", "age": 30},
					map[string]any{"name": "Jane", "age": 25},
				},
				wantErr: true,
			},
			{
				name:  "array in enum",
				value: []any{1, 2, 3},
				enum: []any{
					[]any{1, 2, 3},
					[]any{4, 5, 6},
				},
				wantErr: false,
			},
			{
				name:  "array not in enum - different order",
				value: []any{3, 2, 1},
				enum: []any{
					[]any{1, 2, 3},
					[]any{4, 5, 6},
				},
				wantErr: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Enum(tc.enum...).
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
}

// TestConstValidationComprehensive tests all const validation features across types
func TestConstValidationComprehensive(t *testing.T) {
	t.Run("String Const Validation", func(t *testing.T) {
		testCases := []struct {
			name     string
			value    any
			constVal any
			wantErr  bool
			errMsg   string
		}{
			{
				name:     "matching string const",
				value:    "hello",
				constVal: "hello",
				wantErr:  false,
			},
			{
				name:     "non-matching string const",
				value:    "world",
				constVal: "hello",
				wantErr:  true,
				errMsg:   "const",
			},
			{
				name:     "case sensitive const",
				value:    "Hello",
				constVal: "hello",
				wantErr:  true,
			},
			{
				name:     "empty string const",
				value:    "",
				constVal: "",
				wantErr:  false,
			},
			{
				name:     "unicode string const",
				value:    "café",
				constVal: "café",
				wantErr:  false,
			},
			{
				name:     "whitespace in const",
				value:    "hello world",
				constVal: "hello world",
				wantErr:  false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					Const(tc.constVal).
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

	t.Run("Numeric Const Validation", func(t *testing.T) {
		testCases := []struct {
			name     string
			value    any
			constVal any
			wantErr  bool
			errMsg   string
		}{
			{
				name:     "matching integer const",
				value:    42,
				constVal: 42,
				wantErr:  false,
			},
			{
				name:     "non-matching integer const",
				value:    41,
				constVal: 42,
				wantErr:  true,
				errMsg:   "const",
			},
			{
				name:     "zero integer const",
				value:    0,
				constVal: 0,
				wantErr:  false,
			},
			{
				name:     "negative integer const",
				value:    -5,
				constVal: -5,
				wantErr:  false,
			},
			{
				name:     "matching float const",
				value:    3.14,
				constVal: 3.14,
				wantErr:  false,
			},
			{
				name:     "non-matching float const",
				value:    3.15,
				constVal: 3.14,
				wantErr:  true,
			},
			{
				name:     "integer matching float const",
				value:    5,
				constVal: 5.0,
				wantErr:  false, // 5 should match 5.0
			},
			{
				name:     "float matching integer const",
				value:    5.0,
				constVal: 5,
				wantErr:  false, // 5.0 should match 5
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Determine type based on value
				schemaType := schema.IntegerType
				if _, ok := tc.value.(float64); ok {
					schemaType = schema.NumberType
				}

				s, err := schema.NewBuilder().
					Types(schemaType).
					Const(tc.constVal).
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

	t.Run("Boolean Const Validation", func(t *testing.T) {
		testCases := []struct {
			name     string
			value    bool
			constVal any
			wantErr  bool
			errMsg   string
		}{
			{
				name:     "matching true const",
				value:    true,
				constVal: true,
				wantErr:  false,
			},
			{
				name:     "matching false const",
				value:    false,
				constVal: false,
				wantErr:  false,
			},
			{
				name:     "non-matching true const",
				value:    false,
				constVal: true,
				wantErr:  true,
				errMsg:   "const",
			},
			{
				name:     "non-matching false const",
				value:    true,
				constVal: false,
				wantErr:  true,
				errMsg:   "const",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Types(schema.BooleanType).
					Const(tc.constVal).
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

	t.Run("Complex Object Const Validation", func(t *testing.T) {
		testCases := []struct {
			name     string
			value    any
			constVal any
			wantErr  bool
			errMsg   string
		}{
			{
				name:     "matching object const",
				value:    map[string]any{"name": "John", "age": 30},
				constVal: map[string]any{"name": "John", "age": 30},
				wantErr:  false,
			},
			{
				name:     "non-matching object const - different value",
				value:    map[string]any{"name": "John", "age": 31},
				constVal: map[string]any{"name": "John", "age": 30},
				wantErr:  true,
				errMsg:   "const",
			},
			{
				name:     "non-matching object const - extra property",
				value:    map[string]any{"name": "John", "age": 30, "city": "NYC"},
				constVal: map[string]any{"name": "John", "age": 30},
				wantErr:  true,
			},
			{
				name:     "non-matching object const - missing property",
				value:    map[string]any{"name": "John"},
				constVal: map[string]any{"name": "John", "age": 30},
				wantErr:  true,
			},
			{
				name:     "matching array const",
				value:    []any{1, 2, 3},
				constVal: []any{1, 2, 3},
				wantErr:  false,
			},
			{
				name:     "non-matching array const - different values",
				value:    []any{1, 2, 4},
				constVal: []any{1, 2, 3},
				wantErr:  true,
			},
			{
				name:     "non-matching array const - different length",
				value:    []any{1, 2},
				constVal: []any{1, 2, 3},
				wantErr:  true,
			},
			{
				name:     "non-matching array const - different order",
				value:    []any{3, 2, 1},
				constVal: []any{1, 2, 3},
				wantErr:  true,
			},
			{
				name:     "null const",
				value:    nil,
				constVal: nil,
				wantErr:  false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Const(tc.constVal).
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

	t.Run("Const vs Enum Combination", func(t *testing.T) {
		testCases := []struct {
			name     string
			value    any
			constVal any
			enum     []any
			wantErr  bool
			errMsg   string
		}{
			{
				name:     "const and enum both satisfied",
				value:    "red",
				constVal: "red",
				enum:     []any{"red", "green", "blue"},
				wantErr:  false,
			},
			{
				name:     "const satisfied but not in enum",
				value:    "yellow",
				constVal: "yellow",
				enum:     []any{"red", "green", "blue"},
				wantErr:  true, // Should fail because const conflicts with enum
			},
			{
				name:     "in enum but doesn't match const",
				value:    "green",
				constVal: "red",
				enum:     []any{"red", "green", "blue"},
				wantErr:  true, // Should fail const validation
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					Const(tc.constVal).
					Enum(tc.enum...).
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
}
