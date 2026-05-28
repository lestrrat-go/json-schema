package examples_test

import (
	"bytes"
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// Example_docCodegen compiles a schema and emits Go source that reconstructs the
// validator directly, so production code can skip compilation. This is the
// programmatic form of the `json-schema gen-validator` CLI command.
func Example_docCodegen() {
	s := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(1).
		MustBuild()

	v, err := validator.Compile(context.Background(), s)
	if err != nil {
		fmt.Println("compile failed:", err)
		return
	}

	var buf bytes.Buffer
	if err := validator.NewCodeGenerator().Generate(&buf, v); err != nil {
		fmt.Println("generate failed:", err)
		return
	}
	// Generate emits the raw builder calls (one per line). The gen-validator CLI
	// additionally assigns them to a variable and runs the result through gofmt.
	fmt.Print(buf.String())
	// Output:
	// validator.String().
	// MinLength(1).
	// MustBuild()
}
