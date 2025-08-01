package validator

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

func TestCompile_Simple(t *testing.T) {
	ctx := context.Background()

	// Test empty schema
	emptySchema := schema.NewBuilder().MustBuild()
	validator, err := Compile(ctx, emptySchema)
	require.NoError(t, err)
	require.IsType(t, &EmptyValidator{}, validator)

	// Test simple string schema
	stringSchema := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(1).
		MaxLength(10).
		MustBuild()

	validator, err = Compile(ctx, stringSchema)
	require.NoError(t, err)

	// Should be just the string validator since no unevaluated constraints
	result, err := validator.Validate(ctx, "hello")
	require.NoError(t, err)
	if result != nil {
		t.Logf("Got result: %+v", result)
	}

	// Test validation failure
	_, err = validator.Validate(ctx, "this string is too long")
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

	validator, err := Compile(ctx, allOfSchema)
	require.NoError(t, err)

	// Should be an unevaluatedCoordinator
	require.IsType(t, &unevaluatedCoordinator{}, validator)

	// Test validation of valid object
	validObj := map[string]any{
		"name": "John",
		"age":  30,
	}
	result, err := validator.Validate(ctx, validObj)
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
	_, err = validator.Validate(ctx, invalidObj)
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

	validator, err := Compile(ctx, allOfSchema)
	require.NoError(t, err)

	// Should be just the allOf validator since no unevaluated constraints
	require.IsType(t, &allOfValidator{}, validator)

	// Test validation
	result, err := validator.Validate(ctx, "hello")
	require.NoError(t, err)
	if result != nil {
		t.Logf("Got result: %+v", result)
	}
}
