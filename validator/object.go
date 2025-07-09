package validator

import (
	"fmt"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
)

func compileObjectValidator(s *schema.Schema) (Validator, error) {
	return &ObjectValidator{}, nil
}

type ObjectValidator struct {
}

type objectValidatorContext struct {
	isStruct bool
}

func (c *ObjectValidator) Validate(v interface{}) error {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		rv = rv.Elem()
	}

	var ctx objectValidatorContext
	switch rv.Kind() {
	case reflect.Map:
	case reflect.Struct:
		ctx.isStruct = true
	default:
		return fmt.Errorf(`invalid value passed to ObjectValidator: expected map or a struct, got %T`, v)
	}

	return nil
}
