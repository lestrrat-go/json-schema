# Reference Resolution Specification

## Overview

The reference resolution system handles JSON references ($ref) and anchors within JSON Schema documents. This system must handle circular references, lazy resolution, and efficient caching while supporting both local and external references.

## Core Components

### Resolver Interface

**Primary Methods**:
```go
type Resolver interface {
    ResolveAnchor(ctx context.Context, dst *Schema, anchorName string) error
    ResolveJSONReference(ctx context.Context, dst *Schema, reference string) error
    ResolveReference(ctx context.Context, dst *Schema, reference string) error
}
```

**Unified Resolution**: `ResolveReference` automatically dispatches to appropriate resolver based on reference format

### Reference Types

**JSON References**: Standard JSON Pointer references (#/definitions/User)
**Anchors**: Named anchors within schema documents (#anchor-name)
**External References**: URLs to external schema documents (https://example.com/schema.json)
**Relative References**: Relative paths to other schema files (./user.json)

## Context Dependencies

### BaseSchema Context

**Purpose**: Provide root schema for relative resolution
**Usage**: Pass the containing schema document for reference resolution

```go
ctx = WithBaseSchema(ctx, rootSchema)
err := resolver.ResolveReference(ctx, &dst, "#/definitions/User")
```

### BaseURI Context

**Purpose**: Base URI for resolving relative external references
**Usage**: Establish base URI for the current schema document

```go
ctx = WithBaseURI(ctx, "https://example.com/schemas/")
err := resolver.ResolveReference(ctx, &dst, "./user.json")
```

## Resolution Strategies

### Lazy Resolution

**Principle**: Avoid resolving references until absolutely necessary
**Benefits**: 
- Prevents infinite loops in circular references
- Improves performance for unused references
- Reduces memory usage

**Implementation**: Store references as unresolved identifiers, resolve during validation

### Circular Reference Handling

**Problem**: References can form circular dependencies
**Solution**: Reference tracking with resolution depth limits

```go
type ResolutionContext struct {
    visited    map[string]bool
    maxDepth   int
    currentDepth int
}
```

**Strategy**:
1. Track visited references during resolution
2. Detect circular references and break resolution chain
3. Use placeholder validators for circular references
4. Implement depth limits to prevent stack overflow

### Caching Strategy

**Resolution Cache**: Store resolved schemas to avoid repeated resolution
**Cache Key**: Normalize reference strings for consistent lookup
**Cache Invalidation**: Handle cache invalidation for dynamic schema updates

```go
type ResolverCache struct {
    schemas map[string]*Schema
    mutex   sync.RWMutex
}
```

## Reference Formats

### JSON Pointer References

**Format**: `#/path/to/property`
**Examples**:
- `#/definitions/User` - Definition reference
- `#/properties/name` - Property reference
- `#/items/0` - Array item reference

**Resolution**: Navigate JSON structure using pointer segments

### Anchor References

**Format**: `#anchor-name`
**Requirements**: Anchors must be unique within document
**Resolution**: Search schema tree for matching anchor property

### External References

**Format**: `https://example.com/schema.json` or `./relative.json`
**Requirements**: 
- HTTP client for remote schemas
- File system access for local schemas
- Base URI resolution for relative paths

**Caching**: Cache external schemas to avoid repeated network/file access

## Implementation Architecture

### Resolver Factory

**Create Resolvers**: Different resolver implementations for different contexts
```go
func NewResolver(options ...ResolverOption) Resolver
func WithHTTPClient(client *http.Client) ResolverOption
func WithFileSystem(fs http.FileSystem) ResolverOption
func WithCache(cache ResolverCache) ResolverOption
```

### Resolution Pipeline

**Stages**:
1. **Parse Reference**: Determine reference type and extract components
2. **Context Preparation**: Add base schema/URI to context
3. **Resolution Dispatch**: Route to appropriate resolution method
4. **Schema Retrieval**: Fetch schema from cache, file, or network
5. **Pointer Navigation**: Navigate to specific schema location
6. **Result Assignment**: Assign resolved schema to destination

### Error Handling

**Resolution Errors**:
- Reference not found
- Circular reference detected
- Network/file access errors
- Invalid reference format
- Schema parsing errors

**Error Context**: Provide detailed context about resolution failure location

## Advanced Features

### Schema Registration

**Purpose**: Pre-register schemas for resolution without network access
**API**:
```go
resolver.RegisterSchema("https://example.com/user.json", userSchema)
resolver.RegisterSchemaFromFile("./schemas/user.json")
```

### Reference Validation

**Validate References**: Check that references are valid during schema parsing
**Dead Reference Detection**: Identify references that cannot be resolved
**Reference Linting**: Warn about potentially problematic references

### Dynamic Schema Loading

**On-Demand Loading**: Load schemas dynamically during validation
**Schema Discovery**: Automatically discover and load related schemas
**Version Management**: Handle different schema versions and compatibility

## Testing Strategy

### Unit Testing

**Reference Parsing**: Test all reference format parsing
**Resolution Logic**: Test resolution for each reference type
**Error Handling**: Test all error conditions
**Cache Behavior**: Test cache hit/miss scenarios

### Integration Testing

**Circular References**: Test circular reference detection and handling
**External Schemas**: Test network and file system schema loading
**Complex Documents**: Test resolution in large, complex schema documents
**Performance**: Test resolution performance with large schema sets

### Test Data Organization

**Test Schemas**: Organize test schemas with various reference patterns
**Mock Servers**: HTTP test servers for external reference testing
**File System**: Local file system structure for relative reference testing

## Performance Optimization

### Lazy Loading

- Defer schema loading until validation time
- Cache loaded schemas for reuse
- Use weak references to allow garbage collection

### Parallel Resolution

- Resolve independent references in parallel
- Use worker pools for bulk reference resolution
- Implement resolution priority queues

### Memory Management

- Limit cache size to prevent memory exhaustion
- Use LRU eviction for schema cache
- Clean up circular reference tracking data

## Security Considerations

### External Access Control

**URL Validation**: Validate external URLs to prevent malicious requests
**Protocol Restrictions**: Limit allowed protocols (https, file)
**Domain Filtering**: Optionally restrict allowed domains
**Timeout Configuration**: Set reasonable timeouts for external requests

### Schema Validation

**Malicious Schemas**: Validate schemas before caching/using
**Size Limits**: Impose limits on schema document size
**Depth Limits**: Limit resolution depth to prevent resource exhaustion

## File Organization

### Core Implementation

- `resolver/resolver.go` - Core resolver interface and types
- `resolver/json_pointer.go` - JSON Pointer resolution implementation
- `resolver/anchor.go` - Anchor resolution implementation
- `resolver/external.go` - External reference resolution
- `resolver/cache.go` - Resolution caching implementation

### Integration Points

- `validator/compiler.go` - Integration with validator compilation
- `schema/references.go` - Schema reference utilities
- `internal/schemactx/resolver.go` - Context utilities for resolution