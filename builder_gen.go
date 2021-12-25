package schema

import "fmt"

type Builder struct {
	err                   error
	additionalProperties  *Schema
	allOf                 []*Schema
	anchor                *string
	anyOf                 []*Schema
	comment               *string
	constantValue         *interface{}
	contains              *Schema
	definitions           *string
	dynamicReference      *string
	enum                  []interface{}
	exclusiveMaximum      *float64
	exclusiveMinimum      *float64
	id                    *string
	items                 *Schema
	maxContains           *uint
	maxItems              *uint
	maxLength             *int
	maxProperties         *uint
	maximum               *float64
	minContains           *uint
	minItems              *uint
	minLength             *int
	minProperties         *uint
	minimum               *float64
	multipleOf            *float64
	not                   *Schema
	oneOf                 []*Schema
	pattern               *string
	patternProperties     []*propPair
	properties            []*propPair
	propertyNames         []*propPair
	reference             *string
	required              *bool
	schema                string
	types                 []PrimitiveType
	unevaluatedItems      *Schema
	unevaluatedProperties *Schema
	uniqueItems           *bool
}

func NewBuilder() *Builder {
	return &Builder{
		schema: Version,
	}
}

func (b *Builder) AdditionalProperties(v SchemaOrBool) *Builder {
	if b.err != nil {
		return b
	}
	var tmp Schema
	if err := tmp.Accept(v); err != nil {
		b.err = fmt.Errorf(`failed to accept value for "additionalProperties": %w`, err)
		return b
	}
	b.additionalProperties = &tmp
	return b
}

func (b *Builder) AllOf(v []*Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.allOf = v
	return b
}

func (b *Builder) Anchor(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.anchor = &v
	return b
}

func (b *Builder) AnyOf(v []*Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.anyOf = v
	return b
}

func (b *Builder) Comment(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.comment = &v
	return b
}

func (b *Builder) Const(v interface{}) *Builder {
	if b.err != nil {
		return b
	}

	b.constantValue = &v
	return b
}

func (b *Builder) Contains(v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.contains = v
	return b
}

func (b *Builder) Definitions(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.definitions = &v
	return b
}

func (b *Builder) DynamicReference(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.dynamicReference = &v
	return b
}

func (b *Builder) Enum(v []interface{}) *Builder {
	if b.err != nil {
		return b
	}

	b.enum = v
	return b
}

func (b *Builder) ExclusiveMaximum(v float64) *Builder {
	if b.err != nil {
		return b
	}

	b.exclusiveMaximum = &v
	return b
}

func (b *Builder) ExclusiveMinimum(v float64) *Builder {
	if b.err != nil {
		return b
	}

	b.exclusiveMinimum = &v
	return b
}

func (b *Builder) ID(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.id = &v
	return b
}

func (b *Builder) Items(v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.items = v
	return b
}

func (b *Builder) MaxContains(v uint) *Builder {
	if b.err != nil {
		return b
	}

	b.maxContains = &v
	return b
}

func (b *Builder) MaxItems(v uint) *Builder {
	if b.err != nil {
		return b
	}

	b.maxItems = &v
	return b
}

func (b *Builder) MaxLength(v int) *Builder {
	if b.err != nil {
		return b
	}

	b.maxLength = &v
	return b
}

func (b *Builder) MaxProperties(v uint) *Builder {
	if b.err != nil {
		return b
	}

	b.maxProperties = &v
	return b
}

func (b *Builder) Maximum(v float64) *Builder {
	if b.err != nil {
		return b
	}

	b.maximum = &v
	return b
}

func (b *Builder) MinContains(v uint) *Builder {
	if b.err != nil {
		return b
	}

	b.minContains = &v
	return b
}

func (b *Builder) MinItems(v uint) *Builder {
	if b.err != nil {
		return b
	}

	b.minItems = &v
	return b
}

func (b *Builder) MinLength(v int) *Builder {
	if b.err != nil {
		return b
	}

	b.minLength = &v
	return b
}

func (b *Builder) MinProperties(v uint) *Builder {
	if b.err != nil {
		return b
	}

	b.minProperties = &v
	return b
}

func (b *Builder) Minimum(v float64) *Builder {
	if b.err != nil {
		return b
	}

	b.minimum = &v
	return b
}

func (b *Builder) MultipleOf(v float64) *Builder {
	if b.err != nil {
		return b
	}

	b.multipleOf = &v
	return b
}

func (b *Builder) Not(v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.not = v
	return b
}

func (b *Builder) OneOf(v []*Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.oneOf = v
	return b
}

func (b *Builder) Pattern(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.pattern = &v
	return b
}

func (b *Builder) PatternProperty(n string, v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.patternProperties = append(b.patternProperties, &propPair{Name: n, Schema: v})
	return b
}

