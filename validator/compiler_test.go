package validator

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
)

func TestCompile_Simple(t *testing.T) {
	ctx := context.Background()

	// Test empty schema
	emptySchema := schema.NewBuilder().MustBuild()
	validator, err := Compile(ctx, emptySchema)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if _, ok := validator.(*EmptyValidator); !ok {
		t.Fatalf("Expected EmptyValidator, got %T", validator)
	}

	// Test simple string schema
	stringSchema := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(1).
		MaxLength(10).
		MustBuild()

	validator, err = Compile(ctx, stringSchema)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should be just the string validator since no unevaluated constraints
	result, err := validator.Validate(ctx, "hello")
	if err != nil {
		t.Fatalf("Expected validation to pass, got %v", err)
	}
	if result != nil {
		t.Logf("Got result: %+v", result)
	}

	// Test validation failure
	_, err = validator.Validate(ctx, "this string is too long")
	if err == nil {
		t.Fatalf("Expected validation to fail for string too long")
	}
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
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should be an unevaluatedCoordinator
	if _, ok := validator.(*unevaluatedCoordinator); !ok {
		t.Fatalf("Expected unevaluatedCoordinator, got %T", validator)
	}

	// Test validation of valid object
	validObj := map[string]any{
		"name": "John",
		"age":  30,
	}
	result, err := validator.Validate(ctx, validObj)
	if err != nil {
		t.Fatalf("Expected validation to pass, got %v", err)
	}
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
	if err == nil {
		t.Fatalf("Expected validation to fail for unevaluated property")
	}
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
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should be just the allOf validator since no unevaluated constraints
	if _, ok := validator.(*allOfValidator); !ok {
		t.Fatalf("Expected allOfValidator, got %T", validator)
	}

	// Test validation
	result, err := validator.Validate(ctx, "hello")
	if err != nil {
		t.Fatalf("Expected validation to pass, got %v", err)
	}
	if result != nil {
		t.Logf("Got result: %+v", result)
	}
}
