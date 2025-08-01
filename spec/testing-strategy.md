# Testing Strategy Specification

## Overview

The testing strategy ensures comprehensive validation of JSON Schema functionality through meta-schema testing, performance benchmarking, and systematic validation testing. This strategy minimizes external dependencies while providing thorough coverage.

## Meta Schema Testing

### Local Test Server Architecture

**Problem**: Meta-schema testing requires external network access to schema.org
**Solution**: Local HTTP test server with embedded schema files

### Implementation Pattern

**Embedded File System**:
```go
//go:embed testdata/schemas
var metaSchema embed.FS

func setupTestServer(t *testing.T) *httptest.Server {
    server := httptest.NewUnstartedServer(nil)
    
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Serve files from metaSchema embed.FS
        // Handle template substitution for $id fields
    })
    server.Config.Handler = mux
    server.Start()
    
    return server
}
```

### Template System

**Template Format**: Store meta-schemas as Go templates in `testdata/schemas/meta/`
**Dynamic Values**: Replace `$id` fields with local test server URLs
**Template Variables**: 
- `{{.BaseURL}}` - Test server base URL
- `{{.SchemaPath}}` - Relative path to schema

**Example Template**:
```json
{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "$id": "{{.BaseURL}}/draft/2020-12/meta/core",
    "title": "Core vocabulary meta-schema",
    ...
}
```

### Test Data Organization

**Directory Structure**:
```
testdata/
├── schemas/
│   ├── meta/                 # Meta-schema templates
│   │   ├── core.json.tmpl    # Core vocabulary
│   │   ├── validation.json.tmpl # Validation vocabulary
│   │   └── format.json.tmpl   # Format vocabulary
│   ├── examples/             # Example schemas for testing
│   └── invalid/              # Invalid schemas for error testing
└── data/                     # Test data for validation
```

## Logging Integration

### Context-Based Logging

**Logger Passing**: Pass logger through context for validation tracing
```go
ctx := context.Background()
ctx = validator.WithTraceSlog(ctx, slog.New(slog.NewJSONHandler(os.Stdout, nil)))

validator, err := validator.Compile(ctx, schema)
result, err := validator.Validate(ctx, data)
```

### Logger Access Pattern

**Context Extraction**: Access logger from context with no-op fallback
```go
func TraceSlogFromContext(ctx context.Context) *slog.Logger {
    if logger, ok := ctx.Value(traceSlogKey{}).(*slog.Logger); ok {
        return logger
    }
    return slog.New(noopHandler{}) // No-op handler when no logger present
}
```

### Test Logging Configuration

**Verbose Testing**: Enable detailed logging during test development
```go
func TestValidation(t *testing.T) {
    var logger *slog.Logger
    if testing.Verbose() {
        logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        }))
    } else {
        logger = slog.New(noopHandler{})
    }
    
    ctx := validator.WithTraceSlog(context.Background(), logger)
    // ... test with detailed logging when -v flag is used
}
```

## Test Categories

### Unit Tests

**Validator Testing**: Test each validator type independently
- String validators with all constraint combinations
- Numeric validators with boundary conditions
- Array validators with complex item validation
- Object validators with property patterns
- Composite validators (allOf, anyOf, oneOf)

**Schema Construction**: Test schema building and manipulation
- Builder pattern validation
- SchemaOrBool handling
- JSON serialization/deserialization
- Clone functionality

**Context Management**: Test context passing and data extraction
- Context value storage and retrieval
- Type safety validation
- Error handling for missing values

### Integration Tests

**Complete Validation Workflows**: Test end-to-end validation scenarios
- Schema compilation to validator
- Complex nested validation
- Context data flow through validator chains
- Result merging and aggregation

**Reference Resolution**: Test reference and anchor resolution
- Local references within documents
- External reference loading (using test server)
- Circular reference handling
- Cache behavior validation

### Performance Tests

**Benchmark Suite**:
```go
func BenchmarkStringValidation(b *testing.B) {
    validator := String().MinLength(5).MaxLength(100).MustBuild()
    testString := "valid test string"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := validator.Validate(context.Background(), testString)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

**Performance Comparisons**:
- Schema compilation vs direct validator creation
- Generated validators vs compiled validators
- Memory allocation patterns
- Validation throughput measurements

### Error Testing

**Invalid Schema Handling**: Test error cases in schema construction
- Invalid JSON Schema documents
- Malformed references
- Circular dependencies
- Type mismatches

**Validation Error Cases**: Test comprehensive error reporting
- Constraint violations with detailed messages
- Context information in error messages
- Error aggregation for multiple failures

## Test Utilities

### Helper Functions

**Schema Builders**: Utility functions for common test schemas
```go
func StringSchema(minLen, maxLen int, pattern string) *Schema {
    return schema.NewBuilder().
        Type(schema.StringType).
        MinLength(uint(minLen)).
        MaxLength(uint(maxLen)).
        Pattern(pattern).
        MustBuild()
}
```

**Validator Factories**: Create common validators for testing
```go
func EmailValidator() validator.Interface {
    return validator.String().
        Format("email").
        MinLength(5).
        MaxLength(100).
        MustBuild()
}
```

### Test Data Management

**Shared Test Data**: Reusable test data sets
- Valid inputs for each validator type
- Invalid inputs with expected error patterns
- Edge cases and boundary conditions
- Large datasets for performance testing

**Data Generation**: Programmatic test data generation
- Property-based testing with generated inputs
- Fuzzing for edge case discovery
- Structured data generation for complex schemas

## Continuous Integration

### Test Execution

**Parallel Testing**: Run test suites in parallel for faster feedback
**Coverage Requirements**: Maintain high test coverage (>90%)
**Performance Regression Detection**: Benchmark comparison in CI

### Test Environment

**Isolated Testing**: Each test runs with clean state
**Resource Management**: Proper cleanup of test servers and resources
**Deterministic Testing**: Reproducible test results across environments

### Quality Gates

**Code Quality**: Static analysis and linting
**Performance Thresholds**: Ensure performance doesn't regress
**Memory Usage**: Monitor memory allocation patterns
**Test Reliability**: Identify and fix flaky tests

## Documentation Testing

### Example Validation

**Documentation Examples**: Test all code examples in documentation
**API Examples**: Validate API usage examples work correctly
**Tutorial Validation**: Ensure tutorials produce expected results

### Schema Examples

**Real-World Schemas**: Test with actual JSON Schema documents
- OpenAPI specifications
- JSON Schema store examples
- Community-contributed schemas

## Test Organization

### File Structure

```
validator/
├── string_test.go          # String validator unit tests
├── integer_test.go         # Integer validator unit tests
├── array_test.go           # Array validator unit tests
├── object_test.go          # Object validator unit tests
├── composite_test.go       # AllOf/AnyOf/OneOf tests
├── integration_test.go     # Integration test suite
├── benchmark_test.go       # Performance benchmarks
└── testdata/              # Test data and schemas
```

### Test Naming

**Descriptive Names**: Clear test names indicating what is being tested
**Consistent Patterns**: Follow standard Go testing conventions
**Grouped Tests**: Use subtests for related test cases

### Test Documentation

**Test Purpose**: Comment complex tests with their purpose
**Expected Behavior**: Document expected outcomes for edge cases
**Performance Expectations**: Document performance requirements for benchmarks