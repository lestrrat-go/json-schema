package validator

import (
	"context"
	"fmt"
	"reflect"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/internal/schemactx"
)

// unevaluatedCoordinator orchestrates validation when schemas have unevaluated constraints.
// It executes child validators (composite + base constraints) and applies unevaluated logic
// with complete annotation context.
type unevaluatedCoordinator struct {
	validators       []Interface         // Child validators (allOf, anyOf, base constraints, etc.)
	unevaluatedProps schema.SchemaOrBool // unevaluatedProperties constraint if present
	unevaluatedItems schema.SchemaOrBool // unevaluatedItems constraint if present
	strictArrayType  bool                // True when schema explicitly declares type: array
	strictObjectType bool                // True when schema explicitly declares type: object
}

// Validate orchestrates validation phases: execute all child validators, then apply unevaluated constraints
func (v *unevaluatedCoordinator) Validate(ctx context.Context, in any) (Result, error) {
	// Phase 1: Execute all child validators and collect their annotations
	merger, err := executeValidatorsAndMergeResults(ctx, v.validators, in, "unevaluated coordinator")
	if err != nil {
		return nil, err
	}

	// Phase 2: Apply unevaluated constraints with complete annotation context
	// This returns additional properties/items that were evaluated by unevaluated constraints
	additionalEvaluated, err := v.applyUnevaluatedConstraints(ctx, in, merger)
	if err != nil {
		return nil, err
	}

	// Phase 3: Update final result with any additional evaluated properties/items
	result := merger.FinalResult()
	if additionalEvaluated != nil {
		result = v.mergeAdditionalEvaluated(result, additionalEvaluated, in)
	}

	return result, nil
}

// additionalEvaluations tracks properties/items that were evaluated by unevaluated constraints
type additionalEvaluations struct {
	properties map[string]bool
	items      []bool
}

// applyUnevaluatedConstraints applies unevaluatedProperties and unevaluatedItems with annotation context
func (v *unevaluatedCoordinator) applyUnevaluatedConstraints(ctx context.Context, in any, merger *resultMerger) (*additionalEvaluations, error) {
	additional := &additionalEvaluations{
		properties: make(map[string]bool),
		items:      make([]bool, 0),
	}
	// Apply unevaluatedProperties if present
	if v.unevaluatedProps != nil {
		err := v.validateUnevaluatedProperties(ctx, in, merger.ObjectResult(), additional)
		if err != nil {
			return nil, err
		}
	}

	// Apply unevaluatedItems if present
	if v.unevaluatedItems != nil {
		err := v.validateUnevaluatedItems(ctx, in, merger.ArrayResult(), additional)
		if err != nil {
			return nil, err
		}
	}

	return additional, nil
}

// validateUnevaluatedProperties validates unevaluated object properties
func (v *unevaluatedCoordinator) validateUnevaluatedProperties(ctx context.Context, in any, objectResult *ObjectResult, additional *additionalEvaluations) error {
	// Handle different input types - only apply to objects/maps
	obj, ok := resolveToObjectMap(in)
	if !ok {
		// For non-object types, unevaluatedProperties constraints don't apply
		// unless schema explicitly declares type: object (strict mode)
		if v.strictObjectType {
			return fmt.Errorf("invalid value passed to unevaluated coordinator: expected object, got %T", in)
		}
		return nil // Non-objects pass unevaluatedProperties constraints
	}

	// Get evaluated properties from annotation context
	var evaluatedProps map[string]bool
	if objectResult != nil {
		evaluatedProps = objectResult.EvaluatedProperties()
	} else {
		evaluatedProps = make(map[string]bool)
	}

	// Also check context for previously evaluated properties
	var ec schemactx.EvaluationContext
	_ = schemactx.EvaluationContextFromContext(ctx, &ec)
	contextPropKeys := ec.Properties.Keys()

	// Merge context and current evaluations
	for _, prop := range contextPropKeys {
		if ec.Properties.IsEvaluated(prop) {
			evaluatedProps[prop] = true
		}
	}

	// Check for unevaluated properties
	for propName := range obj {
		if _, evaluated := evaluatedProps[propName]; !evaluated {
			// This property was not evaluated by any validator
			err := v.handleUnevaluatedProperty(ctx, propName, obj[propName], additional)
			if err != nil {
				return fmt.Errorf("unevaluated property %q: %w", propName, err)
			}
		}
	}

	return nil
}

// handleUnevaluatedProperty handles a single unevaluated property based on unevaluatedProperties constraint
func (v *unevaluatedCoordinator) handleUnevaluatedProperty(ctx context.Context, propName string, propValue any, additional *additionalEvaluations) error {
	switch constraint := v.unevaluatedProps.(type) {
	case schema.BoolSchema:
		if !bool(constraint) {
			// unevaluatedProperties: false - unevaluated properties not allowed
			return fmt.Errorf("property not allowed")
		}
		// unevaluatedProperties: true - allow any unevaluated properties
		// Mark this property as evaluated by the unevaluated constraint
		additional.properties[propName] = true
		return nil

	case *schema.Schema:
		// unevaluatedProperties is a schema - validate the property value against it
		validator, err := Compile(ctx, constraint)
		if err != nil {
			return fmt.Errorf("failed to compile unevaluatedProperties schema: %w", err)
		}

		_, err = validator.Validate(ctx, propValue)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
		// Mark this property as evaluated by the unevaluated constraint
		additional.properties[propName] = true
		return nil

	default:
		return fmt.Errorf("unexpected unevaluatedProperties type: %T", constraint)
	}
}

