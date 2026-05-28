package meta

import (
	"context"

	"github.com/lestrrat-go/json-schema/validator"
)

// metaSchemaValidator wraps the generated meta validator so the
// "$dynamicRef": "#meta" nodes inside it resolve back to the meta validator
// itself.
//
// The 2020-12 meta-schema declares "$dynamicAnchor": "meta" at its root and
// applies "$dynamicRef": "#meta" to every subschema, which means "this value
// must itself be a valid schema". Since metaValidator embodies the whole
// meta-schema, that recursion is simply re-entry into metaValidator. The
// generated validator carries no schema document at runtime, so instead of
// resolving the anchor against a document we register metaValidator under the
// "meta" dynamic anchor and let DynamicReferenceValidator pick it up.
type metaSchemaValidator struct{}

func (metaSchemaValidator) Validate(ctx context.Context, v any, options ...validator.ValidateOption) (validator.Result, error) {
	// Register metaValidator under the "meta" $dynamicAnchor so the "#meta"
	// $dynamicRefs inside it resolve back to the whole meta validator.
	options = append(options, validator.WithDynamicAnchorValidator("meta", metaValidator))
	return metaValidator.Validate(ctx, v, options...)
}

// Validator returns a pre-compiled validator for the JSON Schema 2020-12
// meta-schema. This validator can be used to validate JSON Schema documents
// themselves.
//
// Example usage:
//
//	validator := meta.Validator()
//	result, err := validator.Validate(ctx, jsonSchemaDocument)
func Validator() validator.Interface {
	return metaSchemaValidator{}
}

// Validate validates a JSON Schema document against the JSON Schema 2020-12
// meta-schema. This is a convenience function that uses the pre-compiled
// validator.
//
// Example usage:
//
//	err := meta.Validate(ctx, jsonSchemaDocument)
//	if err != nil {
//	    // The document is not a valid JSON Schema
//	}
func Validate(ctx context.Context, jsonSchemaDocument any) error {
	_, err := Validator().Validate(ctx, jsonSchemaDocument)
	return err
}
