# Core Architecture Principles

## Overview

The JSON Schema library follows a clear separation of concerns between schema representation and validation execution. This architectural decision provides performance, maintainability, and clarity benefits.

## Design Principle: Schema vs Validator Separation

### Schema Objects
- **Purpose**: Pure data vessels for JSON Schema metadata
- **Responsibilities**: 
  - Schema construction and manipulation
  - Serialization and deserialization of JSON Schema documents
  - Immutable data representation
- **Characteristics**: Thread-safe, reusable, focused on data operations

### Validator Objects  
- **Purpose**: Executable validation logic compiled from schemas
- **Responsibilities**:
  - High-performance data validation
  - Error reporting and result collection
  - Context-aware validation execution
- **Characteristics**: Optimized for repeated execution, compiled once and reused

### Benefits of Separation
1. **Single Responsibility**: Clear distinction between data and behavior
2. **Performance**: Validators compiled once, executed many times without recompilation
3. **Immutability**: Schemas remain unchanged during validation (thread-safe)
4. **Testability**: Schema construction and validation logic tested independently
5. **API Clarity**: Distinguishes "what to validate against" from "how to validate"

## Core Interfaces

### Schema System
```go
// Schema represents immutable JSON Schema documents
type Schema struct {
    // Internal fields for schema data
    // Generated accessor methods via code generation
}

// Builder provides fluent API for schema construction
type Builder interface {
    // Fluent methods for building schemas
    Build() (*Schema, error)
    MustBuild() *Schema
}
```

### Validation System
```go
// Interface defines the core validation contract
type Interface interface {
    Validate(context.Context, any) (Result, error)
}

// Compile converts schemas to executable validators
func Compile(ctx context.Context, schema *Schema) (Interface, error)
```

## Implementation Guidelines

### Coding Standards
- Avoid multiple exported functions with similar names (e.g., `Foo` and `FooWithContext`)
- Keep exported API simple and discoverable  
- Organize files logically with descriptive names
- Maintain files under 500-800 lines when possible
- Use unexported functions for internal flexibility

### File Organization
Structure code in logical groupings:
- `schema.go` - Core schema types and interfaces
- `builder.go` - Schema construction utilities  
- `validator/` - Validation system implementation
- `internal/` - Shared utilities and context management
- `cmd/` - Command-line tools