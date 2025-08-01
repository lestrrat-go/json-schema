package validator

import (
	"context"
	"fmt"
	"math"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/vocabulary"
)

var _ Builder = (*NumberValidatorBuilder)(nil)
var _ Interface = (*numberValidator)(nil)

func compileNumberValidator(ctx context.Context, s *schema.Schema) (Interface, error) {
	b := Number()

	if s.HasMultipleOf() && vocabulary.IsKeywordEnabledInContext(ctx, "multipleOf") {
		rv := reflect.ValueOf(s.MultipleOf())
		var tmp float64
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = float64(rv.Int())
			b.MultipleOf(tmp)
		case reflect.Float32, reflect.Float64:
			tmp = rv.Float()
		default:
			return nil, fmt.Errorf(`invalid type for multipleOf field: expected numeric type, got %T`, rv.Interface())
		}
		b.MultipleOf(tmp)
	}

	if s.HasMaximum() && vocabulary.IsKeywordEnabledInContext(ctx, "maximum") {
		rv := reflect.ValueOf(s.Maximum())
		var tmp float64
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = float64(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = rv.Float()
		default:
			return nil, fmt.Errorf(`invalid type for maximum field: expected numeric type, got %T`, rv.Interface())
		}
		b.Maximum(tmp)
	}

	if s.HasExclusiveMaximum() && vocabulary.IsKeywordEnabledInContext(ctx, "exclusiveMaximum") {
		rv := reflect.ValueOf(s.ExclusiveMaximum())
		var tmp float64
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = float64(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = rv.Float()
		default:
			return nil, fmt.Errorf(`invalid type for exclusiveMaximum field: expected numeric type, got %T`, rv.Interface())
		}
		b.ExclusiveMaximum(tmp)
	}

	if s.HasMinimum() && vocabulary.IsKeywordEnabledInContext(ctx, "minimum") {
		rv := reflect.ValueOf(s.Minimum())
		var tmp float64
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = float64(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = rv.Float()
		default:
			return nil, fmt.Errorf(`invalid type for minimum field: expected numeric type, got %T`, rv.Interface())
		}
		b.Minimum(tmp)
	}

	if s.HasExclusiveMinimum() && vocabulary.IsKeywordEnabledInContext(ctx, "exclusiveMinimum") {
		rv := reflect.ValueOf(s.ExclusiveMinimum())
		var tmp float64
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = float64(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = rv.Float()
		default:
			return nil, fmt.Errorf(`invalid type for exclusiveMinimum field: expected numeric type, got %T`, rv.Interface())
		}
		b.ExclusiveMinimum(tmp)
	}

	if s.HasConst() && vocabulary.IsKeywordEnabledInContext(ctx, "const") {
		rv := reflect.ValueOf(s.Const())
		var tmp float64
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tmp = float64(rv.Int())
		case reflect.Float32, reflect.Float64:
			tmp = rv.Float()
		default:
			return nil, fmt.Errorf(`invalid type for constantValue field: expected numeric type, got %T`, rv.Interface())
		}
		b.Const(tmp)
	}

	if s.HasEnum() && vocabulary.IsKeywordEnabledInContext(ctx, "enum") {
		enums := s.Enum()
		l := make([]float64, 0, len(enums))
		for i, e := range s.Enum() {
			rv := reflect.ValueOf(e)
			var tmp float64
			switch rv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				tmp = float64(rv.Int())
			case reflect.Float32, reflect.Float64:
				tmp = rv.Float()
			default:
				return nil, fmt.Errorf(`invalid element in enum: expected numeric element, got %T for element %d`, e, i)
			}
			l = append(l, tmp)
		}
		b.Enum(l...)
	}
	return b.Build()
}

type numberValidator struct {
	multipleOf       *float64
	maximum          *float64
	exclusiveMaximum *float64
	minimum          *float64
	exclusiveMinimum *float64
	constantValue    *float64
	enum             []float64
}

type NumberValidatorBuilder struct {
	err error
	c   *numberValidator
}

func Number() *NumberValidatorBuilder {
	return (&NumberValidatorBuilder{}).Reset()
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

func (b *NumberValidatorBuilder) Enum(v ...float64) *NumberValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.enum = make([]float64, len(v))
	copy(b.c.enum, v)
	return b
}

func (b *NumberValidatorBuilder) Build() (Interface, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.c, nil
}

func (b *NumberValidatorBuilder) MustBuild() Interface {
	if b.err != nil {
		panic(b.err)
	}
	return b.c
}

func (b *NumberValidatorBuilder) Reset() *NumberValidatorBuilder {
	b.err = nil
	b.c = &numberValidator{}
	return b
}

func (v *numberValidator) Validate(_ context.Context, in any) (Result, error) {
	rv := reflect.ValueOf(in)

	var n float64
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n = float64(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n = float64(rv.Uint())
	case reflect.Float32, reflect.Float64:
		n = rv.Float()
	default:
		return nil, fmt.Errorf(`invalid value passed to NumberValidator: expected number, got %T`, in)
	}

	// Reject NaN but allow infinity
	if math.IsNaN(n) {
		return nil, fmt.Errorf(`invalid value passed to NumberValidator: value is not a valid number (NaN)`)
	}

	if m := v.maximum; m != nil {
		if n > *m {
			return nil, fmt.Errorf(`invalid value passed to NumberValidator: value is greater than maximum %f`, *m)
		}
	}

	if em := v.exclusiveMaximum; em != nil {
		if n >= *em {
			return nil, fmt.Errorf(`invalid value passed to NumberValidator: value is greater than or equal to exclusiveMaximum %f`, *em)
		}
	}

	if m := v.minimum; m != nil {
		if n < *m {
			return nil, fmt.Errorf(`invalid value passed to NumberValidator: value is less than minimum %f`, *m)
		}
	}

	if em := v.exclusiveMinimum; em != nil {
		if n <= *em {
			return nil, fmt.Errorf(`invalid value passed to NumberValidator: value is less than or equal to exclusiveMinimum %f`, *em)
		}
	}

	if mo := v.multipleOf; mo != nil {
		remainder := math.Mod(n, *mo)
		if math.Abs(remainder) > 1e-9 && math.Abs(remainder-*mo) > 1e-9 {
			return nil, fmt.Errorf(`invalid value passed to NumberValidator: value is not multiple of %f`, *mo)
		}
	}

	if c := v.constantValue; c != nil {
		if *c != n {
			return nil, fmt.Errorf(`invalid value passed to NumberValidator: value must be const value %f`, *c)
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
			return nil, fmt.Errorf(`invalid value passed to NumberValidator: value not found in enum`)
		}
	}
	return nil, nil
}
