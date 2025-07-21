package validator

import (
	"context"
	"strings"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

func TestCodeGeneration(t *testing.T) {
	tests := []struct {
		name            string
		createValidator func(t *testing.T) Interface
		testValue       any
		shouldPass      bool
		checkGenerated  func(t *testing.T, code string)
	}{
		{
			name: "SimpleStringValidator",
			createValidator: func(_ *testing.T) Interface {
				return String().MinLength(5).MaxLength(10).MustBuild()
			},
			testValue:  "hello",
			shouldPass: true,
			checkGenerated: func(t *testing.T, code string) {
				require.Contains(t, code, "validator.String().")
				require.Contains(t, code, "MinLength(5).")
				require.Contains(t, code, "MaxLength(10).")
				require.Contains(t, code, "MustBuild()")
			},
		},
		{
			name: "StringValidatorWithPattern",
			createValidator: func(_ *testing.T) Interface {
				return String().Pattern("^[a-z]+$").MustBuild()
			},
			testValue:  "hello",
			shouldPass: true,
			checkGenerated: func(t *testing.T, code string) {
				require.Contains(t, code, `Pattern("^[a-z]+$").`)
			},
		},
		{
			name: "StringValidatorWithEnum",
			createValidator: func(_ *testing.T) Interface {
				return String().Enum("foo", "bar", "baz").MustBuild()
			},
			testValue:  "foo",
			shouldPass: true,
			checkGenerated: func(t *testing.T, code string) {
				require.Contains(t, code, `Enum(`)
				require.Contains(t, code, `[]string{`)
				require.Contains(t, code, `"foo",`)
				require.Contains(t, code, `"bar",`)
				require.Contains(t, code, `"baz",`)
			},
		},
		{
			name: "IntegerValidator",
			createValidator: func(_ *testing.T) Interface {
				return Integer().Minimum(0).Maximum(100).MustBuild()
			},
			testValue:  42,
			shouldPass: true,
			checkGenerated: func(t *testing.T, code string) {
				require.Contains(t, code, "validator.Integer().")
				require.Contains(t, code, "Minimum(0).")
				require.Contains(t, code, "Maximum(100).")
			},
		},
		{
			name: "NumberValidator",
			createValidator: func(_ *testing.T) Interface {
				return Number().MultipleOf(0.5).MustBuild()
			},
			testValue:  2.5,
			shouldPass: true,
			checkGenerated: func(t *testing.T, code string) {
				require.Contains(t, code, "validator.Number().")
				require.Contains(t, code, "MultipleOf(0.5).")
			},
		},
		{
			name: "BooleanValidator",
			createValidator: func(_ *testing.T) Interface {
				return Boolean().Const(true).MustBuild()
			},
			testValue:  true,
			shouldPass: true,
			checkGenerated: func(t *testing.T, code string) {
				require.Contains(t, code, "validator.Boolean().")
				require.Contains(t, code, "Const(true).")
			},
		},
		{
			name: "ArrayValidator",
			createValidator: func(_ *testing.T) Interface {
				return Array().MinItems(1).MaxItems(5).UniqueItems(true).MustBuild()
			},
			testValue:  []any{1, 2, 3},
			shouldPass: true,
			checkGenerated: func(t *testing.T, code string) {
				require.Contains(t, code, "validator.Array().")
				require.Contains(t, code, "MinItems(1).")
				require.Contains(t, code, "MaxItems(5).")
				require.Contains(t, code, "UniqueItems(true).")
			},
		},
		{
			name: "MultiValidatorAnd",
			createValidator: func(_ *testing.T) Interface {
				v1 := String().MinLength(3).MustBuild()
				v2 := String().MaxLength(10).MustBuild()
				mv := NewMultiValidator(AndMode)
				mv.Append(v1)
				mv.Append(v2)
				return mv
			},
			testValue:  "hello",
			shouldPass: true,
			checkGenerated: func(t *testing.T, code string) {
				require.Contains(t, code, "validator.NewMultiValidator(validator.AndMode)")
				require.Contains(t, code, "validator.String().")
				require.Contains(t, code, "MinLength(3).")
				require.Contains(t, code, "MaxLength(10).")
			},
		},
		{
			name: "EmptyValidator",
			createValidator: func(_ *testing.T) Interface {
				return &EmptyValidator{}
			},
			testValue:  "anything",
			shouldPass: true,
			checkGenerated: func(t *testing.T, code string) {
				require.Contains(t, code, "&validator.EmptyValidator{}")
			},
		},
		{
			name: "NotValidator",
			createValidator: func(_ *testing.T) Interface {
				child := String().MinLength(10).MustBuild()
				return &NotValidator{validator: child}
			},
			testValue:  "short",
			shouldPass: true, // "short" should pass NOT minLength(10)
			checkGenerated: func(t *testing.T, code string) {
				require.Contains(t, code, "validator.String().")
				require.Contains(t, code, "&validator.NotValidator{validator:")
				require.Contains(t, code, "MinLength(10).")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create original validator
			originalValidator := tt.createValidator(t)

			// Test that original validator works
			ctx := context.Background()
			_, err := originalValidator.Validate(ctx, tt.testValue)
			if tt.shouldPass {
				require.NoError(t, err, "Original validator should pass")
			} else {
				require.Error(t, err, "Original validator should fail")
			}

			// Generate code
			generator := NewCodeGenerator()
			var buf strings.Builder
			err = generator.Generate(&buf, originalValidator)
			require.NoError(t, err, "Code generation should succeed")
			code := buf.String()
			require.NotEmpty(t, code, "Generated code should not be empty")

			// Check that generated code contains expected elements
			if tt.checkGenerated != nil {
				tt.checkGenerated(t, code)
			}

			// Print generated code for debugging
			t.Logf("Generated code for %s:\n%s", tt.name, code)
		})
	}
}

func TestCodeGenerationWithCompiledValidators(t *testing.T) {
	// Test code generation with validators compiled from schemas
	tests := []struct {
		name       string
		schemaJSON string
		testValue  any
		shouldPass bool
	}{
		{
			name:       "CompiledStringValidator",
			schemaJSON: `{"type": "string", "minLength": 3, "maxLength": 10}`,
			testValue:  "hello",
			shouldPass: true,
		},
		{
			name:       "CompiledObjectValidator",
			schemaJSON: `{"type": "object", "properties": {"name": {"type": "string"}}, "required": ["name"]}`,
			testValue:  map[string]any{"name": "test"},
			shouldPass: true,
		},
		{
			name:       "CompiledArrayValidator",
			schemaJSON: `{"type": "array", "items": {"type": "string"}, "minItems": 1}`,
			testValue:  []any{"hello", "world"},
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse schema
			var s schema.Schema
			err := s.UnmarshalJSON([]byte(tt.schemaJSON))
			require.NoError(t, err)

			// Compile validator
			ctx := context.Background()
			originalValidator, err := Compile(ctx, &s)
			require.NoError(t, err)

			// Test original validator
			_, err = originalValidator.Validate(ctx, tt.testValue)
			if tt.shouldPass {
				require.NoError(t, err, "Original validator should pass")
			} else {
				require.Error(t, err, "Original validator should fail")
			}

			// Generate code
			generator := NewCodeGenerator()
			var buf strings.Builder
			err = generator.Generate(&buf, originalValidator)
			require.NoError(t, err, "Code generation should succeed")
			code := buf.String()
			require.NotEmpty(t, code, "Generated code should not be empty")

			t.Logf("Generated code for %s:\n%s", tt.name, code)
		})
	}
}


