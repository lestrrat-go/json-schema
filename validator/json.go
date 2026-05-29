package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// ValidateJSON validates raw JSON text against an already-compiled validator,
// saving callers the boilerplate of unmarshaling first. Compile a schema once
// with Compile, then call ValidateJSON per input (the compile-once /
// validate-many pattern).
//
// Numbers are decoded as json.Number to preserve precision: an integer in the
// 2^53..2^63 range survives intact rather than being rounded by float64, so it
// can be validated exactly against integer constraints. (Integer values outside
// the int64 range cannot be validated as integers and are reported as an error.)
//
// The input must contain exactly one top-level JSON value; trailing content
// after it (other than whitespace) is rejected.
func ValidateJSON(ctx context.Context, v Interface, data []byte, options ...ValidateOption) (Result, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var decoded any
	if err := dec.Decode(&decoded); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("failed to decode JSON: empty input")
		}
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Reject trailing content after the first value. Everything from the
	// decoder's current offset onward must be JSON whitespace. This avoids a
	// second Decode pass (and its allocation) just to detect trailing data.
	for _, b := range data[dec.InputOffset():] {
		switch b {
		case ' ', '\t', '\n', '\r':
		default:
			return nil, fmt.Errorf("invalid JSON: trailing data after top-level value")
		}
	}

	return v.Validate(ctx, decoded, options...)
}
