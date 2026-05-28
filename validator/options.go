package validator

import (
	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/vocabulary"
	"github.com/lestrrat-go/option/v3"
)

// CompileOption configures Compile.
type CompileOption interface {
	option.Interface
	compileOption()
}

type compileOption struct{ option.Interface }

func (compileOption) compileOption() {}

type identResolver struct{}
type identVocabularySet struct{}
type identBaseURI struct{}
type identBaseSchema struct{}

// WithResolver supplies the $ref resolver used during compilation. When omitted,
// a fresh resolver is created.
func WithResolver(r *schema.Resolver) CompileOption {
	return compileOption{option.New(identResolver{}, r)}
}

// WithVocabularySet gates which keywords are compiled. When omitted, the JSON
// Schema 2020-12 default vocabulary set is used.
func WithVocabularySet(vs *vocabulary.VocabularySet) CompileOption {
	return compileOption{option.New(identVocabularySet{}, vs)}
}

// WithBaseURI sets the base URI the root schema is considered to live at, used
// to resolve relative references.
func WithBaseURI(u string) CompileOption {
	return compileOption{option.New(identBaseURI{}, u)}
}

// WithBaseSchema declares the document that the schema being compiled belongs to.
// Use it when compiling a fragment whose local references (e.g. "#/$defs/...")
// must resolve against a separate root document rather than the fragment itself.
// The supplied schema becomes both the document root and the base resource for
// reference resolution.
func WithBaseSchema(s *schema.Schema) CompileOption {
	return compileOption{option.New(identBaseSchema{}, s)}
}

// ValidateOption configures a Validate call.
type ValidateOption interface {
	option.Interface
	validateOption()
}

type validateOption struct{ option.Interface }

func (validateOption) validateOption() {}

type identDynamicAnchorValidator struct{}

// dynamicAnchorRegistration pairs a $dynamicAnchor name with the validator that
// stands in for the outermost resource declaring it.
type dynamicAnchorRegistration struct {
	name string
	v    Interface
}

// WithDynamicAnchorValidator registers a validator under a $dynamicAnchor name so
// a "$dynamicRef" to that anchor resolves to it. This lets a precompiled
// validator (e.g. the generated meta-schema validator) satisfy a $dynamicRef
// when no schema document is available to resolve against at validation time.
func WithDynamicAnchorValidator(name string, v Interface) ValidateOption {
	return validateOption{option.New(identDynamicAnchorValidator{}, dynamicAnchorRegistration{name: name, v: v})}
}
