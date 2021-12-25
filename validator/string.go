package validator

import (
	"fmt"
	"reflect"
	"regexp"
)

type StringConstraint struct {
	maxLength *uint
	minLength *uint
	pattern   *regexp.Regexp
}

type StringConstraintBuilder struct {
	err error
	c   *StringConstraint
}

func String() *StringConstraintBuilder {
	return &StringConstraintBuilder{c: &StringConstraint{}}
}

func (b *StringConstraintBuilder) MaxLength(v int) *StringConstraintBuilder {
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

func (b *StringConstraintBuilder) MinLength(v int) *StringConstraintBuilder {
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

func (b *StringConstraintBuilder) Pattern(s string) *StringConstraintBuilder {
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

func (b *StringConstraintBuilder) Build() (*StringConstraint, error) {
	if b.err != nil {
		return nil, b.err
	}

	return b.c, nil
}

func (c *StringConstraint) Check(v interface{}) error {
	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.String:
	default:
		return fmt.Errorf(`invalid value passed to StringConstraint: expected string, got %T`, v)
	}

	str := rv.String()
	l := uint(len(str))

	if ml := c.minLength; ml != nil {
		if l < *ml {
			return fmt.Errorf(`invalid value passed to StringConstraint: string length (%d) shorter then minLength (%d)`, l, *ml)
		}
	}

	if ml := c.maxLength; ml != nil {
		if l > *ml {
			return fmt.Errorf(`invalid value passed to StringConstraint: string length (%d) longer then maxLength (%d)`, l, *ml)
		}
	}

	if pat := c.pattern; pat != nil {
		if !pat.MatchString(str) {
			return fmt.Errorf(`invalide value passed to StringConstraint: string did not match pattern %s`, pat.String())
		}
	}

	return nil
}
