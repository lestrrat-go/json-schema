# Command Line Tool Specification

## Overview

The `json-schema` command line tool provides schema validation, linting, and validator code generation capabilities. Built using github.com/urfave/cli/v3, it offers a comprehensive interface for working with JSON Schema documents.

## Tool Architecture

### Binary Location

**Path**: `cmd/json-schema/main.go`
**Build**: Standard Go build process creates `json-schema` executable
**Installation**: Can be installed via `go install` or distributed as binary

### Framework Choice

**CLI Framework**: github.com/urfave/cli/v3
**Benefits**:
- Mature command line parsing
- Subcommand support with rich help
- Flag parsing and validation
- Consistent help generation

## Command Interface

### Lint Command

**Purpose**: Report formatting and structural errors in schema files
**Usage**: `json-schema lint [filename]`

**Features**:
- JSON Schema syntax validation
- Reference resolution verification
- Structural consistency checking
- Format compliance reporting

**Input Handling**:
- File path: Read schema from specified file
- Stdin: Use "-" as filename to read from stdin
- Multiple files: Support multiple file arguments

**Output Format**:
- Human-readable error reports
- File/line/column information for errors
- Severity levels (error, warning, info)
- Exit codes: 0 for success, non-zero for errors

### Generate Validator Command

**Purpose**: Create pre-compiled validator Go code from schema files
**Usage**: `json-schema gen-validator [options] [filename]`

**Primary Options**:
- `-name [GoVariableName]`: Variable name for generated validator (default: "val")
- `-package [PackageName]`: Go package name for generated code
- `-output [FilePath]`: Output file path (default: stdout)

**Advanced Options**:
- `-optimize`: Enable additional optimizations
- `-imports`: Include necessary import statements
- `-format`: Apply go fmt to generated code

**Output Format**: Bare variable assignment statement (not complete program)
```go
val := validator.String().
    MinLength(10).
    MaxLength(100).
    Pattern("^[a-zA-Z0-9]+$").
    MustBuild()
```

## Implementation Requirements

### CLI Framework Setup

**Main Structure**:
```go
package main

import (
    "github.com/urfave/cli/v3"
    "os"
)

func main() {
    app := &cli.App{
        Name:        "json-schema",
        Usage:       "JSON Schema validation and code generation tool",
        Version:     "1.0.0",
        Description: "A tool for working with JSON Schema documents",
        Commands: []*cli.Command{
            lintCommand(),
            genValidatorCommand(),
        },
    }
    
    app.Run(os.Args)
}
```

### Lint Command Implementation

**Command Structure**:
```go
func lintCommand() *cli.Command {
    return &cli.Command{
        Name:      "lint",
        Usage:     "report formatting errors found in schema file",
        ArgsUsage: "[filename]",
        Description: "Validate JSON Schema documents for syntax and structural errors",
        Action: func(ctx *cli.Context) error {
            filename := ctx.Args().Get(0)
            if filename == "" {
                return cli.Exit("filename required", 1)
            }
            
            return lintSchema(filename)
        },
    }
}
```

**Linting Logic**:
- Parse JSON Schema document
- Validate against meta-schema
- Check reference resolution
- Report structural issues
- Provide actionable error messages

### Gen-Validator Command Implementation

**Command Structure**:
```go
func genValidatorCommand() *cli.Command {
    return &cli.Command{
        Name:      "gen-validator",
        Usage:     "create a pre-compiled validator code from schema file",
        ArgsUsage: "[filename]",
        Description: "Generate Go code that creates optimized validators",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name:    "name",
                Aliases: []string{"n"},
                Value:   "val",
                Usage:   "assign the resulting validator to this variable",
            },
            &cli.StringFlag{
                Name:    "package",
                Aliases: []string{"p"},
                Usage:   "package name for generated code",
            },
            &cli.StringFlag{
                Name:    "output",
                Aliases: []string{"o"},
                Usage:   "output file path (default: stdout)",
            },
        },
        Action: func(ctx *cli.Context) error {
            filename := ctx.Args().Get(0)
            if filename == "" {
                return cli.Exit("filename required", 1)
            }
            
            return generateValidator(ctx, filename)
        },
    }
}
```

