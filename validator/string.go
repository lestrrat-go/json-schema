package validator

import (
	"context"
	"fmt"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"time"
	"unicode/utf8"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/keywords"
	"github.com/lestrrat-go/json-schema/vocabulary"
)

// Format constants for string validation
const (
	FormatEmail    = "email"
	FormatDate     = "date"
	FormatDateTime = "date-time"
	FormatURI      = "uri"
	FormatUUID     = "uuid"
)

var _ Builder = (*StringValidatorBuilder)(nil)
var _ Interface = (*stringValidator)(nil)

// String creates a new StringValidatorBuilder instance that can be used to build a
// Validator for string values according to the JSON Schema specification.
func String() *StringValidatorBuilder {
	return (&StringValidatorBuilder{}).Reset()
}

type stringValidator struct {
	maxLength        *uint
	minLength        *uint
	pattern          *regexp.Regexp
	format           *string
	enum             []any
	constantValue    any
	strictStringType bool // true when schema explicitly declares type: string
}

func (v *stringValidator) Validate(ctx context.Context, in any) (Result, error) {
	logger := TraceSlogFromContext(ctx)
	logger.InfoContext(ctx, "string validator starting", "value", in, "type", fmt.Sprintf("%T", in))
	rv := reflect.ValueOf(in)

	switch rv.Kind() {
	case reflect.String:
		logger.InfoContext(ctx, "string validator processing string value")
		// Continue with string validation
	default:
		// Handle non-string values based on whether this is strict string type validation
		if v.strictStringType {
			// When schema explicitly declares type: string, non-string values should fail
			logger.InfoContext(ctx, "string validator rejecting non-string for strict type", "strict", true)
			return nil, fmt.Errorf(`invalid value passed to StringValidator: expected string, got %T`, in)
		}
		// For non-string values with inferred string type, string constraints don't apply
		// According to JSON Schema spec, string constraints should be ignored for non-strings
		logger.InfoContext(ctx, "string validator skipping non-string for inferred type", "strict", false)
		//nolint: nilnil
		return nil, nil
	}

	str := rv.String()
	// Count Unicode rune length instead of byte length to better handle Unicode text
	// This is closer to the JSON Schema spec's requirement for grapheme clusters
	l := uint(utf8.RuneCountInString(str))
	logger.InfoContext(ctx, "string validator checking constraints", "length", l, "value_preview", truncateString(str, 50))

	if v.constantValue != nil {
		if err := validateConst(ctx, str, v.constantValue); err != nil {
			return nil, fmt.Errorf(`invalid value passed to StringValidator: %w`, err)
		}
	}

	if ml := v.minLength; ml != nil {
		logger.InfoContext(ctx, "string validator checking minLength", "minLength", *ml, "actual", l)
		if l < *ml {
			return nil, fmt.Errorf(`invalid value passed to StringValidator: string length (%d) shorter then minLength (%d)`, l, *ml)
		}
	}

	if ml := v.maxLength; ml != nil {
		logger.InfoContext(ctx, "string validator checking maxLength", "maxLength", *ml, "actual", l)
		if l > *ml {
			return nil, fmt.Errorf(`invalid value passed to StringValidator: string length (%d) longer then maxLength (%d)`, l, *ml)
		}
	}

	if pat := v.pattern; pat != nil {
		logger.InfoContext(ctx, "string validator checking pattern", "pattern", pat.String())
		if !pat.MatchString(str) {
			return nil, fmt.Errorf(`invalid value passed to StringValidator: string did not match pattern %s`, pat.String())
		}
	}

	if len(v.enum) > 0 {
		if err := validateEnum(ctx, str, v.enum); err != nil {
			return nil, fmt.Errorf(`invalid value passed to StringValidator: %w`, err)
		}
	}

	if format := v.format; format != nil {
		logger.InfoContext(ctx, "string validator checking format", "format", *format, "value", str)
		if err := validateFormat(str, *format); err != nil {
			return nil, fmt.Errorf(`invalid value passed to StringValidator: %w`, err)
		}
	}

	//nolint: nilnil
	return nil, nil
}

// validateFormat validates a string against the specified format
func validateFormat(value, format string) error {
	switch format {
	case FormatEmail:
		_, err := mail.ParseAddress(value)
		if err != nil {
			return fmt.Errorf("invalid email format")
		}
	case FormatDate:
		_, err := time.Parse("2006-01-02", value)
		if err != nil {
			return fmt.Errorf("invalid date format")
		}
	case FormatDateTime:
		// Try RFC3339 format first (with timezone)
		_, err := time.Parse(time.RFC3339, value)
		if err != nil {
			// Try ISO 8601 format without timezone
			_, err = time.Parse("2006-01-02T15:04:05", value)
			if err != nil {
				return fmt.Errorf("invalid date-time format")
			}
		}
	case FormatURI:
		_, err := url.ParseRequestURI(value)
		if err != nil {
			return fmt.Errorf("invalid URI format")
		}
	case FormatUUID:
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

// truncateString truncates a string to maxLength runes for logging purposes
func truncateString(s string, maxLength int) string {
	if utf8.RuneCountInString(s) <= maxLength {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLength]) + "..."
}

func compileStringValidator(ctx context.Context, s *schema.Schema, strictType bool) (Interface, error) {
	v := String()
	v.StrictStringType(strictType)
	if s.HasConst() && vocabulary.IsKeywordEnabledInContext(ctx, keywords.Const) {
		v.Const(s.Const())
	}
	if s.HasMaxLength() && vocabulary.IsKeywordEnabledInContext(ctx, keywords.MaxLength) {
		v.MaxLength(s.MaxLength())
	}
	if s.HasMinLength() && vocabulary.IsKeywordEnabledInContext(ctx, keywords.MinLength) {
		v.MinLength(s.MinLength())
	}
	if s.HasPattern() && vocabulary.IsKeywordEnabledInContext(ctx, keywords.Pattern) {
		v.Pattern(s.Pattern())
	}
	// Format validation should only be enforced when format-assertion vocabulary is enabled
	// When only format-annotation is enabled, format should be treated as annotation-only
	if s.HasFormat() {
		vocabSet := vocabulary.SetFromContext(ctx)
		if vocabSet.IsEnabled("https://json-schema.org/draft/2020-12/vocab/format-assertion") {
			v.Format(s.Format())
		}
		// If only format-annotation is enabled, we skip format validation (annotation-only behavior)
	}
	if s.HasEnum() && vocabulary.IsKeywordEnabledInContext(ctx, keywords.Enum) {
		v.Enum(s.Enum()...)
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

	uv := uint(v)
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

	uv := uint(v)
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

func (b *StringValidatorBuilder) Enum(enums ...any) *StringValidatorBuilder {
	if b.err != nil {
		return b
	}

	b.c.enum = make([]any, len(enums))
	copy(b.c.enum, enums)
	return b
}

func (b *StringValidatorBuilder) Const(c any) *StringValidatorBuilder {
	if b.err != nil {
		return b
	}

	b.c.constantValue = c
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
