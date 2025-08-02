package validator

import (
	"context"
	"fmt"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/internal/schemactx"
	"github.com/lestrrat-go/json-schema/vocabulary"
)

var _ Builder = (*ArrayValidatorBuilder)(nil)
var _ Interface = (*arrayValidator)(nil)

func compileArrayValidator(ctx context.Context, s *schema.Schema, strictType bool) (Interface, error) {
	v := Array()

	if s.Has(schema.MinItemsField) && vocabulary.IsKeywordEnabledInContext(ctx, "minItems") {
		v.MinItems(s.MinItems())
	}
	if s.Has(schema.MaxItemsField) && vocabulary.IsKeywordEnabledInContext(ctx, "maxItems") {
		v.MaxItems(s.MaxItems())
	}
	if s.Has(schema.UniqueItemsField) && vocabulary.IsKeywordEnabledInContext(ctx, "uniqueItems") {
		v.UniqueItems(s.UniqueItems())
	}
	if s.Has(schema.PrefixItemsField) {
		if prefixItems := s.PrefixItems(); len(prefixItems) > 0 {
			prefixValidators := make([]Interface, len(prefixItems))
			for i, prefixSchema := range prefixItems {
				prefixValidator, err := Compile(ctx, prefixSchema)
				if err != nil {
					return nil, fmt.Errorf("failed to compile prefixItems[%d] validator: %w", i, err)
				}
				prefixValidators[i] = prefixValidator
			}
			v.PrefixItems(prefixValidators)
		}
	}
	if s.Has(schema.ItemsField) {
		itemsSchema := s.Items()
		if itemsSchema != nil {
			itemValidator, err := Compile(ctx, convertSchemaOrBool(itemsSchema))
			if err != nil {
				return nil, fmt.Errorf("failed to compile items validator: %w", err)
			}
			v.Items(itemValidator)
		}
	}
	if s.Has(schema.AdditionalItemsField) {
		additionalItemsSchema := s.AdditionalItems()
		if additionalItemsSchema != nil {
			additionalItemsValidator, err := Compile(ctx, convertSchemaOrBool(additionalItemsSchema))
			if err != nil {
				return nil, fmt.Errorf("failed to compile additionalItems validator: %w", err)
			}
			v.AdditionalItems(additionalItemsValidator)
		}
	}
	if s.Has(schema.ContainsField) {
		containsSchema := s.Contains()
		if containsSchema != nil {
			// Handle SchemaOrBool types
			switch val := containsSchema.(type) {
			case schema.BoolSchema:
				// Boolean schema: true means any item matches, false means no items should match
				if bool(val) {
					// contains: true - any item matches, so any non-empty array is valid
					// We'll create a validator that always passes
					v.Contains(&EmptyValidator{})
				} else {
					// contains: false - no items should match, so any non-empty array is invalid
					// We'll create a validator that always fails
					v.Contains(&NotValidator{validator: &EmptyValidator{}})
				}
			case *schema.Schema:
				// Regular schema object
				containsValidator, err := Compile(ctx, val)
				if err != nil {
					return nil, fmt.Errorf("failed to compile contains validator: %w", err)
				}
				v.Contains(containsValidator)
			default:
				return nil, fmt.Errorf("unexpected contains type: %T", containsSchema)
			}
		}
	}
	if s.Has(schema.MinContainsField) {
		v.MinContains(s.MinContains())
	}
	if s.Has(schema.MaxContainsField) {
		v.MaxContains(s.MaxContains())
	}
	if s.Has(schema.UnevaluatedItemsField) {
		unevaluatedItems := s.UnevaluatedItems()
		if unevaluatedItems != nil {
			// Handle SchemaOrBool types
			switch val := unevaluatedItems.(type) {
			case schema.BoolSchema:
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
	minItems         *uint
	maxItems         *uint
	uniqueItems      bool
	prefixItems      []Interface
	items            Interface
	additionalItems  Interface
	contains         Interface
	minContains      *uint
	maxContains      *uint
	unevaluatedItems any  // can be bool or Interface
	strictArrayType  bool // true when schema explicitly declares type: array
}

type ArrayValidatorBuilder struct {
	err error
	c   *arrayValidator
}

// Array creates a new ArrayValidatorBuilder instance that can be used to build a
// Validator for array values according to the JSON Schema specification.
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

func (b *ArrayValidatorBuilder) PrefixItems(v []Interface) *ArrayValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.prefixItems = v
	return b
}

func (b *ArrayValidatorBuilder) Items(v Interface) *ArrayValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.items = v
	return b
}

func (b *ArrayValidatorBuilder) AdditionalItems(v Interface) *ArrayValidatorBuilder {
	if b.err != nil {
		return b
	}
	b.c.additionalItems = v
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

	var ec schemactx.EvaluationContext
	_ = schemactx.EvaluationContextFromContext(ctx, &ec)

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
			for i := range rv.Len() {
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
		result := NewArrayResult()

		// Merge evaluated items from previous validators
		var evaluatedItems schemactx.EvaluatedItems
		evaluatedItems.Copy(&ec.Items)

		// Validate items according to prefixItems and items
		arrayLength := rv.Len()
		prefixItemsCount := len(c.prefixItems)

		// First, validate items covered by prefixItems
		for i := 0; i < arrayLength && i < prefixItemsCount; i++ {
			item := rv.Index(i).Interface()
			_, err := c.prefixItems[i].Validate(ctx, item)
			if err != nil {
				return nil, fmt.Errorf(`invalid value passed to ArrayValidator: prefixItems[%d] validation failed: %w`, i, err)
			}
			// Mark this item as evaluated by prefixItems
			result.SetEvaluatedItem(i)
		}

		// Then, validate remaining items with the items schema (if present)
		if c.items != nil {
			for i := prefixItemsCount; i < arrayLength; i++ {
				item := rv.Index(i).Interface()
				_, err := c.items.Validate(ctx, item)
				if err != nil {
					return nil, fmt.Errorf(`invalid value passed to ArrayValidator: item validation failed: %w`, err)
				}
				// Mark this item as evaluated by items
				result.SetEvaluatedItem(i)
			}
		}
		// Note: Items beyond prefixItems that are not validated by items remain unevaluated

		// Validate contains constraint and track evaluation
		// According to JSON Schema spec, minContains and maxContains are ignored when contains is not present
		// No validation needed when contains schema is absent
		if c.contains != nil {
			containsCount := uint(0)
			for i := range rv.Len() {
				item := rv.Index(i).Interface()
				_, err := c.contains.Validate(ctx, item)
				if err == nil {
					containsCount++
					// Mark this item as evaluated by contains
					result.SetEvaluatedItem(i)
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
		}

		// Validate additionalItems for items beyond prefixItems
		if c.additionalItems != nil {
			// additionalItems only applies to indices beyond prefixItems
			for i := prefixItemsCount; i < arrayLength; i++ {
				// Only apply additionalItems if this item wasn't handled by items schema
				if c.items == nil {
					item := rv.Index(i).Interface()
					_, err := c.additionalItems.Validate(ctx, item)
					if err != nil {
						return nil, fmt.Errorf(`invalid value passed to ArrayValidator: additionalItems validation failed: %w`, err)
					}
					result.SetEvaluatedItem(i)
				}
			}
		}

		// Handle unevaluatedItems validation
		if c.unevaluatedItems != nil {
			// Get evaluated items from context (from previous validators) AND current result
			contextEvaluated := evaluatedItems.Values() // From previous validators
			currentEvaluated := result.EvaluatedItems() // From current validator

			// Merge context and current evaluations
			maxLen := len(contextEvaluated)
			if len(currentEvaluated) > maxLen {
				maxLen = len(currentEvaluated)
			}

			mergedEvaluated := make([]bool, maxLen)
			for i := range maxLen {
				var contextVal, currentVal bool
				if i < len(contextEvaluated) {
					contextVal = contextEvaluated[i]
				}
				if i < len(currentEvaluated) {
					currentVal = currentEvaluated[i]
				}
				mergedEvaluated[i] = contextVal || currentVal
			}

			// Validate unevaluated items
			for i := range rv.Len() {
				// Skip items that were already evaluated (by context or current validator)
				if i < len(mergedEvaluated) && mergedEvaluated[i] {
					continue
				}

				item := rv.Index(i).Interface()

				// Handle boolean unevaluatedItems
				if boolVal, ok := c.unevaluatedItems.(bool); ok {
					if !boolVal {
						// false means unevaluated items are not allowed
						return nil, fmt.Errorf(`invalid value passed to ArrayValidator: unevaluated item at index %d not allowed`, i)
					}
					// true means unevaluated items are allowed - mark as evaluated
					result.SetEvaluatedItem(i)
					continue
				}

				// Handle schema unevaluatedItems
				if validator, ok := c.unevaluatedItems.(Interface); ok {
					_, err := validator.Validate(ctx, item)
					if err != nil {
						return nil, fmt.Errorf(`invalid value passed to ArrayValidator: unevaluated item validation failed at index %d: %w`, i, err)
					}
					// Mark as evaluated when schema validation passes
					result.SetEvaluatedItem(i)
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
		//nolint: nilnil
		return nil, nil
	}
}
