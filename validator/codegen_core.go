package validator

import (
	"io"

	"github.com/lestrrat-go/codegen"
)

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

// Generate writes Go code that constructs the given validator to the provided Writer
// The output is just the builder chain, e.g.: validator.String().MinLength(5).MaxLength(100)
func (g *codeGenerator) Generate(dst io.Writer, v Interface) error {
	// Generate the code using the internal method
	return g.generateInternal(dst, v)
}

// generateInternal generates code using the provided codegen.Output
func (g *codeGenerator) generateInternal(dst io.Writer, v Interface) error {
	o := codegen.NewOutput(dst) // only placed for compatibility. methods should receive dst instead
	switch validator := v.(type) {
	case *stringValidator:
		return g.generateString(dst, validator)
	case *integerValidator:
		return g.generateInteger(dst, validator)
	case *numberValidator:
		return g.generateNumber(dst, validator)
	case *booleanValidator:
		return g.generateBoolean(dst, validator)
	case *arrayValidator:
		return g.generateArray(dst, validator)
	case *objectValidator:
		return g.generateObject(dst, validator)
	case *allOfValidator:
		return g.generateAllOf(dst, validator)
	case *anyOfValidator:
		return g.generateAnyOf(dst, validator)
	case *oneOfValidator:
		return g.generateOneOf(dst, validator)
	case *EmptyValidator:
		return g.generateEmpty(dst)
	case *NotValidator:
		return g.generateNot(dst, validator)
	case *nullValidator:
		return g.generateNull(dst)
	case *untypedValidator:
		return g.generateUntyped(dst, validator)
	case *alwaysPassValidator:
		return g.generateAlwaysPass(dst)
	case *alwaysFailValidator:
		return g.generateAlwaysFail(dst)
	case *ReferenceValidator:
		return g.generateReference(dst, validator)
	case *DynamicReferenceValidator:
		return g.generateDynamicReference(dst, validator)
	case *contentValidator:
		return g.generateContent(dst, validator)
	case *dependentSchemasValidator:
		return g.generateDependentSchemas(dst, validator)
	case *inferredNumberValidator:
		return g.generateInferredNumber(dst, validator)
	case *unevaluatedPropertiesValidator:
		return g.generateUnevaluatedPropertiesComposition(dst, validator)
	case *AnyOfUnevaluatedPropertiesCompositionValidator:
		return g.generateAnyOfUnevaluatedPropertiesComposition(dst, validator)
	case *OneOfUnevaluatedPropertiesCompositionValidator:
		return g.generateOneOfUnevaluatedPropertiesComposition(dst, validator)
	case *RefUnevaluatedPropertiesCompositionValidator:
		return g.generateRefUnevaluatedPropertiesComposition(dst, validator)
	case *IfThenElseValidator:
		return g.generateIfThenElse(dst, validator)
	case *IfThenElseUnevaluatedPropertiesCompositionValidator:
		return g.generateIfThenElseUnevaluatedPropertiesComposition(dst, validator)
	default:
		// Unsupported validator type, falling back to EmptyValidator
		o.R("&validator.EmptyValidator{}")
		return nil
	}
}