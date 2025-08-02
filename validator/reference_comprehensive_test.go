package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestReferenceResolutionComprehensive(t *testing.T) {
	t.Run("nested references", func(t *testing.T) {
		// Schema with references that point to other references
		addressSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("street", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("city", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Required("street", "city").
			MustBuild()

		departmentSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("budget", schema.NewBuilder().Types(schema.NumberType).Minimum(0).MustBuild()).
			Required("name").
			MustBuild()

		personSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("address", schema.NewBuilder().Reference("#/$defs/address").MustBuild()).
			Required("name").
			MustBuild()

		employeeExtraSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("employeeId", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("department", schema.NewBuilder().Reference("#/$defs/department").MustBuild()).
			Required("employeeId").
			MustBuild()

		employeeSchema := schema.NewBuilder().
			AllOf(
				schema.NewBuilder().Reference("#/$defs/person").MustBuild(),
				employeeExtraSchema,
			).
			MustBuild()

		s := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("user", schema.NewBuilder().Reference("#/$defs/person").MustBuild()).
			Property("manager", schema.NewBuilder().Reference("#/$defs/employee").MustBuild()).
			Definitions("person", personSchema).
			Definitions("employee", employeeSchema).
			Definitions("address", addressSchema).
			Definitions("department", departmentSchema).
			MustBuild()

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, s)

		v, err := validator.Compile(ctx, s)
		require.NoError(t, err)

		// Valid data
		validData := map[string]any{
			"user": map[string]any{
				"name": "John Doe",
				"address": map[string]any{
					"street": "123 Main St",
					"city":   "Anytown",
				},
			},
			"manager": map[string]any{
				"name":       "Jane Smith",
				"employeeId": "EMP123",
				"address": map[string]any{
					"street": "456 Oak Ave",
					"city":   "Somewhere",
				},
				"department": map[string]any{
					"name":   "Engineering",
					"budget": 50000.0,
				},
			},
		}

		_, err = v.Validate(context.Background(), validData)
		require.NoError(t, err)

		// Invalid data - missing required field in nested reference
		invalidData := map[string]any{
			"user": map[string]any{
				"name": "John Doe",
				"address": map[string]any{
					"street": "123 Main St",
					// missing "city"
				},
			},
		}

		_, err = v.Validate(context.Background(), invalidData)
		require.Error(t, err)
	})

	t.Run("references in arrays and objects", func(t *testing.T) {
		// Schema with references in various contexts
		itemSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("id", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("value", schema.NewBuilder().Types(schema.NumberType).MustBuild()).
			Required("id").
			MustBuild()

		metaValueSchema := schema.NewBuilder().
			OneOf(
				schema.NewBuilder().Types(schema.StringType).MustBuild(),
				schema.NewBuilder().Types(schema.NumberType).MustBuild(),
				schema.NewBuilder().Types(schema.BooleanType).MustBuild(),
			).
			MustBuild()

		basicConfigSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("type", schema.NewBuilder().Const("basic").MustBuild()).
			Property("enabled", schema.NewBuilder().Types(schema.BooleanType).MustBuild()).
			Required("type").
			MustBuild()

		advancedConfigSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("type", schema.NewBuilder().Const("advanced").MustBuild()).
			Property("features", schema.NewBuilder().
				Types(schema.ArrayType).
				Items(schema.NewBuilder().Types(schema.StringType).MustBuild()).
				MustBuild()).
			Required("type", "features").
			MustBuild()

		s := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("items", schema.NewBuilder().
				Types(schema.ArrayType).
				Items(schema.NewBuilder().Reference("#/$defs/item").MustBuild()).
				MustBuild()).
			Property("metadata", schema.NewBuilder().
				Types(schema.ObjectType).
				PatternProperty("^meta_", schema.NewBuilder().Reference("#/$defs/metaValue").MustBuild()).
				MustBuild()).
			Property("config", schema.NewBuilder().
				OneOf(
					schema.NewBuilder().Reference("#/$defs/basicConfig").MustBuild(),
					schema.NewBuilder().Reference("#/$defs/advancedConfig").MustBuild(),
				).
				MustBuild()).
			Definitions("item", itemSchema).
			Definitions("metaValue", metaValueSchema).
			Definitions("basicConfig", basicConfigSchema).
			Definitions("advancedConfig", advancedConfigSchema).
			MustBuild()

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, s)

		v, err := validator.Compile(ctx, s)
		require.NoError(t, err)

		// Valid data
		validData := map[string]any{
			"items": []any{
				map[string]any{"id": "item1", "value": 10.5},
				map[string]any{"id": "item2", "value": 20.3},
			},
			"metadata": map[string]any{
				"meta_string": "hello",
				"meta_number": 42,
				"meta_bool":   true,
			},
			"config": map[string]any{
				"type":     "advanced",
				"features": []any{"feature1", "feature2"},
			},
		}

		_, err = v.Validate(context.Background(), validData)
		require.NoError(t, err)

		// Invalid data - array item missing required field
		invalidData := map[string]any{
			"items": []any{
				map[string]any{"value": 10.5}, // missing "id"
			},
		}

		_, err = v.Validate(context.Background(), invalidData)
		require.Error(t, err)
	})

	t.Run("deep reference chains", func(t *testing.T) {
		// Schema with multiple levels of reference indirection
		actualSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("value", schema.NewBuilder().Types(schema.StringType).MinLength(1).MustBuild()).
			Required("value").
			MustBuild()

		level3Schema := schema.NewBuilder().Reference("#/$defs/actual").MustBuild()
		level2Schema := schema.NewBuilder().Reference("#/$defs/level3").MustBuild()
		level1Schema := schema.NewBuilder().Reference("#/$defs/level2").MustBuild()
		rootSchema := schema.NewBuilder().Reference("#/$defs/level1").MustBuild()

		s := schema.NewBuilder().
			Reference("#/$defs/root").
			Definitions("root", rootSchema).
			Definitions("level1", level1Schema).
			Definitions("level2", level2Schema).
			Definitions("level3", level3Schema).
			Definitions("actual", actualSchema).
			MustBuild()

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, s)

		v, err := validator.Compile(ctx, s)
		require.NoError(t, err)

		// Valid data
		validData := map[string]any{
			"value": "test",
		}

		_, err = v.Validate(context.Background(), validData)
		require.NoError(t, err)

		// Invalid data
		invalidData := map[string]any{
			"value": "", // violates minLength: 1
		}

		_, err = v.Validate(context.Background(), invalidData)
		require.Error(t, err)
	})

	t.Run("self-referencing schema", func(t *testing.T) {
		// Schema that references itself directly
		selfSchema := schema.NewBuilder().Reference("#/$defs/self").MustBuild()

		s := schema.NewBuilder().
			Reference("#/$defs/self").
			Definitions("self", selfSchema).
			MustBuild()

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, s)

		_, err := validator.Compile(ctx, s)
		require.Error(t, err)
		require.Contains(t, err.Error(), "circular reference")
	})

	t.Run("references with complex composition", func(t *testing.T) {
		// Schema combining references with allOf, anyOf, oneOf
		baseSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("id", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("timestamp", schema.NewBuilder().Types(schema.NumberType).MustBuild()).
			Required("id").
			MustBuild()

		typeASchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("type", schema.NewBuilder().Const("A").MustBuild()).
			Property("valueA", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Required("type", "valueA").
			MustBuild()

		typeBSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("type", schema.NewBuilder().Const("B").MustBuild()).
			Property("valueB", schema.NewBuilder().Types(schema.NumberType).MustBuild()).
			Required("type", "valueB").
			MustBuild()

		s := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("data", schema.NewBuilder().
				AllOf(
					schema.NewBuilder().Reference("#/$defs/base").MustBuild(),
					schema.NewBuilder().
						AnyOf(
							schema.NewBuilder().Reference("#/$defs/typeA").MustBuild(),
							schema.NewBuilder().Reference("#/$defs/typeB").MustBuild(),
						).
						MustBuild(),
				).
				MustBuild()).
			Definitions("base", baseSchema).
			Definitions("typeA", typeASchema).
			Definitions("typeB", typeBSchema).
			MustBuild()

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, s)

		v, err := validator.Compile(ctx, s)
		require.NoError(t, err)

		// Valid data - type A
		validDataA := map[string]any{
			"data": map[string]any{
				"id":        "test123",
				"timestamp": 1234567890.0,
				"type":      "A",
				"valueA":    "hello",
			},
		}

		_, err = v.Validate(context.Background(), validDataA)
		require.NoError(t, err)

		// Valid data - type B
		validDataB := map[string]any{
			"data": map[string]any{
				"id":     "test456",
				"type":   "B",
				"valueB": 42.0,
			},
		}

		_, err = v.Validate(context.Background(), validDataB)
		require.NoError(t, err)

		// Invalid data - missing required field from base
		invalidData := map[string]any{
			"data": map[string]any{
				"type":   "A",
				"valueA": "hello",
				// missing "id" from base
			},
		}

		_, err = v.Validate(context.Background(), invalidData)
		require.Error(t, err)
	})

	t.Run("references in conditional schemas", func(t *testing.T) {
		// Schema with references in if/then/else
		isDocumentSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("type", schema.NewBuilder().Const("document").MustBuild()).
			MustBuild()

		documentSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("type", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("title", schema.NewBuilder().Types(schema.StringType).MinLength(1).MustBuild()).
			Property("content", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Required("type", "title", "content").
			MustBuild()

		mediaSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("type", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("url", schema.NewBuilder().Types(schema.StringType).Format("uri").MustBuild()).
			Property("size", schema.NewBuilder().Types(schema.NumberType).Minimum(0).MustBuild()).
			Required("type", "url").
			MustBuild()

		s := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("item", schema.NewBuilder().
				IfSchema(schema.NewBuilder().Reference("#/$defs/isDocument").MustBuild()).
				ThenSchema(schema.NewBuilder().Reference("#/$defs/documentSchema").MustBuild()).
				ElseSchema(schema.NewBuilder().Reference("#/$defs/mediaSchema").MustBuild()).
				MustBuild()).
			Definitions("isDocument", isDocumentSchema).
			Definitions("documentSchema", documentSchema).
			Definitions("mediaSchema", mediaSchema).
			MustBuild()

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, s)

		v, err := validator.Compile(ctx, s)
		require.NoError(t, err)

		// Valid document
		validDocument := map[string]any{
			"item": map[string]any{
				"type":    "document",
				"title":   "My Document",
				"content": "Document content here",
			},
		}

		_, err = v.Validate(context.Background(), validDocument)
		require.NoError(t, err)

		// Valid media
		validMedia := map[string]any{
			"item": map[string]any{
				"type": "image",
				"url":  "https://example.com/image.jpg",
				"size": 1024.0,
			},
		}

		_, err = v.Validate(context.Background(), validMedia)
		require.NoError(t, err)

		// Invalid document - missing required field
		invalidDocument := map[string]any{
			"item": map[string]any{
				"type":  "document",
				"title": "My Document",
				// missing "content"
			},
		}

		_, err = v.Validate(context.Background(), invalidDocument)
		require.Error(t, err)
	})
}

func TestCircularReferenceDetection(t *testing.T) {
	// Schema with circular references should be detected during compilation
	aSchema := schema.NewBuilder().Reference("#/$defs/b").MustBuild()
	bSchema := schema.NewBuilder().Reference("#/$defs/c").MustBuild()
	cSchema := schema.NewBuilder().Reference("#/$defs/a").MustBuild()

	s := schema.NewBuilder().
		Reference("#/$defs/a").
		Definitions("a", aSchema).
		Definitions("b", bSchema).
		Definitions("c", cSchema).
		MustBuild()

	ctx := context.Background()
	ctx = schema.WithResolver(ctx, schema.NewResolver())
	ctx = schema.WithRootSchema(ctx, s)

	_, err := validator.Compile(ctx, s)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular reference")
}

func TestBasicReferenceResolution(t *testing.T) {
	t.Run("local reference", func(t *testing.T) {
		// Schema with a local reference
		stringTypeSchema := schema.NewBuilder().
			Types(schema.StringType).
			MinLength(1).
			MustBuild()

		s := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("name", schema.NewBuilder().Reference("#/$defs/stringType").MustBuild()).
			Definitions("stringType", stringTypeSchema).
			MustBuild()

		// Set up context with resolver and root schema
		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, s)

		v, err := validator.Compile(ctx, s)
		require.NoError(t, err)

		t.Run("valid object", func(t *testing.T) {
			obj := map[string]any{
				"name": "John",
			}
			_, err := v.Validate(context.Background(), obj)
			require.NoError(t, err)
		})

		t.Run("invalid object - empty string", func(t *testing.T) {
			obj := map[string]any{
				"name": "",
			}
			_, err := v.Validate(context.Background(), obj)
			require.Error(t, err)
		})

		t.Run("invalid object - non-string", func(t *testing.T) {
			obj := map[string]any{
				"name": 123,
			}
			_, err := v.Validate(context.Background(), obj)
			require.Error(t, err)
		})
	})

	t.Run("schema with $ref only", func(t *testing.T) {
		// A schema that is just a reference
		personTypeSchema := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("age", schema.NewBuilder().Types(schema.IntegerType).Minimum(0).MustBuild()).
			Required("name").
			MustBuild()

		s := schema.NewBuilder().
			Reference("#/$defs/personType").
			Definitions("personType", personTypeSchema).
			MustBuild()

		// Set up context with resolver and root schema
		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, s)

		v, err := validator.Compile(ctx, s)
		require.NoError(t, err)

		t.Run("valid person", func(t *testing.T) {
			person := map[string]any{
				"name": "Alice",
				"age":  30,
			}
			_, err := v.Validate(context.Background(), person)
			require.NoError(t, err)
		})

		t.Run("invalid person - missing name", func(t *testing.T) {
			person := map[string]any{
				"age": 30,
			}
			_, err := v.Validate(context.Background(), person)
			require.Error(t, err)
		})
	})
}
