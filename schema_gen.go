package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/lestrrat-go/json-schema/internal/field"
	"github.com/lestrrat-go/json-schema/internal/pool"
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
	SchemaField                = field.Schema
	ThenSchemaField            = field.ThenSchema
	TypesField                 = field.Types
	UnevaluatedItemsField      = field.UnevaluatedItems
	UnevaluatedPropertiesField = field.UnevaluatedProperties
	UniqueItemsField           = field.UniqueItems
	VocabularyField            = field.Vocabulary
)

type Schema struct {
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
	schema                *string
	thenSchema            SchemaOrBool
	types                 PrimitiveTypes
	unevaluatedItems      SchemaOrBool
	unevaluatedProperties SchemaOrBool
	uniqueItems           *bool
	vocabulary            map[string]bool
}

func New() *Schema {
	return &Schema{}
}

// Has checks if ALL of the specified field flags are set in the schema.
// It uses bitwise AND to verify that every flag in the parameter is present.
//
// Single field check:
//   if schema.Has(AnchorField) {
//       // Schema has an anchor field set
//   }
//
// Multiple field check (ALL must be present):
//   if schema.Has(AnchorField | PropertiesField | TypesField) {
//       // Schema has anchor AND properties AND types all set
//   }
//
// Common patterns:
//   - schema.Has(PropertiesField) - check if schema defines properties
//   - schema.Has(TypesField) - check if schema specifies allowed types  
//   - schema.Has(MinimumField | MaximumField) - check if both min and max constraints are set
//   - schema.Has(AllOfField | AnyOfField | OneOfField) - check if all composition keywords are present
func (s *Schema) Has(flags FieldFlag) bool {
	return (s.populatedFields & flags) == flags
}

// HasAny checks if ANY of the specified field flags are set in the schema.
// It uses bitwise AND to verify that at least one flag in the parameter is present.
//
// Single field check (same as Has for single fields):
//   if schema.HasAny(AnchorField) {
//       // Schema has an anchor field set
//   }
//
// Multiple field check (ANY can be present):
//   if schema.HasAny(AnchorField | PropertiesField | TypesField) {
//       // Schema has anchor OR properties OR types (or any combination)
//   }
//
// Common patterns:
//   - schema.HasAny(AllOfField | AnyOfField | OneOfField) - check if any composition is used
//   - schema.HasAny(MinimumField | MaximumField) - check if any numeric constraint is set
//   - schema.HasAny(MinLengthField | MaxLengthField | PatternField) - check if any string constraint exists
//   - schema.HasAny(PropertiesField | PatternPropertiesField | AdditionalPropertiesField) - check if any property rules exist
//
// Use HasAny when you want to detect if a schema uses any validation from a group of related fields.
// Use Has when you need to ensure specific combinations of fields are all present together.
func (s *Schema) HasAny(flags FieldFlag) bool {
	return (s.populatedFields & flags) != 0
}

func (s *Schema) AdditionalItems() SchemaOrBool {
	return s.additionalItems
}

func (s *Schema) AdditionalProperties() SchemaOrBool {
	return s.additionalProperties
}

func (s *Schema) AllOf() []SchemaOrBool {
	return s.allOf
}

func (s *Schema) Anchor() string {
	if s.anchor == nil {
		var zero string
		return zero
	}
	return *(s.anchor)
}

func (s *Schema) AnyOf() []SchemaOrBool {
	return s.anyOf
}

func (s *Schema) Comment() string {
	if s.comment == nil {
		var zero string
		return zero
	}
	return *(s.comment)
}

func (s *Schema) Const() any {
	if s.constantValue == nil {
		var zero any
		return zero
	}
	return *(s.constantValue)
}

func (s *Schema) Contains() SchemaOrBool {
	return s.contains
}

func (s *Schema) ContentEncoding() string {
	if s.contentEncoding == nil {
		var zero string
		return zero
	}
	return *(s.contentEncoding)
}

func (s *Schema) ContentMediaType() string {
	if s.contentMediaType == nil {
		var zero string
		return zero
	}
	return *(s.contentMediaType)
}

func (s *Schema) ContentSchema() *Schema {
	return s.contentSchema
}

func (s *Schema) Default() any {
	if s.defaultValue == nil {
		var zero any
		return zero
	}
	return *(s.defaultValue)
}

