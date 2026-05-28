package examples_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

// Example_docTracing attaches a structured trace logger with WithTraceSlog. The
// logger records the validation walk keyword by keyword, which is the fastest way
// to see why an input was rejected. In real use point the handler at os.Stderr;
// here it writes to a buffer that is never printed, so the example output stays
// deterministic.
func Example_docTracing() {
	s := schema.NewBuilder().
		Types(schema.ObjectType).
		Property("id", schema.PositiveInteger().MustBuild()).
		Required("id").
		MustBuild()

	var traceOut bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&traceOut, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ctx := validator.WithTraceSlog(context.Background(), logger)

	v, err := validator.Compile(ctx, s)
	if err != nil {
		fmt.Println("compile failed:", err)
		return
	}

	_, err = v.Validate(ctx, map[string]any{"id": 1})
	fmt.Println("valid:", err == nil)
	// Output:
	// valid: true
}
