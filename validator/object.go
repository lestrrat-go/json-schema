package validator

import (
	"context"
	"fmt"
	"reflect"
	"regexp"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/internal/schemactx"
)

var _ Builder = (*ObjectValidatorBuilder)(nil)
var _ Interface = (*objectValidator)(nil)

func compileObjectValidator(ctx context.Context, s *schema.Schema, strictType bool) (Interface, error) {
	v := Object()

	if s.HasMinProperties() && IsKeywordEnabledInContext(ctx, "minProperties") {
		v.MinProperties(s.MinProperties())
	}
	if s.HasMaxProperties() && IsKeywordEnabledInContext(ctx, "maxProperties") {
		v.MaxProperties(s.MaxProperties())
	}
	if s.HasRequired() && IsKeywordEnabledInContext(ctx, "required") {
		v.Required(s.Required())
	}
	if s.HasDependentRequired() && IsKeywordEnabledInContext(ctx, "dependentRequired") {
		v.DependentRequired(s.DependentRequired())
	}
	if s.HasProperties() {
		properties := make(map[string]Interface)
		for name, propSchema := range s.Properties() {
			propValidator, err := Compile(ctx, propSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to compile property validator for %s: %w", name, err)
			}
			properties[name] = propValidator
		}
		// Convert map to PropertyPair slice
		props := make([]PropertyPair, 0, len(properties))
		for name, validator := range properties {
			props = append(props, PropPair(name, validator))
		}
		v.Properties(props...)
	}
	if s.HasPatternProperties() {
		patternProperties := make(map[*regexp.Regexp]Interface)
		for pattern, propSchema := range s.PatternProperties() {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("failed to compile pattern %s: %w", pattern, err)
			}
			propValidator, err := Compile(ctx, propSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to compile pattern property validator for %s: %w", pattern, err)
			}
			patternProperties[re] = propValidator
		}
		v.PatternProperties(patternProperties)
	}
	if s.HasAdditionalProperties() {
		additionalProps := s.AdditionalProperties()
		if additionalProps != nil {
			// Handle SchemaOrBool types
			switch val := additionalProps.(type) {
			case schema.BoolSchema:
				// This is a boolean value
				v.AdditionalPropertiesBool(bool(val))
			case *schema.Schema:
				// This is a regular schema - validate additional properties with this schema
				propValidator, err := Compile(ctx, val)
				if err != nil {
					return nil, fmt.Errorf("failed to compile additional properties validator: %w", err)
				}
				v.AdditionalPropertiesSchema(propValidator)
			default:
				return nil, fmt.Errorf("unexpected additionalProperties type: %T", additionalProps)
			}
		}
	}
	if s.HasPropertyNames() {
		propertyNamesSchema := s.PropertyNames()
		if propertyNamesSchema != nil {
			propertyNamesValidator, err := Compile(ctx, propertyNamesSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to compile property names validator: %w", err)
			}
			v.PropertyNames(propertyNamesValidator)
		}
	}
	if s.HasUnevaluatedProperties() {
		unevaluatedProps := s.UnevaluatedProperties()
		if unevaluatedProps != nil {
			// Handle SchemaOrBool types
			switch val := unevaluatedProps.(type) {
			case schema.BoolSchema:
				// This is a boolean value
				v.UnevaluatedPropertiesBool(bool(val))
			case *schema.Schema:
				// This is a regular schema - validate unevaluated properties with this schema
				propValidator, err := Compile(ctx, val)
				if err != nil {
					return nil, fmt.Errorf("failed to compile unevaluated properties validator: %w", err)
				}
				v.UnevaluatedPropertiesSchema(propValidator)
			default:
				return nil, fmt.Errorf("unexpected unevaluatedProperties type: %T", unevaluatedProps)
			}
		}
	}

	v.StrictObjectType(strictType)

	// Set dependent schemas if available in context (from compilation phase)
	if dependentValidators := DependentSchemasFromContext(ctx); dependentValidators != nil {
		v.DependentSchemas(dependentValidators)
	}

	return v.Build()
}

type objectValidator struct {
	minProperties         *uint
	maxProperties         *uint
	required              []string
	dependentRequired     map[string][]string // dependent required fields
	properties            map[string]Interface
	patternProperties     map[*regexp.Regexp]Interface
	additionalProperties  any // can be bool or Validator
	unevaluatedProperties any // can be bool or Validator
	propertyNames         Interface
	strictObjectType      bool                 // true when schema explicitly declares type: object
	dependentSchemas      map[string]Interface // compiled dependent schema validators
}

type ObjectValidatorBuilder struct {
	err error
	c   *objectValidator
}

func Object() *ObjectValidatorBuilder {
	return (&ObjectValidatorBuilder{}).Reset()
}

func (b *ObjectValidatorBuilder) MinProperties(v uint) *ObjectValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.minProperties = &v
	return b
}

func (b *ObjectValidatorBuilder) MaxProperties(v uint) *ObjectValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.maxProperties = &v
	return b
}

