package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

type Schema struct {
	isRoot                bool
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
	defaultValue          *interface{}
	definitions           map[string]*Schema
	dependentSchemas      map[string]*Schema
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
	patternProperties     map[string]*Schema
	properties            map[string]*Schema
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

func New() *Schema {
	return &Schema{
		schema: Version,
	}
}

func (s *Schema) HasAdditionalProperties() bool {
	return s.additionalProperties != nil
}

func (s *Schema) AdditionalProperties() SchemaOrBool {
	return s.additionalProperties
}

func (s *Schema) HasAllOf() bool {
	return s.allOf != nil
}

func (s *Schema) AllOf() []SchemaOrBool {
	return s.allOf
}

func (s *Schema) HasAnchor() bool {
	return s.anchor != nil
}

func (s *Schema) Anchor() string {
	return *(s.anchor)
}

func (s *Schema) HasAnyOf() bool {
	return s.anyOf != nil
}

func (s *Schema) AnyOf() []SchemaOrBool {
	return s.anyOf
}

func (s *Schema) HasComment() bool {
	return s.comment != nil
}

func (s *Schema) Comment() string {
	return *(s.comment)
}

func (s *Schema) HasConst() bool {
	return s.constantValue != nil
}

func (s *Schema) Const() interface{} {
	return *(s.constantValue)
}

func (s *Schema) HasContains() bool {
	return s.contains != nil
}

func (s *Schema) Contains() *Schema {
	return s.contains
}

func (s *Schema) HasContentEncoding() bool {
	return s.contentEncoding != nil
}

func (s *Schema) ContentEncoding() string {
	return *(s.contentEncoding)
}

func (s *Schema) HasContentMediaType() bool {
	return s.contentMediaType != nil
}

func (s *Schema) ContentMediaType() string {
	return *(s.contentMediaType)
}

func (s *Schema) HasContentSchema() bool {
	return s.contentSchema != nil
}

func (s *Schema) ContentSchema() *Schema {
	return s.contentSchema
}

func (s *Schema) HasDefault() bool {
	return s.defaultValue != nil
}

func (s *Schema) Default() interface{} {
	return *(s.defaultValue)
}

func (s *Schema) HasDefinitions() bool {
	return s.definitions != nil
}

func (s *Schema) Definitions() map[string]*Schema {
	return s.definitions
}

func (s *Schema) HasDependentSchemas() bool {
	return s.dependentSchemas != nil
}

func (s *Schema) DependentSchemas() map[string]*Schema {
	return s.dependentSchemas
}

func (s *Schema) HasDynamicReference() bool {
	return s.dynamicReference != nil
}

func (s *Schema) DynamicReference() string {
	return *(s.dynamicReference)
}

func (s *Schema) HasElseSchema() bool {
	return s.elseSchema != nil
}

func (s *Schema) ElseSchema() *Schema {
	return s.elseSchema
}

func (s *Schema) HasEnum() bool {
	return s.enum != nil
}

func (s *Schema) Enum() []interface{} {
	return s.enum
}

func (s *Schema) HasExclusiveMaximum() bool {
	return s.exclusiveMaximum != nil
}

func (s *Schema) ExclusiveMaximum() float64 {
	return *(s.exclusiveMaximum)
}

func (s *Schema) HasExclusiveMinimum() bool {
	return s.exclusiveMinimum != nil
}

func (s *Schema) ExclusiveMinimum() float64 {
	return *(s.exclusiveMinimum)
}

func (s *Schema) HasFormat() bool {
	return s.format != nil
}

func (s *Schema) Format() string {
	return *(s.format)
}

func (s *Schema) HasID() bool {
	return s.id != nil
}

func (s *Schema) ID() string {
	return *(s.id)
}

func (s *Schema) HasIfSchema() bool {
	return s.ifSchema != nil
}

func (s *Schema) IfSchema() *Schema {
	return s.ifSchema
}

func (s *Schema) HasItems() bool {
	return s.items != nil
}

func (s *Schema) Items() SchemaOrBool {
	return s.items
}

func (s *Schema) HasMaxContains() bool {
	return s.maxContains != nil
}

func (s *Schema) MaxContains() uint {
	return *(s.maxContains)
}

func (s *Schema) HasMaxItems() bool {
	return s.maxItems != nil
}

func (s *Schema) MaxItems() uint {
	return *(s.maxItems)
}

func (s *Schema) HasMaxLength() bool {
	return s.maxLength != nil
}

func (s *Schema) MaxLength() int {
	return *(s.maxLength)
}

func (s *Schema) HasMaxProperties() bool {
	return s.maxProperties != nil
}

func (s *Schema) MaxProperties() uint {
	return *(s.maxProperties)
}

func (s *Schema) HasMaximum() bool {
	return s.maximum != nil
}

func (s *Schema) Maximum() float64 {
	return *(s.maximum)
}

func (s *Schema) HasMinContains() bool {
	return s.minContains != nil
}

func (s *Schema) MinContains() uint {
	return *(s.minContains)
}

func (s *Schema) HasMinItems() bool {
	return s.minItems != nil
}

func (s *Schema) MinItems() uint {
	return *(s.minItems)
}

func (s *Schema) HasMinLength() bool {
	return s.minLength != nil
}

func (s *Schema) MinLength() int {
	return *(s.minLength)
}

func (s *Schema) HasMinProperties() bool {
	return s.minProperties != nil
}

func (s *Schema) MinProperties() uint {
	return *(s.minProperties)
}

func (s *Schema) HasMinimum() bool {
	return s.minimum != nil
}

func (s *Schema) Minimum() float64 {
	return *(s.minimum)
}

func (s *Schema) HasMultipleOf() bool {
	return s.multipleOf != nil
}

func (s *Schema) MultipleOf() float64 {
	return *(s.multipleOf)
}

func (s *Schema) HasNot() bool {
	return s.not != nil
}

func (s *Schema) Not() *Schema {
	return s.not
}

func (s *Schema) HasOneOf() bool {
	return s.oneOf != nil
}

func (s *Schema) OneOf() []SchemaOrBool {
	return s.oneOf
}

func (s *Schema) HasPattern() bool {
	return s.pattern != nil
}

func (s *Schema) Pattern() string {
	return *(s.pattern)
}

func (s *Schema) HasPatternProperties() bool {
	return s.patternProperties != nil
}

func (s *Schema) PatternProperties() map[string]*Schema {
	return s.patternProperties
}

func (s *Schema) HasProperties() bool {
	return s.properties != nil
}

func (s *Schema) Properties() map[string]*Schema {
	return s.properties
}

func (s *Schema) HasPropertyNames() bool {
	return s.propertyNames != nil
}

func (s *Schema) PropertyNames() *Schema {
	return s.propertyNames
}

func (s *Schema) HasReference() bool {
	return s.reference != nil
}

func (s *Schema) Reference() string {
	return *(s.reference)
}

func (s *Schema) HasRequired() bool {
	return s.required != nil
}

func (s *Schema) Required() []string {
	return s.required
}

func (s *Schema) Schema() string {
	return s.schema
}

func (s *Schema) HasThenSchema() bool {
	return s.thenSchema != nil
}

func (s *Schema) ThenSchema() *Schema {
	return s.thenSchema
}

func (s *Schema) HasTypes() bool {
	return s.types != nil
}

func (s *Schema) Types() PrimitiveTypes {
	return s.types
}

func (s *Schema) HasUnevaluatedItems() bool {
	return s.unevaluatedItems != nil
}

func (s *Schema) UnevaluatedItems() SchemaOrBool {
	return s.unevaluatedItems
}

func (s *Schema) HasUnevaluatedProperties() bool {
	return s.unevaluatedProperties != nil
}

func (s *Schema) UnevaluatedProperties() SchemaOrBool {
	return s.unevaluatedProperties
}

func (s *Schema) HasUniqueItems() bool {
	return s.uniqueItems != nil
}

func (s *Schema) UniqueItems() bool {
	return *(s.uniqueItems)
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
	fields := make([]pair, 0, 47)
	if v := s.additionalProperties; v != nil {
		fields = append(fields, pair{Name: "additionalProperties", Value: v})
	}
	if v := s.allOf; v != nil {
		fields = append(fields, pair{Name: "allOf", Value: v})
	}
	if v := s.anchor; v != nil {
		fields = append(fields, pair{Name: "$anchor", Value: *v})
	}
	if v := s.anyOf; v != nil {
		fields = append(fields, pair{Name: "anyOf", Value: v})
	}
	if v := s.comment; v != nil {
		fields = append(fields, pair{Name: "$comment", Value: *v})
	}
	if v := s.constantValue; v != nil {
		fields = append(fields, pair{Name: "const", Value: *v})
	}
	if v := s.contains; v != nil {
		fields = append(fields, pair{Name: "contains", Value: v})
	}
	if v := s.contentEncoding; v != nil {
		fields = append(fields, pair{Name: "contentEncoding", Value: *v})
	}
	if v := s.contentMediaType; v != nil {
		fields = append(fields, pair{Name: "contentMediaType", Value: *v})
	}
	if v := s.contentSchema; v != nil {
		fields = append(fields, pair{Name: "contentSchema", Value: v})
	}
	if v := s.defaultValue; v != nil {
		fields = append(fields, pair{Name: "default", Value: *v})
	}
	if v := s.definitions; v != nil {
		fields = append(fields, pair{Name: "$defs", Value: v})
	}
	if v := s.dependentSchemas; v != nil {
		fields = append(fields, pair{Name: "dependentSchemas", Value: v})
	}
	if v := s.dynamicReference; v != nil {
		fields = append(fields, pair{Name: "$dynamicRef", Value: *v})
	}
	if v := s.elseSchema; v != nil {
		fields = append(fields, pair{Name: "else", Value: v})
	}
	if v := s.enum; v != nil {
		fields = append(fields, pair{Name: "enum", Value: v})
	}
	if v := s.exclusiveMaximum; v != nil {
		fields = append(fields, pair{Name: "exclusiveMaximum", Value: *v})
	}
	if v := s.exclusiveMinimum; v != nil {
		fields = append(fields, pair{Name: "exclusiveMinimum", Value: *v})
	}
	if v := s.format; v != nil {
		fields = append(fields, pair{Name: "format", Value: *v})
	}
	if v := s.id; v != nil {
		fields = append(fields, pair{Name: "$id", Value: *v})
	}
	if v := s.ifSchema; v != nil {
		fields = append(fields, pair{Name: "if", Value: v})
	}
	if v := s.items; v != nil {
		fields = append(fields, pair{Name: "items", Value: v})
	}
	if v := s.maxContains; v != nil {
		fields = append(fields, pair{Name: "maxContains", Value: *v})
	}
	if v := s.maxItems; v != nil {
		fields = append(fields, pair{Name: "maxItems", Value: *v})
	}
	if v := s.maxLength; v != nil {
		fields = append(fields, pair{Name: "maxLength", Value: *v})
	}
	if v := s.maxProperties; v != nil {
		fields = append(fields, pair{Name: "maxProperties", Value: *v})
	}
	if v := s.maximum; v != nil {
		fields = append(fields, pair{Name: "maximum", Value: *v})
	}
	if v := s.minContains; v != nil {
		fields = append(fields, pair{Name: "minContains", Value: *v})
	}
	if v := s.minItems; v != nil {
		fields = append(fields, pair{Name: "minItems", Value: *v})
	}
	if v := s.minLength; v != nil {
		fields = append(fields, pair{Name: "minLength", Value: *v})
	}
	if v := s.minProperties; v != nil {
		fields = append(fields, pair{Name: "minProperties", Value: *v})
	}
	if v := s.minimum; v != nil {
		fields = append(fields, pair{Name: "minimum", Value: *v})
	}
	if v := s.multipleOf; v != nil {
		fields = append(fields, pair{Name: "multipleOf", Value: *v})
	}
	if v := s.not; v != nil {
		fields = append(fields, pair{Name: "not", Value: v})
	}
	if v := s.oneOf; v != nil {
		fields = append(fields, pair{Name: "oneOf", Value: v})
	}
	if v := s.pattern; v != nil {
		fields = append(fields, pair{Name: "pattern", Value: *v})
	}
	if v := s.patternProperties; v != nil {
		fields = append(fields, pair{Name: "patternProperties", Value: v})
	}
	if v := s.properties; v != nil {
		fields = append(fields, pair{Name: "properties", Value: v})
	}
	if v := s.propertyNames; v != nil {
		fields = append(fields, pair{Name: "propertyNames", Value: v})
	}
	if v := s.reference; v != nil {
		fields = append(fields, pair{Name: "$ref", Value: *v})
	}
	if v := s.required; v != nil {
		fields = append(fields, pair{Name: "required", Value: v})
	}
	if v := s.schema; s.isRoot && v != "" {
		fields = append(fields, pair{Name: "$schema", Value: v})
	}
	if v := s.thenSchema; v != nil {
		fields = append(fields, pair{Name: "then", Value: v})
	}
	if v := s.types; v != nil {
		fields = append(fields, pair{Name: "type", Value: v})
	}
	if v := s.unevaluatedItems; v != nil {
		fields = append(fields, pair{Name: "unevaluatedItems", Value: v})
	}
	if v := s.unevaluatedProperties; v != nil {
		fields = append(fields, pair{Name: "unevaluatedProperties", Value: v})
	}
	if v := s.uniqueItems; v != nil {
		fields = append(fields, pair{Name: "uniqueItems", Value: *v})
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
			case "allOf":
				v, err := unmarshalSchemaOrBoolSlice(dec)
				if err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "allOf" (attempting to unmarshal as []SchemaOrBool slice): %w`, err)
				}
				s.allOf = v
			case "$anchor":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$anchor" (attempting to unmarshal as string): %w`, err)
				}
				s.anchor = &v
			case "anyOf":
				v, err := unmarshalSchemaOrBoolSlice(dec)
				if err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "anyOf" (attempting to unmarshal as []SchemaOrBool slice): %w`, err)
				}
				s.anyOf = v
			case "$comment":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$comment" (attempting to unmarshal as string): %w`, err)
				}
				s.comment = &v
			case "const":
				var v interface{}
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "const" (attempting to unmarshal as interface{}): %w`, err)
				}
				s.constantValue = &v
			case "contains":
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "contains": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					// Convert boolean to Schema object
					if b {
						s.contains = &Schema{} // true schema - allow everything
					} else {
						s.contains = &Schema{not: &Schema{}} // false schema - deny everything
					}
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.contains = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "contains" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
			case "contentEncoding":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "contentEncoding" (attempting to unmarshal as string): %w`, err)
				}
				s.contentEncoding = &v
			case "contentMediaType":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "contentMediaType" (attempting to unmarshal as string): %w`, err)
				}
				s.contentMediaType = &v
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
			case "default":
				var v interface{}
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "default" (attempting to unmarshal as interface{}): %w`, err)
				}
				s.defaultValue = &v
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
			case "dependentSchemas":
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "dependentSchemas": %w`, err)
				}
				// First unmarshal as map[string]json.RawMessage
				var rawMap map[string]json.RawMessage
				if err := json.Unmarshal(rawData, &rawMap); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "dependentSchemas" (attempting to unmarshal as map): %w`, err)
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
							return fmt.Errorf(`json-schema: failed to decode value for field "dependentSchemas" key %q (attempting to unmarshal as Schema after bool failed): %w`, key, err)
						}
					}
				}
				s.dependentSchemas = v
			case "$dynamicRef":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$dynamicRef" (attempting to unmarshal as string): %w`, err)
				}
				s.dynamicReference = &v
			case "else":
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "else": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					// Convert boolean to Schema object
					if b {
						s.elseSchema = &Schema{} // true schema - allow everything
					} else {
						s.elseSchema = &Schema{not: &Schema{}} // false schema - deny everything
					}
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.elseSchema = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "else" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
			case "enum":
				var v []interface{}
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "enum" (attempting to unmarshal as []interface{}): %w`, err)
				}
				s.enum = v
			case "exclusiveMaximum":
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "exclusiveMaximum" (attempting to unmarshal as float64): %w`, err)
				}
				s.exclusiveMaximum = &v
			case "exclusiveMinimum":
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "exclusiveMinimum" (attempting to unmarshal as float64): %w`, err)
				}
				s.exclusiveMinimum = &v
			case "format":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "format" (attempting to unmarshal as string): %w`, err)
				}
				s.format = &v
			case "$id":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$id" (attempting to unmarshal as string): %w`, err)
				}
				s.id = &v
			case "if":
				var rawData json.RawMessage
				if err := dec.Decode(&rawData); err != nil {
					return fmt.Errorf(`json-schema: failed to decode raw data for field "if": %w`, err)
				}
				// Try to decode as boolean first
				var b bool
				if err := json.Unmarshal(rawData, &b); err == nil {
					// Convert boolean to Schema object
					if b {
						s.ifSchema = &Schema{} // true schema - allow everything
					} else {
						s.ifSchema = &Schema{not: &Schema{}} // false schema - deny everything
					}
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.ifSchema = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "if" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
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
			case "maxContains":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "maxContains" (attempting to unmarshal as uint): %w`, err)
				}
				s.maxContains = &v
			case "maxItems":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "maxItems" (attempting to unmarshal as uint): %w`, err)
				}
				s.maxItems = &v
			case "maxLength":
				var v int
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "maxLength" (attempting to unmarshal as int): %w`, err)
				}
				s.maxLength = &v
			case "maxProperties":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "maxProperties" (attempting to unmarshal as uint): %w`, err)
				}
				s.maxProperties = &v
			case "maximum":
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "maximum" (attempting to unmarshal as float64): %w`, err)
				}
				s.maximum = &v
			case "minContains":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "minContains" (attempting to unmarshal as uint): %w`, err)
				}
				s.minContains = &v
			case "minItems":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "minItems" (attempting to unmarshal as uint): %w`, err)
				}
				s.minItems = &v
			case "minLength":
				var v int
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "minLength" (attempting to unmarshal as int): %w`, err)
				}
				s.minLength = &v
			case "minProperties":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "minProperties" (attempting to unmarshal as uint): %w`, err)
				}
				s.minProperties = &v
			case "minimum":
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "minimum" (attempting to unmarshal as float64): %w`, err)
				}
				s.minimum = &v
			case "multipleOf":
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "multipleOf" (attempting to unmarshal as float64): %w`, err)
				}
				s.multipleOf = &v
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
			case "oneOf":
				v, err := unmarshalSchemaOrBoolSlice(dec)
				if err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "oneOf" (attempting to unmarshal as []SchemaOrBool slice): %w`, err)
				}
				s.oneOf = v
			case "pattern":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "pattern" (attempting to unmarshal as string): %w`, err)
				}
				s.pattern = &v
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
			case "$ref":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "$ref" (attempting to unmarshal as string): %w`, err)
				}
				s.reference = &v
			case "required":
				var v []string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "required" (attempting to unmarshal as []string): %w`, err)
				}
				s.required = v
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
					// Convert boolean to Schema object
					if b {
						s.thenSchema = &Schema{} // true schema - allow everything
					} else {
						s.thenSchema = &Schema{not: &Schema{}} // false schema - deny everything
					}
				} else {
					// Try to decode as Schema object
					var schema Schema
					if err := json.Unmarshal(rawData, &schema); err == nil {
						s.thenSchema = &schema
					} else {
						return fmt.Errorf(`json-schema: failed to decode value for field "then" (attempting to unmarshal as Schema after bool failed): %w`, err)
					}
				}
			case "type":
				var v PrimitiveTypes
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "type" (attempting to unmarshal as PrimitiveTypes): %w`, err)
				}
				s.types = v
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
			case "uniqueItems":
				var v bool
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`json-schema: failed to decode value for field "uniqueItems" (attempting to unmarshal as bool): %w`, err)
				}
				s.uniqueItems = &v
			}
		}
	}
	return nil
}
