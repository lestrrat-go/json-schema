//go:generate ./gen.sh

package schema

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
		panic("WHAT?")
	}
	return nil
}

type propPair struct {
	Name   string
	Schema *Schema
}
