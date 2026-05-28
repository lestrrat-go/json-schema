package schema_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

// A nil SchemaOrBool passed to an applicator setter records a builder error.
// Build() must surface that error instead of silently dropping it, and
// MustBuild() must panic rather than return a half-built schema.
func TestBuilderErrorPropagation(t *testing.T) {
	t.Run("Build returns the recorded error", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			build func() (*schema.Schema, error)
		}{
			{"AllOf", func() (*schema.Schema, error) { return schema.NewBuilder().AllOf(nil).Build() }},
			{"AnyOf", func() (*schema.Schema, error) { return schema.NewBuilder().AnyOf(nil).Build() }},
			{"OneOf", func() (*schema.Schema, error) { return schema.NewBuilder().OneOf(nil).Build() }},
			{"PrefixItems", func() (*schema.Schema, error) { return schema.NewBuilder().PrefixItems(nil).Build() }},
		} {
			t.Run(tc.name, func(t *testing.T) {
				s, err := tc.build()
				require.Error(t, err, "Build must propagate the builder error")
				require.Nil(t, s, "Build must not return a schema on error")
			})
		}
	})

	t.Run("error is sticky across subsequent setters", func(t *testing.T) {
		// Once an error is recorded, later setters are no-ops and Build still fails.
		s, err := schema.NewBuilder().AllOf(nil).Types(schema.StringType).Build()
		require.Error(t, err)
		require.Nil(t, s)
	})

	t.Run("MustBuild panics on builder error", func(t *testing.T) {
		require.Panics(t, func() {
			schema.NewBuilder().AllOf(nil).MustBuild()
		})
	})
}
