package examples_test

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// Example_validateJSON validates raw JSON text directly with
// validator.ValidateJSON, skipping a manual json.Unmarshal step. Numbers are
// decoded as json.Number, so a 64-bit identifier larger than 2^53 is validated
// exactly rather than being rounded by float64.
func Example_validateJSON() {
	s := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("id", schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
		Property("role", schema.Enum("admin", "user").MustBuild()).
		Required("id", "role").
		MustBuild()

	ctx := context.Background()
	v, err := validator.Compile(ctx, s)
	if err != nil {
		fmt.Println("compile failed:", err)
		return
	}

	for _, data := range [][]byte{
		[]byte(`{"id": 9007199254740993, "role": "admin"}`), // large id, valid
		[]byte(`{"id": 1, "role": "root"}`),                 // role not in enum
		[]byte(`{"role": "admin"}`),                         // missing required id
	} {
		_, err := validator.ValidateJSON(ctx, v, data)
		fmt.Printf("valid=%t\n", err == nil)
	}
	// Output:
	// valid=true
	// valid=false
	// valid=false
}
