package validator_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// allOf subschemas are "cousins": one branch's evaluated items/properties must
// not be visible to a sibling branch's unevaluatedItems/unevaluatedProperties.
// Annotations only flow upward to the parent, never sideways between branches.
func TestAllOfCousinAnnotationIsolation(t *testing.T) {
	t.Run("unevaluatedItems cannot see a cousin branch's prefixItems", func(t *testing.T) {
		// { "allOf": [ {"prefixItems":[true]}, {"unevaluatedItems":false} ] }
		// prefixItems is in allOf[0]; unevaluatedItems is in allOf[1]. From
		// allOf[1]'s view index 0 is unevaluated, so false must reject it.
		branchA := schema.NewBuilder().PrefixItems(schema.TrueSchema()).MustBuild()
		branchB := schema.NewBuilder().UnevaluatedItems(schema.FalseSchema()).MustBuild()
		s, err := schema.NewBuilder().AllOf(branchA, branchB).Build()
		require.NoError(t, err)
		v, err := validator.Compile(t.Context(), s)
		require.NoError(t, err)

		_, err = v.Validate(t.Context(), []any{1})
		require.Error(t, err, "cousin branch's prefixItems must not satisfy unevaluatedItems:false")
	})

	t.Run("unevaluatedProperties cannot see a cousin branch's properties", func(t *testing.T) {
		// Object analog: properties in allOf[0], unevaluatedProperties:false in allOf[1].
		fooSchema := schema.NewBuilder().Types(schema.StringType).MustBuild()
		branchA := schema.NewBuilder().Property("foo", fooSchema).MustBuild()
		branchB := schema.NewBuilder().UnevaluatedProperties(schema.FalseSchema()).MustBuild()
		s, err := schema.NewBuilder().AllOf(branchA, branchB).Build()
		require.NoError(t, err)
		v, err := validator.Compile(t.Context(), s)
		require.NoError(t, err)

		_, err = v.Validate(t.Context(), map[string]any{"foo": "bar"})
		require.Error(t, err, "cousin branch's properties must not satisfy unevaluatedProperties:false")
	})

	t.Run("parent-level unevaluatedItems still sees allOf branches (upward merge)", func(t *testing.T) {
		// { "prefixItems":[...] hidden in allOf, unevaluatedItems:false at parent }
		// Here unevaluatedItems is the PARENT of allOf, so the branch's annotation
		// must reach it and the item is evaluated -> valid.
		branch := schema.NewBuilder().PrefixItems(schema.TrueSchema()).MustBuild()
		s, err := schema.NewBuilder().
			AllOf(branch).
			UnevaluatedItems(schema.FalseSchema()).
			Build()
		require.NoError(t, err)
		v, err := validator.Compile(t.Context(), s)
		require.NoError(t, err)

		_, err = v.Validate(t.Context(), []any{1})
		require.NoError(t, err, "parent unevaluatedItems must see the allOf branch's prefixItems")
	})
}
