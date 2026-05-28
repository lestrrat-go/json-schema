package validator_test

import (
	"context"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestValidateRespectsCancellation(t *testing.T) {
	const src = `{
		"type": "object",
		"properties": {"a": {"type": "string"}},
		"additionalProperties": true
	}`
	var s schema.Schema
	require.NoError(t, s.UnmarshalJSON([]byte(src)))

	v, err := validator.Compile(t.Context(), &s)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before validating

	_, err = v.Validate(ctx, map[string]any{"a": "x", "b": 1, "c": 2})
	require.ErrorIs(t, err, context.Canceled)
}

func TestCompileRespectsCancellation(t *testing.T) {
	const src = `{"type": "object", "properties": {"a": {"type": "string"}}}`
	var s schema.Schema
	require.NoError(t, s.UnmarshalJSON([]byte(src)))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before compiling

	_, err := validator.Compile(ctx, &s)
	require.ErrorIs(t, err, context.Canceled)
}
