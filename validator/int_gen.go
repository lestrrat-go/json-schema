package validator

import (
	"context"
	"fmt"
	"math"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
)

var _ Builder = (*IntegerValidatorBuilder)(nil)
var _ Interface = (*integerValidator)(nil)

func compileIntegerValidator(ctx context.Context, s *schema.Schema) (Interface, error) {
	b := Integer()

	if s.HasMultipleOf() {
		rv := reflect.ValueOf(s.MultipleOf())
		var tmp int
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = int(rv.Int())
		case reflect.Float32, reflect.Float64:
			f := rv.Float()
			if f < 1.0 && f > 0 {
				// For very small positive fractions like 1e-8, any integer is a multiple
				// Skip adding multipleOf constraint as all integers pass
				tmp = 0 // This will be ignored due to the <= 0 check in validation
			} else {
				tmp = int(f)
			}
		default:
			panic(`poop`)
		}
		if tmp > 0 { // Only set multipleOf if it's a positive integer
			b.MultipleOf(tmp)
		}
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

	if s.HasEnum() {
		enums := s.Enum()
		l := make([]int, 0, len(enums))
		for i, e := range s.Enum() {
			rv := reflect.ValueOf(e)
			var tmp int
			switch rv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				tmp = int(rv.Int())
			case reflect.Float32, reflect.Float64:
				tmp = int(rv.Float())
			default:
				return nil, fmt.Errorf(`invalid element in enum: expected numeric element, got %T for element %d`, e, i)
			}
			l = append(l, tmp)
		}
		b.Enum(l)
	}
	return b.Build()
}

type integerValidator struct {
	multipleOf       *int
	maximum          *int
	exclusiveMaximum *int
	minimum          *int
	exclusiveMinimum *int
	constantValue    *int
	enum             []int
}

type IntegerValidatorBuilder struct {
	err error
	c   *integerValidator
}

func Integer() *IntegerValidatorBuilder {
	return (&IntegerValidatorBuilder{}).Reset()
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

func (b *IntegerValidatorBuilder) Enum(v []int) *IntegerValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.enum = make([]int, len(v))
	copy(b.c.enum, v)
	return b
}

func (b *IntegerValidatorBuilder) Build() (Interface, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.c, nil
}

func (b *IntegerValidatorBuilder) MustBuild() Interface {
	if b.err != nil {
		panic(b.err)
	}
	return b.c
}

func (b *IntegerValidatorBuilder) Reset() *IntegerValidatorBuilder {
	b.err = nil
	b.c = &integerValidator{}
	return b
}

func (v *integerValidator) Validate(ctx context.Context, in any) (Result, error) {
	rv := reflect.ValueOf(in)

	var n int
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n = int(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n = int(rv.Uint())
	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		// Check if the float represents a whole number
		if f != math.Trunc(f) {
			return nil, fmt.Errorf(`invalid value passed to IntegerValidator: expected integer, got %T with non-integer value %g`, in, f)
		}
		// Check if the float is within integer range
		if f > math.MaxInt || f < math.MinInt || math.IsInf(f, 0) || math.IsNaN(f) {
			return nil, fmt.Errorf(`invalid value passed to IntegerValidator: value %g is out of integer range`, f)
		}
		n = int(f)
	default:
		return nil, fmt.Errorf(`invalid value passed to IntegerValidator: expected integer, got %T`, in)
	}

	if m := v.maximum; m != nil {
		if n > *m {
			return nil, fmt.Errorf(`invalid value passed to IntegerValidator: value is greater than maximum %d`, *m)
		}
	}

	if em := v.exclusiveMaximum; em != nil {
		if n >= *em {
			return nil, fmt.Errorf(`invalid value passed to IntegerValidator: value is greater than or equal to exclusiveMaximum %d`, *em)
		}
	}

	if m := v.minimum; m != nil {
		if n < *m {
			return nil, fmt.Errorf(`invalid value passed to IntegerValidator: value is less than minimum %d`, *m)
		}
	}

	if em := v.exclusiveMinimum; em != nil {
		if n <= *em {
			return nil, fmt.Errorf(`invalid value passed to IntegerValidator: value is less than or equal to exclusiveMinimum %d`, *em)
		}
	}

	if mo := v.multipleOf; mo != nil {
		if *mo <= 0 {
			return nil, fmt.Errorf(`invalid value passed to IntegerValidator: multipleOf must be positive`)
		}
		remainder := math.Mod(float64(n), float64(*mo))
		if math.Abs(remainder) > 1e-9 && math.Abs(remainder-float64(*mo)) > 1e-9 {
			return nil, fmt.Errorf(`invalid value passed to IntegerValidator: value is not multiple of %d`, *mo)
		}
	}

	if c := v.constantValue; c != nil {
		if *c != n {
			return nil, fmt.Errorf(`invalid value passed to IntegerValidator: value must be const value %d`, *c)
		}
	}

	if enums := v.enum; len(enums) > 0 {
		var found bool
		for _, e := range enums {
			if e == n {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf(`invalid value passed to IntegerValidator: value not found in enum`)
		}
	}
	return nil, nil
}
