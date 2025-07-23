package validator

import "io"

// CodeGenerator generates Go code that creates equivalent validators
type CodeGenerator interface {
	// Generate writes Go code that constructs the given validator to the provided Writer
	// The output is just the builder chain, e.g.: validator.String().MinLength(5).MaxLength(100)
	Generate(dst io.Writer, v Interface) error
}

// codeGenerator implements the CodeGenerator interface
type codeGenerator struct{}

// NewCodeGenerator creates a new code generator
func NewCodeGenerator() CodeGenerator {
	return &codeGenerator{}
}
