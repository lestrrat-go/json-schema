package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/lestrrat-go/json-schema/vocabulary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringConstraintSanity(t *testing.T) {
	testcases := makeSanityTestCases()
	for _, tc := range testcases {
		switch tc.Name {
		case "String":
		default:
			tc.Error = true
		}
	}

	c := validator.String().MustBuild()
	for _, tc := range testcases {
		t.Run(tc.Name, makeSanityTestFunc(tc, c))
	}
}

func TestStringValidator(t *testing.T) {
	testcases := []struct {
		Name      string
		Object    any
		Validator func() (validator.Interface, error)
		Error     bool
	}{
		{
			Name:   "no maxLength set, no minLength set, value within range",
			Object: "Hello, World!",
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(context.Background(), s)
			},
		},
		{
			Name:   "maxLength set, no minLength set, value within range",
			Object: "Hello, World!",
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					MaxLength(20).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(context.Background(), s)
			},
		},
		{
			Name:   "maxLength set, minLength set, value within range",
			Object: "Hello, World!",
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					MinLength(1).
					MaxLength(20).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(context.Background(), s)
			},
		},
		{
			Name:   "maxLength set, minLength set, value within range",
			Object: "Hello, World!",
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					MinLength(5).
					MaxLength(20).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(context.Background(), s)
			},
		},
		{
			Name:   "maxLength set, no minLength set, value outside range (l > max)",
			Object: "Hello, World!",
			Error:  true,
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					MaxLength(5).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(context.Background(), s)
			},
		},
		{
			Name:   "maxLength set, minLength set, value outside range (l > max)",
			Object: "Hello, World!",
			Error:  true,
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					MinLength(1).
					MaxLength(5).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(context.Background(), s)
			},
		},
		{
			Name:   "maxLength set, minLength set, value outside range (l < min)",
			Object: "Hello, World!",
			Error:  true,
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					MinLength(14).
					MaxLength(20).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(context.Background(), s)
			},
		},
		{
			Name:   "pattern set, valid value",
			Object: "Hello, World!",
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					Pattern(`^Hello, .+$`).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(context.Background(), s)
			},
		},
		{
			Name:   "pattern set, invalid value",
			Object: "Hello, World!",
			Error:  true,
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					Pattern(`^Night, .+$`).
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(context.Background(), s)
			},
		},
		{
			Name:   "enum set, valid value",
			Object: "three",
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					Enum("one", "two", "three").
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(context.Background(), s)
			},
		},
		{
			Name:   "enum set, invalid value",
			Object: "four",
			Error:  true,
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					Enum("one", "two", "three").
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(context.Background(), s)
			},
		},
		{
			Name:   "const set, valid value",
			Object: "Hello, World!",
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					Const("Hello, World!").
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(context.Background(), s)
			},
		},
		{
			Name:   "const set, invalid value",
			Object: "Night, World!",
			Error:  true,
			Validator: func() (validator.Interface, error) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					Const("Hello, World!").
					Build()
				if err != nil {
					return nil, err
				}
				return validator.Compile(context.Background(), s)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			c, err := tc.Validator()
			if !assert.NoError(t, err, `tc.Validator() should succeed`) {
				return
			}
			_, err = c.Validate(context.Background(), tc.Object)

			if tc.Error {
				if !assert.Error(t, err, `c.Validate should fail`) {
					return
				}
			} else {
				if !assert.NoError(t, err, `c.Validate should succeed`) {
					return
				}
			}
		})
	}
}

