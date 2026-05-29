package validator

import (
	"context"
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

type dependentSchemasValidator struct {
	dependentSchemas map[string]Interface
}

// DependentSchemasValidator validates dependent schemas according to JSON Schema 2020-12.
// When a property name exists in the data object, the corresponding schema must also validate the entire object.
func DependentSchemasValidator(ctx context.Context, dependentSchemas map[string]*schema.Schema) (Interface, error) {
	if len(dependentSchemas) == 0 {
		//nolint: nilnil
		return nil, nil
	}

	validators := make(map[string]Interface)
	for propertyName, depSchema := range dependentSchemas {
		validator, err := Compile(ctx, depSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to compile dependent schema for property %s: %w", propertyName, err)
		}
		validators[propertyName] = validator
	}

	return &dependentSchemasValidator{
		dependentSchemas: validators,
	}, nil
}

func (v *dependentSchemasValidator) Validate(ctx context.Context, value any, options ...ValidateOption) (Result, error) {
	return v.evaluate(ctx, value, newEvalState(ctx, options))
}

func (v *dependentSchemasValidator) evaluate(ctx context.Context, value any, st *evalState) (Result, error) {
	// dependentSchemas only applies to objects. Reuse the object validator's
	// extraction so structs and ObjectFieldResolvers are handled identically,
	// instead of only accepting map[string]any.
	obj, isObject, err := extractObjectProperties(value)
	if err != nil {
		return nil, fmt.Errorf("dependent schema validation failed: %w", err)
	}
	if !isObject {
		//nolint: nilnil
		return nil, nil
	}

	// Check each dependent schema
	for propertyName, depValidator := range v.dependentSchemas {
		// If the property exists in the object, validate the entire object with the dependent schema
		if _, exists := obj[propertyName]; exists {
			if _, err := evalChild(ctx, depValidator, value, st); err != nil {
				return nil, fmt.Errorf("dependent schema validation failed for property %s: %w", propertyName, err)
			}
		}
	}

	//nolint: nilnil
	return nil, nil
}
