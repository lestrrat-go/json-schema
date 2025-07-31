# Schema System Specification

## Overview

The Schema system provides immutable data containers for JSON Schema documents. Schemas handle construction, manipulation, serialization, and provide fluent builder APIs for creating complex schema structures.

## Core Components

### Schema Objects

**Primary Type**: `Schema` struct containing all JSON Schema 2020-12 fields
- **Purpose**: Immutable representation of JSON Schema documents
- **Thread Safety**: Fully thread-safe for concurrent access
- **Serialization**: Full JSON marshaling/unmarshaling support

### Builder Pattern

**Builder Interface**: Fluent API for schema construction
```go
type Builder interface {
    // Fluent methods for all schema fields
    Build() (*Schema, error)
    MustBuild() *Schema
}
```

**Key Characteristics**:
- Error accumulation during building
- Validation at Build() time
- Type-safe field setting
- Clone functionality for schema modification

### SchemaOrBool Interface

**Problem**: JSON Schema fields can accept either boolean or schema objects
**Solution**: Type-safe interface with concrete implementations

```go
type SchemaOrBool interface {
    schemaOrBool() // internal identifier
}

type BoolSchema bool     // represents boolean values
type Schema struct {...} // regular schema objects
```

**Implementation Requirements**:
- Custom JSON marshaling/unmarshaling
- Token-based parsing for mixed arrays (allOf, anyOf, etc.)
- Convenience functions: `TrueSchema()`, `FalseSchema()`

## Code Generation System

### Source of Truth

**objects.yml**: Single source defining all Schema fields
- JSON tags for serialization
- Go types for each field  
- Builder method generation rules
- Accessor method specifications

### Generated Files

**schema_gen.go**: Auto-generated content
- Complete Schema struct definition
- All accessor methods (HasX(), X())
- Builder implementation with fluent API
- Field validation logic

**Critical Rule**: Never modify _gen.go files directly - only through generators

### Generation Commands

- `./gen.sh` - Regenerate schema system
- Must run after modifying objects.yml
- Updates both schema and builder code

## Bit Field Optimization

### Field Presence Tracking

**Problem**: Expensive HasXXX() calls for every field check
**Solution**: Bit field tracking which fields are set

```go
// Instead of multiple method calls
if s.HasAnchor() && s.HasDefinitions() && s.HasProperties() { ... }

// Use efficient bit operations  
requiredFields := schema.AnchorField | schema.DefinitionsField | schema.PropertiesField
if s.Has(requiredFields) { ... }
```

### Implementation Strategy

**Internal Package**: `internal/field` defines bit constants
**Public Constants**: `schema` package exports user-facing constants
**Usage**: All module code references internal identifiers

## Clone Builder Pattern

### Use Case

Create builders pre-initialized with existing schema data for modifications:

```go
original := // existing Schema
modified, err := schema.NewBuilder().
    Clone(original).
    ResetReference().  // Remove $ref field
    Build()
```

### Benefits

- Selective field modification
- Reference processing workflows
- Schema transformation pipelines
- Immutable update patterns

## Implementation Guidelines

### File Organization

- `schema.go` - Core Schema type and basic methods
- `schema_gen.go` - Generated struct and accessors (DO NOT EDIT)  
- `builder.go` - Builder implementation utilities
- `bool_schema.go` - SchemaOrBool implementations

### JSON Handling

**Token-Based Parsing**: Required for mixed-type arrays
- Parse container-level JSON structure
- Distinguish boolean vs object tokens
- Handle both types in single field

**Custom Unmarshaling**: Implement for SchemaOrBool types
- BoolSchema handles boolean JSON values
- Schema handles object JSON values
- Error handling for invalid types

### Builder Validation

**Error Accumulation**: Collect errors during building, validate at Build()
**Type Safety**: Ensure field types match schema specifications
**Immutability**: Builders create new schemas, never modify existing ones

## Testing Requirements

### Schema Construction

- Test all field types and combinations
- Validate JSON serialization roundtrips
- Verify builder error handling
- Test clone functionality

### SchemaOrBool Handling

- Test boolean schema parsing
- Test mixed arrays (allOf with bools and schemas)
- Verify type safety and error cases
- Test convenience functions

### Code Generation

- Verify generated code compiles
- Test all accessor methods
- Validate builder pattern completeness
- Check bit field optimizations work correctly

## Integration Points

1. **Validator System**: Schemas compile to validators
2. **Reference Resolution**: Schemas contain anchors and references  
3. **Context Management**: Schema data flows through context
4. **CLI Tools**: Schemas parsed from command-line inputs