// TestStringValidatorComprehensive tests all string validation features
func TestStringValidatorComprehensive(t *testing.T) {
	t.Run("Basic String Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			schema  func() *schema.Schema
			wantErr bool
			errMsg  string
		}{
			{
				name:  "valid string",
				value: "hello",
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().Types(schema.StringType).Build()
					return s
				},
				wantErr: false,
			},
			{
				name:  "non-string value",
				value: 123,
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().Types(schema.StringType).Build()
					return s
				},
				wantErr: true,
				errMsg:  "expected string",
			},
			{
				name:  "nil value",
				value: nil,
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().Types(schema.StringType).Build()
					return s
				},
				wantErr: true,
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

	t.Run("Length Constraints", func(t *testing.T) {
		testCases := []struct {
			name      string
			value     string
			minLength *int
			maxLength *int
			wantErr   bool
			errMsg    string
		}{
			// MinLength tests
			{
				name:      "valid minLength",
				value:     "hello",
				minLength: intPtr(3),
				wantErr:   false,
			},
			{
				name:      "exact minLength",
				value:     "hello",
				minLength: intPtr(5),
				wantErr:   false,
			},
			{
				name:      "below minLength",
				value:     "hi",
				minLength: intPtr(5),
				wantErr:   true,
				errMsg:    "shorter then minLength",
			},
			{
				name:      "empty string with minLength 1",
				value:     "",
				minLength: intPtr(1),
				wantErr:   true,
			},
			{
				name:      "empty string with minLength 0",
				value:     "",
				minLength: intPtr(0),
				wantErr:   false,
			},
			// MaxLength tests
			{
				name:      "valid maxLength",
				value:     "hello",
				maxLength: intPtr(10),
				wantErr:   false,
			},
			{
				name:      "exact maxLength",
				value:     "hello",
				maxLength: intPtr(5),
				wantErr:   false,
			},
			{
				name:      "above maxLength",
				value:     "hello world",
				maxLength: intPtr(5),
				wantErr:   true,
				errMsg:    "longer then maxLength",
			},
			// Combined tests
			{
				name:      "within min and max range",
				value:     "hello",
				minLength: intPtr(3),
				maxLength: intPtr(10),
				wantErr:   false,
			},
			{
				name:      "below minLength with maxLength set",
				value:     "hi",
				minLength: intPtr(3),
				maxLength: intPtr(10),
				wantErr:   true,
			},
			{
				name:      "above maxLength with minLength set",
				value:     "hello world and more",
				minLength: intPtr(3),
				maxLength: intPtr(10),
				wantErr:   true,
			},
			// Unicode and special characters
			{
				name:      "unicode characters within length",
				value:     "héllo",
				minLength: intPtr(3),
				maxLength: intPtr(10),
				wantErr:   false,
			},
			{
				name:      "emoji within length",
				value:     "hello 😀",
				minLength: intPtr(5),
				maxLength: intPtr(10),
				wantErr:   false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := schema.NewBuilder().Types(schema.StringType)
				if tc.minLength != nil {
					builder = builder.MinLength(*tc.minLength)
				}
				if tc.maxLength != nil {
					builder = builder.MaxLength(*tc.maxLength)
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

	t.Run("Pattern Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   string
			pattern string
			wantErr bool
			errMsg  string
		}{
			{
				name:    "simple pattern match",
				value:   "hello",
				pattern: "^hello$",
				wantErr: false,
			},
			{
				name:    "simple pattern no match",
				value:   "world",
				pattern: "^hello$",
				wantErr: true,
				errMsg:  "did not match pattern",
			},
			{
				name:    "pattern with wildcards",
				value:   "hello world",
				pattern: "hello.*",
				wantErr: false,
			},
			{
				name:    "email pattern valid",
				value:   "user@example.com",
				pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
				wantErr: false,
			},
			{
				name:    "email pattern invalid",
				value:   "invalid-email",
				pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
				wantErr: true,
			},
			{
				name:    "phone pattern valid",
				value:   "+1-555-123-4567",
				pattern: `^\+\d{1,3}-\d{3}-\d{3}-\d{4}$`,
				wantErr: false,
			},
			{
				name:    "case sensitive pattern",
				value:   "Hello",
				pattern: "^hello$",
				wantErr: true,
			},
			{
				name:    "case insensitive pattern",
				value:   "Hello",
				pattern: "(?i)^hello$",
				wantErr: false,
			},
			{
				name:    "unicode pattern",
				value:   "café",
				pattern: "^café$",
				wantErr: false,
			},
			{
				name:    "empty string with pattern",
				value:   "",
				pattern: "^$",
				wantErr: false,
			},
			{
				name:    "multiline pattern",
				value:   "line1\nline2",
				pattern: "(?s)line1.*line2",
				wantErr: false,
			},
			{
				name:    "multiline pattern without flag should fail",
				value:   "line1\nline2",
				pattern: "line1.*line2",
				wantErr: true,
			},
			{
				name:    "multiline pattern with \\s\\S alternative",
				value:   "line1\nline2",
				pattern: "line1[\\s\\S]*line2",
				wantErr: false,
			},
			{
				name:    "multiline with CRLF",
				value:   "line1\r\nline2",
				pattern: "(?s)line1.*line2",
				wantErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					Pattern(tc.pattern).
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

	t.Run("Enum Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   string
			enum    []any
			wantErr bool
			errMsg  string
		}{
			{
				name:    "valid enum value",
				value:   "red",
				enum:    []any{"red", "green", "blue"},
				wantErr: false,
			},
			{
				name:    "invalid enum value",
				value:   "yellow",
				enum:    []any{"red", "green", "blue"},
				wantErr: true,
				errMsg:  "not found in enum",
			},
			{
				name:    "enum with single value",
				value:   "only",
				enum:    []any{"only"},
				wantErr: false,
			},
			{
				name:    "enum with numbers as strings",
				value:   "123",
				enum:    []any{"123", "456", "789"},
				wantErr: false,
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
				enum:    []any{"", "non-empty"},
				wantErr: false,
			},
			{
				name:    "unicode in enum",
				value:   "café",
				enum:    []any{"café", "tea", "water"},
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

	t.Run("Const Validation", func(t *testing.T) {
		testCases := []struct {
			name     string
			value    string
			constVal any
			wantErr  bool
			errMsg   string
		}{
			{
				name:     "valid const value",
				value:    "constant",
				constVal: "constant",
				wantErr:  false,
			},
			{
				name:     "invalid const value",
				value:    "variable",
				constVal: "constant",
				wantErr:  true,
				errMsg:   "must be const value",
			},
			{
				name:     "empty string const",
				value:    "",
				constVal: "",
				wantErr:  false,
			},
			{
				name:     "unicode const",
				value:    "café",
				constVal: "café",
				wantErr:  false,
			},
			{
				name:     "case sensitive const",
				value:    "Constant",
				constVal: "constant",
				wantErr:  true,
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

	t.Run("Combined Constraints", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   string
			schema  func() *schema.Schema
			wantErr bool
			errMsg  string
		}{
			{
				name:  "pattern and length constraints valid",
				value: "hello123",
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Types(schema.StringType).
						Pattern("^[a-z]+[0-9]+$").
						MinLength(5).
						MaxLength(10).
						Build()
					return s
				},
				wantErr: false,
			},
			{
				name:  "pattern valid but length invalid",
				value: "hello123456789",
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Types(schema.StringType).
						Pattern("^[a-z]+[0-9]+$").
						MinLength(5).
						MaxLength(10).
						Build()
					return s
				},
				wantErr: true,
			},
			{
				name:  "length valid but pattern invalid",
				value: "HELLO123",
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Types(schema.StringType).
						Pattern("^[a-z]+[0-9]+$").
						MinLength(5).
						MaxLength(10).
						Build()
					return s
				},
				wantErr: true,
			},
			{
				name:  "enum and pattern constraints valid",
				value: "option1",
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Types(schema.StringType).
						Pattern("^option[0-9]$").
						Enum("option1", "option2", "option3").
						Build()
					return s
				},
				wantErr: false,
			},
			{
				name:  "in enum but pattern invalid",
				value: "option1",
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Types(schema.StringType).
						Pattern("^choice[0-9]$").
						Enum("option1", "option2", "option3").
						Build()
					return s
				},
				wantErr: true,
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

	t.Run("Format Validation", func(t *testing.T) {
		// Note: Format validation might not be implemented yet,
		// but adding tests as if it exists
		testCases := []struct {
			name    string
			value   string
			format  string
			wantErr bool
		}{
			{
				name:    "valid email format",
				value:   "test@example.com",
				format:  "email",
				wantErr: false,
			},
			{
				name:    "invalid email format",
				value:   "not-an-email",
				format:  "email",
				wantErr: true,
			},
			{
				name:    "valid date format",
				value:   "2023-12-25",
				format:  "date",
				wantErr: false,
			},
			{
				name:    "invalid date format",
				value:   "not-a-date",
				format:  "date",
				wantErr: true,
			},
			{
				name:    "valid datetime format",
				value:   "2023-12-25T10:30:00Z",
				format:  "date-time",
				wantErr: false,
			},
			{
				name:    "valid URI format",
				value:   "https://example.com/path",
				format:  "uri",
				wantErr: false,
			},
			{
				name:    "valid UUID format",
				value:   "550e8400-e29b-41d4-a716-446655440000",
				format:  "uuid",
				wantErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Types(schema.StringType).
					Format(tc.format).
					Build()

				// If Format() method doesn't exist yet, skip this test
				if err != nil && err.Error() == "Format method not implemented" {
					t.Skip("Format validation not implemented yet")
				}
				require.NoError(t, err)

				// For direct format validation tests, enable format-assertion
				ctx := context.Background()
				ctx = vocabulary.WithSet(ctx, vocabulary.AllEnabled())
				v, err := validator.Compile(ctx, s)
				require.NoError(t, err)

				_, err = v.Validate(context.Background(), tc.value)
				if tc.wantErr {
					require.Error(t, err, "Expected validation to fail for format %s with value %s", tc.format, tc.value)
				} else {
					require.NoError(t, err, "Expected validation to pass for format %s with value %s", tc.format, tc.value)
				}
			})
		}
	})
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}
