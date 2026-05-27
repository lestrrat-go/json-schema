package validator_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// When validating a Go struct, the object validator must read the property name
// from the JSON tag's name portion, ignoring tag options like ",omitempty", and
// exclude json:"-" fields. validation_targets_test.go covers this at the
// extraction-helper level; this exercises the same behavior end-to-end through
// the public Compile/Validate API.
func TestStructJSONTagOptions(t *testing.T) {
	type payload struct {
		Name  string `json:"name,omitempty"`
		Count int    `json:"count,omitempty"`
		Inner string `json:"-"`
		Plain string
	}

	t.Run("required matches the json tag name despite ,omitempty", func(t *testing.T) {
		s, err := schema.NewBuilder().
			Types(schema.ObjectType).
			Required("name", "count").
			Build()
		require.NoError(t, err)
		v, err := validator.Compile(t.Context(), s)
		require.NoError(t, err)
		_, err = v.Validate(t.Context(), payload{Name: "x", Count: 1})
		require.NoError(t, err)
	})

	t.Run("properties keyword keys on the json tag name", func(t *testing.T) {
		nameSchema, err := schema.NewBuilder().Types(schema.StringType).MinLength(1).Build()
		require.NoError(t, err)
		s, err := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("name", nameSchema).
			Build()
		require.NoError(t, err)
		v, err := validator.Compile(t.Context(), s)
		require.NoError(t, err)

		_, err = v.Validate(t.Context(), payload{Name: "ok"})
		require.NoError(t, err)
		_, err = v.Validate(t.Context(), payload{Name: ""})
		require.Error(t, err, "empty name should fail minLength via the json-tag-named property")
	})

	t.Run("json:\"-\" field is excluded", func(t *testing.T) {
		s, err := schema.NewBuilder().
			Types(schema.ObjectType).
			AdditionalProperties(schema.FalseSchema()).
			Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Property("count", schema.NewBuilder().Types(schema.IntegerType).MustBuild()).
			Property("Plain", schema.NewBuilder().Types(schema.StringType).MustBuild()).
			Build()
		require.NoError(t, err)
		v, err := validator.Compile(t.Context(), s)
		require.NoError(t, err)
		// Inner has json:"-", so it must not appear as an (additional) property.
		_, err = v.Validate(t.Context(), payload{Name: "x", Count: 1, Inner: "hidden", Plain: "p"})
		require.NoError(t, err)
	})
}