func (s *Schema) Definitions() *SchemaMap {
	return &SchemaMap{data: s.definitions}
}

func (s *Schema) DependentRequired() map[string][]string {
	return s.dependentRequired
}

func (s *Schema) DependentSchemas() map[string]SchemaOrBool {
	return s.dependentSchemas
}

func (s *Schema) DynamicAnchor() string {
	if s.dynamicAnchor == nil {
		var zero string
		return zero
	}
	return *(s.dynamicAnchor)
}

func (s *Schema) DynamicReference() string {
	if s.dynamicReference == nil {
		var zero string
		return zero
	}
	return *(s.dynamicReference)
}

func (s *Schema) ElseSchema() SchemaOrBool {
	return s.elseSchema
}

func (s *Schema) Enum() []any {
	return s.enum
}

func (s *Schema) ExclusiveMaximum() float64 {
	if s.exclusiveMaximum == nil {
		var zero float64
		return zero
	}
	return *(s.exclusiveMaximum)
}

func (s *Schema) ExclusiveMinimum() float64 {
	if s.exclusiveMinimum == nil {
		var zero float64
		return zero
	}
	return *(s.exclusiveMinimum)
}

func (s *Schema) Format() string {
	if s.format == nil {
		var zero string
		return zero
	}
	return *(s.format)
}

func (s *Schema) ID() string {
	if s.id == nil {
		var zero string
		return zero
	}
	return *(s.id)
}

func (s *Schema) IfSchema() SchemaOrBool {
	return s.ifSchema
}

func (s *Schema) Items() SchemaOrBool {
	return s.items
}

func (s *Schema) MaxContains() uint {
	if s.maxContains == nil {
		var zero uint
		return zero
	}
	return *(s.maxContains)
}

func (s *Schema) MaxItems() uint {
	if s.maxItems == nil {
		var zero uint
		return zero
	}
	return *(s.maxItems)
}

func (s *Schema) MaxLength() int {
	if s.maxLength == nil {
		var zero int
		return zero
	}
	return *(s.maxLength)
}

func (s *Schema) MaxProperties() uint {
	if s.maxProperties == nil {
		var zero uint
		return zero
	}
	return *(s.maxProperties)
}

func (s *Schema) Maximum() float64 {
	if s.maximum == nil {
		var zero float64
		return zero
	}
	return *(s.maximum)
}

func (s *Schema) MinContains() uint {
	if s.minContains == nil {
		var zero uint
		return zero
	}
	return *(s.minContains)
}

func (s *Schema) MinItems() uint {
	if s.minItems == nil {
		var zero uint
		return zero
	}
	return *(s.minItems)
}

func (s *Schema) MinLength() int {
	if s.minLength == nil {
		var zero int
		return zero
	}
	return *(s.minLength)
}

func (s *Schema) MinProperties() uint {
	if s.minProperties == nil {
		var zero uint
		return zero
	}
	return *(s.minProperties)
}

func (s *Schema) Minimum() float64 {
	if s.minimum == nil {
		var zero float64
		return zero
	}
	return *(s.minimum)
}

func (s *Schema) MultipleOf() float64 {
	if s.multipleOf == nil {
		var zero float64
		return zero
	}
	return *(s.multipleOf)
}

func (s *Schema) Not() *Schema {
	return s.not
}

func (s *Schema) OneOf() []SchemaOrBool {
	return s.oneOf
}

func (s *Schema) Pattern() string {
	if s.pattern == nil {
		var zero string
		return zero
	}
	return *(s.pattern)
}

func (s *Schema) PatternProperties() *SchemaMap {
	return &SchemaMap{data: s.patternProperties}
}

func (s *Schema) PrefixItems() []*Schema {
	return s.prefixItems
}

func (s *Schema) Properties() *SchemaMap {
	return &SchemaMap{data: s.properties}
}

func (s *Schema) PropertyNames() *Schema {
	return s.propertyNames
}

func (s *Schema) Reference() string {
	if s.reference == nil {
		var zero string
		return zero
	}
	return *(s.reference)
}

func (s *Schema) Required() []string {
	return s.required
}

func (s *Schema) Schema() string {
	if s.schema == nil {
		var zero string
		return zero
	}
	return *(s.schema)
}

func (s *Schema) ThenSchema() SchemaOrBool {
	return s.thenSchema
}

func (s *Schema) Types() PrimitiveTypes {
	return s.types
}

