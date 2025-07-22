package validator

import (
	"context"
	"fmt"
	"reflect"
)

// validateConst checks if a value exactly matches the expected constant value
func validateConst(ctx context.Context, value any, constValue any) error {
	logger := TraceSlogFromContext(ctx)
	logger.Info("validating const constraint", "expected", constValue, "actual", value)

	if !reflect.DeepEqual(value, constValue) {
		return fmt.Errorf(`must be const value %v`, constValue)
	}
	return nil
}

// validateEnum checks if a value is found in the allowed enum values
func validateEnum(ctx context.Context, value any, enumValues []any) error {
	logger := TraceSlogFromContext(ctx)
	logger.Info("validating enum constraint", "allowed_values", enumValues, "actual", value)

	for _, enumVal := range enumValues {
		if reflect.DeepEqual(value, enumVal) {
			return nil
		}
	}
	return fmt.Errorf(`invalid value: %v not found in enum %v`, value, enumValues)
}