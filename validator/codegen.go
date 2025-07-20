package validator

import "io"

// CodeGenerator generates Go code that creates equivalent validators
type CodeGenerator interface {
	// Generate writes Go code that constructs the given validator to the provided Writer
	// The output is just the builder chain, e.g.: validator.String().MinLength(5).MaxLength(100)
	Generate(dst io.Writer, v Interface) error
}

// CodeGenOption configures code generation behavior
type CodeGenOption func(*codeGenConfig)

// codeGenConfig holds configuration for code generation
type codeGenConfig struct {
	packageImports   []string
	validatorPrefix  string
	includeComments  bool
}

// WithPackageImports specifies additional imports for generated code
func WithPackageImports(imports ...string) CodeGenOption {
	return func(config *codeGenConfig) {
		config.packageImports = append(config.packageImports, imports...)
	}
}

// WithValidatorPrefix sets a prefix for generated validator variable names
func WithValidatorPrefix(prefix string) CodeGenOption {
	return func(config *codeGenConfig) {
		config.validatorPrefix = prefix
	}
}

// WithIncludeComments controls whether to include comments in generated code
func WithIncludeComments(include bool) CodeGenOption {
	return func(config *codeGenConfig) {
		config.includeComments = include
	}
}

// codeGenerator implements the CodeGenerator interface
type codeGenerator struct {
	config codeGenConfig
}

// NewCodeGenerator creates a new code generator with default settings
func NewCodeGenerator(opts ...CodeGenOption) CodeGenerator {
	config := codeGenConfig{
		validatorPrefix: "",
		includeComments: true,
	}
	for _, opt := range opts {
		opt(&config)
	}
	return &codeGenerator{config: config}
}