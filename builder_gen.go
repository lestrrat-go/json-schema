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
	contains              SchemaOrBool
	contentEncoding       *string
	contentMediaType      *string
	contentSchema         *Schema
	defaultValue          *interface{}
	definitions           []*propPair
	dependentRequired     map[string][]string
	dependentSchemas      map[string]SchemaOrBool
	dynamicAnchor         *string
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
	vocabulary            map[string]bool
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

func (b *Builder) Contains(v SchemaOrBool) *Builder {
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

func (b *Builder) Default(v interface{}) *Builder {
	if b.err != nil {
		return b
	}

	b.defaultValue = &v
	return b
}

func (b *Builder) Definitions(n string, v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.definitions = append(b.definitions, &propPair{Name: n, Schema: v})
	return b
}

func (b *Builder) DependentRequired(v map[string][]string) *Builder {
	if b.err != nil {
		return b
	}

	b.dependentRequired = v
	return b
}

func (b *Builder) DependentSchemas(v map[string]SchemaOrBool) *Builder {
	if b.err != nil {
		return b
	}

	b.dependentSchemas = v
	return b
}

func (b *Builder) DynamicAnchor(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.dynamicAnchor = &v
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

func (b *Builder) Vocabulary(v map[string]bool) *Builder {
	if b.err != nil {
		return b
	}

	b.vocabulary = v
	return b
}

func (b *Builder) Clone(original *Schema) *Builder {
	if b.err != nil {
		return b
	}
	if original == nil {
		return b
	}

	if original.HasAdditionalProperties() {
		b.additionalProperties = original.additionalProperties
	}

	if original.HasAllOf() {
		b.allOf = original.allOf
	}

	if original.HasAnchor() {
		b.anchor = original.anchor
	}

	if original.HasAnyOf() {
		b.anyOf = original.anyOf
	}

	if original.HasComment() {
		b.comment = original.comment
	}

	if original.HasConst() {
		b.constantValue = original.constantValue
	}

	if original.HasContains() {
		b.contains = original.contains
	}

	if original.HasContentEncoding() {
		b.contentEncoding = original.contentEncoding
	}

	if original.HasContentMediaType() {
		b.contentMediaType = original.contentMediaType
	}

	if original.HasContentSchema() {
		b.contentSchema = original.contentSchema
	}

	if original.HasDefault() {
		b.defaultValue = original.defaultValue
	}

	if original.HasDefinitions() {
		for name, schema := range original.definitions {
			b.definitions = append(b.definitions, &propPair{Name: name, Schema: schema})
		}
	}

	if original.HasDependentRequired() {
		b.dependentRequired = original.dependentRequired
	}

	if original.HasDependentSchemas() {
		b.dependentSchemas = original.dependentSchemas
	}

	if original.HasDynamicAnchor() {
		b.dynamicAnchor = original.dynamicAnchor
	}

	if original.HasDynamicReference() {
		b.dynamicReference = original.dynamicReference
	}

	if original.HasElseSchema() {
		b.elseSchema = original.elseSchema
	}

	if original.HasEnum() {
		b.enum = original.enum
	}

	if original.HasExclusiveMaximum() {
		b.exclusiveMaximum = original.exclusiveMaximum
	}

	if original.HasExclusiveMinimum() {
		b.exclusiveMinimum = original.exclusiveMinimum
	}

	if original.HasFormat() {
		b.format = original.format
	}

	if original.HasID() {
		b.id = original.id
	}

	if original.HasIfSchema() {
		b.ifSchema = original.ifSchema
	}

	if original.HasItems() {
		b.items = original.items
	}

	if original.HasMaxContains() {
		b.maxContains = original.maxContains
	}

	if original.HasMaxItems() {
		b.maxItems = original.maxItems
	}

	if original.HasMaxLength() {
		b.maxLength = original.maxLength
	}

	if original.HasMaxProperties() {
		b.maxProperties = original.maxProperties
	}

	if original.HasMaximum() {
		b.maximum = original.maximum
	}

	if original.HasMinContains() {
		b.minContains = original.minContains
	}

	if original.HasMinItems() {
		b.minItems = original.minItems
	}

	if original.HasMinLength() {
		b.minLength = original.minLength
	}

	if original.HasMinProperties() {
		b.minProperties = original.minProperties
	}

	if original.HasMinimum() {
		b.minimum = original.minimum
	}

	if original.HasMultipleOf() {
		b.multipleOf = original.multipleOf
	}

	if original.HasNot() {
		b.not = original.not
	}

	if original.HasOneOf() {
		b.oneOf = original.oneOf
	}

	if original.HasPattern() {
		b.pattern = original.pattern
	}

	if original.HasPatternProperties() {
		for name, schema := range original.patternProperties {
			b.patternProperties = append(b.patternProperties, &propPair{Name: name, Schema: schema})
		}
	}

	if original.HasProperties() {
		for name, schema := range original.properties {
			b.properties = append(b.properties, &propPair{Name: name, Schema: schema})
		}
	}

	if original.HasPropertyNames() {
		b.propertyNames = original.propertyNames
	}

	if original.HasReference() {
		b.reference = original.reference
	}

	if original.HasRequired() {
		b.required = original.required
	}

	b.schema = original.schema

	if original.HasThenSchema() {
		b.thenSchema = original.thenSchema
	}

	if original.HasTypes() {
		b.types = original.types
	}

	if original.HasUnevaluatedItems() {
		b.unevaluatedItems = original.unevaluatedItems
	}

	if original.HasUnevaluatedProperties() {
		b.unevaluatedProperties = original.unevaluatedProperties
	}

	if original.HasUniqueItems() {
		b.uniqueItems = original.uniqueItems
	}

	if original.HasVocabulary() {
		b.vocabulary = original.vocabulary
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

func (b *Builder) ResetDefault() *Builder {
	if b.err != nil {
		return b
	}
	b.defaultValue = nil
	return b
}

func (b *Builder) ResetDefinitions() *Builder {
	if b.err != nil {
		return b
	}
	b.definitions = nil
	return b
}

func (b *Builder) ResetDependentRequired() *Builder {
	if b.err != nil {
		return b
	}
	b.dependentRequired = nil
	return b
}

func (b *Builder) ResetDependentSchemas() *Builder {
	if b.err != nil {
		return b
	}
	b.dependentSchemas = nil
	return b
}

func (b *Builder) ResetDynamicAnchor() *Builder {
	if b.err != nil {
		return b
	}
	b.dynamicAnchor = nil
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

func (b *Builder) ResetVocabulary() *Builder {
	if b.err != nil {
		return b
	}
	b.vocabulary = nil
	return b
}

func (b *Builder) Build() (*Schema, error) {
	s := New()
	if b.additionalProperties != nil {
		s.additionalProperties = b.additionalProperties
		s.populatedFields |= AdditionalPropertiesField
	}
	if b.allOf != nil {
		s.allOf = b.allOf
		s.populatedFields |= AllOfField
	}
	if b.anchor != nil {
		s.anchor = b.anchor
		s.populatedFields |= AnchorField
	}
	if b.anyOf != nil {
		s.anyOf = b.anyOf
		s.populatedFields |= AnyOfField
	}
	if b.comment != nil {
		s.comment = b.comment
		s.populatedFields |= CommentField
	}
	if b.constantValue != nil {
		s.constantValue = b.constantValue
		s.populatedFields |= ConstField
	}
	if b.contains != nil {
		s.contains = b.contains
		s.populatedFields |= ContainsField
	}
	if b.contentEncoding != nil {
		s.contentEncoding = b.contentEncoding
		s.populatedFields |= ContentEncodingField
	}
	if b.contentMediaType != nil {
		s.contentMediaType = b.contentMediaType
		s.populatedFields |= ContentMediaTypeField
	}
	if b.contentSchema != nil {
		s.contentSchema = b.contentSchema
		s.populatedFields |= ContentSchemaField
	}
	if b.defaultValue != nil {
		s.defaultValue = b.defaultValue
		s.populatedFields |= DefaultField
	}

	if b.definitions != nil {
		s.definitions = make(map[string]*Schema)
		for _, pair := range b.definitions {
			if _, ok := s.definitions[pair.Name]; ok {
				return nil, fmt.Errorf(`duplicate key %q in "$defs"`, pair.Name)
			}
			s.definitions[pair.Name] = pair.Schema
		}
		s.populatedFields |= DefinitionsField
	}
	if b.dependentRequired != nil {
		s.dependentRequired = b.dependentRequired
		s.populatedFields |= DependentRequiredField
	}
	if b.dependentSchemas != nil {
		s.dependentSchemas = b.dependentSchemas
		s.populatedFields |= DependentSchemasField
	}
	if b.dynamicAnchor != nil {
		s.dynamicAnchor = b.dynamicAnchor
		s.populatedFields |= DynamicAnchorField
	}
	if b.dynamicReference != nil {
		s.dynamicReference = b.dynamicReference
		s.populatedFields |= DynamicReferenceField
	}
	if b.elseSchema != nil {
		s.elseSchema = b.elseSchema
		s.populatedFields |= ElseSchemaField
	}
	if b.enum != nil {
		s.enum = b.enum
		s.populatedFields |= EnumField
	}
	if b.exclusiveMaximum != nil {
		s.exclusiveMaximum = b.exclusiveMaximum
		s.populatedFields |= ExclusiveMaximumField
	}
	if b.exclusiveMinimum != nil {
		s.exclusiveMinimum = b.exclusiveMinimum
		s.populatedFields |= ExclusiveMinimumField
	}
	if b.format != nil {
		s.format = b.format
		s.populatedFields |= FormatField
	}
	if b.id != nil {
		s.id = b.id
		s.populatedFields |= IDField
	}
	if b.ifSchema != nil {
		s.ifSchema = b.ifSchema
		s.populatedFields |= IfSchemaField
	}
	if b.items != nil {
		s.items = b.items
		s.populatedFields |= ItemsField
	}
	if b.maxContains != nil {
		s.maxContains = b.maxContains
		s.populatedFields |= MaxContainsField
	}
	if b.maxItems != nil {
		s.maxItems = b.maxItems
		s.populatedFields |= MaxItemsField
	}
	if b.maxLength != nil {
		s.maxLength = b.maxLength
		s.populatedFields |= MaxLengthField
	}
	if b.maxProperties != nil {
		s.maxProperties = b.maxProperties
		s.populatedFields |= MaxPropertiesField
	}
	if b.maximum != nil {
		s.maximum = b.maximum
		s.populatedFields |= MaximumField
	}
	if b.minContains != nil {
		s.minContains = b.minContains
		s.populatedFields |= MinContainsField
	}
	if b.minItems != nil {
		s.minItems = b.minItems
		s.populatedFields |= MinItemsField
	}
	if b.minLength != nil {
		s.minLength = b.minLength
		s.populatedFields |= MinLengthField
	}
	if b.minProperties != nil {
		s.minProperties = b.minProperties
		s.populatedFields |= MinPropertiesField
	}
	if b.minimum != nil {
		s.minimum = b.minimum
		s.populatedFields |= MinimumField
	}
	if b.multipleOf != nil {
		s.multipleOf = b.multipleOf
		s.populatedFields |= MultipleOfField
	}
	if b.not != nil {
		s.not = b.not
		s.populatedFields |= NotField
	}
	if b.oneOf != nil {
		s.oneOf = b.oneOf
		s.populatedFields |= OneOfField
	}
	if b.pattern != nil {
		s.pattern = b.pattern
		s.populatedFields |= PatternField
	}

	if b.patternProperties != nil {
		s.patternProperties = make(map[string]*Schema)
		for _, pair := range b.patternProperties {
			if _, ok := s.patternProperties[pair.Name]; ok {
				return nil, fmt.Errorf(`duplicate key %q in "patternProperties"`, pair.Name)
			}
			s.patternProperties[pair.Name] = pair.Schema
		}
		s.populatedFields |= PatternPropertiesField
	}

	if b.properties != nil {
		s.properties = make(map[string]*Schema)
		for _, pair := range b.properties {
			if _, ok := s.properties[pair.Name]; ok {
				return nil, fmt.Errorf(`duplicate key %q in "properties"`, pair.Name)
			}
			s.properties[pair.Name] = pair.Schema
		}
		s.populatedFields |= PropertiesField
	}
	if b.propertyNames != nil {
		s.propertyNames = b.propertyNames
		s.populatedFields |= PropertyNamesField
	}
	if b.reference != nil {
		s.reference = b.reference
		s.populatedFields |= ReferenceField
	}
	if b.required != nil {
		s.required = b.required
		s.populatedFields |= RequiredField
	}
	s.schema = b.schema
	if b.thenSchema != nil {
		s.thenSchema = b.thenSchema
		s.populatedFields |= ThenSchemaField
	}
	if b.types != nil {
		s.types = b.types
		s.populatedFields |= TypesField
	}
	if b.unevaluatedItems != nil {
		s.unevaluatedItems = b.unevaluatedItems
		s.populatedFields |= UnevaluatedItemsField
	}
	if b.unevaluatedProperties != nil {
		s.unevaluatedProperties = b.unevaluatedProperties
		s.populatedFields |= UnevaluatedPropertiesField
	}
	if b.uniqueItems != nil {
		s.uniqueItems = b.uniqueItems
		s.populatedFields |= UniqueItemsField
	}
	if b.vocabulary != nil {
		s.vocabulary = b.vocabulary
		s.populatedFields |= VocabularyField
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
