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

func (v *dependentSchemasValidator) Validate(ctx context.Context, value any) (Result, error) {
	// dependentSchemas only applies to objects
	obj, ok := value.(map[string]any)
	if !ok {
		//nolint: nilnil
		return nil, nil
	}

	// Check each dependent schema
	for propertyName, depValidator := range v.dependentSchemas {
		// If the property exists in the object, validate the entire object with the dependent schema
		if _, exists := obj[propertyName]; exists {
			_, err := depValidator.Validate(ctx, value)
			if err != nil {
				return nil, fmt.Errorf("dependent schema validation failed for property %s: %w", propertyName, err)
			}
		}
	}

	//nolint: nilnil
	return nil, nil
}
