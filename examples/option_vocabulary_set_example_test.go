package examples_test

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/lestrrat-go/json-schema/vocabulary"
)

// Example_option_vocabulary_set shows how validator.WithVocabularySet selects which
// vocabularies are enforced. By default "format" is annotation-only, so a
// malformed value still validates; enabling all vocabularies turns on
// format-assertion, making "format" an enforced constraint.
func Example_option_vocabulary_set() {
	s := schema.Email().MustBuild() // type: string, format: email

	const malformed = "not-an-email"

	def, err := validator.Compile(context.Background(), s)
	if err != nil {
		panic(err)
	}
	strict, err := validator.Compile(context.Background(), s,
		validator.WithVocabularySet(vocabulary.AllEnabled()))
	if err != nil {
		panic(err)
	}

	_, defErr := def.Validate(context.Background(), malformed)
	_, strictErr := strict.Validate(context.Background(), malformed)

	fmt.Printf("default (annotation) valid=%t\n", defErr == nil)
	fmt.Printf("strict  (assertion)  valid=%t\n", strictErr == nil)
	// Output:
	// default (annotation) valid=true
	// strict  (assertion)  valid=false
}
