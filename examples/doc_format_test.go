package examples_test

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/lestrrat-go/json-schema/vocabulary"
)

// Example_docFormatAssertion shows that "format" is an annotation by default (it
// does not reject bad values) and only asserts when the format-assertion
// vocabulary is enabled via vocabulary.AllEnabled.
func Example_docFormatAssertion() {
	s := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("email", schema.Email().MustBuild()).
		MustBuild()

	data := map[string]any{"email": "not-an-email"}

	// Default vocabularies: format is annotation-only, so this passes.
	def := context.Background()
	v, _ := validator.Compile(def, s)
	_, err := v.Validate(def, data)
	fmt.Println("default set: valid:", err == nil)

	// All vocabularies enabled: format-assertion is on, so this fails.
	strict := context.Background()
	vs, _ := validator.Compile(strict, s, validator.WithVocabularySet(vocabulary.AllEnabled()))
	_, err = vs.Validate(strict, data)
	fmt.Println("all enabled: valid:", err == nil)
	// Output:
	// default set: valid: true
	// all enabled: valid: false
}
