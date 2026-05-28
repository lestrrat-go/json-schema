package schemactx

import (
	"context"
	"log/slog"
)

type EvaluatedProperties struct {
	props map[string]struct{} // Evaluated properties
}

func (ep *EvaluatedProperties) Keys() []string {
	if ep == nil || ep.props == nil {
		return nil
	}
	keys := make([]string, 0, len(ep.props))
	for k := range ep.props {
		keys = append(keys, k)
	}
	return keys
}

func (ep *EvaluatedProperties) IsEvaluated(prop string) bool {
	if ep == nil || ep.props == nil {
		return false
	}
	_, exists := ep.props[prop]
	return exists
}

func (ep *EvaluatedProperties) MarkEvaluated(prop string) {
	if ep == nil {
		return
	}

	if ep.props == nil {
		ep.props = make(map[string]struct{})
	}
	ep.props[prop] = struct{}{}
}

type EvaluatedItems struct {
	items []bool // Evaluated items
}

func (ei *EvaluatedItems) IsEvaluated(index int) bool {
	if ei == nil || ei.items == nil || index < 0 || index >= len(ei.items) {
		return false
	}
	return ei.items[index]
}

func (ei *EvaluatedItems) Set(index int, value bool) {
	if cap(ei.items) < index+1 {
		allocated := make([]bool, index+1)
		copy(allocated, ei.items)
		ei.items = allocated
	}

	ei.items[index] = value
}

func (ei *EvaluatedItems) Copy(other *EvaluatedItems) {
	for i, v := range other.Values() {
		ei.Set(i, v)
	}
}

func (ei *EvaluatedItems) Values() []bool {
	if ei == nil || ei.items == nil {
		return nil
	}
	return ei.items
}

// EvaluationContext holds the context for validation evaluation, including
// evaluated properties and items. This context is used to track which properties
// and items have been evaluated by previous validators.
type EvaluationContext struct {
	// Properties is a map that tracks properties that have been evaluated
	// by previous validators. The key is the property name, and the value is
	// a struct{} to indicate that the property has been evaluated.
	Properties EvaluatedProperties
	// Items is a slice of booleans indicating whether each item in an array
	// has been evaluated by previous validators. The index corresponds to the item
	// position in the array, and the boolean indicates whether that item has been evaluated.
	Items EvaluatedItems // Evaluated items
}

// Logging context functions

// Context key for trace slog logger
type traceSlogKey struct{}

// WithTraceSlog adds a trace slog logger to the context
func WithTraceSlog(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, traceSlogKey{}, logger)
}

// TraceSlogFromContext retrieves the trace slog logger from context
// Returns a no-op logger if no logger is associated with the context
func TraceSlogFromContext(ctx context.Context) *slog.Logger {
	if v := ctx.Value(traceSlogKey{}); v != nil {
		if logger, ok := v.(*slog.Logger); ok {
			return logger
		}
	}
	// Return no-op logger that discards all output
	return slog.New(slog.DiscardHandler)
}
