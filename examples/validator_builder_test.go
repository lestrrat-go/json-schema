package examples_test

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/json-schema/validator"
)

// Example_validatorBuilder constructs a validator directly with the validator
// package's builders, instead of compiling a *schema.Schema. This is the
// lower-level path: useful when the constraints are known in Go and there is no
// schema document to load. (It has no JSON form — that is exactly what loading a
// schema and calling validator.Compile is for.)
func Example_validatorBuilder() {
	v := validator.String().
		MinLength(3).
		MaxLength(10).
		Pattern("^[a-z]+$").
		MustBuild()

	ctx := context.Background()
	_, errGood := v.Validate(ctx, "hello")
	_, errBad := v.Validate(ctx, "Hi") // too short and not all lowercase

	fmt.Println("hello valid:", errGood == nil)
	fmt.Println("Hi valid:   ", errBad == nil)
	// Output:
	// hello valid: true
	// Hi valid:    false
}
