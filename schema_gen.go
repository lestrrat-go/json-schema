package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/lestrrat-go/json-schema/internal/field"
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
	constantValue         *interface{}
	contains              SchemaOrBool
	contentEncoding       *string
	contentMediaType      *string
	contentSchema         *Schema
	defaultValue          *interface{}
	definitions           map[string]*Schema
	dependentRequired     map[string][]string
	dependentSchemas      map[string]SchemaOrBool
	dynamicAnchor         *string
	dynamicReference      *string
	elseSchema            SchemaOrBool
	enum                  []interface{}
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

func (s *Schema) Const() interface{} {
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

func (s *Schema) Default() interface{} {
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

func (s *Schema) Enum() []interface{} {
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
	Value interface{}
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	s.isRoot = true
	defer func() { s.isRoot = false }()
	fields := make([]pair, 0, 52)
	if s.HasAdditionalItems() {
		fields = append(fields, pair{Name: "additionalItems", Value: s.additionalItems})
	}
	if s.HasAdditionalProperties() {
		fields = append(fields, pair{Name: "additionalProperties", Value: s.additionalProperties})
	}
	if s.HasAllOf() {
		fields = append(fields, pair{Name: "allOf", Value: s.allOf})
	}
	if s.HasAnchor() {
		fields = append(fields, pair{Name: "$anchor", Value: *(s.anchor)})
	}
	if s.HasAnyOf() {
		fields = append(fields, pair{Name: "anyOf", Value: s.anyOf})
	}
	if s.HasComment() {
		fields = append(fields, pair{Name: "$comment", Value: *(s.comment)})
	}
	if s.HasConst() {
		fields = append(fields, pair{Name: "const", Value: *(s.constantValue)})
	}
	if s.HasContains() {
		fields = append(fields, pair{Name: "contains", Value: s.contains})
	}
	if s.HasContentEncoding() {
		fields = append(fields, pair{Name: "contentEncoding", Value: *(s.contentEncoding)})
	}
	if s.HasContentMediaType() {
		fields = append(fields, pair{Name: "contentMediaType", Value: *(s.contentMediaType)})
	}
	if s.HasContentSchema() {
		fields = append(fields, pair{Name: "contentSchema", Value: s.contentSchema})
	}
	if s.HasDefault() {
		fields = append(fields, pair{Name: "default", Value: *(s.defaultValue)})
	}
	if s.HasDefinitions() {
		fields = append(fields, pair{Name: "$defs", Value: s.definitions})
	}
	if s.HasDependentRequired() {
		fields = append(fields, pair{Name: "dependentRequired", Value: s.dependentRequired})
	}
	if s.HasDependentSchemas() {
		fields = append(fields, pair{Name: "dependentSchemas", Value: s.dependentSchemas})
	}
	if s.HasDynamicAnchor() {
		fields = append(fields, pair{Name: "$dynamicAnchor", Value: *(s.dynamicAnchor)})
	}
	if s.HasDynamicReference() {
		fields = append(fields, pair{Name: "$dynamicRef", Value: *(s.dynamicReference)})
	}
	if s.HasElseSchema() {
		fields = append(fields, pair{Name: "else", Value: s.elseSchema})
	}
	if s.HasEnum() {
		fields = append(fields, pair{Name: "enum", Value: s.enum})
	}
	if s.HasExclusiveMaximum() {
		fields = append(fields, pair{Name: "exclusiveMaximum", Value: *(s.exclusiveMaximum)})
	}
	if s.HasExclusiveMinimum() {
		fields = append(fields, pair{Name: "exclusiveMinimum", Value: *(s.exclusiveMinimum)})
	}
	if s.HasFormat() {
		fields = append(fields, pair{Name: "format", Value: *(s.format)})
	}
	if s.HasID() {
		fields = append(fields, pair{Name: "$id", Value: *(s.id)})
	}
	if s.HasIfSchema() {
		fields = append(fields, pair{Name: "if", Value: s.ifSchema})
	}
	if s.HasItems() {
		fields = append(fields, pair{Name: "items", Value: s.items})
	}
	if s.HasMaxContains() {
		fields = append(fields, pair{Name: "maxContains", Value: *(s.maxContains)})
	}
	if s.HasMaxItems() {
		fields = append(fields, pair{Name: "maxItems", Value: *(s.maxItems)})
	}
	if s.HasMaxLength() {
		fields = append(fields, pair{Name: "maxLength", Value: *(s.maxLength)})
	}
	if s.HasMaxProperties() {
		fields = append(fields, pair{Name: "maxProperties", Value: *(s.maxProperties)})
	}
	if s.HasMaximum() {
		fields = append(fields, pair{Name: "maximum", Value: *(s.maximum)})
	}
	if s.HasMinContains() {
		fields = append(fields, pair{Name: "minContains", Value: *(s.minContains)})
	}
	if s.HasMinItems() {
		fields = append(fields, pair{Name: "minItems", Value: *(s.minItems)})
	}
	if s.HasMinLength() {
		fields = append(fields, pair{Name: "minLength", Value: *(s.minLength)})
	}
	if s.HasMinProperties() {
		fields = append(fields, pair{Name: "minProperties", Value: *(s.minProperties)})
	}
	if s.HasMinimum() {
		fields = append(fields, pair{Name: "minimum", Value: *(s.minimum)})
	}
	if s.HasMultipleOf() {
		fields = append(fields, pair{Name: "multipleOf", Value: *(s.multipleOf)})
	}
	if s.HasNot() {
		fields = append(fields, pair{Name: "not", Value: s.not})
	}
	if s.HasOneOf() {
		fields = append(fields, pair{Name: "oneOf", Value: s.oneOf})
	}
	if s.HasPattern() {
		fields = append(fields, pair{Name: "pattern", Value: *(s.pattern)})
	}
	if s.HasPatternProperties() {
		fields = append(fields, pair{Name: "patternProperties", Value: s.patternProperties})
	}
	if s.HasPrefixItems() {
		fields = append(fields, pair{Name: "prefixItems", Value: s.prefixItems})
	}
	if s.HasProperties() {
		fields = append(fields, pair{Name: "properties", Value: s.properties})
	}
	if s.HasPropertyNames() {
		fields = append(fields, pair{Name: "propertyNames", Value: s.propertyNames})
	}
	if s.HasReference() {
		fields = append(fields, pair{Name: "$ref", Value: *(s.reference)})
	}
	if s.HasRequired() {
		fields = append(fields, pair{Name: "required", Value: s.required})
	}
	if v := s.schema; s.isRoot && v != "" {
		fields = append(fields, pair{Name: "$schema", Value: v})
	}
	if s.HasThenSchema() {
		fields = append(fields, pair{Name: "then", Value: s.thenSchema})
	}
	if s.HasTypes() {
		fields = append(fields, pair{Name: "type", Value: s.types})
	}
	if s.HasUnevaluatedItems() {
		fields = append(fields, pair{Name: "unevaluatedItems", Value: s.unevaluatedItems})
	}
	if s.HasUnevaluatedProperties() {
		fields = append(fields, pair{Name: "unevaluatedProperties", Value: s.unevaluatedProperties})
	}
	if s.HasUniqueItems() {
		fields = append(fields, pair{Name: "uniqueItems", Value: *(s.uniqueItems)})
	}
	if s.HasVocabulary() {
		fields = append(fields, pair{Name: "$vocabulary", Value: s.vocabulary})
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
		enc.Encode(field.Name)
		buf.WriteByte(':')
		enc.Encode(field.Value)
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
			case "additionalItems":
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "additionalItems": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.additionalItems = SchemaBool(b)
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
			case "additionalProperties":
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "additionalProperties": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.additionalProperties = SchemaBool(b)
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
			case "allOf":
				v, err := unmarshalSchemaOrBoolSlice(dec)
				if err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "allOf" (attempting to unmarshal as []SchemaOrBool slice): %w`, err)
				}
				s.allOf = v
				s.populatedFields |= AllOfField
			case "$anchor":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$anchor" (attempting to unmarshal as string): %w`, err)
				}
				s.anchor = &v
				s.populatedFields |= AnchorField
			case "anyOf":
				v, err := unmarshalSchemaOrBoolSlice(dec)
				if err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "anyOf" (attempting to unmarshal as []SchemaOrBool slice): %w`, err)
				}
				s.anyOf = v
				s.populatedFields |= AnyOfField
			case "$comment":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$comment" (attempting to unmarshal as string): %w`, err)
				}
				s.comment = &v
				s.populatedFields |= CommentField
			case "const":
				var v interface{}
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "const" (attempting to unmarshal as interface{}): %w`, err)
				}
				s.constantValue = &v
				s.populatedFields |= ConstField
			case "contains":
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "contains": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.contains = SchemaBool(b)
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
			case "contentEncoding":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "contentEncoding" (attempting to unmarshal as string): %w`, err)
				}
				s.contentEncoding = &v
				s.populatedFields |= ContentEncodingField
			case "contentMediaType":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "contentMediaType" (attempting to unmarshal as string): %w`, err)
				}
				s.contentMediaType = &v
				s.populatedFields |= ContentMediaTypeField
			case "contentSchema":
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
						s.contentSchema = &Schema{not: &Schema{}} // false schema - deny everything
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
			case "default":
				var v interface{}
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "default" (attempting to unmarshal as interface{}): %w`, err)
				}
				s.defaultValue = &v
				s.populatedFields |= DefaultField
			case "$defs":
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
							v[key] = &Schema{not: &Schema{}} // false schema - deny everything
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
			case "dependentRequired":
				var v map[string][]string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "dependentRequired" (attempting to unmarshal as map[string][]string): %w`, err)
				}
				s.dependentRequired = v
				s.populatedFields |= DependentRequiredField
			case "dependentSchemas":
				v, err := unmarshalSchemaOrBoolMap(dec)
				if err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "dependentSchemas" (attempting to unmarshal as map[string]SchemaOrBool): %w`, err)
				}
				s.dependentSchemas = v
				s.populatedFields |= DependentSchemasField
			case "$dynamicAnchor":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$dynamicAnchor" (attempting to unmarshal as string): %w`, err)
				}
				s.dynamicAnchor = &v
				s.populatedFields |= DynamicAnchorField
			case "$dynamicRef":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$dynamicRef" (attempting to unmarshal as string): %w`, err)
				}
				s.dynamicReference = &v
				s.populatedFields |= DynamicReferenceField
			case "else":
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "else": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.elseSchema = SchemaBool(b)
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
			case "enum":
				var v []interface{}
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "enum" (attempting to unmarshal as []interface{}): %w`, err)
				}
				s.enum = v
				s.populatedFields |= EnumField
			case "exclusiveMaximum":
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "exclusiveMaximum" (attempting to unmarshal as float64): %w`, err)
				}
				s.exclusiveMaximum = &v
				s.populatedFields |= ExclusiveMaximumField
			case "exclusiveMinimum":
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "exclusiveMinimum" (attempting to unmarshal as float64): %w`, err)
				}
				s.exclusiveMinimum = &v
				s.populatedFields |= ExclusiveMinimumField
			case "format":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "format" (attempting to unmarshal as string): %w`, err)
				}
				s.format = &v
				s.populatedFields |= FormatField
			case "$id":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$id" (attempting to unmarshal as string): %w`, err)
				}
				s.id = &v
				s.populatedFields |= IDField
			case "if":
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "if": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.ifSchema = SchemaBool(b)
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
			case "items":
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "items": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.items = SchemaBool(b)
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
			case "maxContains":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "maxContains" (attempting to unmarshal as uint): %w`, err)
				}
				s.maxContains = &v
				s.populatedFields |= MaxContainsField
			case "maxItems":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "maxItems" (attempting to unmarshal as uint): %w`, err)
				}
				s.maxItems = &v
				s.populatedFields |= MaxItemsField
			case "maxLength":
				var v int
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "maxLength" (attempting to unmarshal as int): %w`, err)
				}
				s.maxLength = &v
				s.populatedFields |= MaxLengthField
			case "maxProperties":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "maxProperties" (attempting to unmarshal as uint): %w`, err)
				}
				s.maxProperties = &v
				s.populatedFields |= MaxPropertiesField
			case "maximum":
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "maximum" (attempting to unmarshal as float64): %w`, err)
				}
				s.maximum = &v
				s.populatedFields |= MaximumField
			case "minContains":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "minContains" (attempting to unmarshal as uint): %w`, err)
				}
				s.minContains = &v
				s.populatedFields |= MinContainsField
			case "minItems":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "minItems" (attempting to unmarshal as uint): %w`, err)
				}
				s.minItems = &v
				s.populatedFields |= MinItemsField
			case "minLength":
				var v int
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "minLength" (attempting to unmarshal as int): %w`, err)
				}
				s.minLength = &v
				s.populatedFields |= MinLengthField
			case "minProperties":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "minProperties" (attempting to unmarshal as uint): %w`, err)
				}
				s.minProperties = &v
				s.populatedFields |= MinPropertiesField
			case "minimum":
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "minimum" (attempting to unmarshal as float64): %w`, err)
				}
				s.minimum = &v
				s.populatedFields |= MinimumField
			case "multipleOf":
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "multipleOf" (attempting to unmarshal as float64): %w`, err)
				}
				s.multipleOf = &v
				s.populatedFields |= MultipleOfField
			case "not":
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
						s.not = &Schema{not: &Schema{}} // false schema - deny everything
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
			case "oneOf":
				v, err := unmarshalSchemaOrBoolSlice(dec)
				if err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "oneOf" (attempting to unmarshal as []SchemaOrBool slice): %w`, err)
				}
				s.oneOf = v
				s.populatedFields |= OneOfField
			case "pattern":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "pattern" (attempting to unmarshal as string): %w`, err)
				}
				s.pattern = &v
				s.populatedFields |= PatternField
			case "patternProperties":
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
							v[key] = &Schema{not: &Schema{}} // false schema - deny everything
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
			case "prefixItems":
				var v []*Schema
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "prefixItems" (attempting to unmarshal as []*Schema): %w`, err)
				}
				s.prefixItems = v
				s.populatedFields |= PrefixItemsField
			case "properties":
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
							v[key] = &Schema{not: &Schema{}} // false schema - deny everything
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
			case "propertyNames":
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
						s.propertyNames = &Schema{not: &Schema{}} // false schema - deny everything
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
			case "$ref":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$ref" (attempting to unmarshal as string): %w`, err)
				}
				s.reference = &v
				s.populatedFields |= ReferenceField
			case "required":
				var v []string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "required" (attempting to unmarshal as []string): %w`, err)
				}
				s.required = v
				s.populatedFields |= RequiredField
			case "$schema":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$schema" (attempting to unmarshal as string): %w`, err)
				}
				s.schema = v
			case "then":
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "then": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.thenSchema = SchemaBool(b)
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
			case "type":
				var v PrimitiveTypes
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "type" (attempting to unmarshal as PrimitiveTypes): %w`, err)
				}
				s.types = v
				s.populatedFields |= TypesField
			case "unevaluatedItems":
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "unevaluatedItems": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.unevaluatedItems = SchemaBool(b)
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
			case "unevaluatedProperties":
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "unevaluatedProperties": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					s.unevaluatedProperties = SchemaBool(b)
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
			case "uniqueItems":
				var v bool
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "uniqueItems" (attempting to unmarshal as bool): %w`, err)
				}
				s.uniqueItems = &v
				s.populatedFields |= UniqueItemsField
			case "$vocabulary":
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
