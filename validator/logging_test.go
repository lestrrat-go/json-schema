package validator_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/require"
)

func TestLoggingIntegration(t *testing.T) {
	// Create a simple schema
	s := schema.NewBuilder().
		Types(schema.StringType).
		MinLength(3).
		MaxLength(10).
		MustBuild()

	// Compile validator
	v, err := validator.Compile(context.Background(), s)
	require.NoError(t, err)

	// Test with trace logging context
	t.Run("with trace logging context", func(t *testing.T) {
		// Create context with trace logger
		logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		ctx := validator.WithTraceSlog(context.Background(), logger)

		// Verify logger can be retrieved
		retrievedLogger := validator.TraceSlogFromContext(ctx)
		require.NotNil(t, retrievedLogger, "should be able to retrieve trace logger from context")

		// Use the context for validation (logging would occur within validators that use it)
		_, err := v.Validate(ctx, "hello")
		require.NoError(t, err)

		_, err = v.Validate(ctx, "x") // too short
		require.Error(t, err)
	})

	t.Run("without trace logging context", func(t *testing.T) {
		ctx := context.Background()

		// Verify no-op logger is returned when no logger in context
		retrievedLogger := validator.TraceSlogFromContext(ctx)
		require.NotNil(t, retrievedLogger, "should return no-op logger when no trace logger in context")

		// The no-op logger should be usable (won't panic)
		retrievedLogger.DebugContext(ctx, "test message - this should be discarded")
		retrievedLogger.InfoContext(ctx, "test info - this should be discarded")

		// Validation should still work
		_, err := v.Validate(ctx, "hello")
		require.NoError(t, err)
	})

	t.Run("no-op logger behavior", func(_ *testing.T) {
		// Test that the no-op logger can be used safely without panics
		ctx := context.Background()
		noopLogger := validator.TraceSlogFromContext(ctx)

		// These should not panic or cause issues
		noopLogger.DebugContext(ctx, "debug message")
		noopLogger.InfoContext(ctx, "info message")
		noopLogger.WarnContext(ctx, "warn message")
		noopLogger.ErrorContext(ctx, "error message")

		// Test with context values
		noopLogger.With("key", "value").InfoContext(ctx, "message with context")
	})
}
