package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/lestrrat-go/json-schema/internal/field"
	"github.com/lestrrat-go/json-schema/keywords"
)

// Field bit flags for tracking populated fields
type FieldFlag = field.Flag

const (
	AdditionalItemsField       = field.AdditionalItems
	AdditionalPropertiesField  = field.AdditionalProperties
	AllOfField                 = field.AllOf
	AnchorField                = field.Anchor
	AnyOfField                 = field.AnyOf
	CommentField               = field.Comment
	ConstField                 = field.Const
	ContainsField              = field.Contains
	ContentEncodingField       = field.ContentEncoding
	ContentMediaTypeField      = field.ContentMediaType
	ContentSchemaField         = field.ContentSchema
	DefaultField               = field.Default
	DefinitionsField           = field.Definitions
	DependentRequiredField     = field.DependentRequired
	DependentSchemasField      = field.DependentSchemas
	DynamicAnchorField         = field.DynamicAnchor
	DynamicReferenceField      = field.DynamicReference
	ElseSchemaField            = field.ElseSchema
	EnumField                  = field.Enum
	ExclusiveMaximumField      = field.ExclusiveMaximum
	ExclusiveMinimumField      = field.ExclusiveMinimum
	FormatField                = field.Format
	IDField                    = field.ID
	IfSchemaField              = field.IfSchema
	ItemsField                 = field.Items
	MaxContainsField           = field.MaxContains
	MaxItemsField              = field.MaxItems
	MaxLengthField             = field.MaxLength
	MaxPropertiesField         = field.MaxProperties
	MaximumField               = field.Maximum
	MinContainsField           = field.MinContains
	MinItemsField              = field.MinItems
	MinLengthField             = field.MinLength
	MinPropertiesField         = field.MinProperties
	MinimumField               = field.Minimum
	MultipleOfField            = field.MultipleOf
	NotField                   = field.Not
	OneOfField                 = field.OneOf
	PatternField               = field.Pattern
	PatternPropertiesField     = field.PatternProperties
	PrefixItemsField           = field.PrefixItems
	PropertiesField            = field.Properties
	PropertyNamesField         = field.PropertyNames
	ReferenceField             = field.Reference
	RequiredField              = field.Required
	ThenSchemaField            = field.ThenSchema
	TypesField                 = field.Types
	UnevaluatedItemsField      = field.UnevaluatedItems
	UnevaluatedPropertiesField = field.UnevaluatedProperties
	UniqueItemsField           = field.UniqueItems
	VocabularyField            = field.Vocabulary
)

type Schema struct {
	isRoot                bool
	populatedFields       field.Flag
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
	definitions           map[string]*Schema
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
	patternProperties     map[string]*Schema
	prefixItems           []*Schema
	properties            map[string]*Schema
	propertyNames         *Schema
	reference             *string
	required              []string
	schema                string
	thenSchema            SchemaOrBool
	types                 PrimitiveTypes
	unevaluatedItems      SchemaOrBool
	unevaluatedProperties SchemaOrBool
	uniqueItems           *bool
	vocabulary            map[string]bool
}

func New() *Schema {
	return &Schema{
		schema: Version,
	}
}

// Has checks if the specified field flags are set
// Usage: schema.Has(AnchorField | PropertiesField) returns true if both anchor and properties are set
func (s *Schema) Has(flags FieldFlag) bool {
	return (s.populatedFields & flags) == flags
}

// HasAny checks if any of the specified field flags are set
// Usage: schema.HasAny(AnchorField | PropertiesField) returns true if either anchor or properties (or both) are set
func (s *Schema) HasAny(flags FieldFlag) bool {
	return (s.populatedFields & flags) != 0
}

func (s *Schema) HasAdditionalItems() bool {
	return s.populatedFields&AdditionalItemsField != 0
}

func (s *Schema) AdditionalItems() SchemaOrBool {
	return s.additionalItems
}

func (s *Schema) HasAdditionalProperties() bool {
	return s.populatedFields&AdditionalPropertiesField != 0
}

func (s *Schema) AdditionalProperties() SchemaOrBool {
	return s.additionalProperties
}

func (s *Schema) HasAllOf() bool {
	return s.populatedFields&AllOfField != 0
}

func (s *Schema) AllOf() []SchemaOrBool {
	return s.allOf
}

func (s *Schema) HasAnchor() bool {
	return s.populatedFields&AnchorField != 0
}

func (s *Schema) Anchor() string {
	return *(s.anchor)
}

func (s *Schema) HasAnyOf() bool {
	return s.populatedFields&AnyOfField != 0
}

