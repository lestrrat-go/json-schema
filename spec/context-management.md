# Context Management Specification

## Overview

Context management provides a type-safe system for passing data between validators during validation execution. This system avoids circular dependencies between packages while enabling complex validation workflows like unevaluated property tracking.

## Architecture Pattern

### Problem: Circular Dependencies

**Challenge**: Both `schema` and `validator` packages need access to context utilities
**Solution**: Internal `schemactx` package for shared context utilities

### Package Structure

```
internal/schemactx/     # Shared context utilities
├── keys.go            # Context key definitions
├── evaluatedprops.go  # Evaluated properties context
├── evaluateditems.go  # Evaluated items context
└── base.go           # Base schema/URI context
```

## Context Key Pattern

### Type-Safe Keys

**Pattern**: Unexported empty structs as context keys
```go
package schemactx

type evaluatedPropsKey struct{}
type baseSchemaKey struct{}
type baseURIKey struct{}
```

**Benefits**:
- Prevents key collisions
- Type safety at compile time
- No external package interference

### Helper Functions

**Internal Package Functions**:
```go
func WithEvaluatedProperties(ctx context.Context, props map[string]struct{}) context.Context {
    return context.WithValue(ctx, evaluatedPropsKey{}, props)
}

func EvaluatedPropertiesFromContext(ctx context.Context, dst any) error {
    v := ctx.Value(evaluatedPropsKey{})
    if v == nil {
        return errors.New("evaluated properties not found in context")
    }
    return blackmagic.AssignIfCompatible(dst, v)
}
```

### Public API Wrappers

**Schema Package Exports**:
```go
package schema

func WithEvaluatedProperties(ctx context.Context, props map[string]struct{}) context.Context {
    return schemactx.WithEvaluatedProperties(ctx, props)
}

func EvaluatedPropertiesFromContext(ctx context.Context) map[string]struct{} {
    var props map[string]struct{}
    if err := schemactx.EvaluatedPropertiesFromContext(ctx, &props); err != nil {
        return nil // or empty map
    }
    return props
}
```

## Context Data Types

### EvaluatedProperties

**Purpose**: Track which object properties have been validated
**Type**: `map[string]struct{}` for efficient set operations
**Usage**: unevaluatedProperties validation

```go
ctx = WithEvaluatedProperties(ctx, evaluatedProps)
result, err := validator.Validate(ctx, object)
```

### EvaluatedItems

**Purpose**: Track which array items have been validated
**Type**: `[]bool` or `map[int]struct{}` for index tracking
**Usage**: unevaluatedItems validation

```go
ctx = WithEvaluatedItems(ctx, evaluatedItems)
result, err := validator.Validate(ctx, array)
```

### BaseSchema

**Purpose**: Pass the root schema for reference resolution
**Type**: `*Schema`
**Usage**: $ref and anchor resolution

```go
ctx = WithBaseSchema(ctx, rootSchema)
resolvedSchema, err := resolver.ResolveReference(ctx, "#/definitions/User")
```

### BaseURI

**Purpose**: Base URI for relative reference resolution
**Type**: `string` (URI)
**Usage**: External reference resolution

```go
ctx = WithBaseURI(ctx, "https://example.com/schema.json")
resolvedSchema, err := resolver.ResolveReference(ctx, "./user.json")
```

### DependentSchemas

**Purpose**: Pass compiled dependent schema validators
**Type**: `map[string]Interface` (property name -> validator)
**Usage**: Object validation with dependent schemas

```go
dependentValidators := map[string]Interface{
    "creditCard": compiledCreditCardValidator,
    "billingAddress": compiledAddressValidator,
}
ctx = WithDependentSchemas(ctx, dependentValidators)
```

## Implementation Requirements

### AssignIfCompatible Integration

**Purpose**: Type-safe value extraction from context
**Pattern**: All context extraction uses `blackmagic.AssignIfCompatible`

```go
func EvaluatedPropertiesFromContext(ctx context.Context, dst any) error {
    v := ctx.Value(evaluatedPropsKey{})
    if v == nil {
        return errors.New("evaluated properties not found")
    }
    return blackmagic.AssignIfCompatible(dst, v)
}
```

### Error Handling

**Missing Values**: Return appropriate errors when context values are missing
**Type Mismatches**: Let `AssignIfCompatible` handle type validation
**Fallback Values**: Public API can return zero values or nil for convenience

### Context Propagation

**Validation Chain**: Context flows through all validator calls
**Result Aggregation**: Context data collected from sub-validators via Result objects
**Immutability**: Context values should not be modified, only replaced

## Usage Patterns

### UnevaluatedProperties Workflow

```go
// 1. Initialize tracking
evaluatedProps := make(map[string]struct{})

// 2. Pass to first validator
ctx = WithEvaluatedProperties(ctx, evaluatedProps)
result1, err := propertiesValidator.Validate(ctx, object)

// 3. Merge evaluated properties from result
if objectResult, ok := result1.(*ObjectResult); ok {
    for prop := range objectResult.EvaluatedProperties() {
        evaluatedProps[prop] = struct{}{}
    }
}

// 4. Pass updated context to next validator
ctx = WithEvaluatedProperties(ctx, evaluatedProps)
result2, err := additionalPropertiesValidator.Validate(ctx, object)

// 5. Check for unevaluated properties
for prop := range objectProperties {
    if _, evaluated := evaluatedProps[prop]; !evaluated {
        // Property not evaluated by any validator
        return fmt.Errorf("unevaluated property: %s", prop)
    }
}
```

### DependentSchemas Workflow

```go
// 1. Compile dependent schemas during validator compilation
dependentValidators := make(map[string]Interface)
for propName, depSchema := range schema.DependentSchemas() {
    validator, err := Compile(ctx, depSchema)
    if err != nil {
        return err
    }
    dependentValidators[propName] = validator
}

// 2. Pass compiled validators through context during validation
ctx = WithDependentSchemas(ctx, dependentValidators)
result, err := objectValidator.Validate(ctx, object)

// 3. Object validator extracts and uses dependent validators
var depValidators map[string]Interface
if err := DependentSchemasFromContext(ctx, &depValidators); err == nil {
    for propName, validator := range depValidators {
        if objectHasProperty(object, propName) {
            _, err := validator.Validate(ctx, object)
            if err != nil {
                return err
            }
        }
    }
}
```

## Testing Requirements

### Context Utilities

- Test context value storage and retrieval
- Verify type safety with invalid types
- Test missing value error handling
- Validate `AssignIfCompatible` integration

### Validation Integration

- Test context data flow through validator chains
- Verify evaluated property/item tracking
- Test dependent schema passing
- Validate complex nested validation scenarios

### Package Isolation

- Ensure no circular dependencies
- Test public API wrapper functions
- Verify internal utilities work correctly
- Test from both schema and validator packages

## File Organization

### Internal Package Structure

- `internal/schemactx/context.go` - Core context utilities and key definitions
- `internal/schemactx/evaluated.go` - Evaluated properties and items handling
- `internal/schemactx/resolution.go` - Base schema and URI context utilities

### Public API Wrappers

- `schema/context.go` - Public context utilities exported from schema package
- `validator/context.go` - Validator-specific context utilities if needed

## Performance Considerations

### Context Overhead

- Minimize context.WithValue calls
- Reuse context objects when possible
- Use efficient data structures (map[string]struct{} for sets)
- Avoid deep context nesting

### Memory Management

- Clear large context values when validation completes
- Use appropriate data structure sizes
- Avoid memory leaks in long-running validations