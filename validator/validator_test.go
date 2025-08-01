package validator_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/internal/schemactx"
	"github.com/lestrrat-go/json-schema/validator"
	"github.com/stretchr/testify/assert"
)

// contextWithLogging creates a context with trace slog when verbose testing is enabled
func contextWithLogging(_ *testing.T) context.Context {
	ctx := context.Background()
	if testing.Verbose() {
		logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		ctx = schemactx.WithTraceSlog(ctx, logger)
	}
	return ctx
}

func makeSanityTestFunc(tc *sanityTestCase, c validator.Interface) func(*testing.T) {
	return func(t *testing.T) {
		ctx := contextWithLogging(t)
		if tc.Error {
			_, err := c.Validate(ctx, tc.Object)
			if !assert.Error(t, err, `c.check should fail`) {
				return
			}
		} else {
			_, err := c.Validate(ctx, tc.Object)
			if !assert.NoError(t, err, `c.Validate should succeed`) {
				return
			}
		}
	}
}

// Some default set of objects used for sanity checking
type sanityTestCase struct {
	Object any
	Name   string
	Error  bool
}

func makeSanityTestCases() []*sanityTestCase {
	return []*sanityTestCase{
		{
			Name:   "Empty Map",
			Object: make(map[string]any),
		},
		{
			Name:   "Empty Object",
			Object: struct{}{},
		},
		{
			Name:   "Integer",
			Object: 1,
		},
	}
}

func TestValidator(t *testing.T) {
	s, err := schema.NewBuilder().
		Types(schema.ObjectType).
		Build()
	if !assert.NoError(t, err, `schema.NewBuilder should succeed`) {
		return
	}
	_ = s
	/*
		v, err := validator.Compile(context.Background(), s)
		if !assert.NoError(t, err, `validator.Build should succeed`) {
			return
		}
		_ = v
	*/
}
