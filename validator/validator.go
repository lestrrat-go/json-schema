package validator

import schema "github.com/lestrrat-go/json-schema"

type Validator struct{}

func Build(s *schema.Schema) (*Validator, error) {
	return &Validator{}, nil
}

type Constraint interface {
	Check(interface{}) error
}
