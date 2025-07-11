package validator

import (
	"fmt"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
)

var _ Builder = (*ArrayValidatorBuilder)(nil)
var _ Interface = (*arrayValidator)(nil)

func compileArrayValidator(s *schema.Schema, strictType bool) (Interface, error) {
	v := Array()

	if s.HasMinItems() {
		v.MinItems(s.MinItems())
	}
	if s.HasMaxItems() {
		v.MaxItems(s.MaxItems())
	}
	if s.HasUniqueItems() {
		v.UniqueItems(s.UniqueItems())
	}
	if s.HasItems() {
		itemsSchema := s.Items()
		if itemsSchema != nil {
			itemValidator, err := Compile(itemsSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to compile items validator: %w", err)
			}
			v.Items(itemValidator)
		}
	}
	if s.HasContains() {
		containsSchema := s.Contains()
		if containsSchema != nil {
			containsValidator, err := Compile(containsSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to compile contains validator: %w", err)
			}
			v.Contains(containsValidator)
		}
	}
	if s.HasMinContains() {
		v.MinContains(s.MinContains())
	}
	if s.HasMaxContains() {
		v.MaxContains(s.MaxContains())
	}

	v.StrictArrayType(strictType)

	return v.Build()
}

type arrayValidator struct {
	minItems         *uint
	maxItems         *uint
	uniqueItems      bool
	items            Interface
	contains         Interface
	minContains      *uint
	maxContains      *uint
	strictArrayType  bool // true when schema explicitly declares type: array
}

type ArrayValidatorBuilder struct {
	err error
	c   *arrayValidator
}

func Array() *ArrayValidatorBuilder {
	return (&ArrayValidatorBuilder{}).Reset()
}

func (b *ArrayValidatorBuilder) MinItems(v uint) *ArrayValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.minItems = &v
	return b
}

func (b *ArrayValidatorBuilder) MaxItems(v uint) *ArrayValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.maxItems = &v
	return b
}

func (b *ArrayValidatorBuilder) UniqueItems(v bool) *ArrayValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.uniqueItems = v
	return b
}

func (b *ArrayValidatorBuilder) Items(v Interface) *ArrayValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.items = v
	return b
}

func (b *ArrayValidatorBuilder) Contains(v Interface) *ArrayValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.contains = v
	return b
}

func (b *ArrayValidatorBuilder) MinContains(v uint) *ArrayValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.minContains = &v
	return b
}

func (b *ArrayValidatorBuilder) MaxContains(v uint) *ArrayValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.maxContains = &v
	return b
}

func (b *ArrayValidatorBuilder) StrictArrayType(v bool) *ArrayValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.strictArrayType = v
	return b
}

func (b *ArrayValidatorBuilder) Build() (Interface, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.c, nil
}

func (b *ArrayValidatorBuilder) MustBuild() Interface {
	if b.err != nil {
		panic(b.err)
	}
	return b.c
}

func (b *ArrayValidatorBuilder) Reset() *ArrayValidatorBuilder {
	b.err = nil
	b.c = &arrayValidator{}
	return b
}

func (c *arrayValidator) Validate(v any) error {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		length := uint(rv.Len())

		// Check minItems constraint
		if c.minItems != nil && length < *c.minItems {
			return fmt.Errorf(`invalid value passed to ArrayValidator: array length %d is below minimum items %d`, length, *c.minItems)
		}

		// Check maxItems constraint
		if c.maxItems != nil && length > *c.maxItems {
			return fmt.Errorf(`invalid value passed to ArrayValidator: array length %d exceeds maximum items %d`, length, *c.maxItems)
		}

		// Check uniqueItems constraint
		if c.uniqueItems && length > 0 {
			for i := 0; i < rv.Len(); i++ {
				item1 := rv.Index(i).Interface()
				for j := i + 1; j < rv.Len(); j++ {
					item2 := rv.Index(j).Interface()
					if reflect.DeepEqual(item1, item2) {
						return fmt.Errorf(`invalid value passed to ArrayValidator: duplicate items found, uniqueItems violation`)
					}
				}
			}
		}

		// Validate each item
		if c.items != nil {
			for i := 0; i < rv.Len(); i++ {
				item := rv.Index(i).Interface()
				if err := c.items.Validate(item); err != nil {
					return fmt.Errorf(`invalid value passed to ArrayValidator: item validation failed: %w`, err)
				}
			}
		}

		// Validate contains constraint
		if c.contains != nil {
			containsCount := uint(0)
			for i := 0; i < rv.Len(); i++ {
				item := rv.Index(i).Interface()
				if err := c.contains.Validate(item); err == nil {
					containsCount++
				}
			}

			// Check if any item matches the contains schema
			if containsCount == 0 {
				return fmt.Errorf(`invalid value passed to ArrayValidator: does not contain required item`)
			}

			// Check minContains constraint
			if c.minContains != nil && containsCount < *c.minContains {
				return fmt.Errorf(`invalid value passed to ArrayValidator: minimum contains constraint failed: found %d, expected at least %d`, containsCount, *c.minContains)
			}

			// Check maxContains constraint
			if c.maxContains != nil && containsCount > *c.maxContains {
				return fmt.Errorf(`invalid value passed to ArrayValidator: maximum contains constraint failed: found %d, expected at most %d`, containsCount, *c.maxContains)
			}
		} else {
			// Check minContains and maxContains without contains schema
			if c.minContains != nil && *c.minContains > 0 {
				return fmt.Errorf(`invalid value passed to ArrayValidator: minimum contains constraint failed: found 0, expected at least %d`, *c.minContains)
			}
		}

		return nil
	default:
		// Handle non-array values based on whether this is strict array type validation
		if c.strictArrayType {
			// When schema explicitly declares type: array, non-array values should fail
			return fmt.Errorf(`invalid value passed to ArrayValidator: expected array or slice, got %T`, v)
		}
		// For non-array values with inferred array type, array constraints don't apply
		// According to JSON Schema spec, array constraints should be ignored for non-arrays
		return nil
	}
}