## Input/Output Handling

### File Input Processing

**File Reading**: Support both file paths and stdin input
```go
func readSchema(filename string) (*schema.Schema, error) {
    var reader io.Reader
    
    if filename == "-" {
        reader = os.Stdin
    } else {
        file, err := os.Open(filename)
        if err != nil {
            return nil, err
        }
        defer file.Close()
        reader = file
    }
    
    data, err := io.ReadAll(reader)
    if err != nil {
        return nil, err
    }
    
    return schema.ParseJSON(data)
}
```

### Output Formatting

**Code Generation Output**:
- Clean, formatted Go code
- Proper indentation and spacing
- Minimal, necessary imports
- Comments for complex validators

**Lint Output**:
- Structured error reporting
- Color-coded severity levels (if terminal supports)
- File location information
- Suggested fixes where applicable

## Error Handling

### Graceful Error Reporting

**User-Friendly Messages**: Clear, actionable error messages
**Exit Codes**: Proper exit codes for different error types
**Validation Errors**: Detailed validation failure information

**Error Categories**:
- File not found/access errors
- JSON parsing errors  
- Schema validation errors
- Code generation errors
- Internal tool errors

### Logging and Debugging

**Verbose Mode**: Optional verbose output for debugging
**Debug Logging**: Internal operation logging for troubleshooting
**Progress Indicators**: Progress reporting for large operations

## Advanced Features

### Batch Processing

**Multiple Files**: Process multiple schema files in single command
**Directory Processing**: Recursively process schema files in directories
**Glob Patterns**: Support file glob patterns for batch operations

### Configuration Files

**Config File Support**: Optional configuration file for default settings
**Profile Support**: Named configuration profiles for different use cases
**Environment Variables**: Environment variable override support

### Integration Features

**CI/CD Integration**: Exit codes and output formats suitable for CI/CD
**Editor Integration**: JSON output format for editor integration
**Build Tool Integration**: Integration with Go build tools and generators

## Testing Requirements

### Command Testing

**CLI Testing**: Test command line parsing and flag handling
**Integration Testing**: Test complete command workflows
**Error Testing**: Test error conditions and edge cases

**Test Structure**:
```go
func TestLintCommand(t *testing.T) {
    tests := []struct {
        name           string
        input          string
        expectedError  bool
        expectedOutput string
    }{
        // Test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Output Validation

**Generated Code Testing**: Ensure generated code compiles and works
**Lint Output Testing**: Verify lint reports are accurate
**Performance Testing**: Test performance with large schema files

## Documentation

### Help System

**Command Help**: Comprehensive help for each command
**Usage Examples**: Clear examples for common use cases
**Flag Documentation**: Detailed flag descriptions

### Man Page Generation

**Manual Pages**: Generate man pages for system installation
**Completion Scripts**: Shell completion scripts for bash/zsh
**Integration Docs**: Documentation for CI/CD and build tool integration

## Distribution

### Binary Distribution

**Multi-Platform Builds**: Binaries for multiple OS/architecture combinations
**Release Automation**: Automated binary builds and releases
**Package Management**: Integration with package managers (brew, apt, etc.)

### Installation Methods

**Go Install**: `go install github.com/lestrrat-go/json-schema/cmd/json-schema@latest`
**Binary Download**: Direct binary download from releases
**Container Images**: Docker images with tool pre-installed

## Performance Considerations

### Large File Handling

**Memory Efficiency**: Handle large schema files without excessive memory usage
**Streaming Processing**: Stream processing for very large schemas
**Parallel Processing**: Parallel processing for multiple files

### Code Generation Optimization

**Fast Generation**: Minimize code generation time
**Optimized Output**: Generate efficient validator code
**Caching**: Cache compiled schemas for repeated operations