func (b *ObjectValidatorBuilder) Required(v []string) *ObjectValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.required = v
	return b
}

func (b *ObjectValidatorBuilder) DependentRequired(v map[string][]string) *ObjectValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.dependentRequired = v
	return b
}

func (b *ObjectValidatorBuilder) Properties(props ...PropertyPair) *ObjectValidatorBuilder {
	if b.err != nil {
		return b
	}
	if b.c.properties == nil {
		b.c.properties = make(map[string]Interface)
	}
	for _, prop := range props {
		b.c.properties[prop.Name] = prop.Validator
	}
	return b
}


func (b *ObjectValidatorBuilder) PatternProperties(v map[*regexp.Regexp]Interface) *ObjectValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.patternProperties = v
	return b
}

func (b *ObjectValidatorBuilder) AdditionalPropertiesBool(v bool) *ObjectValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.additionalProperties = v
	return b
}

func (b *ObjectValidatorBuilder) AdditionalPropertiesSchema(v Interface) *ObjectValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.additionalProperties = v
	return b
}

func (b *ObjectValidatorBuilder) UnevaluatedPropertiesBool(v bool) *ObjectValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.unevaluatedProperties = v
	return b
}

func (b *ObjectValidatorBuilder) UnevaluatedPropertiesSchema(v Interface) *ObjectValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.unevaluatedProperties = v
	return b
}

func (b *ObjectValidatorBuilder) PropertyNames(v Interface) *ObjectValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.propertyNames = v
	return b
}

func (b *ObjectValidatorBuilder) StrictObjectType(v bool) *ObjectValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.strictObjectType = v
	return b
}

func (b *ObjectValidatorBuilder) DependentSchemas(v map[string]Interface) *ObjectValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.dependentSchemas = v
	return b
}

func (b *ObjectValidatorBuilder) Build() (Interface, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.c, nil
}

func (b *ObjectValidatorBuilder) MustBuild() Interface {
	if b.err != nil {
		panic(b.err)
	}
	return b.c
}

func (b *ObjectValidatorBuilder) Reset() *ObjectValidatorBuilder {
	b.err = nil
	b.c = &objectValidator{}
	return b
}

