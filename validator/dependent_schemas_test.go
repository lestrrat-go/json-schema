package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestDependentSchemas(t *testing.T) {
	t.Run("single dependency", func(t *testing.T) {
		// Schema from the JSON Schema test suite
		barDependentSchema := schema.NewBuilder().
			Property("foo", schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
			Property("bar", schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
			MustBuild()

		s := schema.NewBuilder().
			DependentSchemas(map[string]schema.SchemaOrBool{
				"bar": barDependentSchema,
			}).
			MustBuild()

		v, err := validator.Compile(context.Background(), s)
		require.NoError(t, err)

		// Valid case - both properties satisfy the dependent schema
		validData := map[string]any{"foo": 1, "bar": 2}
		_, err = v.Validate(context.Background(), validData)
		require.NoError(t, err)

		// Valid case - no dependency (bar property not present)
		noDependencyData := map[string]any{"foo": "quux"}
		_, err = v.Validate(context.Background(), noDependencyData)
		require.NoError(t, err)

		// Invalid case - wrong type for foo when bar is present
		wrongTypeFoo := map[string]any{"foo": "quux", "bar": 2}
		_, err = v.Validate(context.Background(), wrongTypeFoo)
		require.Error(t, err)

		// Invalid case - wrong type for bar when bar is present
		wrongTypeBar := map[string]any{"foo": 2, "bar": "quux"}
		_, err = v.Validate(context.Background(), wrongTypeBar)
		require.Error(t, err)

		// Invalid case - wrong types for both
		wrongTypeBoth := map[string]any{"foo": "quux", "bar": "quux"}
		_, err = v.Validate(context.Background(), wrongTypeBoth)
		require.Error(t, err)

		// Valid case - ignores arrays
		arrayData := []any{"bar"}
		_, err = v.Validate(context.Background(), arrayData)
		require.NoError(t, err)

		// Valid case - ignores strings
		stringData := "foobar"
		_, err = v.Validate(context.Background(), stringData)
		require.NoError(t, err)
	})

	t.Run("multiple dependencies", func(t *testing.T) {
		quuxDependentSchema := schema.NewBuilder().
			Property("foo", schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
			Property("bar", schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
			MustBuild()

		fooDependentSchema := schema.NewBuilder().
			Property("bar", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			MustBuild()

		s := schema.NewBuilder().
			DependentSchemas(map[string]schema.SchemaOrBool{
				"quux": quuxDependentSchema,
				"foo":  fooDependentSchema,
			}).
			MustBuild()

		v, err := validator.Compile(context.Background(), s)
		require.NoError(t, err)

		// Invalid case - foo dependency requires bar to be string, but quux dependency requires bar to be integer
		// This should fail because both dependencies apply and they conflict
		conflictData := map[string]any{"foo": 1, "bar": 2, "quux": "baz"}
		_, err = v.Validate(context.Background(), conflictData)
		require.Error(t, err)

		// Valid case - only foo dependency applies
		onlyFooData := map[string]any{"foo": 1, "bar": "string"}
		_, err = v.Validate(context.Background(), onlyFooData)
		require.NoError(t, err)

		// Valid case - only quux dependency applies
		onlyQuuxData := map[string]any{"bar": 2, "quux": "baz"}
		_, err = v.Validate(context.Background(), onlyQuuxData)
		require.NoError(t, err)
	})

	t.Run("empty dependent schemas", func(t *testing.T) {
		s := schema.NewBuilder().
			DependentSchemas(map[string]schema.SchemaOrBool{}).
			MustBuild()

		v, err := validator.Compile(context.Background(), s)
		require.NoError(t, err)

		// Any data should be valid
		data := map[string]any{"foo": "bar"}
		_, err = v.Validate(context.Background(), data)
		require.NoError(t, err)
	})

	t.Run("dependent schema with complex validation", func(t *testing.T) {
		creditCardDependentSchema := schema.NewBuilder().
			Property("billing_address", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Required("billing_address").
			MustBuild()

		s := schema.NewBuilder().
			Types(schema.ObjectType).
			DependentSchemas(map[string]schema.SchemaOrBool{
				"credit_card": creditCardDependentSchema,
			}).
			MustBuild()

		v, err := validator.Compile(context.Background(), s)
		require.NoError(t, err)

		// Valid case - credit_card present with required billing_address
		validData := map[string]any{
			"credit_card":     "1234-5678-9012-3456",
			"billing_address": "123 Main St",
		}
		_, err = v.Validate(context.Background(), validData)
		require.NoError(t, err)

		// Invalid case - credit_card present but missing billing_address
		invalidData := map[string]any{
			"credit_card": "1234-5678-9012-3456",
		}
		_, err = v.Validate(context.Background(), invalidData)
		require.Error(t, err)

		// Valid case - no credit_card, so no dependency
		noDependencyData := map[string]any{
			"payment_method": "cash",
		}
		_, err = v.Validate(context.Background(), noDependencyData)
		require.NoError(t, err)
	})

	t.Run("dependent schemas with references", func(t *testing.T) {
		personSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("age", schema.NewBuilder().Types(schema.IntegerType).Minimum(0).MustBuild()).
			Required("name", "age").
			MustBuild()

		nameRefSchema := schema.NewBuilder().
			Reference("#/$defs/person").
			MustBuild()

		s := schema.NewBuilder().
			Types(schema.ObjectType).
			DependentSchemas(map[string]schema.SchemaOrBool{
				"name": nameRefSchema,
			}).
			Definitions("person", personSchema).
			MustBuild()

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, s)

		v, err := validator.Compile(ctx, s)
		require.NoError(t, err)

		// Valid case - name present with required age
		validData := map[string]any{
			"name": "John Doe",
			"age":  30,
		}
		_, err = v.Validate(context.Background(), validData)
		require.NoError(t, err)

		// Invalid case - name present but missing age
		invalidData := map[string]any{
			"name": "John Doe",
		}
		_, err = v.Validate(context.Background(), invalidData)
		require.Error(t, err)

		// Invalid case - name present but invalid age
		invalidAgeData := map[string]any{
			"name": "John Doe",
			"age":  -5,
		}
		_, err = v.Validate(context.Background(), invalidAgeData)
		require.Error(t, err)

		// Valid case - no name, so no dependency
		noDependencyData := map[string]any{
			"title": "Dr.",
		}
		_, err = v.Validate(context.Background(), noDependencyData)
		require.NoError(t, err)
	})

	t.Run("debug: incompatible root and dependent schema", func(t *testing.T) {
		// Test case from JSON Schema compliance test suite
		fooDependentSchema := schema.NewBuilder().
			Property("bar", schema.New()).
			AdditionalProperties(schema.BoolSchema(false)).
			MustBuild()

		s := schema.NewBuilder().
			Property("foo", schema.New()).
			DependentSchemas(map[string]schema.SchemaOrBool{
				"foo": fooDependentSchema,
			}).
			MustBuild()

		v, err := validator.Compile(context.Background(), s)
		require.NoError(t, err)

		// This should FAIL because:
		// 1. Root allows "foo" property
		// 2. But dependent schema (triggered by "foo") only allows "bar" property with additionalProperties: false
		// 3. So "foo" violates the dependent schema's additionalProperties constraint
		data := map[string]any{"foo": 1}
		_, err = v.Validate(context.Background(), data)
		t.Logf("Validation result for {\"foo\": 1}: %v", err)
		require.Error(t, err, "Should fail because foo violates dependent schema's additionalProperties: false")

		// This should FAIL for the same reason
		dataBoth := map[string]any{"foo": 1, "bar": 2}
		_, err = v.Validate(context.Background(), dataBoth)
		t.Logf("Validation result for {\"foo\": 1, \"bar\": 2}: %v", err)
		require.Error(t, err, "Should fail because foo violates dependent schema's additionalProperties: false")

		// This should PASS because no "foo" property, so no dependent schema triggered
		dataBar := map[string]any{"bar": 1}
		_, err = v.Validate(context.Background(), dataBar)
		t.Logf("Validation result for {\"bar\": 1}: %v", err)
		require.NoError(t, err)

		// This should PASS because no dependency triggered
		dataBaz := map[string]any{"baz": 1}
		_, err = v.Validate(context.Background(), dataBaz)
		t.Logf("Validation result for {\"baz\": 1}: %v", err)
		require.NoError(t, err)
	})

	t.Run("debug: schema type inference", func(t *testing.T) {
		// Test case from JSON Schema compliance test suite
		fooDependentSchema := schema.NewBuilder().
			Property("bar", schema.New()).
			AdditionalProperties(schema.BoolSchema(false)).
			MustBuild()

		s := schema.NewBuilder().
			Property("foo", schema.New()).
			DependentSchemas(map[string]schema.SchemaOrBool{
				"foo": fooDependentSchema,
			}).
			MustBuild()

		// Check what types are defined
		t.Logf("Schema types: %v", s.Types())
		t.Logf("Has explicit types: %v", len(s.Types()) > 0)
		t.Logf("Has properties: %v", s.Has(schema.PropertiesField))
		t.Logf("Has dependent schemas: %v", s.Has(schema.DependentSchemasField))

		_, err := validator.Compile(context.Background(), s)
		require.NoError(t, err)

		// Try to test without the root properties constraint
		fooDependentSchema2 := schema.NewBuilder().
			Property("bar", schema.New()).
			AdditionalProperties(schema.BoolSchema(false)).
			MustBuild()

		s2 := schema.NewBuilder().
			Types(schema.ObjectType).
			DependentSchemas(map[string]schema.SchemaOrBool{
				"foo": fooDependentSchema2,
			}).
			MustBuild()

		v2, err := validator.Compile(context.Background(), s2)
		require.NoError(t, err)

		// This should FAIL because dependent schema rejects "foo"
		data := map[string]any{"foo": 1}
		_, err = v2.Validate(context.Background(), data)
		t.Logf("Without root properties - validation result for {\"foo\": 1}: %v", err)
		require.Error(t, err, "Should fail because dependent schema rejects foo")
	})

	t.Run("isolation: dependent schema alone", func(t *testing.T) {
		// Test just the dependent schema part
		depSchema := schema.NewBuilder().
			Property("bar", schema.New()).
			AdditionalProperties(schema.BoolSchema(false)).
			MustBuild()

		v, err := validator.Compile(context.Background(), depSchema)
		require.NoError(t, err)

		// This should FAIL - "foo" is not allowed because additionalProperties: false
		data := map[string]any{"foo": 1}
		_, err = v.Validate(context.Background(), data)
		t.Logf("Dependent schema validation result for {\"foo\": 1}: %v", err)
		require.Error(t, err, "Dependent schema should reject foo")

		// This should PASS - "bar" is allowed
		dataBar := map[string]any{"bar": 1}
		_, err = v.Validate(context.Background(), dataBar)
		t.Logf("Dependent schema validation result for {\"bar\": 1}: %v", err)
		require.NoError(t, err)
	})

	t.Run("isolation: direct dependent schemas validator", func(t *testing.T) {
		// Test the DependentSchemasValidator directly
		depSchema := schema.NewBuilder().
			Property("bar", schema.New()).
			AdditionalProperties(schema.BoolSchema(false)).
			MustBuild()

		ctx := context.Background()
		depSchemas := map[string]*schema.Schema{"foo": depSchema}

		depValidator, err := validator.DependentSchemasValidator(ctx, depSchemas)
		require.NoError(t, err)

		// This should FAIL - when "foo" is present, dependent schema rejects it
		data := map[string]any{"foo": 1}
		_, err = depValidator.Validate(context.Background(), data)
		t.Logf("DependentSchemasValidator result for {\"foo\": 1}: %v", err)
		require.Error(t, err, "Should fail because dependent schema is triggered and rejects foo")

		// This should PASS - no "foo" property, so no dependent schema triggered
		dataBar := map[string]any{"bar": 1}
		_, err = depValidator.Validate(context.Background(), dataBar)
		t.Logf("DependentSchemasValidator result for {\"bar\": 1}: %v", err)
		require.NoError(t, err)
	})
}
