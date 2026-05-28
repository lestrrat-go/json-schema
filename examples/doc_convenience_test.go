package examples_test

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// Example_docConvenience composes several one-line convenience constructors —
// PositiveInteger, Optional, NonEmptyString and Enum — into one object schema.
// Optional(s) accepts either s or null.
func Example_docConvenience() {
	s := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("id", schema.PositiveInteger().MustBuild()).
		Property("nickname", schema.Optional(schema.NonEmptyString().MustBuild()).MustBuild()).
		Property("role", schema.Enum("admin", "user").MustBuild()).
		Required("id", "role").
		MustBuild()

	ctx := context.Background()
	v, _ := validator.Compile(ctx, s)
	check := func(data map[string]any) bool {
		_, err := v.Validate(ctx, data)
		return err == nil
	}

	fmt.Println("admin:        ", check(map[string]any{"id": 1, "role": "admin"}))
	fmt.Println("null nickname:", check(map[string]any{"id": 1, "role": "user", "nickname": nil}))
	fmt.Println("bad role:     ", check(map[string]any{"id": 1, "role": "root"}))
	// Output:
	// admin:         true
	// null nickname: true
	// bad role:      false
}
