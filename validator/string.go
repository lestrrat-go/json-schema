package validator

import (
	"context"
	"fmt"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"slices"
	"time"
	"unicode/utf8"

	schema "github.com/lestrrat-go/json-schema"
)

var _ Builder = (*StringValidatorBuilder)(nil)
var _ Interface = (*stringValidator)(nil)

// String creates a new StringValidatorBuilder instance that can be used to build a
// Validator for string values according to the JSON Schema specification.
func String() *StringValidatorBuilder {
	return (&StringValidatorBuilder{}).Reset()
}

type stringValidator struct {
	maxLength      *uint
	minLength      *uint
	pattern        *regexp.Regexp
	format         *string
	enum           []string
	constantValue  *string
	strictStringType bool // true when schema explicitly declares type: string
}

func (v *stringValidator) Validate(ctx context.Context, in any) (Result, error) {
	rv := reflect.ValueOf(in)

	switch rv.Kind() {
	case reflect.String:
		// Continue with string validation
	default:
		// Handle non-string values based on whether this is strict string type validation
		if v.strictStringType {
			// When schema explicitly declares type: string, non-string values should fail
			return nil, fmt.Errorf(`invalid value passed to StringValidator: expected string, got %T`, in)
		}
		// For non-string values with inferred string type, string constraints don't apply
		// According to JSON Schema spec, string constraints should be ignored for non-strings
		return nil, nil
	}

	str := rv.String()
	// Count Unicode rune length instead of byte length to better handle Unicode text
	// This is closer to the JSON Schema spec's requirement for grapheme clusters
	l := uint(utf8.RuneCountInString(str))

	if v := v.constantValue; v != nil {
		if *v != str {
			return nil, fmt.Errorf(`invalid value passed to StringValidator: string must be const value %q`, *v)
		}
	}

	if ml := v.minLength; ml != nil {
		if l < *ml {
			return nil, fmt.Errorf(`invalid value passed to StringValidator: string length (%d) shorter then minLength (%d)`, l, *ml)
		}
	}

	if ml := v.maxLength; ml != nil {
		if l > *ml {
			return nil, fmt.Errorf(`invalid value passed to StringValidator: string length (%d) longer then maxLength (%d)`, l, *ml)
		}
	}

	if pat := v.pattern; pat != nil {
		if !pat.MatchString(str) {
			return nil, fmt.Errorf(`invalid value passed to StringValidator: string did not match pattern %s`, pat.String())
		}
	}

	if enums := v.enum; len(enums) > 0 {
		if !slices.Contains(enums, str) {
			return nil, fmt.Errorf(`invalid value passed to StringValidator: string not found in enum`)
		}
	}

	if format := v.format; format != nil {
		if err := validateFormat(str, *format); err != nil {
			return nil, fmt.Errorf(`invalid value passed to StringValidator: %w`, err)
		}
	}

	return nil, nil
}

// validateFormat validates a string against the specified format
func validateFormat(value, format string) error {
	switch format {
	case "email":
		_, err := mail.ParseAddress(value)
		if err != nil {
			return fmt.Errorf("invalid email format")
		}
	case "date":
		_, err := time.Parse("2006-01-02", value)
		if err != nil {
			return fmt.Errorf("invalid date format")
		}
	case "date-time":
		// Try RFC3339 format first (with timezone)
		_, err := time.Parse(time.RFC3339, value)
		if err != nil {
			// Try ISO 8601 format without timezone
			_, err = time.Parse("2006-01-02T15:04:05", value)
			if err != nil {
				return fmt.Errorf("invalid date-time format")
			}
		}
	case "uri":
		_, err := url.ParseRequestURI(value)
		if err != nil {
			return fmt.Errorf("invalid URI format")
		}
	case "uuid":
		// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
		uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
		if !uuidRegex.MatchString(value) {
			return fmt.Errorf("invalid UUID format")
		}
	default:
		// Unknown format - just allow it (format validation is optional in JSON Schema)
		return nil
	}
	return nil
}

func compileStringValidator(ctx context.Context, s *schema.Schema, strictType bool) (Interface, error) {
	v := String()
	v.StrictStringType(strictType)
	if s.HasConst() {
		c, ok := s.Const().(string)
		if !ok {
			return nil, fmt.Errorf(`invalid element in const: expected string element, got %T`, s.Const())
		}
		v.Const(c)
	}
	if s.HasMaxLength() {
		v.MaxLength(s.MaxLength())
	}
	if s.HasMinLength() {
		v.MinLength(s.MinLength())
	}
	if s.HasPattern() {
		v.Pattern(s.Pattern())
	}
	if s.HasFormat() {
		v.Format(s.Format())
	}
	if s.HasEnum() {
		enums := s.Enum()
		l := make([]string, 0, len(enums))
		for i, e := range s.Enum() {
			s, ok := e.(string)
			if !ok {
				return nil, fmt.Errorf(`invalid element in enum: expected string element, got %T for element %d`, e, i)
			}
			l = append(l, s)
		}

		v.Enum(l)
	}

	return v.Build()
}

type StringValidatorBuilder struct {
	err error
	c   *stringValidator
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

func (b *StringValidatorBuilder) Format(format string) *StringValidatorBuilder {
	if b.err != nil {
		return b
	}

	b.c.format = &format
	return b
}

func (b *StringValidatorBuilder) Enum(enums []string) *StringValidatorBuilder {
	if b.err != nil {
		return b
	}

	b.c.enum = make([]string, len(enums))
	copy(b.c.enum, enums)
	return b
}

func (b *StringValidatorBuilder) Const(c string) *StringValidatorBuilder {
	if b.err != nil {
		return b
	}

	b.c.constantValue = &c
	return b
}

func (b *StringValidatorBuilder) StrictStringType(v bool) *StringValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.strictStringType = v
	return b
}

func (b *StringValidatorBuilder) Build() (Interface, error) {
	if b.err != nil {
		return nil, b.err
	}

	return b.c, nil
}

func (b *StringValidatorBuilder) MustBuild() Interface {
	if b.err != nil {
		panic(b.err)
	}
	return b.c
}

func (b *StringValidatorBuilder) Reset() *StringValidatorBuilder {
	b.err = nil
	b.c = &stringValidator{
		strictStringType: true, // Default to strict for direct usage
	}
	return b
}
