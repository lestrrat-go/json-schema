package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_ifThenElse applies one subschema when the "if" subschema matches and
// another when it does not. Here: integers must be non-negative, and anything
// that is not an integer must be a string.
func Example_ifThenElse() {
	built := schema.NewBuilder().
		IfSchema(schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
		ThenSchema(schema.NewBuilder().Minimum(0).MustBuild()).
		ElseSchema(schema.NewBuilder().Types(schema.StringType).MustBuild()).
		MustBuild()

	loaded := loadSchema("testdata/conditional.json")
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# 5 (integer branch: non-negative)")
	report(schemas, 5)
	fmt.Println("# -3 (integer branch: negative, fails then)")
	report(schemas, -3)
	fmt.Println("# \"hi\" (else branch: a string)")
	report(schemas, "hi")
	// Output:
	// # 5 (integer branch: non-negative)
	// programmatic valid=true
	// from-json    valid=true
	// # -3 (integer branch: negative, fails then)
	// programmatic valid=false
	// from-json    valid=false
	// # "hi" (else branch: a string)
	// programmatic valid=true
	// from-json    valid=true
}