func (s *Schema) AnyOf() []SchemaOrBool {
	return s.anyOf
}

func (s *Schema) HasComment() bool {
	return s.populatedFields&CommentField != 0
}

func (s *Schema) Comment() string {
	return *(s.comment)
}

func (s *Schema) HasConst() bool {
	return s.populatedFields&ConstField != 0
}

func (s *Schema) Const() any {
	return *(s.constantValue)
}

func (s *Schema) HasContains() bool {
	return s.populatedFields&ContainsField != 0
}

func (s *Schema) Contains() SchemaOrBool {
	return s.contains
}

func (s *Schema) HasContentEncoding() bool {
	return s.populatedFields&ContentEncodingField != 0
}

func (s *Schema) ContentEncoding() string {
	return *(s.contentEncoding)
}

func (s *Schema) HasContentMediaType() bool {
	return s.populatedFields&ContentMediaTypeField != 0
}

func (s *Schema) ContentMediaType() string {
	return *(s.contentMediaType)
}

func (s *Schema) HasContentSchema() bool {
	return s.populatedFields&ContentSchemaField != 0
}

func (s *Schema) ContentSchema() *Schema {
	return s.contentSchema
}

func (s *Schema) HasDefault() bool {
	return s.populatedFields&DefaultField != 0
}

func (s *Schema) Default() any {
	return *(s.defaultValue)
}

func (s *Schema) HasDefinitions() bool {
	return s.populatedFields&DefinitionsField != 0
}

func (s *Schema) Definitions() map[string]*Schema {
	return s.definitions
}

func (s *Schema) HasDependentRequired() bool {
	return s.populatedFields&DependentRequiredField != 0
}

func (s *Schema) DependentRequired() map[string][]string {
	return s.dependentRequired
}

func (s *Schema) HasDependentSchemas() bool {
	return s.populatedFields&DependentSchemasField != 0
}

func (s *Schema) DependentSchemas() map[string]SchemaOrBool {
	return s.dependentSchemas
}

func (s *Schema) HasDynamicAnchor() bool {
	return s.populatedFields&DynamicAnchorField != 0
}

func (s *Schema) DynamicAnchor() string {
	return *(s.dynamicAnchor)
}

func (s *Schema) HasDynamicReference() bool {
	return s.populatedFields&DynamicReferenceField != 0
}

func (s *Schema) DynamicReference() string {
	return *(s.dynamicReference)
}

func (s *Schema) HasElseSchema() bool {
	return s.populatedFields&ElseSchemaField != 0
}

func (s *Schema) ElseSchema() SchemaOrBool {
	return s.elseSchema
}

func (s *Schema) HasEnum() bool {
	return s.populatedFields&EnumField != 0
}

func (s *Schema) Enum() []any {
	return s.enum
}

func (s *Schema) HasExclusiveMaximum() bool {
	return s.populatedFields&ExclusiveMaximumField != 0
}

func (s *Schema) ExclusiveMaximum() float64 {
	return *(s.exclusiveMaximum)
}

func (s *Schema) HasExclusiveMinimum() bool {
	return s.populatedFields&ExclusiveMinimumField != 0
}

func (s *Schema) ExclusiveMinimum() float64 {
	return *(s.exclusiveMinimum)
}

func (s *Schema) HasFormat() bool {
	return s.populatedFields&FormatField != 0
}

func (s *Schema) Format() string {
	return *(s.format)
}

func (s *Schema) HasID() bool {
	return s.populatedFields&IDField != 0
}

func (s *Schema) ID() string {
	return *(s.id)
}

func (s *Schema) HasIfSchema() bool {
	return s.populatedFields&IfSchemaField != 0
}

func (s *Schema) IfSchema() SchemaOrBool {
	return s.ifSchema
}

func (s *Schema) HasItems() bool {
	return s.populatedFields&ItemsField != 0
}

func (s *Schema) Items() SchemaOrBool {
	return s.items
}

func (s *Schema) HasMaxContains() bool {
	return s.populatedFields&MaxContainsField != 0
}

func (s *Schema) MaxContains() uint {
	return *(s.maxContains)
}

func (s *Schema) HasMaxItems() bool {
	return s.populatedFields&MaxItemsField != 0
}

func (s *Schema) MaxItems() uint {
	return *(s.maxItems)
}

func (s *Schema) HasMaxLength() bool {
	return s.populatedFields&MaxLengthField != 0
}

func (s *Schema) MaxLength() int {
	return *(s.maxLength)
}

func (s *Schema) HasMaxProperties() bool {
	return s.populatedFields&MaxPropertiesField != 0
}

