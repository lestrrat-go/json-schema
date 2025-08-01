package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestCompile_Simple(t *testing.T) {
	ctx := context.Background()

	// Test empty schema
	emptySchema := schema.NewBuilder().MustBuild()
	v, err := validator.Compile(ctx, emptySchema)
	require.NoError(t, err)
	require.NotNil(t, v)

	// Test simple string schema
	stringSchema := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(1).
		MaxLength(10).
		MustBuild()

	v, err = validator.Compile(ctx, stringSchema)
	require.NoError(t, err)

	// Should be just the string validator since no unevaluated constraints
	result, err := v.Validate(ctx, "hello")
	require.NoError(t, err)
	if result != nil {
		t.Logf("Got result: %+v", result)
	}

	// Test validation failure
	_, err = v.Validate(ctx, "this string is too long")
	require.Error(t, err)
}

func TestCompile_WithUnevaluated(t *testing.T) {
	ctx := context.Background()

	// Test schema with allOf + unevaluatedProperties
	allOfSchema := schema.NewBuilder().
		AllOf(
			schema.NewBuilder().
				Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
				MustBuild(),
		).
		Property("age", schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
		UnevaluatedProperties(schema.BoolSchema(false)). // No additional properties allowed
		MustBuild()

	v, err := validator.Compile(ctx, allOfSchema)
	require.NoError(t, err)
	require.NotNil(t, v)

	// Test validation of valid object
	validObj := map[string]any{
		"name": "John",
		"age":  30,
	}
	result, err := v.Validate(ctx, validObj)
	require.NoError(t, err)
	if result != nil {
		t.Logf("Got result: %+v", result)
	}

	// Test validation of object with unevaluated property (should fail)
	invalidObj := map[string]any{
		"name":  "John",
		"age":   30,
		"extra": "not allowed",
	}
	_, err = v.Validate(ctx, invalidObj)
	require.Error(t, err)
	t.Logf("Got expected error: %v", err)
}

func TestCompile_AllOfOnly(t *testing.T) {
	ctx := context.Background()

	// Test schema with allOf but no unevaluated constraints
	allOfSchema := schema.NewBuilder().
		AllOf(
			schema.NewBuilder().Types(schema.StringType).MinLength(1).MustBuild(),
			schema.NewBuilder().Types(schema.StringType).MaxLength(10).MustBuild(),
		).
		MustBuild()

	v, err := validator.Compile(ctx, allOfSchema)
	require.NoError(t, err)
	require.NotNil(t, v)

	// Test validation
	result, err := v.Validate(ctx, "hello")
	require.NoError(t, err)
	if result != nil {
		t.Logf("Got result: %+v", result)
	}
}
