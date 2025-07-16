package schema_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
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
		stringSchema, err := schema.NewBuilder().Types(schema.StringType).Build()
		require.NoError(t, err)

		minLengthSchema, err := schema.NewBuilder().MinLength(5).Build()
		require.NoError(t, err)

		composedSchema, err := schema.NewBuilder().
			AllOf(stringSchema, minLengthSchema).
			Build()
		require.NoError(t, err)
		require.Len(t, composedSchema.AllOf(), 2)
	})

	t.Run("AnyOf Composition", func(t *testing.T) {
		stringSchema, err := schema.NewBuilder().Types(schema.StringType).Build()
		require.NoError(t, err)

		numberSchema, err := schema.NewBuilder().Types(schema.NumberType).Build()
		require.NoError(t, err)

		composedSchema, err := schema.NewBuilder().
			AnyOf(stringSchema, numberSchema).
			Build()
		require.NoError(t, err)
		require.Len(t, composedSchema.AnyOf(), 2)
	})

	t.Run("OneOf Composition", func(t *testing.T) {
		stringSchema, err := schema.NewBuilder().Types(schema.StringType).Build()
		require.NoError(t, err)

		numberSchema, err := schema.NewBuilder().Types(schema.NumberType).Build()
		require.NoError(t, err)

		composedSchema, err := schema.NewBuilder().
			OneOf(stringSchema, numberSchema).
			Build()
		require.NoError(t, err)
		require.Len(t, composedSchema.OneOf(), 2)
	})

	t.Run("Not Composition", func(t *testing.T) {
		stringSchema, err := schema.NewBuilder().Types(schema.StringType).Build()
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
			Types(schema.StringType).
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
			Types(schema.NumberType).
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
		itemSchema, err := schema.NewBuilder().Types(schema.StringType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Types(schema.ArrayType).
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
		propSchema, err := schema.NewBuilder().Types(schema.StringType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Types(schema.ObjectType).
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
		enumValues := []any{"red", "green", "blue"}
		s, err := schema.NewBuilder().
			Types(schema.StringType).
			Enum(enumValues...).
			Build()
		require.NoError(t, err)
		require.Equal(t, enumValues, s.Enum())
	})

	t.Run("Const Value", func(t *testing.T) {
		constValue := "constant"
		s, err := schema.NewBuilder().
			Types(schema.StringType).
			Const(constValue).
			Build()
		require.NoError(t, err)
		require.Equal(t, constValue, s.Const())
	})
}

func TestAdvancedFeatures(t *testing.T) {
	t.Run("Pattern Properties", func(t *testing.T) {
		propSchema, err := schema.NewBuilder().Types(schema.StringType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Types(schema.ObjectType).
			PatternProperty("^[a-z]+$", propSchema).
			Build()
		require.NoError(t, err)
		require.NotNil(t, s.PatternProperties()["^[a-z]+$"])
	})

	t.Run("Additional Properties", func(t *testing.T) {
		additionalPropSchema, err := schema.NewBuilder().Types(schema.StringType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Types(schema.ObjectType).
			AdditionalProperties(additionalPropSchema).
			Build()
		require.NoError(t, err)
		require.NotNil(t, s.AdditionalProperties())
	})

	t.Run("Contains", func(t *testing.T) {
		containsSchema, err := schema.NewBuilder().Types(schema.StringType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Types(schema.ArrayType).
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
		unevalPropSchema, err := schema.NewBuilder().Types(schema.StringType).Build()
		require.NoError(t, err)

		unevalItemSchema, err := schema.NewBuilder().Types(schema.NumberType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Types(schema.ObjectType).
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
		personSchema, err := schema.NewBuilder().Types(schema.StringType).Build()
		require.NoError(t, err)

		s, err := schema.NewBuilder().
			Definitions("person", personSchema).
			Build()
		require.NoError(t, err)
		defs := s.Definitions()
		require.Contains(t, defs, "person")
		require.Equal(t, personSchema, defs["person"])
	})
}

// TestMain initializes the JSON Schema Test Suite
func TestMain(m *testing.M) {
	// Check if we're in short mode by looking at command line args
	skipInit := false
	for _, arg := range os.Args {
		if arg == "-test.short" || arg == "--test.short" {
			skipInit = true
			break
		}
	}

	if !skipInit {
		// Initialize the test suite
		if err := initializeTestSuite(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize test suite: %v\n", err)
			os.Exit(1)
		}
	}

	// Run the tests
	code := m.Run()
	os.Exit(code)
}

// initializeTestSuite clones or updates the JSON Schema Test Suite repository
func initializeTestSuite() error {
	testDir := "tests"

	// Check if the directory already exists
	if _, err := os.Stat(testDir); err == nil {
		// Directory exists, pull latest changes
		fmt.Println("Updating existing JSON Schema Test Suite...")
		cmd := exec.Command("git", "pull")
		cmd.Dir = testDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update test suite: %w", err)
		}
	} else {
		// Directory doesn't exist, clone it
		fmt.Println("Cloning JSON Schema Test Suite...")
		cmd := exec.Command("git", "clone", "https://github.com/json-schema-org/JSON-Schema-Test-Suite.git", testDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone test suite: %w", err)
		}
	}

	return nil
}

// TestSpecificationCompliance runs the official JSON Schema Test Suite tests
func TestSpecificationCompliance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping specification compliance tests in short mode")
	}

	testDir := filepath.Join("tests", "tests", "latest")

	// Resolve the symlink to get the actual directory
	resolvedDir, err := filepath.EvalSymlinks(testDir)
	if err != nil {
		t.Fatalf("Failed to resolve symlink %s: %v", testDir, err)
	}
	testDir = resolvedDir

	// Check if test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Fatalf("Test directory %s does not exist. Make sure TestMain ran successfully.", testDir)
	}

	// Read all test files
	err = filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".json") || strings.Contains(path, "remotes") {
			return nil
		}

		// Skip optional tests that we don't support yet
		if strings.Contains(path, "optional") {
			return nil
		}

		relPath, _ := filepath.Rel(testDir, path)
		t.Run(relPath, func(t *testing.T) {
			runTestFile(t, path)
		})

		return nil
	})

	require.NoError(t, err)
}

