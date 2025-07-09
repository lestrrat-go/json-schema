package schema_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

// Test JSON Schema 2020-12 Core Specification Compliance
func TestJSONSchema2020_12_CoreCompliance(t *testing.T) {
	t.Run("Schema Version Declaration", func(t *testing.T) {
		// Test that schemas declare the correct version
		s := schema.New()
		require.Equal(t, schema.Version, s.Schema(), "Schema should declare 2020-12 version")
	})

	t.Run("Schema ID and Identification", func(t *testing.T) {
		testCases := []struct {
			name string
			id   string
		}{
			{"Absolute URI", "https://example.com/schema"},
			{"URI with fragment", "https://example.com/schema#def"},
			{"Relative URI", "/schema"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := schema.NewBuilder().
					ID(tc.id).
					Build()
				require.NoError(t, err)
				require.Equal(t, tc.id, s.ID())
			})
		}
	})

	t.Run("Core Keywords Support", func(t *testing.T) {
		// Test all core keywords are supported
		s, err := schema.NewBuilder().
			ID("https://example.com/test").
			Schema(schema.Version).
			Reference("#/definitions/test").
			Anchor("test-anchor").
			Comment("Test comment").
			Build()
		require.NoError(t, err)
		require.Equal(t, "https://example.com/test", s.ID())
		require.Equal(t, schema.Version, s.Schema())
		require.Equal(t, "#/definitions/test", s.Reference())
		require.Equal(t, "test-anchor", s.Anchor())
		require.Equal(t, "Test comment", s.Comment())
	})

	t.Run("Schema as Boolean", func(t *testing.T) {
		// Test that schemas can be boolean values (true/false)
		trueSchema := schema.New()
		falseSchema := schema.New()

		// Test schema acceptance of boolean values
		err := trueSchema.Accept(true)
		require.NoError(t, err)

		err = falseSchema.Accept(false)
		require.NoError(t, err)

		// false schema should have a 'not' constraint
		require.NotNil(t, falseSchema.Not())
	})
}

