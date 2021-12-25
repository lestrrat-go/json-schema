package validator

import (
	"fmt"
	"reflect"
)

type ObjectConstraint struct {
}

type objectConstraintContext struct {
	isStruct bool
}

func (c *ObjectConstraint) Check(v interface{}) error {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		rv = rv.Elem()
	}

	var ctx objectConstraintContext
	switch rv.Kind() {
	case reflect.Map:
	case reflect.Struct:
		ctx.isStruct = true
	default:
		return fmt.Errorf(`invalid value passed to ObjectConstraint: expected map or a struct, got %T`, v)
	}

	return nil
}