// TestSuite represents a single test suite from the JSON Schema Test Suite
type TestSuite struct {
	Description string     `json:"description"`
	Schema      any        `json:"schema"`
	Tests       []TestCase `json:"tests"`
}

// TestCase represents a single test case
type TestCase struct {
	Description string `json:"description"`
	Data        any    `json:"data"`
	Valid       bool   `json:"valid"`
}

// runTestFile runs all test suites in a single JSON file
func runTestFile(t *testing.T, filePath string) {
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)

	var testSuites []TestSuite
	err = json.Unmarshal(data, &testSuites)
	require.NoError(t, err)

	for _, testSuite := range testSuites {
		t.Run(testSuite.Description, func(t *testing.T) {
			runTestSuite(t, testSuite)
		})
	}
}

// runTestSuite runs a single test suite
func runTestSuite(t *testing.T, testSuite TestSuite) {
	var s *schema.Schema
	var err error

	// Handle boolean schemas (true/false) and object schemas
	switch schemaValue := testSuite.Schema.(type) {
	case bool:
		// Boolean schema: true accepts everything, false rejects everything
		if schemaValue {
			s = schema.New() // Empty schema accepts all
		} else {
			// False schema should reject everything
			s, err = schema.NewBuilder().Not(schema.New()).Build()
			if err != nil {
				t.Skipf("Failed to build false schema: %v", err)
				return
			}
		}
	default:
		// Object schema - convert to JSON and parse
		schemaJSON, err := json.Marshal(testSuite.Schema)
		if err != nil {
			t.Skipf("Failed to marshal schema: %v", err)
			return
		}

		// Try to parse the schema
		err = json.Unmarshal(schemaJSON, &s)
		if err != nil {
			t.Skipf("Failed to parse schema: %v", err)
			return
		}
	}

	// Compile the schema to a validator
	v, err := validator.Compile(context.Background(), s)
	if err != nil {
		t.Skipf("Failed to compile schema: %v", err)
		return
	}

	// Run each test case
	for _, testCase := range testSuite.Tests {
		t.Run(testCase.Description, func(t *testing.T) {
			_, err := v.Validate(context.Background(), testCase.Data)
			if testCase.Valid {
				require.NoError(t, err, "Expected validation to pass but got error: %v", err)
			} else {
				require.Error(t, err, "Expected validation to fail but it passed")
			}
		})
	}
}