// unsupportedValidator is a mock validator type for testing
type unsupportedValidator struct{}

func (v *unsupportedValidator) Validate(_ context.Context, _ any) (Result, error) {
	return nil, nil
}

func TestUnsupportedValidatorType(t *testing.T) {
	generator := NewCodeGenerator()
	var buf strings.Builder
	err := generator.Generate(&buf, &unsupportedValidator{})
	// Note: The new interface doesn't return errors for unsupported types,
	// it falls back to EmptyValidator, so we expect success
	require.NoError(t, err)
	code := buf.String()
	require.Contains(t, code, "EmptyValidator")
}

func TestComplexNestedValidator(t *testing.T) {
	// Create a complex nested validator: allOf with string and integer constraints
	stringValidator := String().MinLength(3).MustBuild()
	integerValidator := Integer().Minimum(0).MustBuild()

	complexValidator := NewMultiValidator(AndMode)
	complexValidator.Append(stringValidator)
	complexValidator.Append(integerValidator)

	generator := NewCodeGenerator()
	var buf strings.Builder
	err := generator.Generate(&buf, complexValidator)
	require.NoError(t, err)
	code := buf.String()
	require.NotEmpty(t, code)

	// Should contain nested validator definitions
	require.Contains(t, code, "validator.NewMultiValidator(validator.AndMode)")
	require.Contains(t, code, "validator.String().")
	require.Contains(t, code, "validator.Integer().")

	t.Logf("Generated complex validator:\n%s", code)
}

// Benchmark code generation performance
func BenchmarkCodeGeneration(b *testing.B) {
	validator := String().MinLength(5).MaxLength(100).Pattern("^[a-z]+$").MustBuild()
	generator := NewCodeGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf strings.Builder
		err := generator.Generate(&buf, validator)
		if err != nil {
			b.Fatal(err)
		}
		buf.Reset() // Reset for next iteration
	}
}

func BenchmarkComplexValidatorGeneration(b *testing.B) {
	// Create a complex validator with multiple levels
	v1 := String().MinLength(1).MustBuild()
	v2 := Integer().Minimum(0).MustBuild()
	v3 := Array().MaxItems(10).MustBuild()

	nestedMulti := NewMultiValidator(OrMode)
	nestedMulti.Append(v2)
	nestedMulti.Append(v3)

	complex := NewMultiValidator(AndMode)
	complex.Append(v1)
	complex.Append(nestedMulti)

	generator := NewCodeGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf strings.Builder
		err := generator.Generate(&buf, complex)
		if err != nil {
			b.Fatal(err)
		}
		buf.Reset() // Reset for next iteration
	}
}
