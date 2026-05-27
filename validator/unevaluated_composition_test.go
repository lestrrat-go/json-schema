package validator_test

import (
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

// These tests exercise the interaction between unevaluatedProperties /
// unevaluatedItems and the in-place applicators anyOf / oneOf / if-then-else
// strictly through the public Compile -> Validate path. Per the JSON Schema
// spec, properties (or items) evaluated by an applicator subschema that
// *successfully applies* are considered evaluated and are therefore exempt
// from "unevaluatedProperties": false; properties seen only by a subschema
// that does not apply remain unevaluated.

func mustBuild(t *testing.T, b *schema.Builder) *schema.Schema {
	t.Helper()
	s, err := b.Build()
	require.NoError(t, err)
	return s
}

func TestUnevaluatedPropertiesWithAnyOf(t *testing.T) {
	// foo via properties; bar/baz only reachable through anyOf branches that
	// must actually match (required + const) to evaluate their property.
	barBranch := mustBuild(t, schema.NewBuilder().
		Property("bar", mustBuild(t, schema.NewBuilder().Const("bar"))).
		Required("bar"))
	bazBranch := mustBuild(t, schema.NewBuilder().
		Property("baz", mustBuild(t, schema.NewBuilder().Const("baz"))).
		Required("baz"))

	s := mustBuild(t, schema.NewBuilder().
		Types(schema.ObjectType).
		Property("foo", mustBuild(t, schema.NewBuilder().Types(schema.StringType))).
		AnyOf(barBranch, bazBranch).
		UnevaluatedProperties(schema.FalseSchema()))

	v, err := validator.Compile(t.Context(), s)
	require.NoError(t, err)

	t.Run("property evaluated by matching anyOf branch is allowed", func(t *testing.T) {
		_, err := v.Validate(t.Context(), map[string]any{"foo": "x", "bar": "bar"})
		require.NoError(t, err)
	})

	t.Run("both branches matching, all properties evaluated", func(t *testing.T) {
		_, err := v.Validate(t.Context(), map[string]any{"foo": "x", "bar": "bar", "baz": "baz"})
		require.NoError(t, err)
	})

	t.Run("truly extra property is unevaluated and rejected", func(t *testing.T) {
		_, err := v.Validate(t.Context(), map[string]any{"foo": "x", "bar": "bar", "extra": "nope"})
		require.Error(t, err)
	})

	t.Run("property seen only by a non-matching branch stays unevaluated", func(t *testing.T) {
		// baz="wrong" fails the baz branch (const mismatch), so that branch does
		// not apply and does not evaluate "baz"; bar branch matches. "baz" is
		// therefore unevaluated and must be rejected.
		_, err := v.Validate(t.Context(), map[string]any{"foo": "x", "bar": "bar", "baz": "wrong"})
		require.Error(t, err)
	})
}

func TestUnevaluatedPropertiesWithOneOf(t *testing.T) {
	barBranch := mustBuild(t, schema.NewBuilder().
		Property("bar", mustBuild(t, schema.NewBuilder().Types(schema.StringType))).
		Required("bar"))
	bazBranch := mustBuild(t, schema.NewBuilder().
		Property("baz", mustBuild(t, schema.NewBuilder().Types(schema.StringType))).
		Required("baz"))

	s := mustBuild(t, schema.NewBuilder().
		Types(schema.ObjectType).
		Property("foo", mustBuild(t, schema.NewBuilder().Types(schema.StringType))).
		OneOf(barBranch, bazBranch).
		UnevaluatedProperties(schema.FalseSchema()))

	v, err := validator.Compile(t.Context(), s)
	require.NoError(t, err)

	t.Run("property from the single matching branch is allowed", func(t *testing.T) {
		_, err := v.Validate(t.Context(), map[string]any{"foo": "x", "bar": "y"})
		require.NoError(t, err)
	})

	t.Run("extra property alongside matching branch is unevaluated", func(t *testing.T) {
		_, err := v.Validate(t.Context(), map[string]any{"foo": "x", "bar": "y", "qux": "z"})
		require.Error(t, err)
	})

	t.Run("matching neither branch fails oneOf", func(t *testing.T) {
		_, err := v.Validate(t.Context(), map[string]any{"foo": "x"})
		require.Error(t, err)
	})
}

func TestUnevaluatedPropertiesWithIfThenElse(t *testing.T) {
	ifBranch := mustBuild(t, schema.NewBuilder().
		Property("kind", mustBuild(t, schema.NewBuilder().Const("withThen"))).
		Required("kind"))
	thenBranch := mustBuild(t, schema.NewBuilder().
		Property("thenProp", mustBuild(t, schema.NewBuilder().Types(schema.StringType))).
		Required("thenProp"))
	elseBranch := mustBuild(t, schema.NewBuilder().
		Property("elseProp", mustBuild(t, schema.NewBuilder().Types(schema.StringType))).
		Required("elseProp"))

	s := mustBuild(t, schema.NewBuilder().
		Types(schema.ObjectType).
		Property("kind", mustBuild(t, schema.NewBuilder().Types(schema.StringType))).
		IfSchema(ifBranch).
		ThenSchema(thenBranch).
		ElseSchema(elseBranch).
		UnevaluatedProperties(schema.FalseSchema()))

	v, err := validator.Compile(t.Context(), s)
	require.NoError(t, err)

	t.Run("then branch evaluates its property", func(t *testing.T) {
		_, err := v.Validate(t.Context(), map[string]any{"kind": "withThen", "thenProp": "ok"})
		require.NoError(t, err)
	})

	t.Run("else branch evaluates its property", func(t *testing.T) {
		_, err := v.Validate(t.Context(), map[string]any{"kind": "other", "elseProp": "ok"})
		require.NoError(t, err)
	})

	t.Run("else-branch property is unevaluated when then applies", func(t *testing.T) {
		// "if" matches, so "then" applies and "else" does not. elseProp is only
		// named by the else branch, so it is unevaluated here.
		_, err := v.Validate(t.Context(), map[string]any{"kind": "withThen", "thenProp": "ok", "elseProp": "leak"})
		require.Error(t, err)
	})
}

func TestUnevaluatedItemsWithAnyOf(t *testing.T) {
	// prefixItems evaluates index 0; an anyOf branch with prefixItems of length 2
	// evaluates index 1 only when that branch applies.
	stringSchema := mustBuild(t, schema.NewBuilder().Types(schema.StringType))
	branch := mustBuild(t, schema.NewBuilder().
		PrefixItems(stringSchema, stringSchema))

	s := mustBuild(t, schema.NewBuilder().
		Types(schema.ArrayType).
		PrefixItems(stringSchema).
		AnyOf(branch).
		UnevaluatedItems(schema.FalseSchema()))

	v, err := validator.Compile(t.Context(), s)
	require.NoError(t, err)

	t.Run("items evaluated across base and matching branch are allowed", func(t *testing.T) {
		_, err := v.Validate(t.Context(), []any{"a", "b"})
		require.NoError(t, err)
	})

	t.Run("item beyond what any applicator evaluates is rejected", func(t *testing.T) {
		_, err := v.Validate(t.Context(), []any{"a", "b", "c"})
		require.Error(t, err)
	})
}
