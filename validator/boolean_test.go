package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// TestBooleanValidatorComprehensive tests all boolean validation features
func TestBooleanValidatorComprehensive(t *testing.T) {
	t.Run("Basic Boolean Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			wantErr bool
			errMsg  string
		}{
			{
				name:    "valid true",
				value:   true,
				wantErr: false,
			},
			{
				name:    "valid false",
				value:   false,
				wantErr: false,
			},
			{
				name:    "string should fail",
				value:   "true",
				wantErr: true,
				errMsg:  "expected boolean",
			},
			{
				name:    "integer should fail",
				value:   1,
				wantErr: true,
				errMsg:  "expected boolean",
			},
			{
				name:    "integer zero should fail",
				value:   0,
				wantErr: true,
				errMsg:  "expected boolean",
			},
			{
				name:    "float should fail",
				value:   1.0,
				wantErr: true,
				errMsg:  "expected boolean",
			},
			{
				name:    "object should fail",
				value:   map[string]any{"key": "value"},
				wantErr: true,
				errMsg:  "expected boolean",
			},
			{
				name:    "array should fail",
				value:   []any{true, false},
				wantErr: true,
				errMsg:  "expected boolean",
			},
			{
				name:    "nil should fail",
				value:   nil,
				wantErr: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().Types(schema.BooleanType).Build()
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
				name:     "const true with true value",
				value:    true,
				constVal: true,
				wantErr:  false,
			},
			{
				name:     "const false with false value",
				value:    false,
				constVal: false,
				wantErr:  false,
			},
			{
				name:     "const true with false value",
				value:    false,
				constVal: true,
				wantErr:  true,
				errMsg:   "const",
			},
			{
				name:     "const false with true value",
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
			{
				name:    "false in enum with only false",
				value:   false,
				enum:    []any{false},
				wantErr: false,
			},
			{
				name:    "true not in enum with only false",
				value:   true,
				enum:    []any{false},
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

	t.Run("Boolean in Complex Schemas", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			schema  func() *schema.Schema
			wantErr bool
			errMsg  string
		}{
			{
				name: "boolean in object property",
				value: map[string]any{
					"enabled": true,
					"name":    "feature",
				},
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Types(schema.ObjectType).
						Property("enabled", schema.NewBuilder().Types(schema.BooleanType).MustBuild()).
						Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
						Build()
					return s
				},
				wantErr: false,
			},
			{
				name: "invalid boolean in object property",
				value: map[string]any{
					"enabled": "true", // should be boolean
					"name":    "feature",
				},
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Types(schema.ObjectType).
						Property("enabled", schema.NewBuilder().Types(schema.BooleanType).MustBuild()).
						Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
						Build()
					return s
				},
				wantErr: true,
				errMsg:  "property validation failed",
			},
			{
				name:  "boolean array",
				value: []any{true, false, true, false},
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Types(schema.ArrayType).
						Items(schema.NewBuilder().Types(schema.BooleanType).MustBuild()).
						Build()
					return s
				},
				wantErr: false,
			},
			{
				name:  "mixed array with boolean validation failure",
				value: []any{true, false, "true"}, // third item should be boolean
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Types(schema.ArrayType).
						Items(schema.NewBuilder().Types(schema.BooleanType).MustBuild()).
						Build()
					return s
				},
				wantErr: true,
				errMsg:  "item validation failed",
			},
			{
				name: "boolean with const in nested structure",
				value: map[string]any{
					"config": map[string]any{
						"debug": true,
					},
				},
				schema: func() *schema.Schema {
					configSchema, _ := schema.NewBuilder().
						Types(schema.ObjectType).
						Property("debug", schema.NewBuilder().
							Types(schema.BooleanType).
							Const(true).
							MustBuild()).
						Build()

					s, _ := schema.NewBuilder().
						Types(schema.ObjectType).
						Property("config", configSchema).
						Build()
					return s
				},
				wantErr: false,
			},
			{
				name: "boolean const violation in nested structure",
				value: map[string]any{
					"config": map[string]any{
						"debug": false, // should be true
					},
				},
				schema: func() *schema.Schema {
					configSchema, _ := schema.NewBuilder().
						Types(schema.ObjectType).
						Property("debug", schema.NewBuilder().
							Types(schema.BooleanType).
							Const(true).
							MustBuild()).
						Build()

					s, _ := schema.NewBuilder().
						Types(schema.ObjectType).
						Property("config", configSchema).
						Build()
					return s
				},
				wantErr: true,
				errMsg:  "const",
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
