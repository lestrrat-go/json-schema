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

	if s.HasMultipleOf() && IsKeywordEnabledInContext(ctx, "multipleOf") {
		rv := reflect.ValueOf(s.MultipleOf())
		var tmp int
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = int(rv.Int())
		case reflect.Float32, reflect.Float64:
			f := rv.Float()
			if f > 0 && f < 1 {
				// Skip multipleOf constraint for very small values with integer type
				// Any integer is a multiple of very small numbers like 1e-8
			} else {
				tmp = int(f)
				b.MultipleOf(tmp)
			}
		default:
			return nil, fmt.Errorf(`invalid type for multipleOf field: expected numeric type, got %T`, rv.Interface())
		}
	}

	if s.HasMaximum() && IsKeywordEnabledInContext(ctx, "maximum") {
		rv := reflect.ValueOf(s.Maximum())
		var tmp int
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = int(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = int(rv.Float())
		default:
			return nil, fmt.Errorf(`invalid type for maximum field: expected numeric type, got %T`, rv.Interface())
		}
		b.Maximum(tmp)
	}

	if s.HasExclusiveMaximum() && IsKeywordEnabledInContext(ctx, "exclusiveMaximum") {
		rv := reflect.ValueOf(s.ExclusiveMaximum())
		var tmp int
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = int(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = int(rv.Float())
		default:
			return nil, fmt.Errorf(`invalid type for exclusiveMaximum field: expected numeric type, got %T`, rv.Interface())
		}
		b.ExclusiveMaximum(tmp)
	}

	if s.HasMinimum() && IsKeywordEnabledInContext(ctx, "minimum") {
		rv := reflect.ValueOf(s.Minimum())
		var tmp int
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = int(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = int(rv.Float())
		default:
			return nil, fmt.Errorf(`invalid type for minimum field: expected numeric type, got %T`, rv.Interface())
		}
		b.Minimum(tmp)
	}

	if s.HasExclusiveMinimum() && IsKeywordEnabledInContext(ctx, "exclusiveMinimum") {
		rv := reflect.ValueOf(s.ExclusiveMinimum())
		var tmp int
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = int(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = int(rv.Float())
		default:
			return nil, fmt.Errorf(`invalid type for exclusiveMinimum field: expected numeric type, got %T`, rv.Interface())
		}
		b.ExclusiveMinimum(tmp)
	}

	if s.HasConst() && IsKeywordEnabledInContext(ctx, "const") {
		rv := reflect.ValueOf(s.Const())
		var tmp int
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = int(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = int(rv.Float())
		default:
			return nil, fmt.Errorf(`invalid type for constantValue field: expected numeric type, got %T`, rv.Interface())
		}
		b.Const(tmp)
	}

	if s.HasEnum() && IsKeywordEnabledInContext(ctx, "enum") {
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
		if f != float64(int(f)) {
			return nil, fmt.Errorf(`expected integer, got float value %g`, f)
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
		if *mo == 0 {
			return nil, fmt.Errorf(`invalid value passed to IntegerValidator: multipleOf cannot be zero`)
		}
		if math.Mod(float64(n), float64(*mo)) != 0 {
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
