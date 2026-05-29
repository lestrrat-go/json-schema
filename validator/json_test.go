package validator_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestValidateJSON(t *testing.T) {
	ctx := t.Context()

	t.Run("malformed and empty input", func(t *testing.T) {
		s := schema.NewBuilder().Types(schema.ObjectType).MustBuild()
		v, err := validator.Compile(ctx, s)
		require.NoError(t, err)

		testCases := []struct {
			name string
			data string
		}{
			{name: "empty", data: ""},
			{name: "whitespace only", data: "   \n\t "},
			{name: "invalid json", data: `{"a":}`},
			{name: "trailing value", data: `{} {}`},
			{name: "trailing scalar", data: `5 6`},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := validator.ValidateJSON(ctx, v, []byte(tc.data))
				require.Error(t, err)
			})
		}
	})

	t.Run("object", func(t *testing.T) {
		s := schema.NewBuilder().
			Types(schema.ObjectType).
			Property("id", schema.PositiveInteger().MustBuild()).
			Property("role", schema.Enum("admin", "user").MustBuild()).
			Required("id", "role").
			MustBuild()
		v, err := validator.Compile(ctx, s)
		require.NoError(t, err)

		_, err = validator.ValidateJSON(ctx, v, []byte(`{"id": 1, "role": "admin"}`))
		require.NoError(t, err)

		_, err = validator.ValidateJSON(ctx, v, []byte(`{"id": 1, "role": "root"}`))
		require.Error(t, err)

		_, err = validator.ValidateJSON(ctx, v, []byte(`{"role": "admin"}`))
		require.Error(t, err, "missing required id")

		// Surrounding whitespace is fine.
		_, err = validator.ValidateJSON(ctx, v, []byte("\n  {\"id\": 1, \"role\": \"user\"}\n"))
		require.NoError(t, err)
	})

	t.Run("integer precision and form", func(t *testing.T) {
		s := schema.NewBuilder().Types(schema.IntegerType).MustBuild()
		v, err := validator.Compile(ctx, s)
		require.NoError(t, err)

		testCases := []struct {
			name    string
			data    string
			wantErr bool
		}{
			{name: "large integer beyond 2^53", data: "9007199254740993"},
			{name: "integral with decimal point", data: "5.0"},
			{name: "exponent form integer", data: "1e2"},
			{name: "fractional fails", data: "5.5", wantErr: true},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := validator.ValidateJSON(ctx, v, []byte(tc.data))
				if tc.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
			})
		}
	})

	t.Run("const and enum", func(t *testing.T) {
		constSchema := schema.NewBuilder().Const(42).MustBuild()
		cv, err := validator.Compile(ctx, constSchema)
		require.NoError(t, err)
		_, err = validator.ValidateJSON(ctx, cv, []byte("42"))
		require.NoError(t, err)
		_, err = validator.ValidateJSON(ctx, cv, []byte("43"))
		require.Error(t, err)

		enumSchema := schema.NewBuilder().Enum(1, 2, 3).MustBuild()
		ev, err := validator.Compile(ctx, enumSchema)
		require.NoError(t, err)
		_, err = validator.ValidateJSON(ctx, ev, []byte("2"))
		require.NoError(t, err)
		_, err = validator.ValidateJSON(ctx, ev, []byte("9"))
		require.Error(t, err)
	})

	t.Run("multi-type integer or string", func(t *testing.T) {
		s := schema.NewBuilder().Types(schema.IntegerType, schema.StringType).MustBuild()
		v, err := validator.Compile(ctx, s)
		require.NoError(t, err)

		_, err = validator.ValidateJSON(ctx, v, []byte("5"))
		require.NoError(t, err, "integer satisfies the integer branch")

		_, err = validator.ValidateJSON(ctx, v, []byte(`"hello"`))
		require.NoError(t, err, "string satisfies the string branch")

		_, err = validator.ValidateJSON(ctx, v, []byte("true"))
		require.Error(t, err, "boolean satisfies neither branch")
	})
}