func (s *Schema) MaxProperties() uint {
	return *(s.maxProperties)
}

func (s *Schema) HasMaximum() bool {
	return s.populatedFields&MaximumField != 0
}

func (s *Schema) Maximum() float64 {
	return *(s.maximum)
}

func (s *Schema) HasMinContains() bool {
	return s.populatedFields&MinContainsField != 0
}

func (s *Schema) MinContains() uint {
	return *(s.minContains)
}

func (s *Schema) HasMinItems() bool {
	return s.populatedFields&MinItemsField != 0
}

func (s *Schema) MinItems() uint {
	return *(s.minItems)
}

func (s *Schema) HasMinLength() bool {
	return s.populatedFields&MinLengthField != 0
}

func (s *Schema) MinLength() int {
	return *(s.minLength)
}

func (s *Schema) HasMinProperties() bool {
	return s.populatedFields&MinPropertiesField != 0
}

func (s *Schema) MinProperties() uint {
	return *(s.minProperties)
}

func (s *Schema) HasMinimum() bool {
	return s.populatedFields&MinimumField != 0
}

func (s *Schema) Minimum() float64 {
	return *(s.minimum)
}

func (s *Schema) HasMultipleOf() bool {
	return s.populatedFields&MultipleOfField != 0
}

func (s *Schema) MultipleOf() float64 {
	return *(s.multipleOf)
}

func (s *Schema) HasNot() bool {
	return s.populatedFields&NotField != 0
}

func (s *Schema) Not() *Schema {
	return s.not
}

func (s *Schema) HasOneOf() bool {
	return s.populatedFields&OneOfField != 0
}

func (s *Schema) OneOf() []SchemaOrBool {
	return s.oneOf
}

func (s *Schema) HasPattern() bool {
	return s.populatedFields&PatternField != 0
}

func (s *Schema) Pattern() string {
	return *(s.pattern)
}

func (s *Schema) HasPatternProperties() bool {
	return s.populatedFields&PatternPropertiesField != 0
}

func (s *Schema) PatternProperties() map[string]*Schema {
	return s.patternProperties
}

func (s *Schema) HasPrefixItems() bool {
	return s.populatedFields&PrefixItemsField != 0
}

func (s *Schema) PrefixItems() []*Schema {
	return s.prefixItems
}

func (s *Schema) HasProperties() bool {
	return s.populatedFields&PropertiesField != 0
}

func (s *Schema) Properties() map[string]*Schema {
	return s.properties
}

func (s *Schema) HasPropertyNames() bool {
	return s.populatedFields&PropertyNamesField != 0
}

func (s *Schema) PropertyNames() *Schema {
	return s.propertyNames
}

func (s *Schema) HasReference() bool {
	return s.populatedFields&ReferenceField != 0
}

func (s *Schema) Reference() string {
	return *(s.reference)
}

func (s *Schema) HasRequired() bool {
	return s.populatedFields&RequiredField != 0
}

func (s *Schema) Required() []string {
	return s.required
}

func (s *Schema) Schema() string {
	return s.schema
}

func (s *Schema) HasThenSchema() bool {
	return s.populatedFields&ThenSchemaField != 0
}

func (s *Schema) ThenSchema() SchemaOrBool {
	return s.thenSchema
}

func (s *Schema) HasTypes() bool {
	return s.populatedFields&TypesField != 0
}

func (s *Schema) Types() PrimitiveTypes {
	return s.types
}

func (s *Schema) HasUnevaluatedItems() bool {
	return s.populatedFields&UnevaluatedItemsField != 0
}

func (s *Schema) UnevaluatedItems() SchemaOrBool {
	return s.unevaluatedItems
}

func (s *Schema) HasUnevaluatedProperties() bool {
	return s.populatedFields&UnevaluatedPropertiesField != 0
}

func (s *Schema) UnevaluatedProperties() SchemaOrBool {
	return s.unevaluatedProperties
}

func (s *Schema) HasUniqueItems() bool {
	return s.populatedFields&UniqueItemsField != 0
}

func (s *Schema) UniqueItems() bool {
	return *(s.uniqueItems)
}

func (s *Schema) HasVocabulary() bool {
	return s.populatedFields&VocabularyField != 0
}

func (s *Schema) Vocabulary() map[string]bool {
	return s.vocabulary
}

func (s *Schema) ContainsType(typ PrimitiveType) bool {
	if s.types == nil {
		return false
	}
	for _, t := range s.types {
		if t == typ {
			return true
		}
	}
	return false
}

