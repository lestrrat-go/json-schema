package schema

import "fmt"

type propPair struct {
	Name   string
	Schema *Schema
}

func validateSchemaOrBool(v SchemaOrBool) error {
	// Basic validation - just check if the value implements the interface
	if v == nil {
		return fmt.Errorf("value cannot be nil")
	}
	return nil
}

type Builder struct {
	err                   error
	additionalItems       SchemaOrBool
	additionalProperties  SchemaOrBool
	allOf                 []SchemaOrBool
	anchor                *string
	anyOf                 []SchemaOrBool
	comment               *string
	constantValue         *any
	contains              SchemaOrBool
	contentEncoding       *string
	contentMediaType      *string
	contentSchema         *Schema
	defaultValue          *any
	definitions           []*propPair
	dependentRequired     map[string][]string
	dependentSchemas      map[string]SchemaOrBool
	dynamicAnchor         *string
	dynamicReference      *string
	elseSchema            SchemaOrBool
	enum                  []any
	exclusiveMaximum      *float64
	exclusiveMinimum      *float64
	format                *string
	id                    *string
	ifSchema              SchemaOrBool
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
	prefixItems           []*Schema
	properties            []*propPair
	propertyNames         *Schema
	reference             *string
	required              []string
	schema                *string
	thenSchema            SchemaOrBool
	types                 PrimitiveTypes
	unevaluatedItems      SchemaOrBool
	unevaluatedProperties SchemaOrBool
	uniqueItems           *bool
	vocabulary            map[string]bool
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) AdditionalItems(v SchemaOrBool) *Builder {
	if b.err != nil {
		return b
	}
	b.additionalItems = v
	return b
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

// Anchor sets the $anchor field of the schema being built.
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

// Comment sets the $comment field of the schema being built.
func (b *Builder) Comment(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.comment = &v
	return b
}

// Const sets the const field of the schema being built.
func (b *Builder) Const(v any) *Builder {
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

// ContentEncoding sets the contentEncoding field of the schema being built.
func (b *Builder) ContentEncoding(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.contentEncoding = &v
	return b
}

// ContentMediaType sets the contentMediaType field of the schema being built.
func (b *Builder) ContentMediaType(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.contentMediaType = &v
	return b
}

// ContentSchema sets the contentSchema field of the schema being built.
func (b *Builder) ContentSchema(v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.contentSchema = v
	return b
}

// Default sets the default field of the schema being built.
func (b *Builder) Default(v any) *Builder {
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

// DependentRequired sets the dependentRequired field of the schema being built.
func (b *Builder) DependentRequired(v map[string][]string) *Builder {
	if b.err != nil {
		return b
	}

	b.dependentRequired = v
	return b
}

// DependentSchemas sets the dependentSchemas field of the schema being built.
func (b *Builder) DependentSchemas(v map[string]SchemaOrBool) *Builder {
	if b.err != nil {
		return b
	}

	b.dependentSchemas = v
	return b
}

// DynamicAnchor sets the $dynamicAnchor field of the schema being built.
func (b *Builder) DynamicAnchor(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.dynamicAnchor = &v
	return b
}

// DynamicReference sets the $dynamicRef field of the schema being built.
func (b *Builder) DynamicReference(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.dynamicReference = &v
	return b
}

func (b *Builder) ElseSchema(v SchemaOrBool) *Builder {
	if b.err != nil {
		return b
	}
	b.elseSchema = v
	return b
}

func (b *Builder) Enum(v ...any) *Builder {
	if b.err != nil {
		return b
	}

	b.enum = v
	return b
}

// ExclusiveMaximum sets the exclusiveMaximum field of the schema being built.
func (b *Builder) ExclusiveMaximum(v float64) *Builder {
	if b.err != nil {
		return b
	}

	b.exclusiveMaximum = &v
	return b
}

// ExclusiveMinimum sets the exclusiveMinimum field of the schema being built.
func (b *Builder) ExclusiveMinimum(v float64) *Builder {
	if b.err != nil {
		return b
	}

	b.exclusiveMinimum = &v
	return b
}

// Format sets the format field of the schema being built.
func (b *Builder) Format(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.format = &v
	return b
}

// ID sets the $id field of the schema being built.
func (b *Builder) ID(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.id = &v
	return b
}

func (b *Builder) IfSchema(v SchemaOrBool) *Builder {
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

// MaxContains sets the maxContains field of the schema being built.
func (b *Builder) MaxContains(v uint) *Builder {
	if b.err != nil {
		return b
	}

	b.maxContains = &v
	return b
}

// MaxItems sets the maxItems field of the schema being built.
func (b *Builder) MaxItems(v uint) *Builder {
	if b.err != nil {
		return b
	}

	b.maxItems = &v
	return b
}

// MaxLength sets the maxLength field of the schema being built.
func (b *Builder) MaxLength(v int) *Builder {
	if b.err != nil {
		return b
	}

	b.maxLength = &v
	return b
}

// MaxProperties sets the maxProperties field of the schema being built.
func (b *Builder) MaxProperties(v uint) *Builder {
	if b.err != nil {
		return b
	}

	b.maxProperties = &v
	return b
}

// Maximum sets the maximum field of the schema being built.
func (b *Builder) Maximum(v float64) *Builder {
	if b.err != nil {
		return b
	}

	b.maximum = &v
	return b
}

// MinContains sets the minContains field of the schema being built.
func (b *Builder) MinContains(v uint) *Builder {
	if b.err != nil {
		return b
	}

	b.minContains = &v
	return b
}

// MinItems sets the minItems field of the schema being built.
func (b *Builder) MinItems(v uint) *Builder {
	if b.err != nil {
		return b
	}

	b.minItems = &v
	return b
}

// MinLength sets the minLength field of the schema being built.
func (b *Builder) MinLength(v int) *Builder {
	if b.err != nil {
		return b
	}

	b.minLength = &v
	return b
}

// MinProperties sets the minProperties field of the schema being built.
func (b *Builder) MinProperties(v uint) *Builder {
	if b.err != nil {
		return b
	}

	b.minProperties = &v
	return b
}

// Minimum sets the minimum field of the schema being built.
func (b *Builder) Minimum(v float64) *Builder {
	if b.err != nil {
		return b
	}

	b.minimum = &v
	return b
}

// MultipleOf sets the multipleOf field of the schema being built.
func (b *Builder) MultipleOf(v float64) *Builder {
	if b.err != nil {
		return b
	}

	b.multipleOf = &v
	return b
}

// Not sets the not field of the schema being built.
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

// Pattern sets the pattern field of the schema being built.
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

func (b *Builder) PrefixItems(v ...*Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.prefixItems = v
	return b
}

func (b *Builder) Property(n string, v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.properties = append(b.properties, &propPair{Name: n, Schema: v})
	return b
}

// PropertyNames sets the propertyNames field of the schema being built.
func (b *Builder) PropertyNames(v *Schema) *Builder {
	if b.err != nil {
		return b
	}

	b.propertyNames = v
	return b
}

// Reference sets the $ref field of the schema being built.
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

// Schema sets the $schema field of the schema being built.
// Please note that this field is not automatically set by the library
// when building a schema, as there is no way to know if the schema
// is intended to be used as a standalone schema or as part of a larger
// schema. Therefore, you must set it explicitly when building a schema.
//
// When unmarshaling from JSON, this field is will be automatically set
// to the value of the `$schema` keyword if it exists.
func (b *Builder) Schema(v string) *Builder {
	if b.err != nil {
		return b
	}

	b.schema = &v
	return b
}

func (b *Builder) ThenSchema(v SchemaOrBool) *Builder {
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

// UniqueItems sets the uniqueItems field of the schema being built.
func (b *Builder) UniqueItems(v bool) *Builder {
	if b.err != nil {
		return b
	}

	b.uniqueItems = &v
	return b
}

// Vocabulary sets the $vocabulary field of the schema being built.
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

	if original.Has(AdditionalItemsField) {
		b.additionalItems = original.additionalItems
	}

	if original.Has(AdditionalPropertiesField) {
		b.additionalProperties = original.additionalProperties
	}

	if original.Has(AllOfField) {
		b.allOf = original.allOf
	}

	if original.Has(AnchorField) {
		b.anchor = original.anchor
	}

	if original.Has(AnyOfField) {
		b.anyOf = original.anyOf
	}

	if original.Has(CommentField) {
		b.comment = original.comment
	}

	if original.Has(ConstField) {
		b.constantValue = original.constantValue
	}

	if original.Has(ContainsField) {
		b.contains = original.contains
	}

	if original.Has(ContentEncodingField) {
		b.contentEncoding = original.contentEncoding
	}

	if original.Has(ContentMediaTypeField) {
		b.contentMediaType = original.contentMediaType
	}

	if original.Has(ContentSchemaField) {
		b.contentSchema = original.contentSchema
	}

	if original.Has(DefaultField) {
		b.defaultValue = original.defaultValue
	}

	if original.Has(DefinitionsField) {
		for name, schema := range original.definitions {
			b.definitions = append(b.definitions, &propPair{Name: name, Schema: schema})
		}
	}

	if original.Has(DependentRequiredField) {
		b.dependentRequired = original.dependentRequired
	}

	if original.Has(DependentSchemasField) {
		b.dependentSchemas = original.dependentSchemas
	}

	if original.Has(DynamicAnchorField) {
		b.dynamicAnchor = original.dynamicAnchor
	}

	if original.Has(DynamicReferenceField) {
		b.dynamicReference = original.dynamicReference
	}

	if original.Has(ElseSchemaField) {
		b.elseSchema = original.elseSchema
	}

	if original.Has(EnumField) {
		b.enum = original.enum
	}

	if original.Has(ExclusiveMaximumField) {
		b.exclusiveMaximum = original.exclusiveMaximum
	}

	if original.Has(ExclusiveMinimumField) {
		b.exclusiveMinimum = original.exclusiveMinimum
	}

	if original.Has(FormatField) {
		b.format = original.format
	}

	if original.Has(IDField) {
		b.id = original.id
	}

	if original.Has(IfSchemaField) {
		b.ifSchema = original.ifSchema
	}

	if original.Has(ItemsField) {
		b.items = original.items
	}

	if original.Has(MaxContainsField) {
		b.maxContains = original.maxContains
	}

	if original.Has(MaxItemsField) {
		b.maxItems = original.maxItems
	}

	if original.Has(MaxLengthField) {
		b.maxLength = original.maxLength
	}

	if original.Has(MaxPropertiesField) {
		b.maxProperties = original.maxProperties
	}

	if original.Has(MaximumField) {
		b.maximum = original.maximum
	}

	if original.Has(MinContainsField) {
		b.minContains = original.minContains
	}

	if original.Has(MinItemsField) {
		b.minItems = original.minItems
	}

	if original.Has(MinLengthField) {
		b.minLength = original.minLength
	}

	if original.Has(MinPropertiesField) {
		b.minProperties = original.minProperties
	}

	if original.Has(MinimumField) {
		b.minimum = original.minimum
	}

	if original.Has(MultipleOfField) {
		b.multipleOf = original.multipleOf
	}

	if original.Has(NotField) {
		b.not = original.not
	}

	if original.Has(OneOfField) {
		b.oneOf = original.oneOf
	}

	if original.Has(PatternField) {
		b.pattern = original.pattern
	}

	if original.Has(PatternPropertiesField) {
		for name, schema := range original.patternProperties {
			b.patternProperties = append(b.patternProperties, &propPair{Name: name, Schema: schema})
		}
	}

	if original.Has(PrefixItemsField) {
		b.prefixItems = original.prefixItems
	}

	if original.Has(PropertiesField) {
		for name, schema := range original.properties {
			b.properties = append(b.properties, &propPair{Name: name, Schema: schema})
		}
	}

	if original.Has(PropertyNamesField) {
		b.propertyNames = original.propertyNames
	}

	if original.Has(ReferenceField) {
		b.reference = original.reference
	}

	if original.Has(RequiredField) {
		b.required = original.required
	}

	if original.Has(SchemaField) {
		b.schema = original.schema
	}

	if original.Has(ThenSchemaField) {
		b.thenSchema = original.thenSchema
	}

	if original.Has(TypesField) {
		b.types = original.types
	}

	if original.Has(UnevaluatedItemsField) {
		b.unevaluatedItems = original.unevaluatedItems
	}

	if original.Has(UnevaluatedPropertiesField) {
		b.unevaluatedProperties = original.unevaluatedProperties
	}

	if original.Has(UniqueItemsField) {
		b.uniqueItems = original.uniqueItems
	}

	if original.Has(VocabularyField) {
		b.vocabulary = original.vocabulary
	}
	return b
}

// Reset clears the specified field flags
// Usage: builder.Reset(AnchorField | PropertiesField) clears both anchor and properties
func (b *Builder) Reset(flags FieldFlag) *Builder {
	if b.err != nil {
		return b
	}

	if (flags & AdditionalItemsField) != 0 {
		b.additionalItems = nil
	}
	if (flags & AdditionalPropertiesField) != 0 {
		b.additionalProperties = nil
	}
	if (flags & AllOfField) != 0 {
		b.allOf = nil
	}
	if (flags & AnchorField) != 0 {
		b.anchor = nil
	}
	if (flags & AnyOfField) != 0 {
		b.anyOf = nil
	}
	if (flags & CommentField) != 0 {
		b.comment = nil
	}
	if (flags & ConstField) != 0 {
		b.constantValue = nil
	}
	if (flags & ContainsField) != 0 {
		b.contains = nil
	}
	if (flags & ContentEncodingField) != 0 {
		b.contentEncoding = nil
	}
	if (flags & ContentMediaTypeField) != 0 {
		b.contentMediaType = nil
	}
	if (flags & ContentSchemaField) != 0 {
		b.contentSchema = nil
	}
	if (flags & DefaultField) != 0 {
		b.defaultValue = nil
	}
	if (flags & DefinitionsField) != 0 {
		b.definitions = nil
	}
	if (flags & DependentRequiredField) != 0 {
		b.dependentRequired = nil
	}
	if (flags & DependentSchemasField) != 0 {
		b.dependentSchemas = nil
	}
	if (flags & DynamicAnchorField) != 0 {
		b.dynamicAnchor = nil
	}
	if (flags & DynamicReferenceField) != 0 {
		b.dynamicReference = nil
	}
	if (flags & ElseSchemaField) != 0 {
		b.elseSchema = nil
	}
	if (flags & EnumField) != 0 {
		b.enum = nil
	}
	if (flags & ExclusiveMaximumField) != 0 {
		b.exclusiveMaximum = nil
	}
	if (flags & ExclusiveMinimumField) != 0 {
		b.exclusiveMinimum = nil
	}
	if (flags & FormatField) != 0 {
		b.format = nil
	}
	if (flags & IDField) != 0 {
		b.id = nil
	}
	if (flags & IfSchemaField) != 0 {
		b.ifSchema = nil
	}
	if (flags & ItemsField) != 0 {
		b.items = nil
	}
	if (flags & MaxContainsField) != 0 {
		b.maxContains = nil
	}
	if (flags & MaxItemsField) != 0 {
		b.maxItems = nil
	}
	if (flags & MaxLengthField) != 0 {
		b.maxLength = nil
	}
	if (flags & MaxPropertiesField) != 0 {
		b.maxProperties = nil
	}
	if (flags & MaximumField) != 0 {
		b.maximum = nil
	}
	if (flags & MinContainsField) != 0 {
		b.minContains = nil
	}
	if (flags & MinItemsField) != 0 {
		b.minItems = nil
	}
	if (flags & MinLengthField) != 0 {
		b.minLength = nil
	}
	if (flags & MinPropertiesField) != 0 {
		b.minProperties = nil
	}
	if (flags & MinimumField) != 0 {
		b.minimum = nil
	}
	if (flags & MultipleOfField) != 0 {
		b.multipleOf = nil
	}
	if (flags & NotField) != 0 {
		b.not = nil
	}
	if (flags & OneOfField) != 0 {
		b.oneOf = nil
	}
	if (flags & PatternField) != 0 {
		b.pattern = nil
	}
	if (flags & PatternPropertiesField) != 0 {
		b.patternProperties = nil
	}
	if (flags & PrefixItemsField) != 0 {
		b.prefixItems = nil
	}
	if (flags & PropertiesField) != 0 {
		b.properties = nil
	}
	if (flags & PropertyNamesField) != 0 {
		b.propertyNames = nil
	}
	if (flags & ReferenceField) != 0 {
		b.reference = nil
	}
	if (flags & RequiredField) != 0 {
		b.required = nil
	}
	if (flags & SchemaField) != 0 {
		b.schema = nil
	}
	if (flags & ThenSchemaField) != 0 {
		b.thenSchema = nil
	}
	if (flags & TypesField) != 0 {
		b.types = nil
	}
	if (flags & UnevaluatedItemsField) != 0 {
		b.unevaluatedItems = nil
	}
	if (flags & UnevaluatedPropertiesField) != 0 {
		b.unevaluatedProperties = nil
	}
	if (flags & UniqueItemsField) != 0 {
		b.uniqueItems = nil
	}
	if (flags & VocabularyField) != 0 {
		b.vocabulary = nil
	}

	return b
}

func (b *Builder) Build() (*Schema, error) {
	s := New()
	if b.additionalItems != nil {
		s.additionalItems = b.additionalItems
		s.populatedFields |= AdditionalItemsField
	}
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
	if b.prefixItems != nil {
		s.prefixItems = b.prefixItems
		s.populatedFields |= PrefixItemsField
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
	if b.schema != nil {
		s.schema = b.schema
		s.populatedFields |= SchemaField
	}
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
