//go:generate ./gen.sh

package schema

import "fmt"

// SchemaOrBool is a visual indicator for those cases where
// a Schema or boolean can be passed, for example, AdditionalProperties
type SchemaOrBool interface{}

// The schema that this implementation supports. We use the name
// `Version` here because `Schema` is confusin with other types
const Version = `https://json-schema.org/draft/2020-12/schema`

func (s *Schema) Accept(v interface{}) error {
	switch v := v.(type) {
	case bool:
		if v {
			*s = Schema{}
		} else {
			*s = Schema{not: &Schema{}}
		}
	case *Schema:
		*s = *v
	default:
		return fmt.Errorf(`invalid value for additionalProperties. Got %T`, v)
	}
	return nil
}

type propPair struct {
	Name   string
	Schema *Schema
}
