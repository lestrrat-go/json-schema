package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_integerConstraints validates an integer against inclusive bounds and a
// multipleOf divisor.
func Example_integerConstraints() {
	built := schema.NewBuilder().
		Types(schema.IntegerType).
		Minimum(0).
		Maximum(100).
		MultipleOf(5).
		MustBuild()

	loaded := loadSchema("testdata/numeric_integer.json")
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# 25 (in range, multiple of 5)")
	report(schemas, 25)
	fmt.Println("# 7 (not a multiple of 5)")
	report(schemas, 7)
	// Output:
	// # 25 (in range, multiple of 5)
	// programmatic valid=true
	// from-json    valid=true
	// # 7 (not a multiple of 5)
	// programmatic valid=false
	// from-json    valid=false
}

// Example_numberExclusiveBounds validates a number against exclusive bounds: the
// value must be strictly between 0 and 1.
func Example_numberExclusiveBounds() {
	built := schema.NewBuilder().
		Types(schema.NumberType).
		ExclusiveMinimum(0).
		ExclusiveMaximum(1).
		MustBuild()

	loaded := loadSchema("testdata/numeric_number.json")
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# 0.5 (strictly between 0 and 1)")
	report(schemas, 0.5)
	fmt.Println("# 0 (not strictly greater than 0)")
	report(schemas, 0)
	// Output:
	// # 0.5 (strictly between 0 and 1)
	// programmatic valid=true
	// from-json    valid=true
	// # 0 (not strictly greater than 0)
	// programmatic valid=false
	// from-json    valid=false
}
