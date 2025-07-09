package validator_test

import (
	"math"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// TestIntegerValidatorComprehensive tests all integer validation features
func TestIntegerValidatorComprehensive(t *testing.T) {
	t.Run("Basic Integer Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			wantErr bool
			errMsg  string
		}{
			{
				name:    "valid positive integer",
				value:   42,
				wantErr: false,
			},
			{
				name:    "valid negative integer",
				value:   -42,
				wantErr: false,
			},
			{
				name:    "zero integer",
				value:   0,
				wantErr: false,
			},
			{
				name:    "large integer",
				value:   1234567890,
				wantErr: false,
			},
			{
				name:    "max int",
				value:   math.MaxInt64,
				wantErr: false,
			},
			{
				name:    "min int",
				value:   math.MinInt64,
				wantErr: false,
			},
			{
				name:    "float value should fail",
				value:   42.5,
				wantErr: true,
				errMsg:  "expected integer",
			},
			{
				name:    "string value should fail",
				value:   "42",
				wantErr: true,
				errMsg:  "expected integer",
			},
			{
				name:    "boolean value should fail",
				value:   true,
				wantErr: true,
			},
			{
				name:    "nil value should fail",
				value:   nil,
				wantErr: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().Type(schema.IntegerType).Build()
				require.NoError(t, err)

				v, err := validator.Compile(s)
				require.NoError(t, err)

				err = v.Validate(tc.value)
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

	t.Run("Range Constraints", func(t *testing.T) {
		testCases := []struct {
			name      string
			value     int
			minimum   *float64
			maximum   *float64
			wantErr   bool
			errMsg    string
		}{
			// Minimum tests
			{
				name:    "valid minimum",
				value:   10,
				minimum: float64Ptr(5),
				wantErr: false,
			},
			{
				name:    "exact minimum",
				value:   5,
				minimum: float64Ptr(5),
				wantErr: false,
			},
			{
				name:    "below minimum",
				value:   3,
				minimum: float64Ptr(5),
				wantErr: true,
				errMsg:  "minimum",
			},
			{
				name:    "negative minimum",
				value:   -5,
				minimum: float64Ptr(-10),
				wantErr: false,
			},
			// Maximum tests
			{
				name:    "valid maximum",
				value:   5,
				maximum: float64Ptr(10),
				wantErr: false,
			},
			{
				name:    "exact maximum",
				value:   10,
				maximum: float64Ptr(10),
				wantErr: false,
			},
			{
				name:    "above maximum",
				value:   15,
				maximum: float64Ptr(10),
				wantErr: true,
				errMsg:  "maximum",
			},
			{
				name:    "negative maximum",
				value:   -15,
				maximum: float64Ptr(-10),
				wantErr: true,
			},
			// Combined tests
			{
				name:    "within range",
				value:   7,
				minimum: float64Ptr(5),
				maximum: float64Ptr(10),
				wantErr: false,
			},
			{
				name:    "below range",
				value:   3,
				minimum: float64Ptr(5),
				maximum: float64Ptr(10),
				wantErr: true,
			},
			{
				name:    "above range",
				value:   15,
				minimum: float64Ptr(5),
				maximum: float64Ptr(10),
				wantErr: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := schema.NewBuilder().Type(schema.IntegerType)
				if tc.minimum != nil {
					builder = builder.Minimum(*tc.minimum)
				}
				if tc.maximum != nil {
					builder = builder.Maximum(*tc.maximum)
				}
				s, err := builder.Build()
				require.NoError(t, err)

				v, err := validator.Compile(s)
				require.NoError(t, err)

				err = v.Validate(tc.value)
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

	t.Run("Exclusive Range Constraints", func(t *testing.T) {
		testCases := []struct {
			name             string
			value            int
			exclusiveMinimum *float64
			exclusiveMaximum *float64
			wantErr          bool
			errMsg           string
		}{
			// Exclusive minimum tests
			{
				name:             "valid exclusive minimum",
				value:            10,
				exclusiveMinimum: float64Ptr(5),
				wantErr:          false,
			},
			{
				name:             "equal to exclusive minimum should fail",
				value:            5,
				exclusiveMinimum: float64Ptr(5),
				wantErr:          true,
				errMsg:           "exclusive",
			},
			{
				name:             "below exclusive minimum",
				value:            3,
				exclusiveMinimum: float64Ptr(5),
				wantErr:          true,
			},
			// Exclusive maximum tests
			{
				name:             "valid exclusive maximum",
				value:            5,
				exclusiveMaximum: float64Ptr(10),
				wantErr:          false,
			},
			{
				name:             "equal to exclusive maximum should fail",
				value:            10,
				exclusiveMaximum: float64Ptr(10),
				wantErr:          true,
				errMsg:           "exclusive",
			},
			{
				name:             "above exclusive maximum",
				value:            15,
				exclusiveMaximum: float64Ptr(10),
				wantErr:          true,
			},
			// Combined tests
			{
				name:             "within exclusive range",
				value:            7,
				exclusiveMinimum: float64Ptr(5),
				exclusiveMaximum: float64Ptr(10),
				wantErr:          false,
			},
			{
				name:             "at exclusive minimum boundary",
				value:            5,
				exclusiveMinimum: float64Ptr(5),
				exclusiveMaximum: float64Ptr(10),
				wantErr:          true,
			},
			{
				name:             "at exclusive maximum boundary",
				value:            10,
				exclusiveMinimum: float64Ptr(5),
				exclusiveMaximum: float64Ptr(10),
				wantErr:          true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := schema.NewBuilder().Type(schema.IntegerType)
				if tc.exclusiveMinimum != nil {
					builder = builder.ExclusiveMinimum(*tc.exclusiveMinimum)
				}
				if tc.exclusiveMaximum != nil {
					builder = builder.ExclusiveMaximum(*tc.exclusiveMaximum)
				}
				s, err := builder.Build()
				require.NoError(t, err)

				v, err := validator.Compile(s)
				require.NoError(t, err)

				err = v.Validate(tc.value)
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

	t.Run("MultipleOf Constraint", func(t *testing.T) {
		testCases := []struct {
			name       string
			value      int
			multipleOf float64
			wantErr    bool
			errMsg     string
		}{
			{
				name:       "valid multiple of 5",
				value:      15,
				multipleOf: 5,
				wantErr:    false,
			},
			{
				name:       "invalid multiple of 5",
				value:      16,
				multipleOf: 5,
				wantErr:    true,
				errMsg:     "multiple",
			},
			{
				name:       "multiple of 1 (any integer)",
				value:      42,
				multipleOf: 1,
				wantErr:    false,
			},
			{
				name:       "multiple of 10",
				value:      100,
				multipleOf: 10,
				wantErr:    false,
			},
			{
				name:       "not multiple of 10",
				value:      105,
				multipleOf: 10,
				wantErr:    true,
			},
			{
				name:       "zero as multiple of any number",
				value:      0,
				multipleOf: 7,
				wantErr:    false,
			},
			{
				name:       "negative number multiple",
				value:      -15,
				multipleOf: 5,
				wantErr:    false,
			},
			{
				name:       "negative number not multiple",
				value:      -16,
				multipleOf: 5,
				wantErr:    true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Type(schema.IntegerType).
					MultipleOf(tc.multipleOf).
					Build()
				require.NoError(t, err)

				v, err := validator.Compile(s)
				require.NoError(t, err)

				err = v.Validate(tc.value)
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

	t.Run("Enum and Const for Integers", func(t *testing.T) {
		testCases := []struct {
			name     string
			value    int
			enum     []any
			constVal any
			wantErr  bool
			errMsg   string
		}{
			{
				name:    "valid enum value",
				value:   2,
				enum:    []any{1, 2, 3, 5, 8},
				wantErr: false,
			},
			{
				name:    "invalid enum value",
				value:   4,
				enum:    []any{1, 2, 3, 5, 8},
				wantErr: true,
				errMsg:  "enum",
			},
			{
				name:     "valid const value",
				value:    42,
				constVal: 42,
				wantErr:  false,
			},
			{
				name:     "invalid const value",
				value:    41,
				constVal: 42,
				wantErr:  true,
				errMsg:   "const",
			},
			{
				name:    "negative numbers in enum",
				value:   -5,
				enum:    []any{-10, -5, 0, 5, 10},
				wantErr: false,
			},
			{
				name:    "zero in enum",
				value:   0,
				enum:    []any{-1, 0, 1},
				wantErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := schema.NewBuilder().Type(schema.IntegerType)
				if tc.enum != nil {
					builder = builder.Enum(tc.enum...)
				}
				if tc.constVal != nil {
					builder = builder.Const(tc.constVal)
				}
				s, err := builder.Build()
				require.NoError(t, err)

				v, err := validator.Compile(s)
				require.NoError(t, err)

				err = v.Validate(tc.value)
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

// TestNumberValidatorComprehensive tests all number validation features
func TestNumberValidatorComprehensive(t *testing.T) {
	t.Run("Basic Number Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			wantErr bool
			errMsg  string
		}{
			{
				name:    "valid integer as number",
				value:   42,
				wantErr: false,
			},
			{
				name:    "valid float as number",
				value:   42.5,
				wantErr: false,
			},
			{
				name:    "valid negative number",
				value:   -42.5,
				wantErr: false,
			},
			{
				name:    "zero as number",
				value:   0,
				wantErr: false,
			},
			{
				name:    "zero float as number",
				value:   0.0,
				wantErr: false,
			},
			{
				name:    "very small number",
				value:   0.000001,
				wantErr: false,
			},
			{
				name:    "very large number",
				value:   1.23e10,
				wantErr: false,
			},
			{
				name:    "infinity should be valid",
				value:   math.Inf(1),
				wantErr: false,
			},
			{
				name:    "negative infinity should be valid",
				value:   math.Inf(-1),
				wantErr: false,
			},
			{
				name:    "NaN should be invalid",
				value:   math.NaN(),
				wantErr: true,
				errMsg:  "NaN",
			},
			{
				name:    "string value should fail",
				value:   "42.5",
				wantErr: true,
				errMsg:  "expected number",
			},
			{
				name:    "boolean value should fail",
				value:   true,
				wantErr: true,
			},
			{
				name:    "nil value should fail",
				value:   nil,
				wantErr: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().Type(schema.NumberType).Build()
				require.NoError(t, err)

				v, err := validator.Compile(s)
				require.NoError(t, err)

				err = v.Validate(tc.value)
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

	t.Run("Number Range Constraints", func(t *testing.T) {
		testCases := []struct {
			name      string
			value     float64
			minimum   *float64
			maximum   *float64
			wantErr   bool
			errMsg    string
		}{
			// Float range tests
			{
				name:    "float within range",
				value:   7.5,
				minimum: float64Ptr(5.0),
				maximum: float64Ptr(10.0),
				wantErr: false,
			},
			{
				name:    "float at minimum boundary",
				value:   5.0,
				minimum: float64Ptr(5.0),
				wantErr: false,
			},
			{
				name:    "float at maximum boundary",
				value:   10.0,
				maximum: float64Ptr(10.0),
				wantErr: false,
			},
			{
				name:    "float below minimum",
				value:   4.9,
				minimum: float64Ptr(5.0),
				wantErr: true,
				errMsg:  "minimum",
			},
			{
				name:    "float above maximum",
				value:   10.1,
				maximum: float64Ptr(10.0),
				wantErr: true,
				errMsg:  "maximum",
			},
			// Precision tests
			{
				name:    "high precision float within range",
				value:   7.123456789,
				minimum: float64Ptr(7.0),
				maximum: float64Ptr(8.0),
				wantErr: false,
			},
			{
				name:    "scientific notation within range",
				value:   1.23e-5,
				minimum: float64Ptr(0.0),
				maximum: float64Ptr(1.0),
				wantErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := schema.NewBuilder().Type(schema.NumberType)
				if tc.minimum != nil {
					builder = builder.Minimum(*tc.minimum)
				}
				if tc.maximum != nil {
					builder = builder.Maximum(*tc.maximum)
				}
				s, err := builder.Build()
				require.NoError(t, err)

				v, err := validator.Compile(s)
				require.NoError(t, err)

				err = v.Validate(tc.value)
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

	t.Run("Number MultipleOf Constraint", func(t *testing.T) {
		testCases := []struct {
			name       string
			value      float64
			multipleOf float64
			wantErr    bool
			errMsg     string
		}{
			{
				name:       "float multiple of 0.5",
				value:      2.5,
				multipleOf: 0.5,
				wantErr:    false,
			},
			{
				name:       "float not multiple of 0.5",
				value:      2.3,
				multipleOf: 0.5,
				wantErr:    true,
				errMsg:     "multiple",
			},
			{
				name:       "integer multiple of float",
				value:      10.0,
				multipleOf: 2.5,
				wantErr:    false,
			},
			{
				name:       "precise decimal multiple",
				value:      0.75,
				multipleOf: 0.25,
				wantErr:    false,
			},
			{
				name:       "imprecise decimal not multiple",
				value:      0.77,
				multipleOf: 0.25,
				wantErr:    true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Type(schema.NumberType).
					MultipleOf(tc.multipleOf).
					Build()
				require.NoError(t, err)

				v, err := validator.Compile(s)
				require.NoError(t, err)

				err = v.Validate(tc.value)
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

	t.Run("Combined Number Constraints", func(t *testing.T) {
		testCases := []struct {
			name   string
			value  float64
			schema func() *schema.Schema
			wantErr bool
			errMsg  string
		}{
			{
				name:  "all constraints satisfied",
				value: 10.0,
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Type(schema.NumberType).
						Minimum(5.0).
						Maximum(15.0).
						MultipleOf(2.5).
						Build()
					return s
				},
				wantErr: false,
			},
			{
				name:  "range satisfied but not multiple",
				value: 11.0,
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Type(schema.NumberType).
						Minimum(5.0).
						Maximum(15.0).
						MultipleOf(2.5).
						Build()
					return s
				},
				wantErr: true,
			},
			{
				name:  "multiple satisfied but out of range",
				value: 20.0,
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Type(schema.NumberType).
						Minimum(5.0).
						Maximum(15.0).
						MultipleOf(2.5).
						Build()
					return s
				},
				wantErr: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				v, err := validator.Compile(tc.schema())
				require.NoError(t, err)

				err = v.Validate(tc.value)
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

// Helper function to create float64 pointers
func float64Ptr(f float64) *float64 {
	return &f
}