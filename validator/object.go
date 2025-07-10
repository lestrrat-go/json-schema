package validator

import (
	"fmt"
	"reflect"
	"regexp"

	schema "github.com/lestrrat-go/json-schema"
)

var _ Builder = (*ObjectValidatorBuilder)(nil)
var _ Interface = (*objectValidator)(nil)

func compileObjectValidator(s *schema.Schema) (Interface, error) {
	v := Object()

	if s.HasMinProperties() {
		v.MinProperties(s.MinProperties())
	}
	if s.HasMaxProperties() {
		v.MaxProperties(s.MaxProperties())
	}
	if s.HasRequired() {
		v.Required(s.Required())
	}
	if s.HasProperties() {
		properties := make(map[string]Interface)
		for name, propSchema := range s.Properties() {
			propValidator, err := Compile(propSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to compile property validator for %s: %w", name, err)
			}
			properties[name] = propValidator
		}
		v.Properties(properties)
	}
	if s.HasPatternProperties() {
		patternProperties := make(map[*regexp.Regexp]Interface)
		for pattern, propSchema := range s.PatternProperties() {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("failed to compile pattern %s: %w", pattern, err)
			}
			propValidator, err := Compile(propSchema)
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
			// Check if it's a boolean schema (has not field) or a regular schema
			if additionalProps.HasNot() && additionalProps.Not() != nil {
				// This represents `false` - disallow additional properties
				v.AdditionalPropertiesBool(false)
			} else if !additionalProps.HasNot() && len(additionalProps.Types()) == 0 {
				// This represents `true` - allow additional properties
				v.AdditionalPropertiesBool(true)
			} else {
				// This is a regular schema - validate additional properties with this schema
				propValidator, err := Compile(additionalProps)
				if err != nil {
					return nil, fmt.Errorf("failed to compile additional properties validator: %w", err)
				}
				v.AdditionalPropertiesSchema(propValidator)
			}
		}
	}
	if s.HasPropertyNames() {
		propertyNamesSchema := s.PropertyNames()
		if propertyNamesSchema != nil {
			propertyNamesValidator, err := Compile(propertyNamesSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to compile property names validator: %w", err)
			}
			v.PropertyNames(propertyNamesValidator)
		}
	}

	return v.Build()
}

type objectValidator struct {
	minProperties        *uint
	maxProperties        *uint
	required             []string
	properties           map[string]Interface
	patternProperties    map[*regexp.Regexp]Interface
	additionalProperties any // can be bool or Validator
	propertyNames        Interface
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

func (b *ObjectValidatorBuilder) Properties(v map[string]Interface) *ObjectValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.properties = v
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

func (b *ObjectValidatorBuilder) PropertyNames(v Interface) *ObjectValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.propertyNames = v
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

func (c *objectValidator) Validate(v any) error {
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
		for i := 0; i < rv.NumField(); i++ {
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
		return fmt.Errorf(`invalid value passed to ObjectValidator: expected map or a struct, got %T`, v)
	}

	// Check minProperties constraint
	if c.minProperties != nil && uint(len(properties)) < *c.minProperties {
		return fmt.Errorf(`invalid value passed to ObjectValidator: object has %d properties, below minimum properties %d`, len(properties), *c.minProperties)
	}

	// Check maxProperties constraint
	if c.maxProperties != nil && uint(len(properties)) > *c.maxProperties {
		return fmt.Errorf(`invalid value passed to ObjectValidator: object has %d properties, exceeds maximum properties %d`, len(properties), *c.maxProperties)
	}

	// Check required properties
	for _, requiredProp := range c.required {
		if val, exists := properties[requiredProp]; !exists || val == nil {
			return fmt.Errorf(`invalid value passed to ObjectValidator: required property %s is missing`, requiredProp)
		}
	}

	// Validate property names
	if c.propertyNames != nil {
		for propName := range properties {
			if err := c.propertyNames.Validate(propName); err != nil {
				return fmt.Errorf(`invalid value passed to ObjectValidator: property name validation failed for %s: %w`, propName, err)
			}
		}
	}

	// Validate properties
	for propName, propValue := range properties {
		validated := false

		// Check explicit properties
		if c.properties != nil {
			if propValidator, exists := c.properties[propName]; exists {
				if err := propValidator.Validate(propValue); err != nil {
					return fmt.Errorf(`invalid value passed to ObjectValidator: property validation failed for %s: %w`, propName, err)
				}
				validated = true
			}
		}

		// Check pattern properties
		if c.patternProperties != nil {
			for pattern, propValidator := range c.patternProperties {
				if pattern.MatchString(propName) {
					if err := propValidator.Validate(propValue); err != nil {
						return fmt.Errorf(`invalid value passed to ObjectValidator: pattern property validation failed for %s: %w`, propName, err)
					}
					validated = true
				}
			}
		}

		// Check additional properties
		if !validated && c.additionalProperties != nil {
			if boolVal, ok := c.additionalProperties.(bool); ok {
				if !boolVal {
					return fmt.Errorf(`invalid value passed to ObjectValidator: additional property not allowed: %s`, propName)
				}
			} else if propValidator, ok := c.additionalProperties.(Interface); ok {
				if err := propValidator.Validate(propValue); err != nil {
					return fmt.Errorf(`invalid value passed to ObjectValidator: additional property validation failed for %s: %w`, propName, err)
				}
			}
		}
	}

	return nil
}
