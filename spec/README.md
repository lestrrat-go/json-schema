# JSON Schema Library Specification

This directory contains high-level architectural specifications for building a Go JSON Schema library. Each file contains implementation guidelines that future AI coding agents can follow to recreate or maintain this project.

## File Organization

- **[architecture.md](architecture.md)** - Core architectural principles and design patterns
- **[schema-system.md](schema-system.md)** - Schema objects, builders, and data structures  
- **[validation-system.md](validation-system.md)** - Validator interfaces, compilation, and execution
- **[context-management.md](context-management.md)** - Context passing patterns and internal utilities
- **[reference-resolution.md](reference-resolution.md)** - JSON reference and anchor resolution systems
- **[code-generation.md](code-generation.md)** - Performance optimization through code generation
- **[testing-strategy.md](testing-strategy.md)** - Testing approaches and meta-schema validation
- **[command-line-tool.md](command-line-tool.md)** - CLI tool specification and requirements

## Implementation Principles

All specifications follow these principles for AI agents:

1. **High-level Instructions**: Focus on "what to build" rather than exact implementation details
2. **Clear Separation of Concerns**: Each system has distinct responsibilities
3. **Performance Considerations**: Explicit guidance on optimization opportunities
4. **Extensibility**: Design patterns that support future JSON Schema evolution
5. **Type Safety**: Leverage Go's type system for compile-time correctness

## Getting Started

For new implementations, follow this order:
1. Read [architecture.md](architecture.md) for foundational principles
2. Implement the schema system from [schema-system.md](schema-system.md)
3. Build the validation system from [validation-system.md](validation-system.md)
4. Add context management from [context-management.md](context-management.md)
5. Implement reference resolution from [reference-resolution.md](reference-resolution.md)
6. Optimize with code generation from [code-generation.md](code-generation.md)
7. Add comprehensive testing per [testing-strategy.md](testing-strategy.md)
8. Build CLI tooling from [command-line-tool.md](command-line-tool.md)