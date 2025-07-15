package schema

import "fmt"

type Builder struct {
	err                   error
	additionalProperties  SchemaOrBool
	allOf                 []SchemaOrBool
	anchor                *string
	anyOf                 []SchemaOrBool
	comment               *string
	constantValue         *interface{}
	contains              *Schema
	contentEncoding       *string
	contentMediaType      *string
	contentSchema         *Schema
	definitions           []*propPair
	dependentSchemas      []*propPair
	dynamicReference      *string
	elseSchema            *Schema
	enum                  []interface{}
	exclusiveMaximum      *float64
	exclusiveMinimum      *float64
	format                *string
	id                    *string
	ifSchema              *Schema
	items                 SchemaOrBool
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
	oneOf                 []SchemaOrBool
	pattern               *string
	patternProperties     []*propPair
	properties            []*propPair
	propertyNames         *Schema
	reference             *string
	required              []string
	schema                string
	thenSchema            *Schema
	types                 PrimitiveTypes
	unevaluatedItems      SchemaOrBool
	unevaluatedProperties SchemaOrBool
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
	b.additionalProperties = v
	return b
}

func (b *Builder) AllOf(v ...SchemaOrBool) *Builder {
	if b.err != nil {
		return b
	}

	for _, item := range v {
		if err := validateSchemaOrBool(item); err != nil {
			b.err = fmt.Errorf(`invalid value in AllOf: %w`, err)
			return b
		}
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

func (b *Builder) AnyOf(v ...SchemaOrBool) *Builder {
	if b.err != nil {
		return b
	}

	for _, item := range v {
		if err := validateSchemaOrBool(item); err != nil {
			b.err = fmt.Errorf(`invalid value in AnyOf: %w`, err)
			return b
		}
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

func (b *Builder) ContentEncoding(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.contentEncoding = &v
	return b
}

func (b *Builder) ContentMediaType(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.contentMediaType = &v
	return b
}

func (b *Builder) ContentSchema(v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.contentSchema = v
	return b
}

func (b *Builder) Definitions(n string, v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.definitions = append(b.definitions, &propPair{Name: n, Schema: v})
	return b
}

func (b *Builder) DependentSchemas(n string, v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.dependentSchemas = append(b.dependentSchemas, &propPair{Name: n, Schema: v})
	return b
}

func (b *Builder) DynamicReference(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.dynamicReference = &v
	return b
}

func (b *Builder) ElseSchema(v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.elseSchema = v
	return b
}

func (b *Builder) Enum(v ...interface{}) *Builder {
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

func (b *Builder) Format(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.format = &v
	return b
}

func (b *Builder) ID(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.id = &v
	return b
}

func (b *Builder) IfSchema(v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.ifSchema = v
	return b
}

func (b *Builder) Items(v SchemaOrBool) *Builder {
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

func (b *Builder) OneOf(v ...SchemaOrBool) *Builder {
	if b.err != nil {
		return b
	}

	for _, item := range v {
		if err := validateSchemaOrBool(item); err != nil {
			b.err = fmt.Errorf(`invalid value in OneOf: %w`, err)
			return b
		}
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

func (b *Builder) PropertyNames(v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.propertyNames = v
	return b
}

func (b *Builder) Reference(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.reference = &v
	return b
}

func (b *Builder) Required(v ...string) *Builder {
	if b.err != nil {
		return b
	}

	b.required = v
	return b
}

func (b *Builder) Schema(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.schema = v
	return b
}

func (b *Builder) ThenSchema(v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.thenSchema = v
	return b
}

func (b *Builder) Types(v ...PrimitiveType) *Builder {
	if b.err != nil {
		return b
	}

	b.types = PrimitiveTypes(v)
	return b
}

func (b *Builder) UnevaluatedItems(v SchemaOrBool) *Builder {
	if b.err != nil {
		return b
	}
	b.unevaluatedItems = v
	return b
}

func (b *Builder) UnevaluatedProperties(v SchemaOrBool) *Builder {
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

func (b *Builder) Clone(original *Schema) *Builder {
	if b.err != nil {
		return b
	}
	if original == nil {
		return b
	}

	if original.additionalProperties != nil {
		b.additionalProperties = original.additionalProperties
	}

	if original.allOf != nil {
		b.allOf = original.allOf
	}

	if original.anchor != nil {
		b.anchor = original.anchor
	}

	if original.anyOf != nil {
		b.anyOf = original.anyOf
	}

	if original.comment != nil {
		b.comment = original.comment
	}

	if original.constantValue != nil {
		b.constantValue = original.constantValue
	}

	if original.contains != nil {
		b.contains = original.contains
	}

	if original.contentEncoding != nil {
		b.contentEncoding = original.contentEncoding
	}

	if original.contentMediaType != nil {
		b.contentMediaType = original.contentMediaType
	}

	if original.contentSchema != nil {
		b.contentSchema = original.contentSchema
	}

	if original.definitions != nil {
		for name, schema := range original.definitions {
			b.definitions = append(b.definitions, &propPair{Name: name, Schema: schema})
		}
	}

	if original.dependentSchemas != nil {
		for name, schema := range original.dependentSchemas {
			b.dependentSchemas = append(b.dependentSchemas, &propPair{Name: name, Schema: schema})
		}
	}

	if original.dynamicReference != nil {
		b.dynamicReference = original.dynamicReference
	}

	if original.elseSchema != nil {
		b.elseSchema = original.elseSchema
	}

	if original.enum != nil {
		b.enum = original.enum
	}

	if original.exclusiveMaximum != nil {
		b.exclusiveMaximum = original.exclusiveMaximum
	}

	if original.exclusiveMinimum != nil {
		b.exclusiveMinimum = original.exclusiveMinimum
	}

	if original.format != nil {
		b.format = original.format
	}

	if original.id != nil {
		b.id = original.id
	}

	if original.ifSchema != nil {
		b.ifSchema = original.ifSchema
	}

	if original.items != nil {
		b.items = original.items
	}

	if original.maxContains != nil {
		b.maxContains = original.maxContains
	}

	if original.maxItems != nil {
		b.maxItems = original.maxItems
	}

	if original.maxLength != nil {
		b.maxLength = original.maxLength
	}

	if original.maxProperties != nil {
		b.maxProperties = original.maxProperties
	}

	if original.maximum != nil {
		b.maximum = original.maximum
	}

	if original.minContains != nil {
		b.minContains = original.minContains
	}

	if original.minItems != nil {
		b.minItems = original.minItems
	}

	if original.minLength != nil {
		b.minLength = original.minLength
	}

	if original.minProperties != nil {
		b.minProperties = original.minProperties
	}

	if original.minimum != nil {
		b.minimum = original.minimum
	}

	if original.multipleOf != nil {
		b.multipleOf = original.multipleOf
	}

	if original.not != nil {
		b.not = original.not
	}

	if original.oneOf != nil {
		b.oneOf = original.oneOf
	}

	if original.pattern != nil {
		b.pattern = original.pattern
	}

	if original.patternProperties != nil {
		for name, schema := range original.patternProperties {
			b.patternProperties = append(b.patternProperties, &propPair{Name: name, Schema: schema})
		}
	}

	if original.properties != nil {
		for name, schema := range original.properties {
			b.properties = append(b.properties, &propPair{Name: name, Schema: schema})
		}
	}

	if original.propertyNames != nil {
		b.propertyNames = original.propertyNames
	}

	if original.reference != nil {
		b.reference = original.reference
	}

	if original.required != nil {
		b.required = original.required
	}

	b.schema = original.schema

	if original.thenSchema != nil {
		b.thenSchema = original.thenSchema
	}

	if original.types != nil {
		b.types = original.types
	}

	if original.unevaluatedItems != nil {
		b.unevaluatedItems = original.unevaluatedItems
	}

	if original.unevaluatedProperties != nil {
		b.unevaluatedProperties = original.unevaluatedProperties
	}

	if original.uniqueItems != nil {
		b.uniqueItems = original.uniqueItems
	}
	return b
}

func (b *Builder) ResetAdditionalProperties() *Builder {
	if b.err != nil {
		return b
	}
	b.additionalProperties = nil
	return b
}

func (b *Builder) ResetAllOf() *Builder {
	if b.err != nil {
		return b
	}
	b.allOf = nil
	return b
}

func (b *Builder) ResetAnchor() *Builder {
	if b.err != nil {
		return b
	}
	b.anchor = nil
	return b
}

func (b *Builder) ResetAnyOf() *Builder {
	if b.err != nil {
		return b
	}
	b.anyOf = nil
	return b
}

func (b *Builder) ResetComment() *Builder {
	if b.err != nil {
		return b
	}
	b.comment = nil
	return b
}

func (b *Builder) ResetConst() *Builder {
	if b.err != nil {
		return b
	}
	b.constantValue = nil
	return b
}

func (b *Builder) ResetContains() *Builder {
	if b.err != nil {
		return b
	}
	b.contains = nil
	return b
}

func (b *Builder) ResetContentEncoding() *Builder {
	if b.err != nil {
		return b
	}
	b.contentEncoding = nil
	return b
}

func (b *Builder) ResetContentMediaType() *Builder {
	if b.err != nil {
		return b
	}
	b.contentMediaType = nil
	return b
}

func (b *Builder) ResetContentSchema() *Builder {
	if b.err != nil {
		return b
	}
	b.contentSchema = nil
	return b
}

func (b *Builder) ResetDefinitions() *Builder {
	if b.err != nil {
		return b
	}
	b.definitions = nil
	return b
}

func (b *Builder) ResetDependentSchemas() *Builder {
	if b.err != nil {
		return b
	}
	b.dependentSchemas = nil
	return b
}

func (b *Builder) ResetDynamicReference() *Builder {
	if b.err != nil {
		return b
	}
	b.dynamicReference = nil
	return b
}

func (b *Builder) ResetElseSchema() *Builder {
	if b.err != nil {
		return b
	}
	b.elseSchema = nil
	return b
}

func (b *Builder) ResetEnum() *Builder {
	if b.err != nil {
		return b
	}
	b.enum = nil
	return b
}

func (b *Builder) ResetExclusiveMaximum() *Builder {
	if b.err != nil {
		return b
	}
	b.exclusiveMaximum = nil
	return b
}

func (b *Builder) ResetExclusiveMinimum() *Builder {
	if b.err != nil {
		return b
	}
	b.exclusiveMinimum = nil
	return b
}

func (b *Builder) ResetFormat() *Builder {
	if b.err != nil {
		return b
	}
	b.format = nil
	return b
}

func (b *Builder) ResetID() *Builder {
	if b.err != nil {
		return b
	}
	b.id = nil
	return b
}

func (b *Builder) ResetIfSchema() *Builder {
	if b.err != nil {
		return b
	}
	b.ifSchema = nil
	return b
}

func (b *Builder) ResetItems() *Builder {
	if b.err != nil {
		return b
	}
	b.items = nil
	return b
}

func (b *Builder) ResetMaxContains() *Builder {
	if b.err != nil {
		return b
	}
	b.maxContains = nil
	return b
}

func (b *Builder) ResetMaxItems() *Builder {
	if b.err != nil {
		return b
	}
	b.maxItems = nil
	return b
}

func (b *Builder) ResetMaxLength() *Builder {
	if b.err != nil {
		return b
	}
	b.maxLength = nil
	return b
}

func (b *Builder) ResetMaxProperties() *Builder {
	if b.err != nil {
		return b
	}
	b.maxProperties = nil
	return b
}

func (b *Builder) ResetMaximum() *Builder {
	if b.err != nil {
		return b
	}
	b.maximum = nil
	return b
}

func (b *Builder) ResetMinContains() *Builder {
	if b.err != nil {
		return b
	}
	b.minContains = nil
	return b
}

func (b *Builder) ResetMinItems() *Builder {
	if b.err != nil {
		return b
	}
	b.minItems = nil
	return b
}

func (b *Builder) ResetMinLength() *Builder {
	if b.err != nil {
		return b
	}
	b.minLength = nil
	return b
}

func (b *Builder) ResetMinProperties() *Builder {
	if b.err != nil {
		return b
	}
	b.minProperties = nil
	return b
}

func (b *Builder) ResetMinimum() *Builder {
	if b.err != nil {
		return b
	}
	b.minimum = nil
	return b
}

func (b *Builder) ResetMultipleOf() *Builder {
	if b.err != nil {
		return b
	}
	b.multipleOf = nil
	return b
}

func (b *Builder) ResetNot() *Builder {
	if b.err != nil {
		return b
	}
	b.not = nil
	return b
}

func (b *Builder) ResetOneOf() *Builder {
	if b.err != nil {
		return b
	}
	b.oneOf = nil
	return b
}

func (b *Builder) ResetPattern() *Builder {
	if b.err != nil {
		return b
	}
	b.pattern = nil
	return b
}

func (b *Builder) ResetPatternProperties() *Builder {
	if b.err != nil {
		return b
	}
	b.patternProperties = nil
	return b
}

func (b *Builder) ResetProperties() *Builder {
	if b.err != nil {
		return b
	}
	b.properties = nil
	return b
}

func (b *Builder) ResetPropertyNames() *Builder {
	if b.err != nil {
		return b
	}
	b.propertyNames = nil
	return b
}

func (b *Builder) ResetReference() *Builder {
	if b.err != nil {
		return b
	}
	b.reference = nil
	return b
}

func (b *Builder) ResetRequired() *Builder {
	if b.err != nil {
		return b
	}
	b.required = nil
	return b
}

func (b *Builder) ResetSchema() *Builder {
	if b.err != nil {
		return b
	}
	b.schema = Version
	return b
}

func (b *Builder) ResetThenSchema() *Builder {
	if b.err != nil {
		return b
	}
	b.thenSchema = nil
	return b
}

func (b *Builder) ResetTypes() *Builder {
	if b.err != nil {
		return b
	}
	b.types = nil
	return b
}

func (b *Builder) ResetUnevaluatedItems() *Builder {
	if b.err != nil {
		return b
	}
	b.unevaluatedItems = nil
	return b
}

func (b *Builder) ResetUnevaluatedProperties() *Builder {
	if b.err != nil {
		return b
	}
	b.unevaluatedProperties = nil
	return b
}

func (b *Builder) ResetUniqueItems() *Builder {
	if b.err != nil {
		return b
	}
	b.uniqueItems = nil
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
	if b.contentEncoding != nil {
		s.contentEncoding = b.contentEncoding
	}
	if b.contentMediaType != nil {
		s.contentMediaType = b.contentMediaType
	}
	if b.contentSchema != nil {
		s.contentSchema = b.contentSchema
	}

	if b.definitions != nil {
		s.definitions = make(map[string]*Schema)
		for _, pair := range b.definitions {
			if _, ok := s.definitions[pair.Name]; ok {
				return nil, fmt.Errorf(`duplicate key %q in "$defs"`, pair.Name)
			}
			s.definitions[pair.Name] = pair.Schema
		}
	}

	if b.dependentSchemas != nil {
		s.dependentSchemas = make(map[string]*Schema)
		for _, pair := range b.dependentSchemas {
			if _, ok := s.dependentSchemas[pair.Name]; ok {
				return nil, fmt.Errorf(`duplicate key %q in "dependentSchemas"`, pair.Name)
			}
			s.dependentSchemas[pair.Name] = pair.Schema
		}
	}
	if b.dynamicReference != nil {
		s.dynamicReference = b.dynamicReference
	}
	if b.elseSchema != nil {
		s.elseSchema = b.elseSchema
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
	if b.format != nil {
		s.format = b.format
	}
	if b.id != nil {
		s.id = b.id
	}
	if b.ifSchema != nil {
		s.ifSchema = b.ifSchema
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
		s.propertyNames = b.propertyNames
	}
	if b.reference != nil {
		s.reference = b.reference
	}
	if b.required != nil {
		s.required = b.required
	}
	s.schema = b.schema
	if b.thenSchema != nil {
		s.thenSchema = b.thenSchema
	}
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

func (b *Builder) MustBuild() *Schema {
	s, err := b.Build()
	if err != nil {
		panic(fmt.Errorf(`failed to build schema: %w`, err))
	}
	return s
}
