package schema_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/stretchr/testify/require"
)

func TestBuilderReset(t *testing.T) {
	build := func(t *testing.T, b *schema.Builder) *schema.Schema {
		t.Helper()
		s, err := b.Build()
		require.NoError(t, err)
		return s
	}

	newBuilder := func() *schema.Builder {
		return schema.NewBuilder().
			Anchor("a").
			Comment("c").
			Property("name", schema.NewBuilder().Types(schema.StringType).MustBuild())
	}

	t.Run("clears a single field, leaving others", func(t *testing.T) {
		s := build(t, newBuilder().Reset(schema.AnchorField))
		require.False(t, s.HasAnchor(), "anchor should be cleared")
		require.True(t, s.HasComment(), "comment should remain")
		require.True(t, s.HasProperties(), "properties should remain")
	})

	t.Run("clears multiple fields via OR-ed flags", func(t *testing.T) {
		s := build(t, newBuilder().Reset(schema.AnchorField|schema.PropertiesField))
		require.False(t, s.HasAnchor())
		require.False(t, s.HasProperties())
		require.True(t, s.HasComment(), "comment should remain")
	})

	t.Run("matches the per-field ResetXXX method", func(t *testing.T) {
		viaGeneric := build(t, newBuilder().Reset(schema.AnchorField))
		viaSpecific := build(t, newBuilder().ResetAnchor())
		require.Equal(t, viaSpecific.HasAnchor(), viaGeneric.HasAnchor())
		require.Equal(t, viaSpecific.HasProperties(), viaGeneric.HasProperties())
	})

	t.Run("is chainable and a no-op for unset flags", func(t *testing.T) {
		s := build(t, schema.NewBuilder().Comment("c").Reset(schema.AnchorField))
		require.True(t, s.HasComment())
		require.False(t, s.HasAnchor())
	})
}