func TestPrimitiveTypes(t *testing.T) {
	testCases := []struct {
		name     string
		typeStr  string
		expected schema.PrimitiveType
		valid    bool
	}{
		{"Null type", "null", schema.NullType, true},
		{"Boolean type", "boolean", schema.BooleanType, true},
		{"Object type", "object", schema.ObjectType, true},
		{"Array type", "array", schema.ArrayType, true},
		{"Number type", "number", schema.NumberType, true},
		{"String type", "string", schema.StringType, true},
		{"Integer type", "integer", schema.IntegerType, true},
		{"Invalid type", "invalid", schema.PrimitiveType(0), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pt, err := schema.NewPrimitiveType(tc.typeStr)
			if tc.valid {
				require.NoError(t, err)
				require.Equal(t, tc.expected, pt)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestSchemaComposition(t *testing.T) {
	t.Run("AllOf Composition", func(t *testing.T) {
		stringSchema, err := schema.NewBuilder().Type(schema.StringType).Build()
		require.NoError(t, err)

		minLengthSchema, err := schema.NewBuilder().MinLength(5).Build()
		require.NoError(t, err)

		composedSchema, err := schema.NewBuilder().
			AllOf([]*schema.Schema{stringSchema, minLengthSchema}).
			Build()
		require.NoError(t, err)
		require.Len(t, composedSchema.AllOf(), 2)
	})

	t.Run("AnyOf Composition", func(t *testing.T) {
		stringSchema, err := schema.NewBuilder().Type(schema.StringType).Build()
		require.NoError(t, err)

		numberSchema, err := schema.NewBuilder().Type(schema.NumberType).Build()
		require.NoError(t, err)

		composedSchema, err := schema.NewBuilder().
			AnyOf([]*schema.Schema{stringSchema, numberSchema}).
			Build()
		require.NoError(t, err)
		require.Len(t, composedSchema.AnyOf(), 2)
	})

	t.Run("OneOf Composition", func(t *testing.T) {
		stringSchema, err := schema.NewBuilder().Type(schema.StringType).Build()
		require.NoError(t, err)

		numberSchema, err := schema.NewBuilder().Type(schema.NumberType).Build()
		require.NoError(t, err)

		composedSchema, err := schema.NewBuilder().
			OneOf([]*schema.Schema{stringSchema, numberSchema}).
			Build()
		require.NoError(t, err)
		require.Len(t, composedSchema.OneOf(), 2)
	})

	t.Run("Not Composition", func(t *testing.T) {
		stringSchema, err := schema.NewBuilder().Type(schema.StringType).Build()
		require.NoError(t, err)

		notSchema, err := schema.NewBuilder().
			Not(stringSchema).
			Build()
		require.NoError(t, err)
		require.NotNil(t, notSchema.Not())
	})
}

func TestSchemaConstraints(t *testing.T) {
	t.Run("String Constraints", func(t *testing.T) {
		s, err := schema.NewBuilder().
			Type(schema.StringType).
			MinLength(1).
			MaxLength(100).
			Pattern("^[a-zA-Z]+$").
			Build()
		require.NoError(t, err)
		require.Equal(t, 1, s.MinLength())
		require.Equal(t, 100, s.MaxLength())
		require.Equal(t, "^[a-zA-Z]+$", s.Pattern())
	})

	t.Run("Numeric Constraints", func(t *testing.T) {
		s, err := schema.NewBuilder().
			Type(schema.NumberType).
			Minimum(0.0).
			Maximum(100.0).
			ExclusiveMinimum(0.0).
			ExclusiveMaximum(100.0).
			MultipleOf(0.5).
			Build()
		require.NoError(t, err)
		require.Equal(t, 0.0, s.Minimum())
		require.Equal(t, 100.0, s.Maximum())
		require.Equal(t, 0.0, s.ExclusiveMinimum())
		require.Equal(t, 100.0, s.ExclusiveMaximum())
		require.Equal(t, 0.5, s.MultipleOf())
	})

	t.Run("Array Constraints", func(t *testing.T) {
		itemSchema, err := schema.NewBuilder().Type(schema.StringType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Type(schema.ArrayType).
			Items(itemSchema).
			MinItems(1).
			MaxItems(10).
			UniqueItems(true).
			Build()
		require.NoError(t, err)
		require.NotNil(t, s.Items())
		require.Equal(t, uint(1), s.MinItems())
		require.Equal(t, uint(10), s.MaxItems())
		require.True(t, s.UniqueItems())
	})

	t.Run("Object Constraints", func(t *testing.T) {
		propSchema, err := schema.NewBuilder().Type(schema.StringType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Type(schema.ObjectType).
			Property("name", propSchema).
			MinProperties(1).
			MaxProperties(10).
			Build()
		require.NoError(t, err)
		require.NotNil(t, s.Properties()["name"])
		require.Equal(t, uint(1), s.MinProperties())
		require.Equal(t, uint(10), s.MaxProperties())
	})
}

func TestEnumAndConst(t *testing.T) {
	t.Run("Enum Values", func(t *testing.T) {
		enumValues := []interface{}{"red", "green", "blue"}
		s, err := schema.NewBuilder().
			Type(schema.StringType).
			Enum(enumValues).
			Build()
		require.NoError(t, err)
		require.Equal(t, enumValues, s.Enum())
	})

	t.Run("Const Value", func(t *testing.T) {
		constValue := "constant"
		s, err := schema.NewBuilder().
			Type(schema.StringType).
			Const(constValue).
			Build()
		require.NoError(t, err)
		require.Equal(t, constValue, s.Const())
	})
}

func TestAdvancedFeatures(t *testing.T) {
	t.Run("Pattern Properties", func(t *testing.T) {
		propSchema, err := schema.NewBuilder().Type(schema.StringType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Type(schema.ObjectType).
			PatternProperty("^[a-z]+$", propSchema).
			Build()
		require.NoError(t, err)
		require.NotNil(t, s.PatternProperties()["^[a-z]+$"])
	})

	t.Run("Additional Properties", func(t *testing.T) {
		additionalPropSchema, err := schema.NewBuilder().Type(schema.StringType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Type(schema.ObjectType).
			AdditionalProperties(additionalPropSchema).
			Build()
		require.NoError(t, err)
		require.NotNil(t, s.AdditionalProperties())
	})

	t.Run("Contains", func(t *testing.T) {
		containsSchema, err := schema.NewBuilder().Type(schema.StringType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Type(schema.ArrayType).
			Contains(containsSchema).
			MinContains(1).
			MaxContains(5).
			Build()
		require.NoError(t, err)
		require.NotNil(t, s.Contains())
		require.Equal(t, uint(1), s.MinContains())
		require.Equal(t, uint(5), s.MaxContains())
	})

	t.Run("Unevaluated Properties and Items", func(t *testing.T) {
		unevalPropSchema, err := schema.NewBuilder().Type(schema.StringType).Build()
		require.NoError(t, err)

		unevalItemSchema, err := schema.NewBuilder().Type(schema.NumberType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Type(schema.ObjectType).
			UnevaluatedProperties(unevalPropSchema).
			UnevaluatedItems(unevalItemSchema).
			Build()
		require.NoError(t, err)
		require.NotNil(t, s.UnevaluatedProperties())
		require.NotNil(t, s.UnevaluatedItems())
	})
}

func TestSchemaBasicReferences(t *testing.T) {
	t.Run("Schema Reference", func(t *testing.T) {
		s, err := schema.NewBuilder().
			Reference("#/definitions/person").
			Build()
		require.NoError(t, err)
		require.Equal(t, "#/definitions/person", s.Reference())
	})

	t.Run("Dynamic Reference", func(t *testing.T) {
		s, err := schema.NewBuilder().
			DynamicReference("#person").
			Build()
		require.NoError(t, err)
		require.Equal(t, "#person", s.DynamicReference())
	})

	t.Run("Definitions", func(t *testing.T) {
		s, err := schema.NewBuilder().
			Definitions("person").
			Build()
		require.NoError(t, err)
		require.Equal(t, "person", s.Definitions())
	})
}
