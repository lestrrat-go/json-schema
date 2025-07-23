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
		jsonSchema := `{
			"type": "object",
			"properties": {
				"user": {"$ref": "#/$defs/person"},
				"manager": {"$ref": "#/$defs/employee"}
			},
			"$defs": {
				"person": {
					"type": "object",
					"properties": {
						"name": {"type": "string"},
						"address": {"$ref": "#/$defs/address"}
					},
					"required": ["name"]
				},
				"employee": {
					"allOf": [
						{"$ref": "#/$defs/person"},
						{
							"type": "object",
							"properties": {
								"employeeId": {"type": "string"},
								"department": {"$ref": "#/$defs/department"}
							},
							"required": ["employeeId"]
						}
					]
				},
				"address": {
					"type": "object",
					"properties": {
						"street": {"type": "string"},
						"city": {"type": "string"}
					},
					"required": ["street", "city"]
				},
				"department": {
					"type": "object",
					"properties": {
						"name": {"type": "string"},
						"budget": {"type": "number", "minimum": 0}
					},
					"required": ["name"]
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, &s)

		v, err := validator.Compile(ctx, &s)
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
		jsonSchema := `{
			"type": "object",
			"properties": {
				"items": {
					"type": "array",
					"items": {"$ref": "#/$defs/item"}
				},
				"metadata": {
					"type": "object",
					"patternProperties": {
						"^meta_": {"$ref": "#/$defs/metaValue"}
					}
				},
				"config": {
					"oneOf": [
						{"$ref": "#/$defs/basicConfig"},
						{"$ref": "#/$defs/advancedConfig"}
					]
				}
			},
			"$defs": {
				"item": {
					"type": "object",
					"properties": {
						"id": {"type": "string"},
						"value": {"type": "number"}
					},
					"required": ["id"]
				},
				"metaValue": {
					"oneOf": [
						{"type": "string"},
						{"type": "number"},
						{"type": "boolean"}
					]
				},
				"basicConfig": {
					"type": "object",
					"properties": {
						"type": {"const": "basic"},
						"enabled": {"type": "boolean"}
					},
					"required": ["type"]
				},
				"advancedConfig": {
					"type": "object",
					"properties": {
						"type": {"const": "advanced"},
						"features": {
							"type": "array",
							"items": {"type": "string"}
						}
					},
					"required": ["type", "features"]
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, &s)

		v, err := validator.Compile(ctx, &s)
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
		jsonSchema := `{
			"$ref": "#/$defs/root",
			"$defs": {
				"root": {"$ref": "#/$defs/level1"},
				"level1": {"$ref": "#/$defs/level2"},
				"level2": {"$ref": "#/$defs/level3"},
				"level3": {"$ref": "#/$defs/actual"},
				"actual": {
					"type": "object",
					"properties": {
						"value": {"type": "string", "minLength": 1}
					},
					"required": ["value"]
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, &s)

		v, err := validator.Compile(ctx, &s)
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
		jsonSchema := `{
			"$ref": "#/$defs/self",
			"$defs": {
				"self": {"$ref": "#/$defs/self"}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, &s)

		_, err := validator.Compile(ctx, &s)
		require.Error(t, err)
		require.Contains(t, err.Error(), "circular reference")
	})

	t.Run("references with complex composition", func(t *testing.T) {
		// Schema combining references with allOf, anyOf, oneOf
		jsonSchema := `{
			"type": "object",
			"properties": {
				"data": {
					"allOf": [
						{"$ref": "#/$defs/base"},
						{
							"anyOf": [
								{"$ref": "#/$defs/typeA"},
								{"$ref": "#/$defs/typeB"}
							]
						}
					]
				}
			},
			"$defs": {
				"base": {
					"type": "object",
					"properties": {
						"id": {"type": "string"},
						"timestamp": {"type": "number"}
					},
					"required": ["id"]
				},
				"typeA": {
					"type": "object",
					"properties": {
						"type": {"const": "A"},
						"valueA": {"type": "string"}
					},
					"required": ["type", "valueA"]
				},
				"typeB": {
					"type": "object",
					"properties": {
						"type": {"const": "B"},
						"valueB": {"type": "number"}
					},
					"required": ["type", "valueB"]
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, &s)

		v, err := validator.Compile(ctx, &s)
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
		jsonSchema := `{
			"type": "object",
			"properties": {
				"item": {
					"if": {"$ref": "#/$defs/isDocument"},
					"then": {"$ref": "#/$defs/documentSchema"},
					"else": {"$ref": "#/$defs/mediaSchema"}
				}
			},
			"$defs": {
				"isDocument": {
					"type": "object",
					"properties": {
						"type": {"const": "document"}
					}
				},
				"documentSchema": {
					"type": "object",
					"properties": {
						"type": {"type": "string"},
						"title": {"type": "string", "minLength": 1},
						"content": {"type": "string"}
					},
					"required": ["type", "title", "content"]
				},
				"mediaSchema": {
					"type": "object", 
					"properties": {
						"type": {"type": "string"},
						"url": {"type": "string", "format": "uri"},
						"size": {"type": "number", "minimum": 0}
					},
					"required": ["type", "url"]
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, &s)

		v, err := validator.Compile(ctx, &s)
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
	jsonSchema := `{
		"$ref": "#/$defs/a",
		"$defs": {
			"a": {"$ref": "#/$defs/b"},
			"b": {"$ref": "#/$defs/c"},
			"c": {"$ref": "#/$defs/a"}
		}
	}`

	var s schema.Schema
	require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

	ctx := context.Background()
	ctx = schema.WithResolver(ctx, schema.NewResolver())
	ctx = schema.WithRootSchema(ctx, &s)

	_, err := validator.Compile(ctx, &s)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular reference")
}

func TestBasicReferenceResolution(t *testing.T) {
	t.Run("local reference", func(t *testing.T) {
		// Schema with a local reference
		jsonSchema := `{
			"type": "object",
			"properties": {
				"name": {"$ref": "#/$defs/stringType"}
			},
			"$defs": {
				"stringType": {"type": "string", "minLength": 1}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		// Set up context with resolver and root schema
		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, &s)

		v, err := validator.Compile(ctx, &s)
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
		jsonSchema := `{
			"$ref": "#/$defs/personType",
			"$defs": {
				"personType": {
					"type": "object",
					"properties": {
						"name": {"type": "string"},
						"age": {"type": "integer", "minimum": 0}
					},
					"required": ["name"]
				}
			}
		}`

		var s schema.Schema
		require.NoError(t, s.UnmarshalJSON([]byte(jsonSchema)))

		// Set up context with resolver and root schema
		ctx := context.Background()
		ctx = schema.WithResolver(ctx, schema.NewResolver())
		ctx = schema.WithRootSchema(ctx, &s)

		v, err := validator.Compile(ctx, &s)
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
