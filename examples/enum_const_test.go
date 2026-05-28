package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_enum restricts a value to a fixed set, using the schema.Enum helper.
func Example_enum() {
	built := schema.Enum("red", "green", "blue").MustBuild()

	loaded := loadSchema("testdata/enum.json")
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# member of the set")
	report(schemas, "green")
	fmt.Println("# not a member")
	report(schemas, "purple")
	// Output:
	// # member of the set
	// programmatic valid=true
	// from-json    valid=true
	// # not a member
	// programmatic valid=false
	// from-json    valid=false
}

// Example_const pins a value to a single constant.
func Example_const() {
	built := schema.NewBuilder().Const("v1").MustBuild()

	loaded := loadSchema("testdata/const.json")
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# the constant value")
	report(schemas, "v1")
	fmt.Println("# any other value")
	report(schemas, "v2")
	// Output:
	// # the constant value
	// programmatic valid=true
	// from-json    valid=true
	// # any other value
	// programmatic valid=false
	// from-json    valid=false
}
