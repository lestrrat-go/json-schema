package validator

import (
	"fmt"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
)

func compileArrayValidator(s *schema.Schema) (Validator, error) {
	return &ArrayValidator{}, nil
}

type ArrayValidator struct {
}

func (c *ArrayValidator) Validate(v interface{}) error {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		return nil
	default:
		return fmt.Errorf(`invalid value passed to ArrayValidator: expected array or slice, got %T`, v)
	}
}