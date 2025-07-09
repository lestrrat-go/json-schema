package validator_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// TestSchemaCompositionComprehensive tests all schema composition features (allOf, anyOf, oneOf, not)
func TestSchemaCompositionComprehensive(t *testing.T) {
	t.Run("AllOf Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			schemas []*schema.Schema
			wantErr bool
			errMsg  string
		}{
			{
				name:  "all schemas pass",
				value: "hello",
				schemas: []*schema.Schema{
					schema.NewBuilder().Type(schema.StringType).MustBuild(),
					schema.NewBuilder().MinLength(3).MustBuild(),
					schema.NewBuilder().MaxLength(10).MustBuild(),
				},
				wantErr: false,
			},
			{
				name:  "one schema fails",
				value: "hi",
				schemas: []*schema.Schema{
					schema.NewBuilder().Type(schema.StringType).MustBuild(),
					schema.NewBuilder().MinLength(3).MustBuild(), // fails here
					schema.NewBuilder().MaxLength(10).MustBuild(),
				},
				wantErr: true,
				errMsg:  "allOf validation failed",
			},
			{
				name:  "type mismatch fails first schema",
				value: 123,
				schemas: []*schema.Schema{
					schema.NewBuilder().Type(schema.StringType).MustBuild(), // fails here
					schema.NewBuilder().MinLength(3).MustBuild(),
				},
				wantErr: true,
			},
			{
				name:  "complex allOf with object constraints",
				value: map[string]any{"name": "John", "age": 30, "email": "john@example.com"},
				schemas: []*schema.Schema{
					schema.NewBuilder().
						Type(schema.ObjectType).
						Property("name", schema.NewBuilder().Type(schema.StringType).MustBuild()).MustBuild(),
					schema.NewBuilder().
						Type(schema.ObjectType).
						Property("age", schema.NewBuilder().Type(schema.IntegerType).MustBuild()).MustBuild(),
					schema.NewBuilder().
						Type(schema.ObjectType).
						Property("email", schema.NewBuilder().Type(schema.StringType).MustBuild()).MustBuild(),
				},
				wantErr: false,
			},
			{
				name:  "allOf with overlapping but compatible constraints",
				value: 50,
				schemas: []*schema.Schema{
					schema.NewBuilder().Type(schema.IntegerType).Minimum(10).MustBuild(),
					schema.NewBuilder().Type(schema.IntegerType).Maximum(100).MustBuild(),
					schema.NewBuilder().Type(schema.IntegerType).MultipleOf(5).MustBuild(),
				},
				wantErr: false,
			},
			{
				name:  "allOf with conflicting constraints",
				value: 7,
				schemas: []*schema.Schema{
					schema.NewBuilder().Type(schema.IntegerType).Minimum(10).MustBuild(),
					schema.NewBuilder().Type(schema.IntegerType).Maximum(5).MustBuild(), // impossible
				},
				wantErr: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					AllOf(tc.schemas).
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

	t.Run("AnyOf Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			schemas []*schema.Schema
			wantErr bool
			errMsg  string
		}{
			{
				name:  "first schema passes",
				value: "hello",
				schemas: []*schema.Schema{
					schema.NewBuilder().Type(schema.StringType).MustBuild(),
					schema.NewBuilder().Type(schema.IntegerType).MustBuild(),
				},
				wantErr: false,
			},
			{
				name:  "second schema passes",
				value: 123,
				schemas: []*schema.Schema{
					schema.NewBuilder().Type(schema.StringType).MustBuild(),
					schema.NewBuilder().Type(schema.IntegerType).MustBuild(),
				},
				wantErr: false,
			},
			{
				name:  "multiple schemas pass",
				value: 42,
				schemas: []*schema.Schema{
					schema.NewBuilder().Type(schema.IntegerType).MustBuild(),
					schema.NewBuilder().Type(schema.NumberType).MustBuild(),
					schema.NewBuilder().Minimum(0).MustBuild(),
				},
				wantErr: false,
			},
			{
				name:  "no schemas pass",
				value: true,
				schemas: []*schema.Schema{
					schema.NewBuilder().Type(schema.StringType).MustBuild(),
					schema.NewBuilder().Type(schema.IntegerType).MustBuild(),
					schema.NewBuilder().Type(schema.ArrayType).MustBuild(),
				},
				wantErr: true,
				errMsg:  "anyOf validation failed",
			},
			{
				name:  "anyOf with different constraint types",
				value: []any{1, 2, 3},
				schemas: []*schema.Schema{
					schema.NewBuilder().Type(schema.StringType).MustBuild(),
					schema.NewBuilder().
						Type(schema.ArrayType).
						Items(schema.NewBuilder().Type(schema.IntegerType).MustBuild()).MustBuild(),
				},
				wantErr: false,
			},
			{
				name:  "anyOf with object alternatives",
				value: map[string]any{"name": "John", "age": 30},
				schemas: []*schema.Schema{
					// Person schema
					schema.NewBuilder().
						Type(schema.ObjectType).
						Property("name", schema.NewBuilder().Type(schema.StringType).MustBuild()).
						Property("age", schema.NewBuilder().Type(schema.IntegerType).MustBuild()).MustBuild(),
					// Product schema
					schema.NewBuilder().
						Type(schema.ObjectType).
						Property("title", schema.NewBuilder().Type(schema.StringType).MustBuild()).
						Property("price", schema.NewBuilder().Type(schema.NumberType).MustBuild()).MustBuild(),
				},
				wantErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					AnyOf(tc.schemas).
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

	t.Run("OneOf Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			schemas []*schema.Schema
			wantErr bool
			errMsg  string
		}{
			{
				name:  "exactly one schema passes",
				value: "hello",
				schemas: []*schema.Schema{
					schema.NewBuilder().Type(schema.StringType).MustBuild(),
					schema.NewBuilder().Type(schema.IntegerType).MustBuild(),
				},
				wantErr: false,
			},
			{
				name:  "no schemas pass",
				value: true,
				schemas: []*schema.Schema{
					schema.NewBuilder().Type(schema.StringType).MustBuild(),
					schema.NewBuilder().Type(schema.IntegerType).MustBuild(),
				},
				wantErr: true,
				errMsg:  "oneOf validation failed",
			},
			{
				name:  "multiple schemas pass - should fail",
				value: 42,
				schemas: []*schema.Schema{
					schema.NewBuilder().Type(schema.IntegerType).MustBuild(),
					schema.NewBuilder().Type(schema.NumberType).MustBuild(), // integers are also numbers
					schema.NewBuilder().Minimum(0).MustBuild(),
				},
				wantErr: true,
				errMsg:  "oneOf validation failed", // multiple matches
			},
			{
				name:  "oneOf with mutually exclusive constraints",
				value: 15,
				schemas: []*schema.Schema{
					schema.NewBuilder().Type(schema.IntegerType).Maximum(10).MustBuild(),
					schema.NewBuilder().Type(schema.IntegerType).Minimum(12).MustBuild(),
				},
				wantErr: false, // only second schema passes
			},
			{
				name:  "oneOf with complex object types",
				value: map[string]any{"type": "circle", "radius": 5},
				schemas: []*schema.Schema{
					// Circle schema
					schema.NewBuilder().
						Type(schema.ObjectType).
						Property("type", schema.NewBuilder().Type(schema.StringType).Const("circle").MustBuild()).
						Property("radius", schema.NewBuilder().Type(schema.NumberType).MustBuild()).
						MustBuild(),
					// Rectangle schema
					schema.NewBuilder().
						Type(schema.ObjectType).
						Property("type", schema.NewBuilder().Type(schema.StringType).Const("rectangle").MustBuild()).
						Property("width", schema.NewBuilder().Type(schema.NumberType).MustBuild()).
						Property("height", schema.NewBuilder().Type(schema.NumberType).MustBuild()).MustBuild(),
				},
				wantErr: false,
			},
			{
				name:  "oneOf with ambiguous object",
				value: map[string]any{"name": "John"},
				schemas: []*schema.Schema{
					// Person schema
					schema.NewBuilder().
						Type(schema.ObjectType).
						Property("name", schema.NewBuilder().Type(schema.StringType).MustBuild()).MustBuild(),
					// Product schema with optional name
					schema.NewBuilder().
						Type(schema.ObjectType).
						Property("name", schema.NewBuilder().Type(schema.StringType).MustBuild()).MustBuild(),
				},
				wantErr: true, // both schemas match
				errMsg:  "oneOf validation failed",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					OneOf(tc.schemas).
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

	t.Run("Not Validation", func(t *testing.T) {
		testCases := []struct {
			name      string
			value     any
			notSchema *schema.Schema
			wantErr   bool
			errMsg    string
		}{
			{
				name:      "value doesn't match not schema - valid",
				value:     123,
				notSchema: schema.NewBuilder().Type(schema.StringType).MustBuild(),
				wantErr:   false,
			},
			{
				name:      "value matches not schema - invalid",
				value:     "hello",
				notSchema: schema.NewBuilder().Type(schema.StringType).MustBuild(),
				wantErr:   true,
				errMsg:    "not validation failed",
			},
			{
				name:      "not with minimum constraint",
				value:     5,
				notSchema: schema.NewBuilder().Type(schema.IntegerType).Minimum(10).MustBuild(),
				wantErr:   false, // 5 doesn't match "integer >= 10"
			},
			{
				name:      "not with minimum constraint - fails",
				value:     15,
				notSchema: schema.NewBuilder().Type(schema.IntegerType).Minimum(10).MustBuild(),
				wantErr:   true, // 15 matches "integer >= 10"
			},
			{
				name:  "not with complex object schema",
				value: map[string]any{"name": "John", "type": "user"},
				notSchema: schema.NewBuilder().
					Type(schema.ObjectType).
					Property("type", schema.NewBuilder().Type(schema.StringType).Const("admin").MustBuild()).MustBuild(),
				wantErr: false, // doesn't match admin schema
			},
			{
				name:  "not with complex object schema - fails",
				value: map[string]any{"name": "John", "type": "admin"},
				notSchema: schema.NewBuilder().
					Type(schema.ObjectType).
					Property("type", schema.NewBuilder().Type(schema.StringType).Const("admin").MustBuild()).MustBuild(),
				wantErr: true, // matches admin schema
			},
			{
				name:      "not with array schema",
				value:     "not an array",
				notSchema: schema.NewBuilder().Type(schema.ArrayType).MustBuild(),
				wantErr:   false,
			},
			{
				name:      "not with array schema - fails",
				value:     []any{1, 2, 3},
				notSchema: schema.NewBuilder().Type(schema.ArrayType).MustBuild(),
				wantErr:   true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Not(tc.notSchema).
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

	t.Run("Nested Composition", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			schema  func() *schema.Schema
			wantErr bool
			errMsg  string
		}{
			{
				name:  "allOf containing anyOf",
				value: "hello",
				schema: func() *schema.Schema {
					// Must be string AND (minLength 3 OR maxLength 10)
					anyOfSchema := schema.NewBuilder().
						AnyOf([]*schema.Schema{
							schema.NewBuilder().MinLength(3).MustBuild(),
							schema.NewBuilder().MaxLength(10).MustBuild(),
						}).MustBuild()

					return schema.NewBuilder().
						AllOf([]*schema.Schema{
							schema.NewBuilder().Type(schema.StringType).MustBuild(),
							anyOfSchema,
						}).MustBuild()
				},
				wantErr: false,
			},
			{
				name:  "oneOf with allOf alternatives",
				value: 42,
				schema: func() *schema.Schema {
					// Either (string AND minLength 5) OR (integer AND minimum 0)
					stringConstraints := schema.NewBuilder().
						AllOf([]*schema.Schema{
							schema.NewBuilder().Type(schema.StringType).MustBuild(),
							schema.NewBuilder().MinLength(5).MustBuild(),
						}).MustBuild()

					integerConstraints := schema.NewBuilder().
						AllOf([]*schema.Schema{
							schema.NewBuilder().Type(schema.IntegerType).MustBuild(),
							schema.NewBuilder().Minimum(0).MustBuild(),
						}).MustBuild()

					return schema.NewBuilder().
						OneOf([]*schema.Schema{stringConstraints, integerConstraints}).
						MustBuild()
				},
				wantErr: false,
			},
			{
				name:  "not containing oneOf",
				value: "medium",
				schema: func() *schema.Schema {
					// Not (short string OR long string) = medium length strings
					oneOfSchema := schema.NewBuilder().
						OneOf([]*schema.Schema{
							schema.NewBuilder().Type(schema.StringType).MaxLength(3).MustBuild(),  // short
							schema.NewBuilder().Type(schema.StringType).MinLength(10).MustBuild(), // long
						}).MustBuild()

					return schema.NewBuilder().
						AllOf([]*schema.Schema{
							schema.NewBuilder().Type(schema.StringType).MustBuild(),
							schema.NewBuilder().Not(oneOfSchema).MustBuild(),
						}).MustBuild()
				},
				wantErr: false,
			},
			{
				name:  "complex nested validation failure",
				value: "hi", // too short
				schema: func() *schema.Schema {
					oneOfSchema := schema.NewBuilder().
						OneOf([]*schema.Schema{
							schema.NewBuilder().Type(schema.StringType).MaxLength(3).MustBuild(),  // short
							schema.NewBuilder().Type(schema.StringType).MinLength(10).MustBuild(), // long
						}).MustBuild()

					return schema.NewBuilder().
						AllOf([]*schema.Schema{
							schema.NewBuilder().Type(schema.StringType).MustBuild(),
							schema.NewBuilder().Not(oneOfSchema).MustBuild(), // should NOT be short or long
						}).MustBuild()
				},
				wantErr: true, // "hi" is short, so it matches oneOf, so not(oneOf) fails
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

	t.Run("Composition with Type Constraints", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			schema  func() *schema.Schema
			wantErr bool
			errMsg  string
		}{
			{
				name:  "allOf with type and value constraints",
				value: "hello world",
				schema: func() *schema.Schema {
					return schema.NewBuilder().
						Type(schema.StringType).
						AllOf([]*schema.Schema{
							schema.NewBuilder().MinLength(5).MustBuild(),
							schema.NewBuilder().Pattern("^hello").MustBuild(),
						}).MustBuild()
				},
				wantErr: false,
			},
			{
				name:  "anyOf allowing multiple types",
				value: 42,
				schema: func() *schema.Schema {
					return schema.NewBuilder().
						AnyOf([]*schema.Schema{
							schema.NewBuilder().Type(schema.StringType).MinLength(5).MustBuild(),
							schema.NewBuilder().Type(schema.IntegerType).Minimum(0).MustBuild(),
							schema.NewBuilder().Type(schema.BooleanType).MustBuild(),
						}).MustBuild()
				},
				wantErr: false,
			},
			{
				name:  "oneOf with type discrimination",
				value: map[string]any{"type": "user", "name": "John"},
				schema: func() *schema.Schema {
					userSchema := schema.NewBuilder().
						Type(schema.ObjectType).
						Property("type", schema.NewBuilder().Type(schema.StringType).Const("user").MustBuild()).
						Property("name", schema.NewBuilder().Type(schema.StringType).MustBuild()).
						MustBuild()

					adminSchema := schema.NewBuilder().
						Type(schema.ObjectType).
						Property("type", schema.NewBuilder().Type(schema.StringType).Const("admin").MustBuild()).
						Property("permissions", schema.NewBuilder().Type(schema.ArrayType).MustBuild()).
						MustBuild()

					return schema.NewBuilder().
						OneOf([]*schema.Schema{userSchema, adminSchema}).
						MustBuild()
				},
				wantErr: false,
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