func (s *Schema) UnevaluatedItems() SchemaOrBool {
	return s.unevaluatedItems
}

func (s *Schema) UnevaluatedProperties() SchemaOrBool {
	return s.unevaluatedProperties
}

func (s *Schema) UniqueItems() bool {
	if s.uniqueItems == nil {
		var zero bool
		return zero
	}
	return *(s.uniqueItems)
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

func (s *Schema) MarshalJSON() ([]byte, error) {
	fields := pool.PairSlice().GetCapacity(52)
	defer pool.PairSlice().Put(fields)
	if s.Has(AdditionalItemsField) {
		fields = append(fields, pool.Pair{Name: keywords.AdditionalItems, Value: s.additionalItems})
	}
	if s.Has(AdditionalPropertiesField) {
		fields = append(fields, pool.Pair{Name: keywords.AdditionalProperties, Value: s.additionalProperties})
	}
	if s.Has(AllOfField) {
		fields = append(fields, pool.Pair{Name: keywords.AllOf, Value: s.allOf})
	}
	if s.Has(AnchorField) {
		fields = append(fields, pool.Pair{Name: keywords.Anchor, Value: *(s.anchor)})
	}
	if s.Has(AnyOfField) {
		fields = append(fields, pool.Pair{Name: keywords.AnyOf, Value: s.anyOf})
	}
	if s.Has(CommentField) {
		fields = append(fields, pool.Pair{Name: keywords.Comment, Value: *(s.comment)})
	}
	if s.Has(ConstField) {
		fields = append(fields, pool.Pair{Name: keywords.Const, Value: *(s.constantValue)})
	}
	if s.Has(ContainsField) {
		fields = append(fields, pool.Pair{Name: keywords.Contains, Value: s.contains})
	}
	if s.Has(ContentEncodingField) {
		fields = append(fields, pool.Pair{Name: keywords.ContentEncoding, Value: *(s.contentEncoding)})
	}
	if s.Has(ContentMediaTypeField) {
		fields = append(fields, pool.Pair{Name: keywords.ContentMediaType, Value: *(s.contentMediaType)})
	}
	if s.Has(ContentSchemaField) {
		fields = append(fields, pool.Pair{Name: keywords.ContentSchema, Value: s.contentSchema})
	}
	if s.Has(DefaultField) {
		fields = append(fields, pool.Pair{Name: keywords.Default, Value: *(s.defaultValue)})
	}
	if s.Has(DefinitionsField) {
		fields = append(fields, pool.Pair{Name: keywords.Definitions, Value: s.definitions})
	}
	if s.Has(DependentRequiredField) {
		fields = append(fields, pool.Pair{Name: keywords.DependentRequired, Value: s.dependentRequired})
	}
	if s.Has(DependentSchemasField) {
		fields = append(fields, pool.Pair{Name: keywords.DependentSchemas, Value: s.dependentSchemas})
	}
	if s.Has(DynamicAnchorField) {
		fields = append(fields, pool.Pair{Name: keywords.DynamicAnchor, Value: *(s.dynamicAnchor)})
	}
	if s.Has(DynamicReferenceField) {
		fields = append(fields, pool.Pair{Name: keywords.DynamicReference, Value: *(s.dynamicReference)})
	}
	if s.Has(ElseSchemaField) {
		fields = append(fields, pool.Pair{Name: keywords.Else, Value: s.elseSchema})
	}
	if s.Has(EnumField) {
		fields = append(fields, pool.Pair{Name: keywords.Enum, Value: s.enum})
	}
	if s.Has(ExclusiveMaximumField) {
		fields = append(fields, pool.Pair{Name: keywords.ExclusiveMaximum, Value: *(s.exclusiveMaximum)})
	}
	if s.Has(ExclusiveMinimumField) {
		fields = append(fields, pool.Pair{Name: keywords.ExclusiveMinimum, Value: *(s.exclusiveMinimum)})
	}
	if s.Has(FormatField) {
		fields = append(fields, pool.Pair{Name: keywords.Format, Value: *(s.format)})
	}
	if s.Has(IDField) {
		fields = append(fields, pool.Pair{Name: keywords.ID, Value: *(s.id)})
	}
	if s.Has(IfSchemaField) {
		fields = append(fields, pool.Pair{Name: keywords.If, Value: s.ifSchema})
	}
	if s.Has(ItemsField) {
		fields = append(fields, pool.Pair{Name: keywords.Items, Value: s.items})
	}
	if s.Has(MaxContainsField) {
		fields = append(fields, pool.Pair{Name: keywords.MaxContains, Value: *(s.maxContains)})
	}
	if s.Has(MaxItemsField) {
		fields = append(fields, pool.Pair{Name: keywords.MaxItems, Value: *(s.maxItems)})
	}
	if s.Has(MaxLengthField) {
		fields = append(fields, pool.Pair{Name: keywords.MaxLength, Value: *(s.maxLength)})
	}
	if s.Has(MaxPropertiesField) {
		fields = append(fields, pool.Pair{Name: keywords.MaxProperties, Value: *(s.maxProperties)})
	}
	if s.Has(MaximumField) {
		fields = append(fields, pool.Pair{Name: keywords.Maximum, Value: *(s.maximum)})
	}
	if s.Has(MinContainsField) {
		fields = append(fields, pool.Pair{Name: keywords.MinContains, Value: *(s.minContains)})
	}
	if s.Has(MinItemsField) {
		fields = append(fields, pool.Pair{Name: keywords.MinItems, Value: *(s.minItems)})
	}
	if s.Has(MinLengthField) {
		fields = append(fields, pool.Pair{Name: keywords.MinLength, Value: *(s.minLength)})
	}
	if s.Has(MinPropertiesField) {
		fields = append(fields, pool.Pair{Name: keywords.MinProperties, Value: *(s.minProperties)})
	}
	if s.Has(MinimumField) {
		fields = append(fields, pool.Pair{Name: keywords.Minimum, Value: *(s.minimum)})
	}
	if s.Has(MultipleOfField) {
		fields = append(fields, pool.Pair{Name: keywords.MultipleOf, Value: *(s.multipleOf)})
	}
	if s.Has(NotField) {
		fields = append(fields, pool.Pair{Name: keywords.Not, Value: s.not})
	}
	if s.Has(OneOfField) {
		fields = append(fields, pool.Pair{Name: keywords.OneOf, Value: s.oneOf})
	}
	if s.Has(PatternField) {
		fields = append(fields, pool.Pair{Name: keywords.Pattern, Value: *(s.pattern)})
	}
	if s.Has(PatternPropertiesField) {
		fields = append(fields, pool.Pair{Name: keywords.PatternProperties, Value: s.patternProperties})
	}
	if s.Has(PrefixItemsField) {
		fields = append(fields, pool.Pair{Name: keywords.PrefixItems, Value: s.prefixItems})
	}
	if s.Has(PropertiesField) {
		fields = append(fields, pool.Pair{Name: keywords.Properties, Value: s.properties})
	}
	if s.Has(PropertyNamesField) {
		fields = append(fields, pool.Pair{Name: keywords.PropertyNames, Value: s.propertyNames})
	}
	if s.Has(ReferenceField) {
		fields = append(fields, pool.Pair{Name: keywords.Reference, Value: *(s.reference)})
	}
	if s.Has(RequiredField) {
		fields = append(fields, pool.Pair{Name: keywords.Required, Value: s.required})
	}
	if s.Has(SchemaField) {
		fields = append(fields, pool.Pair{Name: keywords.Schema, Value: *(s.schema)})
	}
	if s.Has(ThenSchemaField) {
		fields = append(fields, pool.Pair{Name: keywords.Then, Value: s.thenSchema})
	}
	if s.Has(TypesField) {
		fields = append(fields, pool.Pair{Name: keywords.Type, Value: s.types})
	}
	if s.Has(UnevaluatedItemsField) {
		fields = append(fields, pool.Pair{Name: keywords.UnevaluatedItems, Value: s.unevaluatedItems})
	}
	if s.Has(UnevaluatedPropertiesField) {
		fields = append(fields, pool.Pair{Name: keywords.UnevaluatedProperties, Value: s.unevaluatedProperties})
	}
	if s.Has(UniqueItemsField) {
		fields = append(fields, pool.Pair{Name: keywords.UniqueItems, Value: *(s.uniqueItems)})
	}
	if s.Has(VocabularyField) {
		fields = append(fields, pool.Pair{Name: keywords.Vocabulary, Value: s.vocabulary})
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
				s.schema = &v
				s.populatedFields |= SchemaField
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
