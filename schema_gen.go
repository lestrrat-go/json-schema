package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

type Schema struct {
	isRoot                bool
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
	patternProperties     map[string]*Schema
	properties            map[string]*Schema
	propertyNames         map[string]*Schema
	reference             *string
	required              *bool
	schema                string
	types                 []PrimitiveType
	unevaluatedItems      *Schema
	unevaluatedProperties *Schema
	uniqueItems           *bool
}

func New() *Schema {
	return &Schema{
		schema: Version,
	}
}

func (s *Schema) AdditionalProperties() *Schema {
	return s.additionalProperties
}

func (s *Schema) AllOf() []*Schema {
	return s.allOf
}

func (s *Schema) Anchor() string {
	return *(s.anchor)
}

func (s *Schema) AnyOf() []*Schema {
	return s.anyOf
}

func (s *Schema) Comment() string {
	return *(s.comment)
}

func (s *Schema) Const() interface{} {
	return *(s.constantValue)
}

func (s *Schema) Contains() *Schema {
	return s.contains
}

func (s *Schema) Definitions() string {
	return *(s.definitions)
}

func (s *Schema) DynamicReference() string {
	return *(s.dynamicReference)
}

func (s *Schema) Enum() []interface{} {
	return s.enum
}

func (s *Schema) ExclusiveMaximum() float64 {
	return *(s.exclusiveMaximum)
}

func (s *Schema) ExclusiveMinimum() float64 {
	return *(s.exclusiveMinimum)
}

func (s *Schema) ID() string {
	return *(s.id)
}

func (s *Schema) Items() *Schema {
	return s.items
}

func (s *Schema) MaxContains() uint {
	return *(s.maxContains)
}

func (s *Schema) MaxItems() uint {
	return *(s.maxItems)
}

func (s *Schema) MaxLength() int {
	return *(s.maxLength)
}

func (s *Schema) MaxProperties() uint {
	return *(s.maxProperties)
}

func (s *Schema) Maximum() float64 {
	return *(s.maximum)
}

func (s *Schema) MinContains() uint {
	return *(s.minContains)
}

func (s *Schema) MinItems() uint {
	return *(s.minItems)
}

func (s *Schema) MinLength() int {
	return *(s.minLength)
}

func (s *Schema) MinProperties() uint {
	return *(s.minProperties)
}

func (s *Schema) Minimum() float64 {
	return *(s.minimum)
}

func (s *Schema) MultipleOf() float64 {
	return *(s.multipleOf)
}

func (s *Schema) Not() *Schema {
	return s.not
}

func (s *Schema) OneOf() []*Schema {
	return s.oneOf
}

func (s *Schema) Pattern() string {
	return *(s.pattern)
}

func (s *Schema) PatternProperties() map[string]*Schema {
	return s.patternProperties
}

func (s *Schema) Properties() map[string]*Schema {
	return s.properties
}

func (s *Schema) PropertyNames() map[string]*Schema {
	return s.propertyNames
}

func (s *Schema) Reference() string {
	return *(s.reference)
}

func (s *Schema) Required() bool {
	return *(s.required)
}

func (s *Schema) Schema() string {
	return s.schema
}

func (s *Schema) Types() []PrimitiveType {
	return s.types
}

func (s *Schema) UnevaluatedItems() *Schema {
	return s.unevaluatedItems
}

func (s *Schema) UnevaluatedProperties() *Schema {
	return s.unevaluatedProperties
}

func (s *Schema) UniqueItems() bool {
	return *(s.uniqueItems)
}

