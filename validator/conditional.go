package validator

import (
	"context"
	"fmt"
	"log/slog"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/internal/schemactx"
)

// IfThenElseValidator handles if/then/else conditional validation
type IfThenElseValidator struct {
	ifValidator   Interface
	thenValidator Interface
	elseValidator Interface
}

func compileIfThenElseValidator(ctx context.Context, s *schema.Schema, cs compileState) (Interface, error) {
	v := &IfThenElseValidator{}

	// Compile 'if' validator (required)
	ifSchema := convertSchemaOrBool(s.IfSchema())
	ifValidator, err := compile(ctx, ifSchema, cs)
	if err != nil {
		return nil, fmt.Errorf(`failed to compile if validator: %w`, err)
	}
	v.ifValidator = ifValidator

	// Compile 'then' validator (optional)
	if s.HasThenSchema() {
		thenSchema := convertSchemaOrBool(s.ThenSchema())
		thenValidator, err := compile(ctx, thenSchema, cs)
		if err != nil {
			return nil, fmt.Errorf(`failed to compile then validator: %w`, err)
		}
		v.thenValidator = thenValidator
	}

	// Compile 'else' validator (optional)
	if s.HasElseSchema() {
		elseSchema := convertSchemaOrBool(s.ElseSchema())
		elseValidator, err := compile(ctx, elseSchema, cs)
		if err != nil {
			return nil, fmt.Errorf(`failed to compile else validator: %w`, err)
		}
		v.elseValidator = elseValidator
	}

	return v, nil
}

func (v *IfThenElseValidator) Validate(ctx context.Context, in any, options ...ValidateOption) (Result, error) {
	return v.evaluate(ctx, in, newEvalState(ctx, options))
}

func (v *IfThenElseValidator) evaluate(ctx context.Context, in any, st *evalState) (Result, error) {
	// First, check the 'if' condition and collect its annotations
	ifResult, ifErr := evalChild(ctx, v.ifValidator, in, st)

	// The 'if' schema contributes annotations regardless of whether it passes or fails
	var conditionalResult Result

	if ifErr == nil {
		// 'if' condition passed, validate against 'then' if it exists
		if v.thenValidator != nil {
			thenResult, err := evalChild(ctx, v.thenValidator, in, st)
			if err != nil {
				return nil, err
			}
			// Merge 'if' and 'then' results
			conditionalResult = mergeGenericResults(ifResult, thenResult)
		} else {
			// Only 'if' result
			conditionalResult = ifResult
		}
	} else {
		// 'if' condition failed, validate against 'else' if it exists
		if v.elseValidator != nil {
			elseResult, err := evalChild(ctx, v.elseValidator, in, st)
			if err != nil {
				return nil, err
			}
			// Merge 'if' and 'else' results
			conditionalResult = mergeGenericResults(ifResult, elseResult)
		} else {
			// Only 'if' result (even though it failed validation, it may have annotations)
			conditionalResult = ifResult
		}
	}

	return conditionalResult, nil
}

// Logging context functions

// WithTraceSlog adds a trace slog logger to the context for debugging purposes
func WithTraceSlog(ctx context.Context, logger *slog.Logger) context.Context {
	return schemactx.WithTraceSlog(ctx, logger)
}

// TraceSlogFromContext retrieves the trace slog logger from context
// Returns a no-op logger if no logger is associated with the context
func TraceSlogFromContext(ctx context.Context) *slog.Logger {
	return schemactx.TraceSlogFromContext(ctx)
}
