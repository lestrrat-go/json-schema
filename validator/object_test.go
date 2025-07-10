package validator_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestObjectConstrainctSanity(t *testing.T) {
	testcases := makeSanityTestCases()
	for _, tc := range testcases {
		switch tc.Name {
		case "Empty Map", "Empty Object":
		default:
			tc.Error = true
		}
	}

	c := validator.Object().MustBuild()
	for _, tc := range testcases {
		t.Run(tc.Name, makeSanityTestFunc(tc, c))
	}
}

// TestObjectValidatorComprehensive tests all object validation features
func TestObjectValidatorComprehensive(t *testing.T) {
	t.Run("Basic Object Validation", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   any
			wantErr bool
			errMsg  string
		}{
			{
				name:    "valid empty map",
				value:   map[string]any{},
				wantErr: false,
			},
			{
				name: "valid map with properties",
				value: map[string]any{
					"name": "John",
					"age":  30,
				},
				wantErr: false,
			},
			{
				name:    "valid empty struct",
				value:   struct{}{},
				wantErr: false,
			},
			{
				name: "valid struct with fields",
				value: struct {
					Name string
					Age  int
				}{Name: "John", Age: 30},
				wantErr: false,
			},
			{
				name:    "pointer to struct",
				value:   &struct{ Name string }{Name: "John"},
				wantErr: false,
			},
			{
				name:    "string should fail",
				value:   "not an object",
				wantErr: true,
				errMsg:  "expected map or a struct",
			},
			{
				name:    "integer should fail",
				value:   123,
				wantErr: true,
				errMsg:  "expected map or a struct",
			},
			{
				name:    "array should fail",
				value:   []string{"a", "b"},
				wantErr: true,
				errMsg:  "expected map or a struct",
			},
			{
				name:    "nil should fail",
				value:   nil,
				wantErr: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().Type(schema.ObjectType).Build()
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

	t.Run("Properties Validation", func(t *testing.T) {
		testCases := []struct {
			name       string
			value      map[string]any
			properties map[string]*schema.Schema
			wantErr    bool
			errMsg     string
		}{
			{
				name: "all properties valid",
				value: map[string]any{
					"name": "John",
					"age":  30,
				},
				properties: map[string]*schema.Schema{
					"name": schema.NewBuilder().Type(schema.StringType).MustBuild(),
					"age":  schema.NewBuilder().Type(schema.IntegerType).MustBuild(),
				},
				wantErr: false,
			},
			{
				name: "missing property allowed",
				value: map[string]any{
					"name": "John",
				},
				properties: map[string]*schema.Schema{
					"name": schema.NewBuilder().Type(schema.StringType).MustBuild(),
					"age":  schema.NewBuilder().Type(schema.IntegerType).MustBuild(),
				},
				wantErr: false,
			},
			{
				name: "extra property allowed by default",
				value: map[string]any{
					"name":  "John",
					"extra": "value",
				},
				properties: map[string]*schema.Schema{
					"name": schema.NewBuilder().Type(schema.StringType).MustBuild(),
				},
				wantErr: false,
			},
			{
				name: "invalid property type",
				value: map[string]any{
					"name": 123, // should be string
					"age":  30,
				},
				properties: map[string]*schema.Schema{
					"name": schema.NewBuilder().Type(schema.StringType).MustBuild(),
					"age":  schema.NewBuilder().Type(schema.IntegerType).MustBuild(),
				},
				wantErr: true,
				errMsg:  "property validation failed",
			},
			{
				name: "nested object properties",
				value: map[string]any{
					"person": map[string]any{
						"name": "John",
						"age":  30,
					},
				},
				properties: map[string]*schema.Schema{
					"person": schema.NewBuilder().
						Type(schema.ObjectType).
						Property("name", schema.NewBuilder().Type(schema.StringType).MustBuild()).
						Property("age", schema.NewBuilder().Type(schema.IntegerType).MustBuild()).
						MustBuild(),
				},
				wantErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := schema.NewBuilder().Type(schema.ObjectType)
				for propName, propSchema := range tc.properties {
					builder = builder.Property(propName, propSchema)
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

	t.Run("Required Properties", func(t *testing.T) {
		testCases := []struct {
			name     string
			value    map[string]any
			required []string
			wantErr  bool
			errMsg   string
		}{
			{
				name: "all required properties present",
				value: map[string]any{
					"name": "John",
					"age":  30,
				},
				required: []string{"name", "age"},
				wantErr:  false,
			},
			{
				name: "missing required property",
				value: map[string]any{
					"name": "John",
				},
				required: []string{"name", "age"},
				wantErr:  true,
				errMsg:   "required property",
			},
			{
				name: "required property with null value",
				value: map[string]any{
					"name": "John",
					"age":  nil,
				},
				required: []string{"name", "age"},
				wantErr:  true,
				errMsg:   "required property",
			},
			{
				name: "extra properties with required",
				value: map[string]any{
					"name":  "John",
					"age":   30,
					"extra": "value",
				},
				required: []string{"name", "age"},
				wantErr:  false,
			},
			{
				name:     "empty object with no required",
				value:    map[string]any{},
				required: []string{},
				wantErr:  false,
			},
			{
				name:     "empty object with required properties",
				value:    map[string]any{},
				required: []string{"name"},
				wantErr:  true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := schema.NewBuilder().Type(schema.ObjectType).Required(tc.required...)
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

	t.Run("Property Count Constraints", func(t *testing.T) {
		testCases := []struct {
			name          string
			value         map[string]any
			minProperties *uint
			maxProperties *uint
			wantErr       bool
			errMsg        string
		}{
			{
				name: "within property count range",
				value: map[string]any{
					"a": 1,
					"b": 2,
					"c": 3,
				},
				minProperties: uintPtr(2),
				maxProperties: uintPtr(5),
				wantErr:       false,
			},
			{
				name: "exact minimum properties",
				value: map[string]any{
					"a": 1,
					"b": 2,
				},
				minProperties: uintPtr(2),
				wantErr:       false,
			},
			{
				name: "exact maximum properties",
				value: map[string]any{
					"a": 1,
					"b": 2,
					"c": 3,
				},
				maxProperties: uintPtr(3),
				wantErr:       false,
			},
			{
				name: "below minimum properties",
				value: map[string]any{
					"a": 1,
				},
				minProperties: uintPtr(2),
				wantErr:       true,
				errMsg:        "minimum properties",
			},
			{
				name: "above maximum properties",
				value: map[string]any{
					"a": 1,
					"b": 2,
					"c": 3,
					"d": 4,
				},
				maxProperties: uintPtr(3),
				wantErr:       true,
				errMsg:        "maximum properties",
			},
			{
				name:          "empty object with min properties",
				value:         map[string]any{},
				minProperties: uintPtr(1),
				wantErr:       true,
			},
			{
				name:          "empty object with max properties 0",
				value:         map[string]any{},
				maxProperties: uintPtr(0),
				wantErr:       false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := schema.NewBuilder().Type(schema.ObjectType)
				if tc.minProperties != nil {
					builder = builder.MinProperties(*tc.minProperties)
				}
				if tc.maxProperties != nil {
					builder = builder.MaxProperties(*tc.maxProperties)
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

	t.Run("Additional Properties", func(t *testing.T) {
		testCases := []struct {
			name                 string
			value                map[string]any
			properties           map[string]*schema.Schema
			additionalProperties any // can be bool or *schema.Schema
			wantErr              bool
			errMsg               string
		}{
			{
				name: "additional properties allowed (default)",
				value: map[string]any{
					"name":  "John",
					"extra": "value",
				},
				properties: map[string]*schema.Schema{
					"name": schema.NewBuilder().Type(schema.StringType).MustBuild(),
				},
				additionalProperties: true,
				wantErr:              false,
			},
			{
				name: "additional properties forbidden",
				value: map[string]any{
					"name":  "John",
					"extra": "value",
				},
				properties: map[string]*schema.Schema{
					"name": schema.NewBuilder().Type(schema.StringType).MustBuild(),
				},
				additionalProperties: false,
				wantErr:              true,
				errMsg:               "additional property not allowed",
			},
			{
				name: "no additional properties when forbidden",
				value: map[string]any{
					"name": "John",
				},
				properties: map[string]*schema.Schema{
					"name": schema.NewBuilder().Type(schema.StringType).MustBuild(),
				},
				additionalProperties: false,
				wantErr:              false,
			},
			{
				name: "additional properties with schema validation",
				value: map[string]any{
					"name":  "John",
					"extra": "string_value", // must be string
				},
				properties: map[string]*schema.Schema{
					"name": schema.NewBuilder().Type(schema.StringType).MustBuild(),
				},
				additionalProperties: schema.NewBuilder().Type(schema.StringType).MustBuild(),
				wantErr:              false,
			},
			{
				name: "additional properties violate schema",
				value: map[string]any{
					"name":  "John",
					"extra": 123, // should be string
				},
				properties: map[string]*schema.Schema{
					"name": schema.NewBuilder().Type(schema.StringType).MustBuild(),
				},
				additionalProperties: schema.NewBuilder().Type(schema.StringType).MustBuild(),
				wantErr:              true,
				errMsg:               "additional property validation failed",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := schema.NewBuilder().Type(schema.ObjectType)
				for propName, propSchema := range tc.properties {
					builder = builder.Property(propName, propSchema)
				}

				builder = builder.AdditionalProperties(tc.additionalProperties)

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

	t.Run("Pattern Properties", func(t *testing.T) {
		testCases := []struct {
			name              string
			value             map[string]any
			patternProperties map[string]*schema.Schema
			wantErr           bool
			errMsg            string
		}{
			{
				name: "pattern properties match",
				value: map[string]any{
					"str_name": "John",
					"str_city": "NYC",
					"num_age":  30,
					"num_id":   12345,
				},
				patternProperties: map[string]*schema.Schema{
					"^str_": schema.NewBuilder().Type(schema.StringType).MustBuild(),
					"^num_": schema.NewBuilder().Type(schema.IntegerType).MustBuild(),
				},
				wantErr: false,
			},
			{
				name: "pattern property type mismatch",
				value: map[string]any{
					"str_name": 123, // should be string
					"num_age":  30,
				},
				patternProperties: map[string]*schema.Schema{
					"^str_": schema.NewBuilder().Type(schema.StringType).MustBuild(),
					"^num_": schema.NewBuilder().Type(schema.IntegerType).MustBuild(),
				},
				wantErr: true,
				errMsg:  "pattern property validation failed",
			},
			{
				name: "property matches multiple patterns",
				value: map[string]any{
					"prefix_suffix": "value",
				},
				patternProperties: map[string]*schema.Schema{
					"^prefix_": schema.NewBuilder().Type(schema.StringType).MustBuild(),
					"_suffix$": schema.NewBuilder().Type(schema.StringType).MustBuild(),
				},
				wantErr: false,
			},
			{
				name: "no pattern match allows any type",
				value: map[string]any{
					"random_prop": "any_value",
					"other":       123,
				},
				patternProperties: map[string]*schema.Schema{
					"^specific_": schema.NewBuilder().Type(schema.StringType).MustBuild(),
				},
				wantErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := schema.NewBuilder().Type(schema.ObjectType)
				for pattern, patternSchema := range tc.patternProperties {
					builder = builder.PatternProperty(pattern, patternSchema)
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

	t.Run("Property Names Validation", func(t *testing.T) {
		testCases := []struct {
			name          string
			value         map[string]any
			propertyNames *schema.Schema
			wantErr       bool
			errMsg        string
		}{
			{
				name: "property names match pattern",
				value: map[string]any{
					"valid_name_1": "value1",
					"valid_name_2": "value2",
				},
				propertyNames: schema.NewBuilder().
					Type(schema.StringType).
					Pattern("^valid_name_[0-9]+$").
					MustBuild(),
				wantErr: false,
			},
			{
				name: "property name doesn't match pattern",
				value: map[string]any{
					"valid_name_1": "value1",
					"invalid_name": "value2",
				},
				propertyNames: schema.NewBuilder().
					Type(schema.StringType).
					Pattern("^valid_name_[0-9]+$").
					MustBuild(),
				wantErr: true,
				errMsg:  "property name validation failed",
			},
			{
				name: "property names with length constraint",
				value: map[string]any{
					"short": "value1",
					"a":     "value2",
				},
				propertyNames: schema.NewBuilder().
					Type(schema.StringType).
					MinLength(2).
					MaxLength(10).
					MustBuild(),
				wantErr: true,
				errMsg:  "property name validation failed",
			},
			{
				name: "all property names within length constraint",
				value: map[string]any{
					"name1": "value1",
					"name2": "value2",
				},
				propertyNames: schema.NewBuilder().
					Type(schema.StringType).
					MinLength(2).
					MaxLength(10).
					MustBuild(),
				wantErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					Type(schema.ObjectType).
					PropertyNames(tc.propertyNames).
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

	t.Run("Complex Object Scenarios", func(t *testing.T) {
		testCases := []struct {
			name    string
			value   map[string]any
			schema  func() *schema.Schema
			wantErr bool
			errMsg  string
		}{
			{
				name: "deeply nested object structure",
				value: map[string]any{
					"user": map[string]any{
						"profile": map[string]any{
							"name": "John",
							"age":  30,
						},
						"settings": map[string]any{
							"theme":         "dark",
							"notifications": true,
						},
					},
				},
				schema: func() *schema.Schema {
					profileSchema, _ := schema.NewBuilder().
						Type(schema.ObjectType).
						Property("name", schema.NewBuilder().Type(schema.StringType).MustBuild()).
						Property("age", schema.NewBuilder().Type(schema.IntegerType).MustBuild()).
						Build()

					settingsSchema, _ := schema.NewBuilder().
						Type(schema.ObjectType).
						Property("theme", schema.NewBuilder().Type(schema.StringType).MustBuild()).
						Property("notifications", schema.NewBuilder().Type(schema.BooleanType).MustBuild()).
						Build()

					userSchema, _ := schema.NewBuilder().
						Type(schema.ObjectType).
						Property("profile", profileSchema).
						Property("settings", settingsSchema).
						Build()

					s, _ := schema.NewBuilder().
						Type(schema.ObjectType).
						Property("user", userSchema).
						Build()
					return s
				},
				wantErr: false,
			},
			{
				name: "object with all constraint types",
				value: map[string]any{
					"required_prop": "value",
					"str_optional":  "string_value",
					"num_count":     42,
				},
				schema: func() *schema.Schema {
					s, _ := schema.NewBuilder().
						Type(schema.ObjectType).
						Property("required_prop", schema.NewBuilder().Type(schema.StringType).MustBuild()).
						Required("required_prop").
						PatternProperty("^str_", schema.NewBuilder().Type(schema.StringType).MustBuild()).
						PatternProperty("^num_", schema.NewBuilder().Type(schema.IntegerType).MustBuild()).
						MinProperties(1).
						MaxProperties(10).
						AdditionalProperties(false).
						Build()
					return s
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

// Helper functions
func uintPtr(u uint) *uint {
	return &u
}
