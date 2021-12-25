package validator

import (
	"fmt"
	"reflect"
	"regexp"

	schema "github.com/lestrrat-go/json-schema"
)

func compileStringValidator(s *schema.Schema) (Validator, error) {
	v := String()
	if s.HasMaxLength() {
		v.MaxLength(s.MaxLength())
	}
	if s.HasMinLength() {
		v.MinLength(s.MinLength())
	}
	if s.HasPattern() {
		v.Pattern(s.Pattern())
	}

	return v.Build()
}

type StringValidator struct {
	maxLength *uint
	minLength *uint
	pattern   *regexp.Regexp
}

type StringValidatorBuilder struct {
	err error
	c   *StringValidator
}

func String() *StringValidatorBuilder {
	return &StringValidatorBuilder{c: &StringValidator{}}
}

func (b *StringValidatorBuilder) MaxLength(v int) *StringValidatorBuilder {
	if b.err != nil {
		return b
	}

	if v < 0 {
		b.err = fmt.Errorf(`invalid value passed to MaxLength: value (%d) may not be less than zero`, v)
		return b
	}

	var uv uint = uint(v)
	b.c.maxLength = &uv
	return b
}

func (b *StringValidatorBuilder) MinLength(v int) *StringValidatorBuilder {
	if b.err != nil {
		return b
	}

	if v < 0 {
		b.err = fmt.Errorf(`invalid value passed to MinLength: value (%d) may not be less than zero`, v)
		return b
	}

	var uv uint = uint(v)
	b.c.minLength = &uv
	return b
}

func (b *StringValidatorBuilder) Pattern(s string) *StringValidatorBuilder {
	if b.err != nil {
		return b
	}

	// https://json-schema.org/draft/2020-12/json-schema-validation.html#rfc.section.6.3.3
	// says "ECMA-262 regular expression dialect, but there's little we can do here :/
	re, err := regexp.Compile(s)
	if err != nil {
		b.err = err
		return b
	}

	b.c.pattern = re
	return b
}

func (b *StringValidatorBuilder) Build() (*StringValidator, error) {
	if b.err != nil {
		return nil, b.err
	}

	return b.c, nil
}

func (c *StringValidator) Validate(v interface{}) error {
	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.String:
	default:
		return fmt.Errorf(`invalid value passed to StringValidator: expected string, got %T`, v)
	}

	str := rv.String()
	l := uint(len(str))

	if ml := c.minLength; ml != nil {
		if l < *ml {
			return fmt.Errorf(`invalid value passed to StringValidator: string length (%d) shorter then minLength (%d)`, l, *ml)
		}
	}

	if ml := c.maxLength; ml != nil {
		if l > *ml {
			return fmt.Errorf(`invalid value passed to StringValidator: string length (%d) longer then maxLength (%d)`, l, *ml)
		}
	}

	if pat := c.pattern; pat != nil {
		if !pat.MatchString(str) {
			return fmt.Errorf(`invalide value passed to StringValidator: string did not match pattern %s`, pat.String())
		}
	}

	return nil
}