type pair struct {
	Name  string
	Value interface{}
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	s.isRoot = true
	defer func() { s.isRoot = false }()
	fields := make([]pair, 0, 38)
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
	if v := s.definitions; v != nil {
		fields = append(fields, pair{Name: "$defs", Value: *v})
	}
	if v := s.dynamicReference; v != nil {
		fields = append(fields, pair{Name: "$dynamicRef", Value: *v})
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
	if v := s.id; v != nil {
		fields = append(fields, pair{Name: "$id", Value: *v})
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
		fields = append(fields, pair{Name: "$reference", Value: *v})
	}
	if v := s.required; v != nil {
		fields = append(fields, pair{Name: "required", Value: *v})
	}
	if v := s.schema; s.isRoot && v != "" {
		fields = append(fields, pair{Name: "$schema", Value: v})
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
		return fields[i].Name < fields[j].Name
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
			return fmt.Errorf(`error reading token: %w`, err)
		}
		switch tok := tok.(type) {
		case json.Delim:
			// Assuming we're doing everything correctly, we should ONLY
			// get either '{' or '}' here.
			if tok == '}' { // End of object
				break LOOP
			} else if tok != '{' {
				return fmt.Errorf(`expected '{', but got '%c'`, tok)
			}
		case string: // Objects can only have string keys
			switch tok {
			case "additionalProperties":
				var v *Schema
				var tmp Schema
				if err := dec.Decode(&tmp); err == nil {
					v = &tmp
				} else {
					var b bool
					if err = dec.Decode(&b); err != nil {
						return fmt.Errorf(`failed to decode value for field "additionalProperties": %w`, err)
					}
					if b {
						v = &Schema{}
					}
				}
				s.additionalProperties = v
			case "allOf":
				var v []*Schema
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "allOf": %w`, err)
				}
				s.allOf = v
			case "$anchor":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "$anchor": %w`, err)
				}
				s.anchor = &v
			case "anyOf":
				var v []*Schema
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "anyOf": %w`, err)
				}
				s.anyOf = v
			case "$comment":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "$comment": %w`, err)
				}
				s.comment = &v
			case "const":
				var v interface{}
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "const": %w`, err)
				}
				s.constantValue = &v
			case "contains":
				var v *Schema
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "contains": %w`, err)
				}
				s.contains = v
			case "$defs":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "$defs": %w`, err)
				}
				s.definitions = &v
			case "$dynamicRef":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "$dynamicRef": %w`, err)
				}
				s.dynamicReference = &v
			case "enum":
				var v []interface{}
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "enum": %w`, err)
				}
				s.enum = v
			case "exclusiveMaximum":
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "exclusiveMaximum": %w`, err)
				}
				s.exclusiveMaximum = &v
			case "exclusiveMinimum":
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "exclusiveMinimum": %w`, err)
				}
				s.exclusiveMinimum = &v
			case "$id":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "$id": %w`, err)
				}
				s.id = &v
			case "items":
				var v *Schema
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "items": %w`, err)
				}
				s.items = v
			case "maxContains":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "maxContains": %w`, err)
				}
				s.maxContains = &v
			case "maxItems":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "maxItems": %w`, err)
				}
				s.maxItems = &v
			case "maxLength":
				var v int
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "maxLength": %w`, err)
				}
				s.maxLength = &v
			case "maxProperties":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "maxProperties": %w`, err)
				}
				s.maxProperties = &v
			case "maximum":
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "maximum": %w`, err)
				}
				s.maximum = &v
			case "minContains":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "minContains": %w`, err)
				}
				s.minContains = &v
			case "minItems":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "minItems": %w`, err)
				}
				s.minItems = &v
			case "minLength":
				var v int
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "minLength": %w`, err)
				}
				s.minLength = &v
			case "minProperties":
				var v uint
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "minProperties": %w`, err)
				}
				s.minProperties = &v
			case "minimum":
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "minimum": %w`, err)
				}
				s.minimum = &v
			case "multipleOf":
				var v float64
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "multipleOf": %w`, err)
				}
				s.multipleOf = &v
			case "not":
				var v *Schema
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "not": %w`, err)
				}
				s.not = v
			case "oneOf":
				var v []*Schema
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "oneOf": %w`, err)
				}
				s.oneOf = v
			case "pattern":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "pattern": %w`, err)
				}
				s.pattern = &v
			case "patternProperties":
				var v map[string]*Schema
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "patternProperties": %w`, err)
				}
				s.patternProperties = v
			case "properties":
				var v map[string]*Schema
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "properties": %w`, err)
				}
				s.properties = v
			case "propertyNames":
				var v map[string]*Schema
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "propertyNames": %w`, err)
				}
				s.propertyNames = v
			case "$reference":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "$reference": %w`, err)
				}
				s.reference = &v
			case "required":
				var v bool
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "required": %w`, err)
				}
				s.required = &v
			case "$schema":
				var v string
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "$schema": %w`, err)
				}
				s.schema = v
			case "type":
				var v []PrimitiveType
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "type": %w`, err)
				}
				s.types = v
			case "unevaluatedItems":
				var v *Schema
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "unevaluatedItems": %w`, err)
				}
				s.unevaluatedItems = v
			case "unevaluatedProperties":
				var v *Schema
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "unevaluatedProperties": %w`, err)
				}
				s.unevaluatedProperties = v
			case "uniqueItems":
				var v bool
				if err := dec.Decode(&v); err != nil {
					return fmt.Errorf(`failed to decode value for field "uniqueItems": %w`, err)
				}
				s.uniqueItems = &v
			}
		}
	}
	return nil
}
