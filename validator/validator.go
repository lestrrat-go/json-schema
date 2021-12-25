package validator

import (
	"fmt"

	schema "github.com/lestrrat-go/json-schema"
)

func Compile(s *schema.Schema) (Validator, error) {
	for _, typ := range s.Types() {
		// This is a placeholder code. In reality we need to
		// OR all types
		switch typ {
		case schema.StringType:
			return compileStringValidator(s)
		}
	}
	return nil, fmt.Errorf(`unimplemented`)
}

type Validator interface {
	Validate(interface{}) error
}