type pair struct {
	Name  string
	Value any
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	s.isRoot = true
	defer func() { s.isRoot = false }()
	fields := make([]pair, 0, 52)
	if s.HasAdditionalItems() {
		fields = append(fields, pair{Name: keywords.AdditionalItems, Value: s.additionalItems})
	}
	if s.HasAdditionalProperties() {
		fields = append(fields, pair{Name: keywords.AdditionalProperties, Value: s.additionalProperties})
	}
	if s.HasAllOf() {
		fields = append(fields, pair{Name: keywords.AllOf, Value: s.allOf})
	}
	if s.HasAnchor() {
		fields = append(fields, pair{Name: keywords.Anchor, Value: *(s.anchor)})
	}
	if s.HasAnyOf() {
		fields = append(fields, pair{Name: keywords.AnyOf, Value: s.anyOf})
	}
	if s.HasComment() {
		fields = append(fields, pair{Name: keywords.Comment, Value: *(s.comment)})
	}
	if s.HasConst() {
		fields = append(fields, pair{Name: keywords.Const, Value: *(s.constantValue)})
	}
	if s.HasContains() {
		fields = append(fields, pair{Name: keywords.Contains, Value: s.contains})
	}
	if s.HasContentEncoding() {
		fields = append(fields, pair{Name: keywords.ContentEncoding, Value: *(s.contentEncoding)})
	}
	if s.HasContentMediaType() {
		fields = append(fields, pair{Name: keywords.ContentMediaType, Value: *(s.contentMediaType)})
	}
	if s.HasContentSchema() {
		fields = append(fields, pair{Name: keywords.ContentSchema, Value: s.contentSchema})
	}
	if s.HasDefault() {
		fields = append(fields, pair{Name: keywords.Default, Value: *(s.defaultValue)})
	}
	if s.HasDefinitions() {
		fields = append(fields, pair{Name: keywords.Definitions, Value: s.definitions})
	}
	if s.HasDependentRequired() {
		fields = append(fields, pair{Name: keywords.DependentRequired, Value: s.dependentRequired})
	}
	if s.HasDependentSchemas() {
		fields = append(fields, pair{Name: keywords.DependentSchemas, Value: s.dependentSchemas})
	}
	if s.HasDynamicAnchor() {
		fields = append(fields, pair{Name: keywords.DynamicAnchor, Value: *(s.dynamicAnchor)})
	}
	if s.HasDynamicReference() {
		fields = append(fields, pair{Name: keywords.DynamicReference, Value: *(s.dynamicReference)})
	}
	if s.HasElseSchema() {
		fields = append(fields, pair{Name: keywords.Else, Value: s.elseSchema})
	}
	if s.HasEnum() {
		fields = append(fields, pair{Name: keywords.Enum, Value: s.enum})
	}
	if s.HasExclusiveMaximum() {
		fields = append(fields, pair{Name: keywords.ExclusiveMaximum, Value: *(s.exclusiveMaximum)})
	}
	if s.HasExclusiveMinimum() {
		fields = append(fields, pair{Name: keywords.ExclusiveMinimum, Value: *(s.exclusiveMinimum)})
	}
	if s.HasFormat() {
		fields = append(fields, pair{Name: keywords.Format, Value: *(s.format)})
	}
	if s.HasID() {
		fields = append(fields, pair{Name: keywords.ID, Value: *(s.id)})
	}
	if s.HasIfSchema() {
		fields = append(fields, pair{Name: keywords.If, Value: s.ifSchema})
	}
	if s.HasItems() {
		fields = append(fields, pair{Name: keywords.Items, Value: s.items})
	}
	if s.HasMaxContains() {
		fields = append(fields, pair{Name: keywords.MaxContains, Value: *(s.maxContains)})
	}
	if s.HasMaxItems() {
		fields = append(fields, pair{Name: keywords.MaxItems, Value: *(s.maxItems)})
	}
	if s.HasMaxLength() {
		fields = append(fields, pair{Name: keywords.MaxLength, Value: *(s.maxLength)})
	}
	if s.HasMaxProperties() {
		fields = append(fields, pair{Name: keywords.MaxProperties, Value: *(s.maxProperties)})
	}
	if s.HasMaximum() {
		fields = append(fields, pair{Name: keywords.Maximum, Value: *(s.maximum)})
	}
	if s.HasMinContains() {
		fields = append(fields, pair{Name: keywords.MinContains, Value: *(s.minContains)})
	}
	if s.HasMinItems() {
		fields = append(fields, pair{Name: keywords.MinItems, Value: *(s.minItems)})
	}
	if s.HasMinLength() {
		fields = append(fields, pair{Name: keywords.MinLength, Value: *(s.minLength)})
	}
	if s.HasMinProperties() {
		fields = append(fields, pair{Name: keywords.MinProperties, Value: *(s.minProperties)})
	}
	if s.HasMinimum() {
		fields = append(fields, pair{Name: keywords.Minimum, Value: *(s.minimum)})
	}
	if s.HasMultipleOf() {
		fields = append(fields, pair{Name: keywords.MultipleOf, Value: *(s.multipleOf)})
	}
	if s.HasNot() {
		fields = append(fields, pair{Name: keywords.Not, Value: s.not})
	}
	if s.HasOneOf() {
		fields = append(fields, pair{Name: keywords.OneOf, Value: s.oneOf})
	}
	if s.HasPattern() {
		fields = append(fields, pair{Name: keywords.Pattern, Value: *(s.pattern)})
	}
	if s.HasPatternProperties() {
		fields = append(fields, pair{Name: keywords.PatternProperties, Value: s.patternProperties})
	}
	if s.HasPrefixItems() {
		fields = append(fields, pair{Name: keywords.PrefixItems, Value: s.prefixItems})
	}
	if s.HasProperties() {
		fields = append(fields, pair{Name: keywords.Properties, Value: s.properties})
	}
	if s.HasPropertyNames() {
		fields = append(fields, pair{Name: keywords.PropertyNames, Value: s.propertyNames})
	}
	if s.HasReference() {
		fields = append(fields, pair{Name: keywords.Reference, Value: *(s.reference)})
	}
	if s.HasRequired() {
		fields = append(fields, pair{Name: keywords.Required, Value: s.required})
	}
	if v := s.schema; s.isRoot && v != "" {
		fields = append(fields, pair{Name: keywords.Schema, Value: v})
	}
	if s.HasThenSchema() {
		fields = append(fields, pair{Name: keywords.Then, Value: s.thenSchema})
	}
	if s.HasTypes() {
		fields = append(fields, pair{Name: keywords.Type, Value: s.types})
	}
	if s.HasUnevaluatedItems() {
		fields = append(fields, pair{Name: keywords.UnevaluatedItems, Value: s.unevaluatedItems})
	}
	if s.HasUnevaluatedProperties() {
		fields = append(fields, pair{Name: keywords.UnevaluatedProperties, Value: s.unevaluatedProperties})
	}
	if s.HasUniqueItems() {
		fields = append(fields, pair{Name: keywords.UniqueItems, Value: *(s.uniqueItems)})
	}
	if s.HasVocabulary() {
		fields = append(fields, pair{Name: keywords.Vocabulary, Value: s.vocabulary})
	}
	sort.Slice(fields, func(i, j int) bool {
		return compareFieldNames(fields[i].Name, fields[j].Name)
	})
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	buf.WriteByte('{')
	for i, field := range fields {
		if i > 0 {
			buf.WriteByte(',')
		}
		if err := enc.Encode(field.Name); err != nil {
			return nil, fmt.Errorf("json-schema: Schema.MarshalJSON: failed to encode field name: %w", err)
		}
		buf.WriteByte(':')
		if err := enc.Encode(field.Value); err != nil {
			return nil, fmt.Errorf("json-schema: Schema.MarshalJSON: failed to encode field value: %w", err)
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (s *Schema) UnmarshalJSON(buf []byte) error {
	dec := json.NewDecoder(bytes.NewReader(buf))
LOOP:
	for {
		tok, err := dec.Token()
		if err != nil {
			return fmt.Errorf(`json-schema: failed to read JSON token: %w`, err)
		}
		switch tok := tok.(type) {
		case json.Delim:
			// Assuming we're doing everything correctly, we should ONLY
			// get either '{' or '}' here.
			if tok == '}' { // End of object
				break LOOP
			} else if tok != '{' {
				return fmt.Errorf(`json-schema: failed to parse JSON structure: expected '{', but got '%c'`, tok)
			}
		case string: // Objects can only have string keys
			switch tok {
			case keywords.AdditionalItems:
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "additionalItems": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.additionalItems = BoolSchema(b)
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.additionalItems = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "additionalItems" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
				s.populatedFields |= AdditionalItemsField
			case keywords.AdditionalProperties:
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "additionalProperties": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.additionalProperties = BoolSchema(b)
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.additionalProperties = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "additionalProperties" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
				s.populatedFields |= AdditionalPropertiesField
			case keywords.AllOf:
				v, err := unmarshalSchemaOrBoolSlice(dec)
				if err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "allOf" (attempting to unmarshal as []SchemaOrBool slice): %w`, err)
				}
				s.allOf = v
				s.populatedFields |= AllOfField
			case keywords.Anchor:
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$anchor" (attempting to unmarshal as string): %w`, err)
				}
				s.anchor = &v
				s.populatedFields |= AnchorField
			case keywords.AnyOf:
				v, err := unmarshalSchemaOrBoolSlice(dec)
				if err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "anyOf" (attempting to unmarshal as []SchemaOrBool slice): %w`, err)
				}
				s.anyOf = v
				s.populatedFields |= AnyOfField
			case keywords.Comment:
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$comment" (attempting to unmarshal as string): %w`, err)
				}
				s.comment = &v
				s.populatedFields |= CommentField
			case keywords.Const:
				var v any
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "const" (attempting to unmarshal as any): %w`, err)
				}
				s.constantValue = &v
				s.populatedFields |= ConstField
			case keywords.Contains:
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "contains": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.contains = BoolSchema(b)
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.contains = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "contains" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
				s.populatedFields |= ContainsField
			case keywords.ContentEncoding:
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "contentEncoding" (attempting to unmarshal as string): %w`, err)
				}
				s.contentEncoding = &v
				s.populatedFields |= ContentEncodingField
			case keywords.ContentMediaType:
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "contentMediaType" (attempting to unmarshal as string): %w`, err)
				}
				s.contentMediaType = &v
				s.populatedFields |= ContentMediaTypeField
			case keywords.ContentSchema:
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "contentSchema": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					// Convert boolean to Schema object
					if b {
						s.contentSchema = &Schema{} // true schema - allow everything
					} else {
						// false schema - deny everything using "not": {}
						falseSchema := &Schema{not: &Schema{}}
						falseSchema.populatedFields |= NotField
						s.contentSchema = falseSchema
					}
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.contentSchema = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "contentSchema" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
				s.populatedFields |= ContentSchemaField
			case keywords.Default:
				var v any
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "default" (attempting to unmarshal as any): %w`, err)
				}
				s.defaultValue = &v
				s.populatedFields |= DefaultField
			case keywords.Definitions:
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "$defs": %w`, err)
				}
				// First unmarshal as map[string]json.RawMessage
				var rawMap map[string]json.RawMessage
				if err := json.Unmarshal(rawData, &rawMap); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$defs" (attempting to unmarshal as map): %w`, err)
				}
				// Convert each value to *Schema
				v := make(map[string]*Schema)
				for key, rawValue := range rawMap {
					// Try to decode as boolean first
					var b bool
					if err := json.Unmarshal(rawValue, &b); err == nil {
						// Convert boolean to Schema object
						if b {
							v[key] = &Schema{} // true schema - allow everything
						} else {
							// false schema - deny everything using "not": {}
							falseSchema := &Schema{not: &Schema{}}
							falseSchema.populatedFields |= NotField
							v[key] = falseSchema
						}
					} else {
						// Try to decode as Schema object
						var schema Schema
						if err := json.Unmarshal(rawValue, &schema); err == nil {
							v[key] = &schema
						} else {
							return fmt.Errorf(`json-schema: failed to decode value for field "$defs" key %q (attempting to unmarshal as Schema after bool failed): %w`, key, err)
						}
					}
				}
				s.definitions = v
				s.populatedFields |= DefinitionsField
			case keywords.DependentRequired:
				var v map[string][]string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "dependentRequired" (attempting to unmarshal as map[string][]string): %w`, err)
				}
				s.dependentRequired = v
				s.populatedFields |= DependentRequiredField
			case keywords.DependentSchemas:
				v, err := unmarshalSchemaOrBoolMap(dec)
				if err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "dependentSchemas" (attempting to unmarshal as map[string]SchemaOrBool): %w`, err)
				}
				s.dependentSchemas = v
				s.populatedFields |= DependentSchemasField
			case keywords.DynamicAnchor:
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$dynamicAnchor" (attempting to unmarshal as string): %w`, err)
				}
				s.dynamicAnchor = &v
				s.populatedFields |= DynamicAnchorField
			case keywords.DynamicReference:
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$dynamicRef" (attempting to unmarshal as string): %w`, err)
				}
				s.dynamicReference = &v
				s.populatedFields |= DynamicReferenceField
			case keywords.Else:
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "else": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.elseSchema = BoolSchema(b)
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.elseSchema = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "else" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
				s.populatedFields |= ElseSchemaField
			case keywords.Enum:
				var v []any
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "enum" (attempting to unmarshal as []any): %w`, err)
				}
				s.enum = v
				s.populatedFields |= EnumField
			case keywords.ExclusiveMaximum:
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "exclusiveMaximum" (attempting to unmarshal as float64): %w`, err)
				}
				s.exclusiveMaximum = &v
				s.populatedFields |= ExclusiveMaximumField
			case keywords.ExclusiveMinimum:
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "exclusiveMinimum" (attempting to unmarshal as float64): %w`, err)
				}
				s.exclusiveMinimum = &v
				s.populatedFields |= ExclusiveMinimumField
			case keywords.Format:
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "format" (attempting to unmarshal as string): %w`, err)
				}
				s.format = &v
				s.populatedFields |= FormatField
			case keywords.ID:
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$id" (attempting to unmarshal as string): %w`, err)
				}
				s.id = &v
				s.populatedFields |= IDField
			case keywords.If:
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "if": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.ifSchema = BoolSchema(b)
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.ifSchema = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "if" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
				s.populatedFields |= IfSchemaField
			case keywords.Items:
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "items": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.items = BoolSchema(b)
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.items = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "items" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
				s.populatedFields |= ItemsField
			case keywords.MaxContains:
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "maxContains" (attempting to unmarshal as uint): %w`, err)
				}
				s.maxContains = &v
				s.populatedFields |= MaxContainsField
			case keywords.MaxItems:
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "maxItems" (attempting to unmarshal as uint): %w`, err)
				}
				s.maxItems = &v
				s.populatedFields |= MaxItemsField
			case keywords.MaxLength:
				var v int
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "maxLength" (attempting to unmarshal as int): %w`, err)
				}
				s.maxLength = &v
				s.populatedFields |= MaxLengthField
			case keywords.MaxProperties:
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "maxProperties" (attempting to unmarshal as uint): %w`, err)
				}
				s.maxProperties = &v
				s.populatedFields |= MaxPropertiesField
			case keywords.Maximum:
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "maximum" (attempting to unmarshal as float64): %w`, err)
				}
				s.maximum = &v
				s.populatedFields |= MaximumField
			case keywords.MinContains:
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "minContains" (attempting to unmarshal as uint): %w`, err)
				}
				s.minContains = &v
				s.populatedFields |= MinContainsField
			case keywords.MinItems:
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "minItems" (attempting to unmarshal as uint): %w`, err)
				}
				s.minItems = &v
				s.populatedFields |= MinItemsField
			case keywords.MinLength:
				var v int
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "minLength" (attempting to unmarshal as int): %w`, err)
				}
				s.minLength = &v
				s.populatedFields |= MinLengthField
			case keywords.MinProperties:
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "minProperties" (attempting to unmarshal as uint): %w`, err)
				}
				s.minProperties = &v
				s.populatedFields |= MinPropertiesField
			case keywords.Minimum:
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "minimum" (attempting to unmarshal as float64): %w`, err)
				}
				s.minimum = &v
				s.populatedFields |= MinimumField
			case keywords.MultipleOf:
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "multipleOf" (attempting to unmarshal as float64): %w`, err)
				}
				s.multipleOf = &v
				s.populatedFields |= MultipleOfField
			case keywords.Not:
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "not": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					// Convert boolean to Schema object
					if b {
						s.not = &Schema{} // true schema - allow everything
					} else {
						// false schema - deny everything using "not": {}
						falseSchema := &Schema{not: &Schema{}}
						falseSchema.populatedFields |= NotField
						s.not = falseSchema
					}
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.not = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "not" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
				s.populatedFields |= NotField
			case keywords.OneOf:
				v, err := unmarshalSchemaOrBoolSlice(dec)
				if err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "oneOf" (attempting to unmarshal as []SchemaOrBool slice): %w`, err)
				}
				s.oneOf = v
				s.populatedFields |= OneOfField
			case keywords.Pattern:
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "pattern" (attempting to unmarshal as string): %w`, err)
				}
				s.pattern = &v
				s.populatedFields |= PatternField
			case keywords.PatternProperties:
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "patternProperties": %w`, err)
				}
				// First unmarshal as map[string]json.RawMessage
				var rawMap map[string]json.RawMessage
				if err := json.Unmarshal(rawData, &rawMap); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "patternProperties" (attempting to unmarshal as map): %w`, err)
				}
				// Convert each value to *Schema
				v := make(map[string]*Schema)
				for key, rawValue := range rawMap {
					// Try to decode as boolean first
					var b bool
					if err := json.Unmarshal(rawValue, &b); err == nil {
						// Convert boolean to Schema object
						if b {
							v[key] = &Schema{} // true schema - allow everything
						} else {
							// false schema - deny everything using "not": {}
							falseSchema := &Schema{not: &Schema{}}
							falseSchema.populatedFields |= NotField
							v[key] = falseSchema
						}
					} else {
						// Try to decode as Schema object
						var schema Schema
						if err := json.Unmarshal(rawValue, &schema); err == nil {
							v[key] = &schema
						} else {
							return fmt.Errorf(`json-schema: failed to decode value for field "patternProperties" key %q (attempting to unmarshal as Schema after bool failed): %w`, key, err)
						}
					}
				}
				s.patternProperties = v
				s.populatedFields |= PatternPropertiesField
			case keywords.PrefixItems:
				var v []*Schema
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "prefixItems" (attempting to unmarshal as []*Schema): %w`, err)
				}
				s.prefixItems = v
				s.populatedFields |= PrefixItemsField
			case keywords.Properties:
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "properties": %w`, err)
				}
				// First unmarshal as map[string]json.RawMessage
				var rawMap map[string]json.RawMessage
				if err := json.Unmarshal(rawData, &rawMap); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "properties" (attempting to unmarshal as map): %w`, err)
				}
				// Convert each value to *Schema
				v := make(map[string]*Schema)
				for key, rawValue := range rawMap {
					// Try to decode as boolean first
					var b bool
					if err := json.Unmarshal(rawValue, &b); err == nil {
						// Convert boolean to Schema object
						if b {
							v[key] = &Schema{} // true schema - allow everything
						} else {
							// false schema - deny everything using "not": {}
							falseSchema := &Schema{not: &Schema{}}
							falseSchema.populatedFields |= NotField
							v[key] = falseSchema
						}
					} else {
						// Try to decode as Schema object
						var schema Schema
						if err := json.Unmarshal(rawValue, &schema); err == nil {
							v[key] = &schema
						} else {
							return fmt.Errorf(`json-schema: failed to decode value for field "properties" key %q (attempting to unmarshal as Schema after bool failed): %w`, key, err)
						}
					}
				}
				s.properties = v
				s.populatedFields |= PropertiesField
			case keywords.PropertyNames:
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "propertyNames": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					// Convert boolean to Schema object
					if b {
						s.propertyNames = &Schema{} // true schema - allow everything
					} else {
						// false schema - deny everything using "not": {}
						falseSchema := &Schema{not: &Schema{}}
						falseSchema.populatedFields |= NotField
						s.propertyNames = falseSchema
					}
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.propertyNames = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "propertyNames" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
				s.populatedFields |= PropertyNamesField
			case keywords.Reference:
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$ref" (attempting to unmarshal as string): %w`, err)
				}
				s.reference = &v
				s.populatedFields |= ReferenceField
			case keywords.Required:
				var v []string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "required" (attempting to unmarshal as []string): %w`, err)
				}
				s.required = v
				s.populatedFields |= RequiredField
			case keywords.Schema:
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$schema" (attempting to unmarshal as string): %w`, err)
				}
				s.schema = v
			case keywords.Then:
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "then": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.thenSchema = BoolSchema(b)
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.thenSchema = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "then" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
				s.populatedFields |= ThenSchemaField
			case keywords.Type:
				var v PrimitiveTypes
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "type" (attempting to unmarshal as PrimitiveTypes): %w`, err)
				}
				s.types = v
				s.populatedFields |= TypesField
			case keywords.UnevaluatedItems:
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "unevaluatedItems": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.unevaluatedItems = BoolSchema(b)
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.unevaluatedItems = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "unevaluatedItems" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
				s.populatedFields |= UnevaluatedItemsField
			case keywords.UnevaluatedProperties:
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "unevaluatedProperties": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.unevaluatedProperties = BoolSchema(b)
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.unevaluatedProperties = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "unevaluatedProperties" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
				s.populatedFields |= UnevaluatedPropertiesField
			case keywords.UniqueItems:
				var v bool
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "uniqueItems" (attempting to unmarshal as bool): %w`, err)
				}
				s.uniqueItems = &v
				s.populatedFields |= UniqueItemsField
			case keywords.Vocabulary:
				var v map[string]bool
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$vocabulary" (attempting to unmarshal as map[string]bool): %w`, err)
				}
				s.vocabulary = v
				s.populatedFields |= VocabularyField
			default:
				// Skip unknown fields by consuming their values
				var discard json.RawMessage
				if err := dec.Decode(&discard); err != nil {
					return fmt.Errorf(`json-schema: failed to decode unknown field %q: %w`, tok, err)
				}
			}
		}
	}
	return nil
}