// Validate implements the Interface
func (c *objectValidator) Validate(ctx context.Context, v any) (Result, error) {
	// Get previously evaluated properties from context
	var previouslyEvaluated map[string]bool
	var contextProps schemactx.EvaluatedProperties
	if err := schemactx.EvaluatedPropertiesFromContext(ctx, &contextProps); err == nil {
		// Convert from map[string]struct{} to map[string]bool
		previouslyEvaluated = make(map[string]bool)
		for prop := range contextProps {
			previouslyEvaluated[prop] = true
		}
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		rv = rv.Elem()
	}

	var properties map[string]any

	switch rv.Kind() {
	case reflect.Map:
		properties = make(map[string]any)
		for _, key := range rv.MapKeys() {
			keyStr := key.String()
			properties[keyStr] = rv.MapIndex(key).Interface()
		}
	case reflect.Struct:
		properties = make(map[string]any)
		t := rv.Type()
		for i := range rv.NumField() {
			field := t.Field(i)
			if field.IsExported() {
				fieldName := field.Name
				if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
					fieldName = jsonTag
				}
				properties[fieldName] = rv.Field(i).Interface()
			}
		}
	default:
		// Handle non-object values based on whether this is strict object type validation
		if c.strictObjectType {
			// When schema explicitly declares type: object, non-object values should fail
			return nil, fmt.Errorf(`invalid value passed to ObjectValidator: expected map or a struct, got %T`, v)
		}
		// For non-object values with inferred object type, object constraints don't apply
		// According to JSON Schema spec, object constraints should be ignored for non-objects
		//nolint: nilnil
		return nil, nil
	}

	// Check minProperties constraint
	if c.minProperties != nil && uint(len(properties)) < *c.minProperties {
		return nil, fmt.Errorf(`invalid value passed to ObjectValidator: object has %d properties, below minimum properties %d`, len(properties), *c.minProperties)
	}

	// Check maxProperties constraint
	if c.maxProperties != nil && uint(len(properties)) > *c.maxProperties {
		return nil, fmt.Errorf(`invalid value passed to ObjectValidator: object has %d properties, exceeds maximum properties %d`, len(properties), *c.maxProperties)
	}

	// Check required properties
	for _, requiredProp := range c.required {
		if _, exists := properties[requiredProp]; !exists {
			return nil, fmt.Errorf(`invalid value passed to ObjectValidator: required property %s is missing`, requiredProp)
		}
	}

	// Check dependent required properties
	for triggerProp, dependentProps := range c.dependentRequired {
		if _, exists := properties[triggerProp]; exists {
			// If the trigger property is present, all dependent properties must be present
			for _, dependentProp := range dependentProps {
				if _, exists := properties[dependentProp]; !exists {
					return nil, fmt.Errorf(`invalid value passed to ObjectValidator: dependent required property %s is missing when %s is present`, dependentProp, triggerProp)
				}
			}
		}
	}

	// Validate property names
	if c.propertyNames != nil {
		for propName := range properties {
			_, err := c.propertyNames.Validate(ctx, propName)
			if err != nil {
				return nil, fmt.Errorf(`invalid value passed to ObjectValidator: property name validation failed for %s: %w`, propName, err)
			}
		}
	}

	// Track evaluated properties for result reporting
	evaluatedProperties := make(map[string]bool)

	// Include previously evaluated properties from earlier validators (e.g., allOf subschemas)
	for prop := range previouslyEvaluated {
		evaluatedProperties[prop] = true
	}

	// Validate properties
	var unevaluatedProps []string
	for propName, propValue := range properties {
		validated := false

		// Check if this property was already evaluated by a previous validator
		if previouslyEvaluated != nil && previouslyEvaluated[propName] {
			validated = true
			evaluatedProperties[propName] = true
		}

		// Check explicit properties
		if c.properties != nil {
			if propValidator, exists := c.properties[propName]; exists {
				_, err := propValidator.Validate(ctx, propValue)
				if err != nil {
					return nil, fmt.Errorf(`invalid value passed to ObjectValidator: property validation failed for %s: %w`, propName, err)
				}
				validated = true
				evaluatedProperties[propName] = true
			}
		}

		// Check pattern properties
		if c.patternProperties != nil {
			for pattern, propValidator := range c.patternProperties {
				if pattern.MatchString(propName) {
					_, err := propValidator.Validate(ctx, propValue)
					if err != nil {
						return nil, fmt.Errorf(`invalid value passed to ObjectValidator: pattern property validation failed for %s: %w`, propName, err)
					}
					validated = true
					evaluatedProperties[propName] = true
				}
			}
		}

		// Check additional properties
		if !validated && c.additionalProperties != nil {
			if boolVal, ok := c.additionalProperties.(bool); ok {
				if !boolVal {
					return nil, fmt.Errorf(`invalid value passed to ObjectValidator: additional property not allowed: %s`, propName)
				}
				// If additionalProperties is true, it means this property is now "evaluated"
				validated = true
				evaluatedProperties[propName] = true
			} else if propValidator, ok := c.additionalProperties.(Interface); ok {
				_, err := propValidator.Validate(ctx, propValue)
				if err != nil {
					return nil, fmt.Errorf(`invalid value passed to ObjectValidator: additional property validation failed for %s: %w`, propName, err)
				}
				// Property was validated by additionalProperties schema, so it's "evaluated"
				validated = true
				evaluatedProperties[propName] = true
			}
		}

		// Track unevaluated properties for later processing
		if !validated {
			unevaluatedProps = append(unevaluatedProps, propName)
		}
	}

	// Handle dependent schemas if stored in this validator (must happen before unevaluated properties)
	if len(c.dependentSchemas) > 0 {
		// Pass the stored dependent schemas through context during execution phase
		depCtx := WithDependentSchemas(ctx, c.dependentSchemas)

		for propertyName, depValidator := range c.dependentSchemas {
			// If the property exists in the object, validate the entire object with the dependent schema
			if _, exists := properties[propertyName]; exists {
				result, err := depValidator.Validate(depCtx, v)
				if err != nil {
					return nil, fmt.Errorf("dependent schema validation failed for property %s: %w", propertyName, err)
				}

				// Merge evaluated properties from dependent schema validation
				if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
					evaluatedProps := objResult.EvaluatedProperties()
					for prop := range evaluatedProps {
						evaluatedProperties[prop] = true
						// Remove from unevaluated list if it was marked as evaluated by dependent schema
						for i, unevalProp := range unevaluatedProps {
							if unevalProp == prop {
								unevaluatedProps = append(unevaluatedProps[:i], unevaluatedProps[i+1:]...)
								break
							}
						}
					}
				}
			}
		}
	}

	// Handle unevaluated properties (after dependent schemas have been processed)
	if len(unevaluatedProps) > 0 && c.unevaluatedProperties != nil {
		for _, propName := range unevaluatedProps {
			propValue := properties[propName]
			if boolVal, ok := c.unevaluatedProperties.(bool); ok {
				if !boolVal {
					return nil, fmt.Errorf(`invalid value passed to ObjectValidator: unevaluated property not allowed: %s`, propName)
				}
				// If unevaluatedProperties is true, mark this property as evaluated
				evaluatedProperties[propName] = true
			} else if propValidator, ok := c.unevaluatedProperties.(Interface); ok {
				_, err := propValidator.Validate(ctx, propValue)
				if err != nil {
					return nil, fmt.Errorf(`invalid value passed to ObjectValidator: unevaluated property validation failed for %s: %w`, propName, err)
				}
				// If property passes unevaluatedProperties schema validation, mark it as evaluated
				evaluatedProperties[propName] = true
			}
		}
	}

	// Always return ObjectResult with evaluated properties information for annotation tracking
	result := NewObjectResult()
	for prop := range evaluatedProperties {
		result.SetEvaluatedProperty(prop)
	}
	return result, nil
}
