package validator

import (
	"context"
	"fmt"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
)

var _ Builder = (*ArrayValidatorBuilder)(nil)
var _ Interface = (*arrayValidator)(nil)

func compileArrayValidator(ctx context.Context, s *schema.Schema, strictType bool) (Interface, error) {
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
			itemValidator, err := Compile(ctx, ConvertSchemaOrBool(itemsSchema))
			if err != nil {
				return nil, fmt.Errorf("failed to compile items validator: %w", err)
			}
			v.Items(itemValidator)
		}
	}
	if s.HasContains() {
		containsSchema := s.Contains()
		if containsSchema != nil {
			containsValidator, err := Compile(ctx, containsSchema)
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
	if s.HasUnevaluatedItems() {
		unevaluatedItems := s.UnevaluatedItems()
		if unevaluatedItems != nil {
			// Handle SchemaOrBool types
			switch val := unevaluatedItems.(type) {
			case schema.SchemaBool:
				// This is a boolean value
				v.UnevaluatedItemsBool(bool(val))
			case *schema.Schema:
				// This is a regular schema - validate unevaluated items with this schema
				itemValidator, err := Compile(ctx, val)
				if err != nil {
					return nil, fmt.Errorf("failed to compile unevaluated items validator: %w", err)
				}
				v.UnevaluatedItemsSchema(itemValidator)
			default:
				return nil, fmt.Errorf("unexpected unevaluatedItems type: %T", unevaluatedItems)
			}
		}
	}

	v.StrictArrayType(strictType)

	return v.Build()
}

type arrayValidator struct {
	minItems          *uint
	maxItems          *uint
	uniqueItems       bool
	items             Interface
	contains          Interface
	minContains       *uint
	maxContains       *uint
	unevaluatedItems  any // can be bool or Interface
	strictArrayType   bool // true when schema explicitly declares type: array
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

func (b *ArrayValidatorBuilder) UnevaluatedItemsBool(v bool) *ArrayValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.unevaluatedItems = v
	return b
}

func (b *ArrayValidatorBuilder) UnevaluatedItemsSchema(v Interface) *ArrayValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.unevaluatedItems = v
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

func (c *arrayValidator) Validate(ctx context.Context, v any) (Result, error) {
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
			return nil, fmt.Errorf(`invalid value passed to ArrayValidator: array length %d is below minimum items %d`, length, *c.minItems)
		}

		// Check maxItems constraint
		if c.maxItems != nil && length > *c.maxItems {
			return nil, fmt.Errorf(`invalid value passed to ArrayValidator: array length %d exceeds maximum items %d`, length, *c.maxItems)
		}

		// Check uniqueItems constraint
		if c.uniqueItems && length > 0 {
			for i := 0; i < rv.Len(); i++ {
				item1 := rv.Index(i).Interface()
				for j := i + 1; j < rv.Len(); j++ {
					item2 := rv.Index(j).Interface()
					if reflect.DeepEqual(item1, item2) {
						return nil, fmt.Errorf(`invalid value passed to ArrayValidator: duplicate items found, uniqueItems violation`)
					}
				}
			}
		}

		// Initialize result for tracking evaluated items
		result := &ArrayResult{
			EvaluatedItems: make([]bool, rv.Len()),
		}

		// Validate each item and track evaluation
		if c.items != nil {
			for i := 0; i < rv.Len(); i++ {
				item := rv.Index(i).Interface()
				_, err := c.items.Validate(ctx, item)
				if err != nil {
					return nil, fmt.Errorf(`invalid value passed to ArrayValidator: item validation failed: %w`, err)
				}
				// Mark this item as evaluated
				result.EvaluatedItems[i] = true
			}
		}

		// Validate contains constraint and track evaluation
		if c.contains != nil {
			containsCount := uint(0)
			for i := 0; i < rv.Len(); i++ {
				item := rv.Index(i).Interface()
				_, err := c.contains.Validate(ctx, item)
				if err == nil {
					containsCount++
					// Mark this item as evaluated by contains
					result.EvaluatedItems[i] = true
				}
			}

			// Check minContains constraint first
			if c.minContains != nil && containsCount < *c.minContains {
				return nil, fmt.Errorf(`invalid value passed to ArrayValidator: minimum contains constraint failed: found %d, expected at least %d`, containsCount, *c.minContains)
			}

			// Check if any item matches the contains schema (only if minContains is not explicitly set to 0)
			if containsCount == 0 && (c.minContains == nil || *c.minContains > 0) {
				return nil, fmt.Errorf(`invalid value passed to ArrayValidator: does not contain required item`)
			}

			// Check maxContains constraint
			if c.maxContains != nil && containsCount > *c.maxContains {
				return nil, fmt.Errorf(`invalid value passed to ArrayValidator: maximum contains constraint failed: found %d, expected at most %d`, containsCount, *c.maxContains)
			}
		} else {
			// Check minContains and maxContains without contains schema
			if c.minContains != nil && *c.minContains > 0 {
				return nil, fmt.Errorf(`invalid value passed to ArrayValidator: minimum contains constraint failed: found 0, expected at least %d`, *c.minContains)
			}
		}

		// Handle unevaluatedItems validation
		if c.unevaluatedItems != nil {
			// Get stash from context to see what items were already evaluated by other validators
			stash := StashFromContext(ctx)
			if stash != nil && stash.EvaluatedItems != nil {
				// Merge stash evaluated items with our current result
				maxLen := len(result.EvaluatedItems)
				if len(stash.EvaluatedItems) > maxLen {
					maxLen = len(stash.EvaluatedItems)
				}
				
				// Extend our result if necessary
				if len(result.EvaluatedItems) < maxLen {
					newEvaluated := make([]bool, maxLen)
					copy(newEvaluated, result.EvaluatedItems)
					result.EvaluatedItems = newEvaluated
				}
				
				// Mark items as evaluated if they were evaluated by previous validators
				for i := 0; i < len(stash.EvaluatedItems) && i < maxLen; i++ {
					if stash.EvaluatedItems[i] {
						result.EvaluatedItems[i] = true
					}
				}
			}

			// Validate unevaluated items
			for i := 0; i < rv.Len(); i++ {
				// Skip items that were already evaluated
				if i < len(result.EvaluatedItems) && result.EvaluatedItems[i] {
					continue
				}
				
				item := rv.Index(i).Interface()
				
				// Handle boolean unevaluatedItems
				if boolVal, ok := c.unevaluatedItems.(bool); ok {
					if !boolVal {
						// false means unevaluated items are not allowed
						return nil, fmt.Errorf(`invalid value passed to ArrayValidator: unevaluated item at index %d not allowed`, i)
					}
					// true means unevaluated items are allowed - no validation needed
					continue
				}
				
				// Handle schema unevaluatedItems
				if validator, ok := c.unevaluatedItems.(Interface); ok {
					_, err := validator.Validate(ctx, item)
					if err != nil {
						return nil, fmt.Errorf(`invalid value passed to ArrayValidator: unevaluated item validation failed at index %d: %w`, i, err)
					}
				}
			}
		}

		return result, nil
	default:
		// Handle non-array values based on whether this is strict array type validation
		if c.strictArrayType {
			// When schema explicitly declares type: array, non-array values should fail
			return nil, fmt.Errorf(`invalid value passed to ArrayValidator: expected array or slice, got %T`, v)
		}
		// For non-array values with inferred array type, array constraints don't apply
		// According to JSON Schema spec, array constraints should be ignored for non-arrays
		return nil, nil
	}
}
