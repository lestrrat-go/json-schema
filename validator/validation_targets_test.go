package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestValidationTargets exercises the instance-access helpers that back object
// and array validation: extractObjectProperties and newArrayAccessor.
func TestValidationTargets(t *testing.T) {
	t.Run("ObjectFieldResolver drives extractObjectProperties", func(t *testing.T) {
		obj := &testObjectResolver{fields: map[string]any{"name": "John", "age": 30}}
		props, ok, err := extractObjectProperties(obj)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, "John", props["name"])
		require.Equal(t, 30, props["age"])
		_, exists := props["missing"]
		require.False(t, exists)
	})

	t.Run("ArrayIndexResolver drives newArrayAccessor", func(t *testing.T) {
		arr := &testArrayResolver{items: []any{"first", "second", "third"}}
		acc, ok := newArrayAccessor(arr)
		require.True(t, ok)
		require.Equal(t, 3, acc.length)

		v0, err := acc.at(0)
		require.NoError(t, err)
		require.Equal(t, "first", v0)

		v2, err := acc.at(2)
		require.NoError(t, err)
		require.Equal(t, "third", v2)

		_, err = acc.at(10)
		require.Error(t, err)
	})

	t.Run("map object extraction", func(t *testing.T) {
		props, ok, err := extractObjectProperties(map[string]any{"foo": "bar", "baz": 42})
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, "bar", props["foo"])
		require.Equal(t, 42, props["baz"])
	})

	t.Run("struct extraction keys on json tag names", func(t *testing.T) {
		type TestStruct struct {
			Name   string `json:"name"`
			Age    int    `json:"age"`
			Bar    string `json:"baz"`           // renamed via tag
			Opt    string `json:"opt,omitempty"` // tag options must be ignored
			Hidden string `json:"-"`             // excluded entirely
			Qux    string // no tag -> exact field name
		}
		obj := TestStruct{Name: "Alice", Age: 25, Bar: "hello", Opt: "o", Hidden: "secret", Qux: "world"}

		props, ok, err := extractObjectProperties(obj)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, "Alice", props["name"])
		require.Equal(t, 25, props["age"])
		require.Equal(t, "hello", props["baz"])
		require.Equal(t, "o", props["opt"], "tag option ,omitempty must not be part of the key")
		require.Equal(t, "world", props["Qux"])

		_, hidden := props["Hidden"]
		require.False(t, hidden, `json:"-" field must be excluded`)
	})

	t.Run("slice array accessor", func(t *testing.T) {
		acc, ok := newArrayAccessor([]any{"a", "b", "c"})
		require.True(t, ok)
		require.Equal(t, 3, acc.length)
		v, err := acc.at(0)
		require.NoError(t, err)
		require.Equal(t, "a", v)
	})

	t.Run("non-object and non-array are reported via the bool", func(t *testing.T) {
		_, ok, err := extractObjectProperties(42)
		require.NoError(t, err)
		require.False(t, ok)

		_, ok = newArrayAccessor(42)
		require.False(t, ok)
	})

	t.Run("NewArrayResult with size parameter", func(t *testing.T) {
		require.Equal(t, 0, len(NewArrayResult().EvaluatedItems()))
		require.Equal(t, 0, len(NewArrayResult(10).EvaluatedItems()))
	})
}

// testObjectResolver implements ObjectFieldResolver for testing
type testObjectResolver struct {
	fields map[string]any
}

func (t *testObjectResolver) FieldNames() []string {
	names := make([]string, 0, len(t.fields))
	for name := range t.fields {
		names = append(names, name)
	}
	return names
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

func (t *testArrayResolver) Len() int { return len(t.items) }

func (t *testArrayResolver) ResolveArrayIndex(index int) (any, error) {
	if index < 0 || index >= len(t.items) {
		return nil, fmt.Errorf("index %d out of bounds", index)
	}
	return t.items[index], nil
}
