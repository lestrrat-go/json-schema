package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestValidationTargets tests the validation target resolution functionality
func TestValidationTargets(t *testing.T) {
	t.Run("ObjectFieldResolver interface", func(t *testing.T) {
		obj := &testObjectResolver{
			fields: map[string]any{
				"name": "John",
				"age":  30,
			},
		}

		value, err := resolveObjectField(obj, "name")
		require.NoError(t, err)
		require.Equal(t, "John", value)

		value, err = resolveObjectField(obj, "age")
		require.NoError(t, err)
		require.Equal(t, 30, value)

		_, err = resolveObjectField(obj, "missing")
		require.Error(t, err)
	})

	t.Run("ArrayIndexResolver interface", func(t *testing.T) {
		arr := &testArrayResolver{
			items: []any{"first", "second", "third"},
		}

		value, err := resolveArrayIndex(arr, 0)
		require.NoError(t, err)
		require.Equal(t, "first", value)

		value, err = resolveArrayIndex(arr, 2)
		require.NoError(t, err)
		require.Equal(t, "third", value)

		_, err = resolveArrayIndex(arr, 10)
		require.Error(t, err)
	})

	t.Run("Map object field resolution", func(t *testing.T) {
		obj := map[string]any{
			"foo": "bar",
			"baz": 42,
		}

		value, err := resolveObjectField(obj, "foo")
		require.NoError(t, err)
		require.Equal(t, "bar", value)

		value, err = resolveObjectField(obj, "baz")
		require.NoError(t, err)
		require.Equal(t, 42, value)

		_, err = resolveObjectField(obj, "missing")
		require.Error(t, err)
	})

	t.Run("Struct field resolution with JSON tags", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
			Bar  string `json:"baz"` // JSON field name differs from struct field
			Qux  string // No JSON tag, should match field name
		}

		obj := TestStruct{
			Name: "Alice",
			Age:  25,
			Bar:  "hello",
			Qux:  "world",
		}

		// Test JSON tag resolution
		value, err := resolveObjectField(obj, "name")
		require.NoError(t, err)
		require.Equal(t, "Alice", value)

		value, err = resolveObjectField(obj, "age")
		require.NoError(t, err)
		require.Equal(t, 25, value)

		// Test JSON tag with different field name
		value, err = resolveObjectField(obj, "baz")
		require.NoError(t, err)
		require.Equal(t, "hello", value)

		// Test field name without JSON tag
		value, err = resolveObjectField(obj, "qux")
		require.NoError(t, err)
		require.Equal(t, "world", value)

		// Test case-insensitive matching
		value, err = resolveObjectField(obj, "Qux")
		require.NoError(t, err)
		require.Equal(t, "world", value)

		// Test missing field
		_, err = resolveObjectField(obj, "missing")
		require.Error(t, err)
	})

	t.Run("Slice array index resolution", func(t *testing.T) {
		arr := []any{"a", "b", "c"}

		value, err := resolveArrayIndex(arr, 0)
		require.NoError(t, err)
		require.Equal(t, "a", value)

		value, err = resolveArrayIndex(arr, 2)
		require.NoError(t, err)
		require.Equal(t, "c", value)

		_, err = resolveArrayIndex(arr, 5)
		require.Error(t, err)

		_, err = resolveArrayIndex(arr, -1)
		require.Error(t, err)
	})

	t.Run("NewArrayResult with size parameter", func(t *testing.T) {
		// Test without size parameter
		result1 := NewArrayResult()
		require.NotNil(t, result1)
		require.Equal(t, 0, len(result1.EvaluatedItems()))

		// Test with size parameter
		result2 := NewArrayResult(10)
		require.NotNil(t, result2)
		require.Equal(t, 0, len(result2.EvaluatedItems()))
	})
}

// testObjectResolver implements ObjectFieldResolver for testing
type testObjectResolver struct {
	fields map[string]any
}

func (t *testObjectResolver) ResolveObjectField(field string) (any, error) {
	if value, exists := t.fields[field]; exists {
		return value, nil
	}
	return nil, fmt.Errorf("field %q not found", field)
}

// testArrayResolver implements ArrayIndexResolver for testing
type testArrayResolver struct {
	items []any
}

func (t *testArrayResolver) ResolveArrayIndex(index int) (any, error) {
	if index < 0 || index >= len(t.items) {
		return nil, fmt.Errorf("index %d out of bounds", index)
	}
	return t.items[index], nil
}
