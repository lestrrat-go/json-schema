package validator_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// opaqueObject is a custom type whose JSON representation the validator can only
// learn through the ObjectFieldResolver interface (it is neither a map nor a
// struct with the relevant fields). It demonstrates that, with FieldNames added
// to the interface, enumeration-based keywords (additionalProperties,
// minProperties, propertyNames) work through the resolver.
type opaqueObject struct {
	store map[string]any
}

func (o opaqueObject) FieldNames() []string {
	names := make([]string, 0, len(o.store))
	for k := range o.store {
		names = append(names, k)
	}
	return names
}

func (o opaqueObject) ResolveObjectField(name string) (any, error) {
	return o.store[name], nil
}

func TestObjectFieldResolverDrivesValidation(t *testing.T) {
	nameSchema := schema.NewBuilder().Types(schema.StringType).MinLength(1).MustBuild()
	s, err := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("name", nameSchema).
		Required("name").
		AdditionalProperties(schema.FalseSchema()).
		Build()
	require.NoError(t, err)
	v, err := validator.Compile(t.Context(), s)
	require.NoError(t, err)

	t.Run("resolver-only object validates via properties + required", func(t *testing.T) {
		_, err := v.Validate(t.Context(), opaqueObject{store: map[string]any{"name": "ok"}})
		require.NoError(t, err)
	})

	t.Run("enumeration sees an extra field and additionalProperties:false rejects it", func(t *testing.T) {
		_, err := v.Validate(t.Context(), opaqueObject{store: map[string]any{"name": "ok", "extra": 1}})
		require.Error(t, err, "FieldNames must expose 'extra' so additionalProperties:false can reject it")
	})

	t.Run("required missing is detected through the resolver", func(t *testing.T) {
		_, err := v.Validate(t.Context(), opaqueObject{store: map[string]any{}})
		require.Error(t, err)
	})
}

// TestUnevaluatedPropertiesThroughResolver covers the unevaluatedProperties
// keyword for object values that are not map[string]any. Previously the
// unevaluated coordinator only accepted map[string]any, so it skipped (or, under
// strict object type, wrongly rejected) ObjectFieldResolver and struct values.
// It now reuses extractObjectProperties, matching the object validator and
// dependentSchemas.
func TestUnevaluatedPropertiesThroughResolver(t *testing.T) {
	s, err := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
		UnevaluatedProperties(schema.FalseSchema()).
		Build()
	require.NoError(t, err)
	v, err := validator.Compile(t.Context(), s)
	require.NoError(t, err)

	t.Run("resolver: evaluated-only object passes", func(t *testing.T) {
		_, err := v.Validate(t.Context(), opaqueObject{store: map[string]any{"name": "ok"}})
		require.NoError(t, err)
	})

	t.Run("resolver: unevaluated property rejected", func(t *testing.T) {
		_, err := v.Validate(t.Context(), opaqueObject{store: map[string]any{"name": "ok", "extra": 1}})
		require.Error(t, err, "extra is unevaluated and unevaluatedProperties:false")
	})

	t.Run("map sanity: unevaluated property rejected", func(t *testing.T) {
		_, err := v.Validate(t.Context(), map[string]any{"name": "ok", "extra": 1})
		require.Error(t, err)
	})

	t.Run("struct: unevaluated field rejected", func(t *testing.T) {
		type person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		_, err := v.Validate(t.Context(), person{Name: "ok", Age: 1})
		require.Error(t, err, "age is unevaluated and unevaluatedProperties:false")
	})
}

// opaqueArray is reachable only through the ArrayIndexResolver interface,
// demonstrating that Len() lets length-based keywords (minItems/maxItems) and
// per-item keywords (prefixItems/items) work through the resolver.
type opaqueArray struct {
	elems []any
}

func (o opaqueArray) Len() int { return len(o.elems) }

func (o opaqueArray) ResolveArrayIndex(i int) (any, error) { return o.elems[i], nil }

func TestArrayIndexResolverDrivesValidation(t *testing.T) {
	elemSchema := schema.NewBuilder().Types(schema.StringType).MustBuild()
	s, err := schema.NewBuilder().
		Types(schema.ArrayType).
		Items(elemSchema).
		MinItems(2).
		Build()
	require.NoError(t, err)
	v, err := validator.Compile(t.Context(), s)
	require.NoError(t, err)

	t.Run("items + minItems satisfied through the resolver", func(t *testing.T) {
		_, err := v.Validate(t.Context(), opaqueArray{elems: []any{"a", "b"}})
		require.NoError(t, err)
	})

	t.Run("minItems uses Len() from the resolver", func(t *testing.T) {
		_, err := v.Validate(t.Context(), opaqueArray{elems: []any{"a"}})
		require.Error(t, err)
	})

	t.Run("items schema applied to resolver elements", func(t *testing.T) {
		_, err := v.Validate(t.Context(), opaqueArray{elems: []any{"a", 42}})
		require.Error(t, err, "non-string element must fail the items schema")
	})
}
