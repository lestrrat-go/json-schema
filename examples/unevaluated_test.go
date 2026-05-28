package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_propertyNames constrains the names of an object's properties (rather
// than their values) with a subschema.
func Example_propertyNames() {
	built := schema.NewBuilder().
		Types(schema.ObjectType).
		PropertyNames(schema.NewBuilder().Pattern("^[a-z]+$").MustBuild()).
		MustBuild()

	loaded := loadSchema("testdata/property_names.json")
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# all names lowercase")
	report(schemas, map[string]any{"foo": 1, "bar": 2})
	fmt.Println("# a name with an uppercase letter")
	report(schemas, map[string]any{"Foo": 1})
	// Output:
	// # all names lowercase
	// programmatic valid=true
	// from-json    valid=true
	// # a name with an uppercase letter
	// programmatic valid=false
	// from-json    valid=false
}

// Example_unevaluatedProperties rejects properties that no other keyword has
// already evaluated. The "name" property is evaluated by the allOf branch, so it
// is allowed; any other property is "unevaluated" and rejected.
func Example_unevaluatedProperties() {
	base := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
		MustBuild()

	built := schema.NewBuilder().
		AllOf(base).
		UnevaluatedProperties(schema.FalseSchema()).
		MustBuild()

	loaded := loadSchema("testdata/unevaluated_properties.json")
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# only the evaluated property")
	report(schemas, map[string]any{"name": "Ada"})
	fmt.Println("# an unevaluated property is present")
	report(schemas, map[string]any{"name": "Ada", "extra": true})
	// Output:
	// # only the evaluated property
	// programmatic valid=true
	// from-json    valid=true
	// # an unevaluated property is present
	// programmatic valid=false
	// from-json    valid=false
}

// Example_unevaluatedItems is the array counterpart of unevaluatedProperties:
// the prefixItems branch evaluates index 0, so any further element is
// "unevaluated" and rejected.
func Example_unevaluatedItems() {
	base := schema.NewBuilder().
		Types(schema.ArrayType).
		PrefixItems(schema.NewBuilder().Types(schema.StringType).MustBuild()).
		MustBuild()

	built := schema.NewBuilder().
		AllOf(base).
		UnevaluatedItems(schema.FalseSchema()).
		MustBuild()

	loaded := loadSchema("testdata/unevaluated_items.json")
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println("# only the evaluated element")
	report(schemas, []any{"first"})
	fmt.Println("# an unevaluated element is present")
	report(schemas, []any{"first", 2})
	// Output:
	// # only the evaluated element
	// programmatic valid=true
	// from-json    valid=true
	// # an unevaluated element is present
	// programmatic valid=false
	// from-json    valid=false
}
