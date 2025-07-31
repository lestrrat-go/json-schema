# Code Generation Specification

## Overview

The code generation system optimizes performance by generating Go code that directly creates compiled validators, eliminating the "build schema → compile validator" pipeline for production use. This system uses direct field access to generate equivalent builder code from existing validators.

## Architecture Approach

### Direct Field Access Strategy

**Philosophy**: Access validator struct fields directly to generate equivalent builder code without intermediate data structures.

**Benefits**:
- No intermediate map creation or copying
- Type-safe compile-time field access
- Simpler implementation with fewer abstractions
- Better performance with fewer allocations
- Easier maintenance with clear generation logic

### Code Generation Pipeline

**Input**: Compiled validator object
**Output**: Go code that creates equivalent validator using builder API
**Process**: Direct field access → Go code generation → Formatted output

```go
// Input: Compiled validator
validator := String().MinLength(5).Pattern("^[a-z]+$").MustBuild()

// Output: Generated Go code
func NewEmailValidator() validator.Interface {
    return validator.String().
        MinLength(5).
        Pattern("^[a-z]+$").
        MustBuild()
}
```

## Core Interface

### CodeGenerator Interface

```go
package validator

type CodeGenerator interface {
    GenerateCode(validatorName string, v Interface) (string, error)
    GeneratePackage(packageName string, validators map[string]Interface) (string, error)
}

func NewCodeGenerator() CodeGenerator
```

### Generation Methods

**Type Switch Pattern**: Handle each validator type with dedicated generation method
```go
func (g *codeGenerator) GenerateCode(validatorName string, v Interface) (string, error) {
    switch validator := v.(type) {
    case *stringValidator:
        return g.generateStringValidator(validatorName, validator)
    case *integerValidator:
        return g.generateIntegerValidator(validatorName, validator)
    case *arrayValidator:
        return g.generateArrayValidator(validatorName, validator)
    case *allOfValidator:
        return g.generateAllOf(validatorName, validator)
    // ... all validator types
    default:
        return "", fmt.Errorf("unsupported validator type: %T", v)
    }
}
```

## Implementation Requirements

### Complete Validator Support

**Required Generation Methods** (must implement ALL):
- `generateStringValidator()` - String constraints and format validation
- `generateIntegerValidator()` - Numeric constraints for integers
- `generateNumberValidator()` - Numeric constraints for floats
- `generateBooleanValidator()` - Boolean type validation
- `generateArrayValidator()` - Array/slice validation with complex logic
- `generateObjectValidator()` - Object/map validation with properties
- `generateAllOf()` - All-of composite validator
- `generateAnyOf()` - Any-of composite validator  
- `generateOneOf()` - One-of composite validator
- `generateEmptyValidator()` - Always-pass validator
- `generateNotValidator()` - Negation wrapper validator
- `generateNullValidator()` - Null type validator
- `generateEnumValidator()` - Enumeration constraint validator
- `generateConstValidator()` - Constant value validator
- `generateReferenceValidator()` - Reference resolution validator
- `generateContentValidator()` - Content encoding/media type validator
- `generateDependentSchemasValidator()` - Object dependent schemas

### Field Access Patterns

**String Validator Example**:
```go
func (g *codeGenerator) generateStringValidator(name string, v *stringValidator) (string, error) {
    var parts []string
    
    // Direct field access - no intermediate map needed
    if v.minLength != nil {
        parts = append(parts, fmt.Sprintf("MinLength(%d)", *v.minLength))
    }
    if v.maxLength != nil {
        parts = append(parts, fmt.Sprintf("MaxLength(%d)", *v.maxLength))
    }
    if v.pattern != nil {
        parts = append(parts, fmt.Sprintf("Pattern(%q)", v.pattern.String()))
    }
    if v.format != nil {
        parts = append(parts, fmt.Sprintf("Format(%q)", *v.format))
    }
    if v.enum != nil {
        enumStr := formatStringSlice(v.enum)
        parts = append(parts, fmt.Sprintf("Enum(%s)", enumStr))
    }
    
    builderCalls := strings.Join(parts, ".")
    if builderCalls != "" {
        builderCalls = "." + builderCalls
    }
    
    return fmt.Sprintf(`func New%s() validator.Interface {
    return validator.String()%s.MustBuild()
}`, name, builderCalls), nil
}
```

### Complex Validator Handling