func (b *Builder) Property(n string, v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.properties = append(b.properties, &propPair{Name: n, Schema: v})
	return b
}

func (b *Builder) PropertyName(n string, v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.propertyNames = append(b.propertyNames, &propPair{Name: n, Schema: v})
	return b
}

func (b *Builder) Reference(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.reference = &v
	return b
}

func (b *Builder) Required(v bool) *Builder {
	if b.err != nil {
		return b
	}

	b.required = &v
	return b
}

func (b *Builder) Schema(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.schema = v
	return b
}

func (b *Builder) Type(v PrimitiveType) *Builder {
	if b.err != nil {
		return b
	}
	b.types = append(b.types, v)
	return b
}

func (b *Builder) UnevaluatedItems(v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.unevaluatedItems = v
	return b
}

func (b *Builder) UnevaluatedProperties(v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.unevaluatedProperties = v
	return b
}

func (b *Builder) UniqueItems(v bool) *Builder {
	if b.err != nil {
		return b
	}

	b.uniqueItems = &v
	return b
}

func (b *Builder) Build() (*Schema, error) {
	s := New()
	if b.additionalProperties != nil {
		s.additionalProperties = b.additionalProperties
	}
	if b.allOf != nil {
		s.allOf = b.allOf
	}
	if b.anchor != nil {
		s.anchor = b.anchor
	}
	if b.anyOf != nil {
		s.anyOf = b.anyOf
	}
	if b.comment != nil {
		s.comment = b.comment
	}
	if b.constantValue != nil {
		s.constantValue = b.constantValue
	}
	if b.contains != nil {
		s.contains = b.contains
	}
	if b.definitions != nil {
		s.definitions = b.definitions
	}
	if b.dynamicReference != nil {
		s.dynamicReference = b.dynamicReference
	}
	if b.enum != nil {
		s.enum = b.enum
	}
	if b.exclusiveMaximum != nil {
		s.exclusiveMaximum = b.exclusiveMaximum
	}
	if b.exclusiveMinimum != nil {
		s.exclusiveMinimum = b.exclusiveMinimum
	}
	if b.id != nil {
		s.id = b.id
	}
	if b.items != nil {
		s.items = b.items
	}
	if b.maxContains != nil {
		s.maxContains = b.maxContains
	}
	if b.maxItems != nil {
		s.maxItems = b.maxItems
	}
	if b.maxLength != nil {
		s.maxLength = b.maxLength
	}
	if b.maxProperties != nil {
		s.maxProperties = b.maxProperties
	}
	if b.maximum != nil {
		s.maximum = b.maximum
	}
	if b.minContains != nil {
		s.minContains = b.minContains
	}
	if b.minItems != nil {
		s.minItems = b.minItems
	}
	if b.minLength != nil {
		s.minLength = b.minLength
	}
	if b.minProperties != nil {
		s.minProperties = b.minProperties
	}
	if b.minimum != nil {
		s.minimum = b.minimum
	}
	if b.multipleOf != nil {
		s.multipleOf = b.multipleOf
	}
	if b.not != nil {
		s.not = b.not
	}
	if b.oneOf != nil {
		s.oneOf = b.oneOf
	}
	if b.pattern != nil {
		s.pattern = b.pattern
	}

	if b.patternProperties != nil {
		s.patternProperties = make(map[string]*Schema)
		for _, pair := range b.patternProperties {
			if _, ok := s.patternProperties[pair.Name]; ok {
				return nil, fmt.Errorf(`duplicate key %q in "patternProperties"`, pair.Name)
			}
			s.patternProperties[pair.Name] = pair.Schema
		}
	}

	if b.properties != nil {
		s.properties = make(map[string]*Schema)
		for _, pair := range b.properties {
			if _, ok := s.properties[pair.Name]; ok {
				return nil, fmt.Errorf(`duplicate key %q in "properties"`, pair.Name)
			}
			s.properties[pair.Name] = pair.Schema
		}
	}

	if b.propertyNames != nil {
		s.propertyNames = make(map[string]*Schema)
		for _, pair := range b.propertyNames {
			if _, ok := s.propertyNames[pair.Name]; ok {
				return nil, fmt.Errorf(`duplicate key %q in "propertyNames"`, pair.Name)
			}
			s.propertyNames[pair.Name] = pair.Schema
		}
	}
	if b.reference != nil {
		s.reference = b.reference
	}
	if b.required != nil {
		s.required = b.required
	}
	s.schema = b.schema
	if b.types != nil {
		s.types = b.types
	}
	if b.unevaluatedItems != nil {
		s.unevaluatedItems = b.unevaluatedItems
	}
	if b.unevaluatedProperties != nil {
		s.unevaluatedProperties = b.unevaluatedProperties
	}
	if b.uniqueItems != nil {
		s.uniqueItems = b.uniqueItems
	}
	return s, nil
}
