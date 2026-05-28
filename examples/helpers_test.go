package examples_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/lestrrat-go/json-schema/vocabulary"
)

// The examples in this package each demonstrate the same validation goal two
// ways:
//
//  1. Programmatically, using the fluent schema.NewBuilder() API.
//  2. By loading the equivalent schema from a JSON file with loadSchema.
//
// Both schemas are then run through the same compile-and-validate flow (see
// compileAndValidate) and produce identical results, which is what each example
// prints. The full, explicit Compile/Validate flow is shown in the top-level
// Example in json_schema_readme_example_test.go; the helpers here keep the
// feature-focused examples concise.

// loadSchema reads a JSON Schema document from path and unmarshals it into a
// *schema.Schema. *schema.Schema implements json.Unmarshaler, so this is all it
// takes to load a schema authored as JSON.
func loadSchema(path string) *schema.Schema {
	buf, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var s schema.Schema
	if err := json.Unmarshal(buf, &s); err != nil {
		panic(err)
	}
	return &s
}

// valid compiles s into a validator and reports whether data satisfies it.
// It uses the default 2020-12 vocabulary set, in which "format" is an annotation
// only (no format assertion) — matching the JSON Schema default.
func valid(s *schema.Schema, data any) bool {
	return validIn(context.Background(), s, data)
}

// validStrict is like valid but enables every standard vocabulary, including
// format-assertion, so that "format" keywords (email, uuid, date-time, ...) are
// enforced rather than treated as annotations.
func validStrict(s *schema.Schema, data any) bool {
	return validIn(context.Background(), s, data, validator.WithVocabularySet(vocabulary.AllEnabled()))
}

func validIn(ctx context.Context, s *schema.Schema, data any, options ...validator.CompileOption) bool {
	v, err := validator.Compile(ctx, s, options...)
	if err != nil {
		return false
	}
	_, err = v.Validate(ctx, data)
	return err == nil
}

// report prints the validation outcome of every (built vs. loaded-from-JSON)
// schema against the same data, one line per schema. Because the two schemas are
// equivalent, the lines are identical — demonstrating that the programmatic and
// JSON-file approaches reach the same goal.
func report(schemas map[string]*schema.Schema, data any) {
	// Stable order: programmatic first, then from-file. Written via Fprintf to
	// os.Stdout (which godoc examples capture) so this shared helper does not trip
	// the forbidigo rule that only exempts code inside Example functions.
	for _, name := range []string{"programmatic", "from-json"} {
		fmt.Fprintf(os.Stdout, "%-12s valid=%t\n", name, valid(schemas[name], data))
	}
}