**Composite Validators**: Handle validators with child validators
```go
func (g *codeGenerator) generateAllOf(name string, v *allOfValidator) (string, error) {
    var childDefs []string
    var childVars []string
    
    // Generate code for child validators recursively
    for i, child := range v.validators {
        childVar := fmt.Sprintf("child%d", i)
        childCode, err := g.GenerateCode("", child)
        if err != nil {
            return "", fmt.Errorf("failed to generate child validator %d: %w", i, err)
        }
        
        creation := extractValidatorCreation(childCode)
        childDefs = append(childDefs, fmt.Sprintf("    %s := %s", childVar, creation))
        childVars = append(childVars, childVar)
    }
    
    template := `func New%s() validator.Interface {
%s
    
    return validator.AllOf(
%s
    )
}`
    
    childDefsStr := strings.Join(childDefs, "\n")
    childArgsStr := formatChildArgs(childVars)
    
    return fmt.Sprintf(template, name, childDefsStr, childArgsStr), nil
}
```

### Error Handling Requirements

**Comprehensive Error Coverage**:
- Unsupported validator types in type switch
- Recursive generation failures for child validators
- Invalid field values that can't be formatted  
- Generated code that fails Go formatting validation

**Error Context**: Provide detailed error messages with validator type and field information

## Template System

### Go Template Integration

**Template-Based Generation**: Use Go templates for clean, readable generated code
```go
const validatorTemplate = `
func New{{.Name}}() validator.Interface {
    {{- if .IsSimple}}
    return validator.{{.Type}}().
        {{- range .Constraints}}
        {{.Method}}({{.Value}}).
        {{- end}}
        MustBuild()
    {{- else}}
    {{- range .Children}}
    {{.VarName}} := {{.GenerateCode}}
    {{- end}}
    
    return validator.{{.Type}}().
        {{- range .Children}}
        Add({{.VarName}}).
        {{- end}}
        MustBuild()
    {{- end}}
}
`
```

### Template Data Structures

**Generation Context**: Structured data for template execution
```go
type GenerationContext struct {
    Name        string
    Type        string
    IsSimple    bool
    Constraints []Constraint
    Children    []ChildValidator
}

type Constraint struct {
    Method string
    Value  string
}

type ChildValidator struct {
    VarName      string
    GenerateCode string
}
```

## Output Optimization

### Import Management

**Minimal Imports**: Generate only necessary imports
**Standard Packages**: regexp, context, etc.
**Project Packages**: validator package imports
**Import Deduplication**: Remove duplicate imports

### Code Formatting

**Go Format Integration**: Use `go/format` for clean output
**Package-Level Variables**: Generate regex patterns as package variables
**Function Organization**: Logical grouping of generated functions

### Pre-compiled Optimizations

**Regex Pre-compilation**: Generate package-level regex variables
```go
var emailPatternRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func NewEmailValidator() validator.Interface {
    return validator.String().
        Pattern(emailPatternRegex.String()).
        Format("email").
        MustBuild()
}
```

**Shared Validators**: Detect and reuse identical sub-validators

## Usage Integration

### Build Tool Integration

**go generate**: Integration with Go generate command
```go
//go:generate json-schema gen-validator -name UserValidator user-schema.json
```

**CI/CD Integration**: Generate validators during build process
**Deployment Optimization**: Use generated validators in production

### Testing Requirements

**Generated Code Validation**: 
- Generated code must compile without errors
- Generated validators must produce identical results to original
- Performance comparison between generated and compiled validators

**Comprehensive Test Coverage**:
```go
func TestCodeGeneration(t *testing.T) {
    // Create complex validator
    original := createComplexValidator()
    
    // Generate code
    generator := validator.NewCodeGenerator()
    code, err := generator.GenerateCode("TestValidator", original)
    require.NoError(t, err)
    
    // Verify generated code compiles and behaves identically
    // (requires runtime compilation testing)
}
```

## File Organization

### Core Implementation

- `validator/codegen.go` - Core interfaces and types (minimal)
- `validator/generator.go` - Main generator implementation with direct field access
- `validator/codegen_test.go` - Comprehensive testing of generation

### Generated Code Structure

- Package declaration and imports
- Pre-compiled regex patterns and constants
- Validator constructor functions
- Helper functions for complex types

## Performance Requirements

### Generation Performance

**Fast Generation**: Minimize code generation time
**Memory Efficiency**: Avoid large intermediate data structures  
**Parallel Generation**: Support concurrent validator generation

### Runtime Performance

**Generated Code Performance**: Generated validators should match or exceed compiled validator performance
**Startup Time**: Faster application startup by avoiding compilation phase
**Memory Usage**: Lower memory footprint by eliminating schema objects

## Meta Schema Integration

### Meta Schema Validator Generation

**Purpose**: Generate high-performance JSON Schema meta-schema validator
**Location**: `meta` package with generated validator
**Usage**: `meta.Validate(ctx, schemaData)` for schema validation

**Implementation**: Generate optimized validator for JSON Schema 2020-12 meta-schema using code generation system

### CLI Integration

**json-schema gen-validator**: Command-line tool for validator generation
**Input**: JSON Schema files or stdin
**Output**: Go code with validator constructor functions
**Options**: Validator name, package name, output format