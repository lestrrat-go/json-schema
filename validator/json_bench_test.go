package validator_test

import (
	"context"
	"encoding/json"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// benchData is a representative object with nested structure and numbers,
// exercising both the object/array traversal and the numeric validators.
var benchData = []byte(`{
	"id": 9007199254740993,
	"role": "admin",
	"score": 87.5,
	"tags": ["a", "b", "c"],
	"profile": {"age": 42, "active": true}
}`)

func benchSchema() *schema.Schema {
	return schema.NewBuilder().
		Types(schema.ObjectType).
		Property("id", schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
		Property("role", schema.Enum("admin", "user").MustBuild()).
		Property("score", schema.NewBuilder().Types(schema.NumberType).MustBuild()).
		Property("tags", schema.NewBuilder().
			Types(schema.ArrayType).
			Items(schema.NewBuilder().Types(schema.StringType).MustBuild()).
			MustBuild()).
		Property("profile", schema.NewBuilder().
			Types(schema.ObjectType).
			Property("age", schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
			Property("active", schema.NewBuilder().Types(schema.BooleanType).MustBuild()).
			MustBuild()).
		Required("id", "role").
		MustBuild()
}

// BenchmarkValidate_UnmarshalThenValidate measures the "decode to a Go object
// first" path: json.Unmarshal into any, then Validate. The unmarshal is inside
// the loop because it is part of the cost being compared.
func BenchmarkValidate_UnmarshalThenValidate(b *testing.B) {
	ctx := context.Background()
	v, err := validator.Compile(ctx, benchSchema())
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		var decoded any
		if err := json.Unmarshal(benchData, &decoded); err != nil {
			b.Fatal(err)
		}
		if _, err := v.Validate(ctx, decoded); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidate_ValidateJSON measures the direct raw-bytes path
// (UseNumber decode + validate) via ValidateJSON.
func BenchmarkValidate_ValidateJSON(b *testing.B) {
	ctx := context.Background()
	v, err := validator.Compile(ctx, benchSchema())
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := validator.ValidateJSON(ctx, v, benchData); err != nil {
			b.Fatal(err)
		}
	}
}
