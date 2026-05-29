package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/internal/schemactx"
)

var _ Builder = (*ArrayValidatorBuilder)(nil)
var _ Interface = (*arrayValidator)(nil)

func compileArrayValidator(ctx context.Context, s *schema.Schema, cs compileState, strictType bool) (Interface, error) {
	// Array keywords (prefixItems, items, contains) apply their subschemas to
	// child elements, so crossing into them is a data boundary for recursion
	// classification.
	cs = cs.incDataDepth()
	v := Array()

	if s.HasMinItems() && cs.cfg.vocab.IsKeywordEnabled("minItems") {
		v.MinItems(s.MinItems())
	}
	if s.HasMaxItems() && cs.cfg.vocab.IsKeywordEnabled("maxItems") {
		v.MaxItems(s.MaxItems())
	}
	if s.HasUniqueItems() && cs.cfg.vocab.IsKeywordEnabled("uniqueItems") {
		v.UniqueItems(s.UniqueItems())
	}
	if s.HasPrefixItems() {
		if prefixItems := s.PrefixItems(); len(prefixItems) > 0 {
			prefixValidators := make([]Interface, len(prefixItems))
			for i, prefixSchema := range prefixItems {
				prefixValidator, err := compile(ctx, convertSchemaOrBool(prefixSchema), cs)
				if err != nil {
					return nil, fmt.Errorf("failed to compile prefixItems[%d] validator: %w", i, err)
				}
				prefixValidators[i] = prefixValidator
			}
			v.PrefixItems(prefixValidators)
		}
	}
	if s.HasItems() {
		itemsSchema := s.Items()
		if itemsSchema != nil {
			itemValidator, err := compile(ctx, convertSchemaOrBool(itemsSchema), cs)
			if err != nil {
				return nil, fmt.Errorf("failed to compile items validator: %w", err)
			}
			v.Items(itemValidator)
		}
	}
	if s.HasAdditionalItems() {
		additionalItemsSchema := s.AdditionalItems()
		if additionalItemsSchema != nil {
			additionalItemsValidator, err := compile(ctx, convertSchemaOrBool(additionalItemsSchema), cs)
			if err != nil {
				return nil, fmt.Errorf("failed to compile additionalItems validator: %w", err)
			}
			v.AdditionalItems(additionalItemsValidator)
		}
	}
	if s.HasContains() {
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
				containsValidator, err := compile(ctx, val, cs)
				if err != nil {
					return nil, fmt.Errorf("failed to compile contains validator: %w", err)
				}
				v.Contains(containsValidator)
			default:
				return nil, fmt.Errorf("unexpected contains type: %T", containsSchema)
			}
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
			case schema.BoolSchema:
				// This is a boolean value
				v.UnevaluatedItemsBool(bool(val))
			case *schema.Schema:
				// This is a regular schema - validate unevaluated items with this schema
				itemValidator, err := compile(ctx, val, cs)
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

// arrayAccessor provides uniform length/element access over an array instance,
// backed either by a custom ArrayIndexResolver or by reflection on a slice/array.
type arrayAccessor struct {
	length int
	at     func(int) (any, error)
}

// newArrayAccessor honors a custom ArrayIndexResolver first, then falls back to
// reflection. The bool reports whether v is array-like at all.
func newArrayAccessor(v any) (arrayAccessor, bool) {
	// Fast path for the standard JSON-decoded shape: index the slice directly
	// instead of reflecting on each element.
	if s, ok := v.([]any); ok {
		return arrayAccessor{length: len(s), at: func(i int) (any, error) { return s[i], nil }}, true
	}

	if resolver, ok := v.(ArrayIndexResolver); ok {
		return arrayAccessor{length: resolver.Len(), at: resolver.ResolveArrayIndex}, true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		return arrayAccessor{
			length: rv.Len(),
			at:     func(i int) (any, error) { return rv.Index(i).Interface(), nil },
		}, true
	default:
		return arrayAccessor{}, false
	}
}

func (c *arrayValidator) Validate(ctx context.Context, v any, options ...ValidateOption) (Result, error) {
	return c.evaluate(ctx, v, newEvalState(ctx, options))
}

func (c *arrayValidator) evaluate(ctx context.Context, v any, st *evalState) (Result, error) {
	acc, isArray := newArrayAccessor(v)

	// Annotations from sibling applicators flow in via returned Results; this
	// starts from an empty evaluated-item set.
	var ec schemactx.EvaluationContext

	if !isArray {
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

	length := uint(acc.length)

	// Check minItems constraint
	if c.minItems != nil && length < *c.minItems {
		return nil, fmt.Errorf(`invalid value passed to ArrayValidator: array length %d is below minimum items %d`, length, *c.minItems)
	}

	// Check maxItems constraint
	if c.maxItems != nil && length > *c.maxItems {
		return nil, fmt.Errorf(`invalid value passed to ArrayValidator: array length %d exceeds maximum items %d`, length, *c.maxItems)
	}

	// Check uniqueItems constraint.
	//
	// Rather than the naive O(n^2) pairwise comparison, bucket items by their
	// canonical JSON encoding and only compare within a bucket. reflect.DeepEqual
	// implies identical json.Marshal output, so the key never yields a false
	// negative (no missed duplicate); it may collide for DeepEqual-unequal values
	// (e.g. native int(1) vs float64(1.0), both encode as "1"), which is why a real
	// duplicate is still confirmed with reflect.DeepEqual to preserve existing
	// semantics. For JSON-decoded data (the untrusted path) a shared key implies
	// equality, so the first within-bucket comparison returns immediately and the
	// scan stays linear.
	if c.uniqueItems && acc.length > 1 {
		// json.Marshal never returns empty bytes for a valid value, so this
		// sentinel cannot collide with a real key. Items that fail to marshal
		// (exotic non-JSON values from reflection or a custom ArrayIndexResolver)
		// land here and are compared among themselves.
		const unmarshalableKey = "\x00unmarshalable"
		seen := make(map[string][]any, acc.length)
		for i := range acc.length {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			item, err := acc.at(i)
			if err != nil {
				return nil, fmt.Errorf(`invalid value passed to ArrayValidator: failed to resolve item %d: %w`, i, err)
			}
			key := unmarshalableKey
			if b, err := json.Marshal(item); err == nil {
				key = string(b)
			}
			for _, prev := range seen[key] {
				if reflect.DeepEqual(prev, item) {
					return nil, fmt.Errorf(`invalid value passed to ArrayValidator: duplicate items found, uniqueItems violation`)
				}
			}
			seen[key] = append(seen[key], item)
		}
	}

	// Initialize result for tracking evaluated items
	result := NewArrayResult()

	// Merge evaluated items from previous validators
	var evaluatedItems schemactx.EvaluatedItems
	evaluatedItems.Copy(&ec.Items)

	// Validate items according to prefixItems and items
	arrayLength := acc.length
	prefixItemsCount := len(c.prefixItems)

	// First, validate items covered by prefixItems
	for i := 0; i < arrayLength && i < prefixItemsCount; i++ {
		item, err := acc.at(i)
		if err != nil {
			return nil, fmt.Errorf(`invalid value passed to ArrayValidator: failed to resolve item %d: %w`, i, err)
		}
		_, err = evalChild(ctx, c.prefixItems[i], item, st)
		if err != nil {
			return nil, fmt.Errorf(`invalid value passed to ArrayValidator: prefixItems[%d] validation failed: %w`, i, err)
		}
		// Mark this item as evaluated by prefixItems
		result.SetEvaluatedItem(i)
	}

	// Then, validate remaining items with the items schema (if present)
	if c.items != nil {
		for i := prefixItemsCount; i < arrayLength; i++ {
			item, err := acc.at(i)
			if err != nil {
				return nil, fmt.Errorf(`invalid value passed to ArrayValidator: failed to resolve item %d: %w`, i, err)
			}
			_, err = evalChild(ctx, c.items, item, st)
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
		for i := range acc.length {
			item, err := acc.at(i)
			if err != nil {
				return nil, fmt.Errorf(`invalid value passed to ArrayValidator: failed to resolve item %d: %w`, i, err)
			}
			_, err = evalChild(ctx, c.contains, item, st)
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
				item, err := acc.at(i)
				if err != nil {
					return nil, fmt.Errorf(`invalid value passed to ArrayValidator: failed to resolve item %d: %w`, i, err)
				}
				_, err = evalChild(ctx, c.additionalItems, item, st)
				if err != nil {
					return nil, fmt.Errorf(`invalid value passed to ArrayValidator: additionalItems validation failed: %w`, err)
				}
				result.SetEvaluatedItem(i)
			}
		}
	}

	// Handle unevaluatedItems validation
	if c.unevaluatedItems != nil {
		// Merge any inherited evaluated-item annotations with this validator's.
		contextEvaluated := evaluatedItems.Values() // inherited (empty unless seeded)
		currentEvaluated := result.EvaluatedItems() // from this validator

		// Merge context and current evaluations
		maxLen := max(len(contextEvaluated), len(currentEvaluated))

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
		for i := range acc.length {
			// Skip items that were already evaluated (by context or current validator)
			if i < len(mergedEvaluated) && mergedEvaluated[i] {
				continue
			}

			item, err := acc.at(i)
			if err != nil {
				return nil, fmt.Errorf(`invalid value passed to ArrayValidator: failed to resolve item %d: %w`, i, err)
			}

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
				_, err := evalChild(ctx, validator, item, st)
				if err != nil {
					return nil, fmt.Errorf(`invalid value passed to ArrayValidator: unevaluated item validation failed at index %d: %w`, i, err)
				}
				// Mark as evaluated when schema validation passes
				result.SetEvaluatedItem(i)
			}
		}
	}

	return result, nil
}
