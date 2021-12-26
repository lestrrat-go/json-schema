package validator

import (
	"fmt"
	"math"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
)

func compileIntegerValidator(s *schema.Schema) (Validator, error) {
	b := Integer()

	if s.HasMultipleOf() {
		rv := reflect.ValueOf(s.MultipleOf())
		var tmp int
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = int(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = int(rv.Float())
		default:
			panic(`poop`)
		}
		b.MultipleOf(tmp)
	}

	if s.HasMaximum() {
		rv := reflect.ValueOf(s.Maximum())
		var tmp int
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = int(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = int(rv.Float())
		default:
			panic(`poop`)
		}
		b.Maximum(tmp)
	}

	if s.HasExclusiveMaximum() {
		rv := reflect.ValueOf(s.ExclusiveMaximum())
		var tmp int
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = int(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = int(rv.Float())
		default:
			panic(`poop`)
		}
		b.ExclusiveMaximum(tmp)
	}

	if s.HasMinimum() {
		rv := reflect.ValueOf(s.Minimum())
		var tmp int
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = int(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = int(rv.Float())
		default:
			panic(`poop`)
		}
		b.Minimum(tmp)
	}

	if s.HasExclusiveMinimum() {
		rv := reflect.ValueOf(s.ExclusiveMinimum())
		var tmp int
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = int(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = int(rv.Float())
		default:
			panic(`poop`)
		}
		b.ExclusiveMinimum(tmp)
	}

	if s.HasConst() {
		rv := reflect.ValueOf(s.Const())
		var tmp int
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = int(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = int(rv.Float())
		default:
			panic(`poop`)
		}
		b.Const(tmp)
	}
	return b.Build()
}

type IntegerValidator struct {
	multipleOf       *int
	maximum          *int
	exclusiveMaximum *int
	minimum          *int
	exclusiveMinimum *int
	constantValue    *int
}

type IntegerValidatorBuilder struct {
	err error
	c   *IntegerValidator
}

func Integer() *IntegerValidatorBuilder {
	return &IntegerValidatorBuilder{c: &IntegerValidator{}}
}

func (b *IntegerValidatorBuilder) MultipleOf(v int) *IntegerValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.multipleOf = &v
	return b
}

func (b *IntegerValidatorBuilder) Maximum(v int) *IntegerValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.maximum = &v
	return b
}

func (b *IntegerValidatorBuilder) ExclusiveMaximum(v int) *IntegerValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.exclusiveMaximum = &v
	return b
}

func (b *IntegerValidatorBuilder) Minimum(v int) *IntegerValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.minimum = &v
	return b
}

func (b *IntegerValidatorBuilder) ExclusiveMinimum(v int) *IntegerValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.exclusiveMinimum = &v
	return b
}

func (b *IntegerValidatorBuilder) Const(v int) *IntegerValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.constantValue = &v
	return b
}

func (b *IntegerValidatorBuilder) Build() (*IntegerValidator, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.c, nil
}

func (v *IntegerValidator) Validate(in interface{}) error {
	rv := reflect.ValueOf(in)

	var n int
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n = int(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n = int(rv.Uint())
	default:
		return fmt.Errorf(`invalid value passed to IntegerValidator: value is not an integer type (%T)`, in)
	}

	if m := v.maximum; m != nil {
		if *m <= n {
			return fmt.Errorf(`invalid value passed to IntegerValidator: value is greater than %d`, *m)
		}
	}

	if em := v.exclusiveMaximum; em != nil {
		if *em < n {
			return fmt.Errorf(`invalid value passed to IntegerValidator: value is greater than or equal to %d`, *em)
		}
	}

	if m := v.minimum; m != nil {
		if *m >= n {
			return fmt.Errorf(`invalid value passed to IntegerValidator: value is less than %d`, *m)
		}
	}

	if em := v.exclusiveMinimum; em != nil {
		if *em > n {
			return fmt.Errorf(`invalid value passed to IntegerValidator: value is less than or equal to %d`, *em)
		}
	}

	if mo := v.multipleOf; mo != nil {
		if math.Mod(float64(n), float64(*mo)) != 0 {
			return fmt.Errorf(`invalid value passed to IntegerValidator: value is not multiple of %d`, *mo)
		}
	}
	return nil
}