// validateUnevaluatedItems validates unevaluated array items
func (v *unevaluatedCoordinator) validateUnevaluatedItems(ctx context.Context, in any, arrayResult *ArrayResult, additional *additionalEvaluations) error {
	// Handle different input types - only apply to arrays/slices
	arr, length, ok := resolveToArray(in)
	if !ok {
		// For non-array types, unevaluatedItems constraints don't apply
		// unless schema explicitly declares type: array (strict mode)
		if v.strictArrayType {
			return fmt.Errorf("invalid value passed to unevaluated coordinator: expected array, got %T", in)
		}
		return nil // Non-arrays pass unevaluatedItems constraints
	}

	// Get evaluated items from annotation context
	var evaluatedItems []bool
	if arrayResult != nil {
		evaluatedItems = arrayResult.EvaluatedItems()
	}

	// Also check context for previously evaluated items
	var ec schemactx.EvaluationContext
	_ = schemactx.EvaluationContextFromContext(ctx, &ec)
	contextItems := ec.Items.Values()

	// Merge context and current evaluations
	maxLen := maxInt(maxInt(len(evaluatedItems), len(contextItems)), length)
	mergedEvaluated := make([]bool, maxLen)

	for i := range maxLen {
		var contextVal, currentVal bool
		if i < len(contextItems) {
			contextVal = contextItems[i]
		}
		if i < len(evaluatedItems) {
			currentVal = evaluatedItems[i]
		}
		mergedEvaluated[i] = contextVal || currentVal
	}

	// Check for unevaluated items
	for i := range length {
		if i >= len(mergedEvaluated) || !mergedEvaluated[i] {
			// This item was not evaluated by any validator
			itemValue := arr.Index(i).Interface()
			err := v.handleUnevaluatedItem(ctx, i, itemValue, additional)
			if err != nil {
				return fmt.Errorf("unevaluated item at index %d: %w", i, err)
			}
		}
	}

	return nil
}

// handleUnevaluatedItem handles a single unevaluated item based on unevaluatedItems constraint
func (v *unevaluatedCoordinator) handleUnevaluatedItem(ctx context.Context, index int, itemValue any, additional *additionalEvaluations) error {
	switch constraint := v.unevaluatedItems.(type) {
	case schema.BoolSchema:
		if !bool(constraint) {
			// unevaluatedItems: false - unevaluated items not allowed
			return fmt.Errorf("item not allowed")
		}
		// unevaluatedItems: true - allow any unevaluated items
		// Mark this item as evaluated by the unevaluated constraint
		// Expand items slice if needed
		for len(additional.items) <= index {
			additional.items = append(additional.items, false)
		}
		additional.items[index] = true
		return nil

	case *schema.Schema:
		// unevaluatedItems is a schema - validate the item value against it
		validator, err := Compile(ctx, constraint)
		if err != nil {
			return fmt.Errorf("failed to compile unevaluatedItems schema: %w", err)
		}

		_, err = validator.Validate(ctx, itemValue)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
		// Mark this item as evaluated by the unevaluated constraint
		// Expand items slice if needed
		for len(additional.items) <= index {
			additional.items = append(additional.items, false)
		}
		additional.items[index] = true
		return nil

	default:
		return fmt.Errorf("unexpected unevaluatedItems type: %T", constraint)
	}
}

// Helper functions for type resolution

func resolveToObjectMap(in any) (map[string]any, bool) {
	switch obj := in.(type) {
	case map[string]any:
		return obj, true
	default:
		// Could add more object types here if needed
		return nil, false
	}
}

func resolveToArray(in any) (reflect.Value, int, bool) {
	rv := reflect.ValueOf(in)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		return rv, rv.Len(), true
	default:
		return reflect.Value{}, 0, false
	}
}

// mergeAdditionalEvaluated merges additional evaluated properties/items into the final result
func (v *unevaluatedCoordinator) mergeAdditionalEvaluated(result Result, additional *additionalEvaluations, in any) Result {
	// If no additional evaluations, return original result
	if len(additional.properties) == 0 && len(additional.items) == 0 {
		return result
	}

	// Create new result based on input type and merge additional evaluations
	switch result := result.(type) {
	case *ObjectResult:
		// Clone the existing object result and add additional properties
		newResult := &ObjectResult{
			evaluatedProperties: make(map[string]bool),
		}
		// Copy existing evaluated properties
		for prop, eval := range result.EvaluatedProperties() {
			newResult.evaluatedProperties[prop] = eval
		}
		// Add additional evaluated properties
		for prop, eval := range additional.properties {
			newResult.evaluatedProperties[prop] = eval
		}
		return newResult

	case *ArrayResult:
		// Clone the existing array result and add additional items
		existingItems := result.EvaluatedItems()
		maxLen := maxInt(len(existingItems), len(additional.items))

		newEvaluatedItems := make([]bool, maxLen)
		// Copy existing evaluated items
		for i := range maxLen {
			var existingVal, additionalVal bool
			if i < len(existingItems) {
				existingVal = existingItems[i]
			}
			if i < len(additional.items) {
				additionalVal = additional.items[i]
			}
			newEvaluatedItems[i] = existingVal || additionalVal
		}

		return &ArrayResult{
			evaluatedItems: newEvaluatedItems,
		}

	case nil:
		// No existing result - create new result if we have additional evaluations
		if len(additional.properties) > 0 {
			// Check if input is actually an object
			if _, ok := resolveToObjectMap(in); ok {
				return &ObjectResult{
					evaluatedProperties: additional.properties,
				}
			}
		}
		if len(additional.items) > 0 {
			// Check if input is actually an array
			if _, _, ok := resolveToArray(in); ok {
				return &ArrayResult{
					evaluatedItems: additional.items,
				}
			}
		}
		return nil

	default:
		// Unknown result type, return as-is
		return result
	}
}

// maxInt returns the maximum of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
