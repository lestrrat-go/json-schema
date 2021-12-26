package validator

import (
	"fmt"
	"math"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
)

func compileNumberValidator(s *schema.Schema) (Validator, error) {
	b := Number()

	if s.HasMultipleOf() {
		rv := reflect.ValueOf(s.MultipleOf())
		var tmp float64
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = float64(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = rv.Float()
		default:
			panic(`poop`)
		}
		b.MultipleOf(tmp)
	}

	if s.HasMaximum() {
		rv := reflect.ValueOf(s.Maximum())
		var tmp float64
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = float64(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = rv.Float()
		default:
			panic(`poop`)
		}
		b.Maximum(tmp)
	}

	if s.HasExclusiveMaximum() {
		rv := reflect.ValueOf(s.ExclusiveMaximum())
		var tmp float64
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = float64(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = rv.Float()
		default:
			panic(`poop`)
		}
		b.ExclusiveMaximum(tmp)
	}

	if s.HasMinimum() {
		rv := reflect.ValueOf(s.Minimum())
		var tmp float64
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = float64(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = rv.Float()
		default:
			panic(`poop`)
		}
		b.Minimum(tmp)
	}

	if s.HasExclusiveMinimum() {
		rv := reflect.ValueOf(s.ExclusiveMinimum())
		var tmp float64
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = float64(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = rv.Float()
		default:
			panic(`poop`)
		}
		b.ExclusiveMinimum(tmp)
	}

	if s.HasConst() {
		rv := reflect.ValueOf(s.Const())
		var tmp float64
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = float64(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = rv.Float()
		default:
			panic(`poop`)
		}
		b.Const(tmp)
	}
	return b.Build()
}

type NumberValidator struct {
	multipleOf       *float64
	maximum          *float64
	exclusiveMaximum *float64
	minimum          *float64
	exclusiveMinimum *float64
	constantValue    *float64
}

type NumberValidatorBuilder struct {
	err error
	c   *NumberValidator
}

func Number() *NumberValidatorBuilder {
	return &NumberValidatorBuilder{c: &NumberValidator{}}
}

func (b *NumberValidatorBuilder) MultipleOf(v float64) *NumberValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.multipleOf = &v
	return b
}

func (b *NumberValidatorBuilder) Maximum(v float64) *NumberValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.maximum = &v
	return b
}

func (b *NumberValidatorBuilder) ExclusiveMaximum(v float64) *NumberValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.exclusiveMaximum = &v
	return b
}

func (b *NumberValidatorBuilder) Minimum(v float64) *NumberValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.minimum = &v
	return b
}

func (b *NumberValidatorBuilder) ExclusiveMinimum(v float64) *NumberValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.exclusiveMinimum = &v
	return b
}

func (b *NumberValidatorBuilder) Const(v float64) *NumberValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.constantValue = &v
	return b
}

func (b *NumberValidatorBuilder) Build() (*NumberValidator, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.c, nil
}

func (v *NumberValidator) Validate(in interface{}) error {
	rv := reflect.ValueOf(in)
	n := rv.Float()

	if m := v.maximum; m != nil {
		if *m <= n {
			return fmt.Errorf(`invalid value passed to NumberValidator: value is greater than %f`, *m)
		}
	}

	if em := v.exclusiveMaximum; em != nil {
		if *em < n {
			return fmt.Errorf(`invalid value passed to NumberValidator: value is greater than or equal to %f`, *em)
		}
	}

	if m := v.minimum; m != nil {
		if *m >= n {
			return fmt.Errorf(`invalid value passed to NumberValidator: value is less than %f`, *m)
		}
	}

	if em := v.exclusiveMinimum; em != nil {
		if *em > n {
			return fmt.Errorf(`invalid value passed to NumberValidator: value is less than or equal to %f`, *em)
		}
	}

	if mo := v.multipleOf; mo != nil {
		if math.Mod(n, *mo) != 0 {
			return fmt.Errorf(`invalid value passed to NumberValidator: value is not multiple of %f`, *mo)
		}
	}
	return nil
}
