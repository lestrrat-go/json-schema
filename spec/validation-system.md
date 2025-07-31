# Validation System Specification

## Overview

The validation system provides executable validation logic compiled from Schema objects. This system handles data validation, error reporting, result collection, and context-aware validation execution with high performance.

## Core Interface

### Validator Interface

```go
type Interface interface {
    Validate(context.Context, any) (Result, error)
}
```

**Key Characteristics**:
- Context-aware validation
- Any-type input handling  
- Result objects for complex validation tracking
- Error reporting with detailed messages

### Compilation Process

**Schema to Validator**: `validator.Compile(ctx, schema)` converts schemas to optimized validators
- One-time compilation cost
- Reusable validator objects
- Performance optimized execution
- Context parameter handling

## Validation Targets

### Supported Types

**Simple Values**: string, bool, int, []slice, map[...]..., struct
**Reflection-Based**: Automatic struct field resolution with JSON tag support
**Custom Resolvers**: Objects implementing query interfaces

```go
type ObjectFieldResolver interface {
    ResolveObjectField(string) (any, error)
}

type ArrayIndexResolver interface {
    ResolveArrayIndex(int) (any, error)
}
```

### Struct Field Resolution

**JSON Tag Priority**: Respect JSON tags over Go field names
**Example**: `type Foo { Bar string `json:"baz"` }` resolves field `baz` to `Bar`
**Reflection**: Use `reflect` package for field access

## Result Objects

### Result Types

**ObjectResult**: For object validation results
**ArrayResult**: For array validation results with size pre-allocation

```go
func NewObjectResult() *ObjectResult
func NewArrayResult(size int) *ArrayResult
```

### Result Merging

**Utility Function**: `MergeResults(dst Result, list ...Result) error`
- Type-based merging logic
- Uses `blackmagic.AssignIfCompatible` for final assignment
- Helper functions: `mergeObjectResults()`, `mergeArrayResults()`

**Critical**: Results do NOT implement their own merge methods

## Context Passing

### Data Flow Patterns

**To Sub-validators**: Pass data down using context values
```go
ctx = schemactx.WithEvaluatedProperties(ctx, ...)
subValidator.Validate(ctx, ...)
```

**From Sub-validators**: Extract data using Result objects
```go
var props map[string]struct{}
err := validator.EvaluatedPropertiesFromContext(ctx, &props)
```

### Use Cases

- **unevaluatedProperties**: Track which properties have been validated
- **unevaluatedItems**: Track which array items have been evaluated
- **Dependent Schemas**: Pass compiled validators through context

## Composite Validators

### Reference Handling

**Problem**: Schema with $ref + additional specifications
**Solution**: Composite validator combining:
1. Validator compiled from reference
2. Validator compiled from remaining specification
3. Combined result merging

### AllOf/AnyOf/OneOf

**Distinct Validator Types**: Separate implementations for each composite type

| Validator | Success Requirement | Implementation Strategy |
|-----------|-------------------|------------------------|
| allOf | All validators pass | Fail on first failure, merge all results |
| anyOf | At least one passes | Succeed on first success |
| oneOf | Exactly one passes | Count successes, fail if count ≠ 1 |

**Construction Pattern**:
```go
allOfValidator := validator.AllOf(
    validator.String().MinLength(3).MustBuild(),
    validator.String().MaxLength(10).MustBuild(),
)
```

## Specialized Validators

### String Validation

**Pattern**: Reference implementation for constraint validators
- MinLength, MaxLength constraints
- Pattern matching with compiled regexes
- Format validation (email, uri, etc.)
- Enum and const value checking

### Numeric Validation

**Generated Validators**: Integer and number validators created via code generation
- MultipleOf, minimum, maximum constraints
- Exclusive bounds handling
- Type-specific optimizations

### Array Validation

**Complex Logic**: Multiple validation phases
- Length constraints (minItems, maxItems)
- Uniqueness validation
- Prefix items validation (tuple support)
- Items and additionalItems validation
- Contains validation with min/max counts
- Unevaluated items handling

### Object Validation

**Property Validation**: 
- Properties and patternProperties
- AdditionalProperties handling
- DependentSchemas compilation and execution
- UnevaluatedProperties tracking

## Builder Pattern Implementation

### Validator Builders

**Error Accumulation**: Collect errors during building, validate at Build()
```go
validator := String().
    MinLength(5).
    MaxLength(10).
    Pattern("^[a-z]+$").
    Build()
```

**Build vs MustBuild**:
- `Build()` returns (Interface, error)
- `MustBuild()` panics on error - use for known-good configurations

### Validation Rules

**Build-Time Validation**: Validate constraints make sense
- MinLength ≤ MaxLength
- Pattern compiles to valid regex
- Enum values match expected type

## Advanced Features

### EmptyValidator

**Always Pass**: Returns (nil, nil) for any input
**Use Cases**: True boolean schemas, reference placeholders

### NotValidator

**Negation**: Wraps another validator and inverts its result
**Pattern**: `NotValidator{validator: &EmptyValidator{}}` for false boolean schemas

### DependentSchemas

**Compilation Phase**: Convert dependent schemas to validators during initial compilation
**Execution Phase**: Pass compiled validators through context
**Two-Pass Validation**: Properties validation, then dependent schema validation

## Performance Considerations

### Compilation Optimization

- Pre-compile regex patterns
- Optimize validator chains  
- Minimize reflection usage
- Cache frequently-used validators

### Execution Optimization

- Early termination for failing validators
- Minimal memory allocation
- Efficient result object reuse
- Context value optimization

## Testing Strategy

### Unit Testing

- Test each validator type independently
- Verify error message accuracy
- Test edge cases and boundary conditions
- Validate context passing behavior

### Integration Testing

- Test composite validator behavior
- Verify result merging correctness
- Test complex schema compilation
- Validate context data flow

### Performance Testing

- Benchmark compilation vs execution
- Compare with direct validation approaches
- Measure memory allocation patterns
- Test with large data sets