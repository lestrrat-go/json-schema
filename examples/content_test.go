package examples_test

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

// Example_content attaches contentEncoding, contentMediaType and contentSchema to
// a string. In JSON Schema 2020-12 these are annotations: they describe how to
// interpret the string but do not, on their own, cause validation to fail. Both
// a well-formed and a malformed payload therefore validate successfully.
func Example_content() {
	built := schema.NewBuilder().
		Types(schema.StringType).
		ContentEncoding("base64").
		ContentMediaType("application/json").
		ContentSchema(schema.NewBuilder().Types(schema.ObjectType).MustBuild()).
		MustBuild()

	loaded := loadSchema("testdata/content.json")
	schemas := map[string]*schema.Schema{"programmatic": built, "from-json": loaded}

	fmt.Println(`# base64 of {"a":1}`)
	report(schemas, "eyJhIjoxfQ==")
	fmt.Println("# malformed base64 (still valid: content is annotation-only)")
	report(schemas, "!!!not-base64!!!")
	// Output:
	// # base64 of {"a":1}
	// programmatic valid=true
	// from-json    valid=true
	// # malformed base64 (still valid: content is annotation-only)
	// programmatic valid=true
	// from-json    valid=true
}
