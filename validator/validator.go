//go:generate ./gen.sh

package validator

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/lestrrat-go/blackmagic"
	schema "github.com/lestrrat-go/json-schema"
)

// Interface is the interface that all validators must implement.
type Interface interface {
	Validate(context.Context, any) (Result, error)
}

// Result contains annotation information from validation that may be used
// by other validators (e.g., for unevaluatedProperties tracking)
type Result any

// ObjectResult contains information about which object properties were evaluated
type ObjectResult struct {
	evaluatedProperties map[string]bool // property name -> true if evaluated
}

// ArrayResult contains information about which array items were evaluated
type ArrayResult struct {
	evaluatedItems []bool // index -> true if evaluated, length determines max evaluated index
}

// NewObjectResult creates a new ObjectResult with an initialized map
func NewObjectResult() *ObjectResult {
	return &ObjectResult{
		evaluatedProperties: make(map[string]bool),
	}
}

// NewArrayResult creates a new ArrayResult with an initialized slice
func NewArrayResult() *ArrayResult {
	return &ArrayResult{
		evaluatedItems: make([]bool, 0),
	}
}

// EvaluatedProperties returns a copy of the evaluated properties map
func (r *ObjectResult) EvaluatedProperties() map[string]bool {
	if r == nil || r.evaluatedProperties == nil {
		return make(map[string]bool)
	}
	result := make(map[string]bool, len(r.evaluatedProperties))
	for k, v := range r.evaluatedProperties {
		result[k] = v
	}
	return result
}

// SetEvaluatedProperty marks a property as evaluated
func (r *ObjectResult) SetEvaluatedProperty(prop string) {
	if r != nil && r.evaluatedProperties != nil {
		r.evaluatedProperties[prop] = true
	}
}

// EvaluatedItems returns a copy of the evaluated items slice
func (r *ArrayResult) EvaluatedItems() []bool {
	if r == nil || r.evaluatedItems == nil {
		return make([]bool, 0)
	}
	result := make([]bool, len(r.evaluatedItems))
	copy(result, r.evaluatedItems)
	return result
}

// SetEvaluatedItem marks an item at the given index as evaluated
func (r *ArrayResult) SetEvaluatedItem(index int) {
	if r == nil {
		return
	}
	// Extend slice if necessary
	for len(r.evaluatedItems) <= index {
		r.evaluatedItems = append(r.evaluatedItems, false)
	}
	r.evaluatedItems[index] = true
}

// SetEvaluatedItems sets the entire slice of evaluated items
func (r *ArrayResult) SetEvaluatedItems(items []bool) {
	if r == nil {
		return
	}
	r.evaluatedItems = make([]bool, len(items))
	copy(r.evaluatedItems, items)
}

// mergeObjectResults merges multiple ObjectResult instances into a single result
func mergeObjectResults(results ...*ObjectResult) *ObjectResult {
	merged := NewObjectResult()
	for _, result := range results {
		if result != nil && result.evaluatedProperties != nil {
			for prop := range result.evaluatedProperties {
				merged.evaluatedProperties[prop] = true
			}
		}
	}
	return merged
}

// mergeArrayResults merges multiple ArrayResult instances into a single result
func mergeArrayResults(results ...*ArrayResult) *ArrayResult {
	merged := NewArrayResult()
	maxLen := 0

	// Find the maximum length needed
	for _, result := range results {
		if result != nil && len(result.evaluatedItems) > maxLen {
			maxLen = len(result.evaluatedItems)
		}
	}

	// Initialize with the correct length
	merged.evaluatedItems = make([]bool, maxLen)

	// Merge all results
	for _, result := range results {
		if result != nil {
			for i, evaluated := range result.evaluatedItems {
				if evaluated {
					merged.evaluatedItems[i] = true
				}
			}
		}
	}

	return merged
}

// MergeResults merges multiple validation results and assigns the final result to dst
// using blackmagic.AssignIfCompatible for type-safe assignment
func MergeResults(dst interface{}, results ...interface{}) error {
	var objectResults []*ObjectResult
	var arrayResults []*ArrayResult

	// Collect results by type
	for _, result := range results {
		switch r := result.(type) {
		case *ObjectResult:
			if r != nil {
				objectResults = append(objectResults, r)
			}
		case *ArrayResult:
			if r != nil {
				arrayResults = append(arrayResults, r)
			}
		}
	}

	// Merge based on destination type
	switch dst.(type) {
	case **ObjectResult:
		merged := mergeObjectResults(objectResults...)
		return blackmagic.AssignIfCompatible(dst, merged)
	case **ArrayResult:
		merged := mergeArrayResults(arrayResults...)
		return blackmagic.AssignIfCompatible(dst, merged)
	default:
		return fmt.Errorf("unsupported destination type: %T", dst)
	}
}

// Stash contains annotation data passed between validators via context
type Stash struct {
	EvaluatedProperties map[string]bool // properties already evaluated by previous validators
	EvaluatedItems      []bool          // items already evaluated by previous validators
}

type stashKey struct{}
type dependentSchemasKey struct{}

// WithStash adds a Stash to the context for passing annotation data to sub-validators
func WithStash(ctx context.Context, stash *Stash) context.Context {
	return context.WithValue(ctx, stashKey{}, stash)
}

// StashFromContext extracts the Stash from context, returns nil if no stash is associated with ctx
func StashFromContext(ctx context.Context) *Stash {
	if stash, ok := ctx.Value(stashKey{}).(*Stash); ok {
		return stash
	}
	return nil
}

// WithDependentSchemas adds compiled dependent schema validators to the context
func WithDependentSchemas(ctx context.Context, dependentSchemas map[string]Interface) context.Context {
	return context.WithValue(ctx, dependentSchemasKey{}, dependentSchemas)
}

// DependentSchemasFromContext extracts compiled dependent schema validators from context, returns nil if none are associated with ctx
func DependentSchemasFromContext(ctx context.Context) map[string]Interface {
	if deps, ok := ctx.Value(dependentSchemasKey{}).(map[string]Interface); ok {
		return deps
	}
	return nil
}

type Builder interface {
	Build() (Interface, error)
	MustBuild() Interface
}

// convertSchemaOrBool converts a SchemaOrBool to a *Schema.
// When the value is true, it returns an empty Schema which accepts everything.
// When the value is false, it returns a Schema with "not": {} which rejects everything.
// When the value is already a *Schema, it returns the schema as-is.
// When the value is a map[string]interface{} from JSON unmarshaling, it converts it to a Schema.
func convertSchemaOrBool(v schema.SchemaOrBool) *schema.Schema {
	switch val := v.(type) {
	case schema.SchemaBool:
		if bool(val) {
			// true schema accepts everything
			return schema.New()
		} else {
			// false schema rejects everything
			return schema.NewBuilder().Not(schema.New()).MustBuild()
		}
	case *schema.Schema:
		return val
	default:
		// This shouldn't happen if validation is working correctly
		panic(fmt.Sprintf("invalid SchemaOrBool type: %T", v))
	}
}

// hasOtherConstraints checks if a schema has constraints other than $ref/$dynamicRef
func hasOtherConstraints(s *schema.Schema) bool {
	// Check for types separately since it's not a bit field check
	if len(s.Types()) > 0 {
		return true
	}

	// Use bit field approach for efficient checking of multiple constraints
	constraintFields := schema.AllOfField | schema.AnyOfField | schema.OneOfField | schema.NotField |
		schema.IfSchemaField | schema.ThenSchemaField | schema.ElseSchemaField |
		schema.PropertiesField | schema.PatternPropertiesField | schema.AdditionalPropertiesField |
		schema.UnevaluatedPropertiesField | schema.RequiredField | schema.MinPropertiesField | schema.MaxPropertiesField |
		schema.DependentSchemasField | schema.ItemsField | schema.ContainsField | schema.MinItemsField | schema.MaxItemsField |
		schema.UniqueItemsField | schema.UnevaluatedItemsField | schema.MinLengthField | schema.MaxLengthField |
		schema.PatternField | schema.FormatField | schema.MinimumField | schema.MaximumField | schema.ExclusiveMinimumField |
		schema.ExclusiveMaximumField | schema.MultipleOfField | schema.EnumField | schema.ConstField |
		schema.ContentEncodingField | schema.ContentMediaTypeField | schema.ContentSchemaField |
		schema.PropertyNamesField

	// Returns true if ANY of the constraint fields are set
	return s.HasAny(constraintFields)
}

// createSchemaWithoutRef creates a copy of the schema without the $ref/$dynamicRef constraint
func createSchemaWithoutRef(s *schema.Schema) *schema.Schema {
	// Use the new Clone Builder pattern to create a copy without the $ref/$dynamicRef field
	builder := schema.NewBuilder().Clone(s).ResetReference()
	if s.HasDynamicReference() {
		builder = builder.ResetDynamicReference()
	}
	return builder.MustBuild()
}

func Compile(ctx context.Context, s *schema.Schema) (Interface, error) {
	// Set up context with default resolver if none provided
	if ResolverFromContext(ctx) == nil {
		ctx = WithResolver(ctx, schema.NewResolver())
	}

	// Set up context with root schema if none provided
	if RootSchemaFromContext(ctx) == nil {
		ctx = WithRootSchema(ctx, s)
	}

	// Set up base URI context from schema's $id field if present
	if s.HasID() {
		schemaID := s.ID()
		if schemaID != "" {
			// Extract base URI from $id for resolving relative references within this schema
			if baseURI := extractBaseURI(schemaID); baseURI != "" {
				ctx = WithBaseURI(ctx, baseURI)
			}
		}
	}

	// Handle $ref and $dynamicRef first - if schema has a reference, resolve it immediately
	var reference string
	if s.HasReference() {
		reference = s.Reference()
	} else if s.HasDynamicReference() {
		// For now, treat $dynamicRef like a normal $ref (simple case)
		reference = s.DynamicReference()
	}

	if reference != "" {
		// Get resolver from context (guaranteed to be present)
		resolver := ResolverFromContext(ctx)

		// Get root schema from context (guaranteed to be present)
		rootSchema := RootSchemaFromContext(ctx)

		// Check for circular references by looking at context
		if stack := ctx.Value(referenceStackKey); stack != nil {
			refStack := stack.([]string)
			for _, ref := range refStack {
				if ref == reference {
					return nil, fmt.Errorf("circular reference detected: %s", reference)
				}
			}
			// Add current reference to stack
			newStack := make([]string, len(refStack)+1)
			copy(newStack, refStack)
			newStack[len(refStack)] = reference
			ctx = context.WithValue(ctx, referenceStackKey, newStack)
		} else {
			// Start new reference stack
			ctx = context.WithValue(ctx, referenceStackKey, []string{reference})
		}

		// Resolve the reference to get the target schema
		var targetSchema schema.Schema
		baseURI := BaseURIFromContext(ctx)
		if err := resolver.ResolveReferenceWithBaseURI(&targetSchema, rootSchema, reference, baseURI); err != nil {
			return nil, fmt.Errorf("failed to resolve reference %s: %w", reference, err)
		}

		// If the target schema has relative references, we need to ensure they're resolved
		// against the correct base URI. For metaschema, this is crucial.
		if targetSchema.HasID() && targetSchema.ID() != "" {
			// Set the base URI from the target schema's $id
			if baseURI := extractBaseURI(targetSchema.ID()); baseURI != "" {
				ctx = WithBaseURI(ctx, baseURI)
			}
		}

		// Compile the reference validator with the target schema as the new root
		// This ensures that relative references within the target schema are resolved correctly
		refCtx := WithRootSchema(ctx, &targetSchema)

		// Set base URI context for resolving relative references within the target schema
		// This is crucial for metaschema which has relative references like "meta/validation"
		if targetSchema.HasID() && targetSchema.ID() != "" {
			if baseURI := extractBaseURI(targetSchema.ID()); baseURI != "" {
				refCtx = WithBaseURI(refCtx, baseURI)
			}
		} else if len(reference) > 0 && reference[0] != '#' {
			// Extract base URI from remote reference if target schema has no $id
			if baseURI := extractBaseURI(reference); baseURI != "" {
				refCtx = WithBaseURI(refCtx, baseURI)
			}
		}

		refValidator, err := Compile(refCtx, &targetSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to compile reference validator: %w", err)
		}

		// Check if the schema has other constraints besides $ref
		if hasOtherConstraints(s) {
			// Special handling for $ref + unevaluatedProperties
			if s.HasUnevaluatedProperties() && (s.HasProperties() || s.HasPatternProperties() || s.HasAdditionalProperties()) {
				// Create a composition validator that properly handles annotation flow
				compositionValidator := NewRefUnevaluatedPropertiesCompositionValidator(ctx, s, refValidator)
				return compositionValidator, nil
			}

			// Create a composite validator that combines $ref with other constraints
			// First, create a schema without the $ref for other constraints
			otherSchema := createSchemaWithoutRef(s)
			otherValidator, err := Compile(ctx, otherSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to compile other constraints validator: %w", err)
			}

			// Create a MultiValidator with allOf logic to combine both
			compositeValidator := NewMultiValidator(AndMode)
			compositeValidator.Append(refValidator)
			compositeValidator.Append(otherValidator)
			return compositeValidator, nil
		}

		// Only $ref constraint, return the reference validator
		return refValidator, nil
	}

	var allValidators []Interface

	// Handle schema composition first
	if s.HasAllOf() {
		// Special handling for allOf with unevaluatedProperties in base schema
		if hasBaseConstraints(s) && s.HasUnevaluatedProperties() {
			// Create a special validator that evaluates allOf first, then base constraints with annotation context
			compositionValidator, err := NewUnevaluatedPropertiesCompositionValidatorWithResolver(ctx, s, ResolverFromContext(ctx))
			if err != nil {
				return nil, fmt.Errorf(`failed to compile allOf composition validator: %w`, err)
			}
			allValidators = append(allValidators, compositionValidator)
		} else {
			allOfValidators := make([]Interface, 0, len(s.AllOf())+1)

			// If the schema has base properties/constraints, create a base validator first
			if hasBaseConstraints(s) {
				baseSchema := createBaseSchema(s)
				baseValidator, err := Compile(ctx, baseSchema)
				if err != nil {
					return nil, fmt.Errorf(`failed to compile base schema for allOf: %w`, err)
				}
				allOfValidators = append(allOfValidators, baseValidator)
			}

			for _, subSchema := range s.AllOf() {
				convertedSchema := convertSchemaOrBool(subSchema)
				v, err := Compile(ctx, convertedSchema)
				if err != nil {
					return nil, fmt.Errorf(`failed to compile allOf validator: %w`, err)
				}
				allOfValidators = append(allOfValidators, v)
			}
			allOfValidator := NewMultiValidator(AndMode)
			for _, v := range allOfValidators {
				allOfValidator.Append(v)
			}
			allValidators = append(allValidators, allOfValidator)
		}
	}

	if s.HasAnyOf() {
		anyOfValidators := make([]Interface, 0, len(s.AnyOf()))
		for _, subSchema := range s.AnyOf() {
			v, err := Compile(ctx, convertSchemaOrBool(subSchema))
			if err != nil {
				return nil, fmt.Errorf(`failed to compile anyOf validator: %w`, err)
			}
			anyOfValidators = append(anyOfValidators, v)
		}

		if hasBaseConstraints(s) && s.HasUnevaluatedProperties() {
			// Special anyOf composition validator for unevaluatedProperties
			compositionValidator, err := NewAnyOfUnevaluatedPropertiesCompositionValidatorWithResolver(ctx, s, anyOfValidators, ResolverFromContext(ctx))
			if err != nil {
				return nil, fmt.Errorf(`failed to compile anyOf composition validator: %w`, err)
			}
			allValidators = append(allValidators, compositionValidator)
		} else {
			anyOfValidator := NewMultiValidator(OrMode)
			for _, v := range anyOfValidators {
				anyOfValidator.Append(v)
			}
			allValidators = append(allValidators, anyOfValidator)
		}
	}

	if s.HasOneOf() {
		oneOfValidators := make([]Interface, 0, len(s.OneOf()))
		for _, subSchema := range s.OneOf() {
			v, err := Compile(ctx, convertSchemaOrBool(subSchema))
			if err != nil {
				return nil, fmt.Errorf(`failed to compile oneOf validator: %w`, err)
			}
			oneOfValidators = append(oneOfValidators, v)
		}

		if hasBaseConstraints(s) && s.HasUnevaluatedProperties() {
			// Special oneOf composition validator for unevaluatedProperties
			compositionValidator, err := NewOneOfUnevaluatedPropertiesCompositionValidatorWithResolver(ctx, s, oneOfValidators, ResolverFromContext(ctx))
			if err != nil {
				return nil, fmt.Errorf(`failed to compile oneOf composition validator: %w`, err)
			}
			allValidators = append(allValidators, compositionValidator)
		} else {
			oneOfValidator := NewMultiValidator(OneOfMode)
			for _, v := range oneOfValidators {
				oneOfValidator.Append(v)
			}
			allValidators = append(allValidators, oneOfValidator)
		}
	}

	if s.HasNot() {
		notValidator, err := Compile(ctx, s.Not())
		if err != nil {
			return nil, fmt.Errorf(`failed to compile not validator: %w`, err)
		}
		allValidators = append(allValidators, &NotValidator{validator: notValidator})
	}

	// Handle if/then/else conditional validation
	if s.HasIfSchema() {
		// Special handling for if/then/else with unevaluatedProperties in base schema
		if hasBaseConstraints(s) && s.HasUnevaluatedProperties() {
			// Create a special validator that evaluates if/then/else first, then base constraints with annotation context
			compositionValidator := NewIfThenElseUnevaluatedPropertiesCompositionValidator(ctx, s)
			allValidators = append(allValidators, compositionValidator)
		} else {
			ifThenElseValidator, err := compileIfThenElseValidator(ctx, s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile if/then/else validator: %w`, err)
			}
			allValidators = append(allValidators, ifThenElseValidator)
		}
	}

	// Compile dependent schemas and pass in context for two-pass validation
	if s.HasDependentSchemas() {
		compiledDependentSchemas := make(map[string]Interface)
		for propertyName, depSchema := range s.DependentSchemas() {
			// Handle SchemaOrBool types
			switch val := depSchema.(type) {
			case schema.SchemaBool:
				// Boolean schema: true means always valid, false means always invalid
				if bool(val) {
					compiledDependentSchemas[propertyName] = &alwaysPassValidator{}
				} else {
					compiledDependentSchemas[propertyName] = &alwaysFailValidator{}
				}
			case *schema.Schema:
				// Regular schema object
				depValidator, err := Compile(ctx, val)
				if err != nil {
					return nil, fmt.Errorf("failed to compile dependent schema for property %s: %w", propertyName, err)
				}
				compiledDependentSchemas[propertyName] = depValidator
			default:
				return nil, fmt.Errorf("unexpected dependent schema type for property %s: %T", propertyName, depSchema)
			}
		}
		ctx = WithDependentSchemas(ctx, compiledDependentSchemas)
	}

	// Handle type-specific validators
	explicitTypes := s.Types()
	types := make([]schema.PrimitiveType, len(explicitTypes))
	copy(types, explicitTypes)
	var validatorsByType []Interface

	// Track which types were inferred (not explicitly declared)
	inferredTypes := make(map[schema.PrimitiveType]bool)

	// If no types are specified but type-specific constraints are present,
	// infer the type from the constraints
	// Skip this if allOf is present and has base constraints (they'll be handled in allOf)
	// Also skip if we have anyOf/oneOf composition validators that will handle these constraints
	hasCompositionValidator := (s.HasAllOf() && hasBaseConstraints(s)) ||
		(s.HasAnyOf() && hasBaseConstraints(s) && s.HasUnevaluatedProperties()) ||
		(s.HasOneOf() && hasBaseConstraints(s) && s.HasUnevaluatedProperties()) ||
		(s.HasIfSchema() && hasBaseConstraints(s) && s.HasUnevaluatedProperties())

	if len(types) == 0 && !hasCompositionValidator {
		if s.HasMinLength() || s.HasMaxLength() || s.HasPattern() {
			types = append(types, schema.StringType)
			inferredTypes[schema.StringType] = true
		}
		if s.HasMinimum() || s.HasMaximum() || s.HasExclusiveMinimum() || s.HasExclusiveMaximum() || s.HasMultipleOf() {
			// For inferred numeric types, create a non-strict validator that only validates numeric constraints when values are numbers
			v, err := compileInferredNumberValidator(ctx, s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile inferred number validator: %w`, err)
			}
			allValidators = append(allValidators, v)
		}
		if s.HasMinItems() || s.HasMaxItems() || s.HasUniqueItems() || s.HasItems() || s.HasContains() || s.HasUnevaluatedItems() {
			// For inferred array types, create a non-strict array validator
			v, err := compileArrayValidator(ctx, s, false)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile inferred array validator: %w`, err)
			}
			allValidators = append(allValidators, v)
		}
		if s.HasMinProperties() || s.HasMaxProperties() || s.HasRequired() || s.HasDependentRequired() || s.HasProperties() || s.HasPatternProperties() || s.HasAdditionalProperties() || s.HasUnevaluatedProperties() || s.HasDependentSchemas() || s.HasPropertyNames() {
			// For inferred object types, create non-strict object validator
			v, err := compileObjectValidator(ctx, s, false)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile inferred object validator: %w`, err)
			}
			allValidators = append(allValidators, v)
		}
	}

	// Handle general enum/const validation when no specific type is set
	// Skip this if we have composition validators that will handle these constraints
	if len(types) == 0 && (s.HasEnum() || s.HasConst()) && !hasCompositionValidator {
		validator, err := compileGeneralValidator(ctx, s)
		if err != nil {
			return nil, fmt.Errorf(`failed to compile general validator: %w`, err)
		}
		allValidators = append(allValidators, validator)
	}

	for _, typ := range types {
		// This is a placeholder code. In reality we need to
		// OR all types
		switch typ {
		case schema.StringType:
			// Use strict type checking only for explicitly declared string types
			strictType := !inferredTypes[schema.StringType]
			v, err := compileStringValidator(ctx, s, strictType)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile string validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.IntegerType:
			v, err := compileIntegerValidator(ctx, s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile integer validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.NumberType:
			v, err := compileNumberValidator(ctx, s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile number validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.BooleanType:
			v, err := compileBooleanValidator(ctx, s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile boolean validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.ArrayType:
			v, err := compileArrayValidator(ctx, s, true)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile array validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.ObjectType:
			v, err := compileObjectValidator(ctx, s, true)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile object validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		case schema.NullType:
			v, err := compileNullValidator(ctx, s)
			if err != nil {
				return nil, fmt.Errorf(`failed to compile null validator: %w`, err)
			}
			validatorsByType = append(validatorsByType, v)
		}
	}

	// Combine type validators if multiple types
	if len(validatorsByType) > 1 {
		typeValidator := NewMultiValidator(OrMode)
		for _, v := range validatorsByType {
			typeValidator.Append(v)
		}
		allValidators = append(allValidators, typeValidator)
	} else if len(validatorsByType) == 1 {
		allValidators = append(allValidators, validatorsByType[0])
	}

	// Handle content validation (contentEncoding, contentMediaType, contentSchema)
	if contentValidator, err := compileContentValidator(ctx, s); err != nil {
		return nil, fmt.Errorf(`failed to compile content validator: %w`, err)
	} else if contentValidator != nil {
		allValidators = append(allValidators, contentValidator)
	}

	// Return the appropriate validator
	if len(allValidators) == 0 {
		// Empty schema - allows anything
		return &EmptyValidator{}, nil
	}

	if len(allValidators) == 1 {
		return allValidators[0], nil
	}

	// Multiple validators - combine with AND
	mv := NewMultiValidator(AndMode)
	for _, v := range allValidators {
		mv.Append(v)
	}

	return mv, nil
}

// inferredNumberValidator validates numeric constraints only when the value is a number,
// ignoring non-numeric values (for inferred number types without explicit type declaration)
type inferredNumberValidator struct {
	numberValidator Interface
}

func compileInferredNumberValidator(ctx context.Context, s *schema.Schema) (Interface, error) {
	// Create the underlying number validator
	numValidator, err := compileNumberValidator(ctx, s)
	if err != nil {
		return nil, err
	}

	return &inferredNumberValidator{
		numberValidator: numValidator,
	}, nil
}

func (v *inferredNumberValidator) Validate(ctx context.Context, in any) (Result, error) {
	// Check if the value is numeric
	rv := reflect.ValueOf(in)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		// Value is numeric, apply number validation
		return v.numberValidator.Validate(ctx, in)
	default:
		// Value is not numeric, ignore numeric constraints (per JSON Schema spec)
		return nil, nil
	}
}

type EmptyValidator struct{}

func (e *EmptyValidator) Validate(ctx context.Context, v any) (Result, error) {
	// Empty schema allows anything
	return nil, nil
}

type NotValidator struct {
	validator Interface
}

func (n *NotValidator) Validate(ctx context.Context, v any) (Result, error) {
	_, err := n.validator.Validate(ctx, v)
	if err == nil {
		return nil, fmt.Errorf(`not validation failed: value should not validate against the schema`)
	}
	return nil, nil
}

type NullValidator struct{}

func (n *NullValidator) Validate(ctx context.Context, v any) (Result, error) {
	if v == nil {
		return nil, nil
	}
	return nil, fmt.Errorf(`invalid value passed to NullValidator: expected null, got %T`, v)
}

func compileNullValidator(ctx context.Context, s *schema.Schema) (Interface, error) {
	return &NullValidator{}, nil
}

// GeneralValidator handles enum and const validation for schemas without specific types
type GeneralValidator struct {
	enum     []any
	const_   any
	hasConst bool
}

func compileGeneralValidator(ctx context.Context, s *schema.Schema) (Interface, error) {
	v := &GeneralValidator{}

	if s.HasEnum() {
		v.enum = s.Enum()
	}

	if s.HasConst() {
		v.const_ = s.Const()
		v.hasConst = true
	}

	return v, nil
}

func (g *GeneralValidator) Validate(ctx context.Context, value any) (Result, error) {
	// Check const first
	if g.hasConst {
		if !reflect.DeepEqual(value, g.const_) {
			return nil, fmt.Errorf(`invalid value: must equal const value %v, got %v`, g.const_, value)
		}
		return nil, nil
	}

	// Check enum
	if g.enum != nil {
		for _, enumVal := range g.enum {
			if reflect.DeepEqual(value, enumVal) {
				return nil, nil
			}
		}
		return nil, fmt.Errorf(`invalid value: %v not found in enum %v`, value, g.enum)
	}

	return nil, nil
}

// ReferenceValidator handles schema references ($ref) with lazy resolution and circular detection
type ReferenceValidator struct {
	reference    string
	resolvedOnce sync.Once
	resolved     Interface
	resolveErr   error
}

func (r *ReferenceValidator) Validate(ctx context.Context, v any) (Result, error) {
	// Lazy resolution - only resolve when actually needed for validation
	r.resolvedOnce.Do(func() {
		r.resolved, r.resolveErr = r.resolveReference(ctx)
	})

	if r.resolveErr != nil {
		return nil, fmt.Errorf("reference resolution failed for %s: %w", r.reference, r.resolveErr)
	}

	return r.resolved.Validate(ctx, v)
}

func (r *ReferenceValidator) resolveReference(ctx context.Context) (Interface, error) {
	// Get resolver from context - create default if none provided
	resolver := ResolverFromContext(ctx)
	if resolver == nil {
		resolver = schema.NewResolver()
	}

	// Get root schema from context
	rootSchema := RootSchemaFromContext(ctx)
	if rootSchema == nil {
		return nil, fmt.Errorf("no root schema available in context for reference resolution: %s", r.reference)
	}

	// Check for circular references by looking at context
	if stack := ctx.Value(referenceStackKey); stack != nil {
		refStack := stack.([]string)
		for _, ref := range refStack {
			if ref == r.reference {
				return nil, fmt.Errorf("circular reference detected: %s", r.reference)
			}
		}
		// Add current reference to stack
		newStack := make([]string, len(refStack)+1)
		copy(newStack, refStack)
		newStack[len(refStack)] = r.reference
		ctx = context.WithValue(ctx, referenceStackKey, newStack)
	} else {
		// Start new reference stack
		ctx = context.WithValue(ctx, referenceStackKey, []string{r.reference})
	}

	// Resolve the reference to get the target schema
	var targetSchema schema.Schema
	baseURI := BaseURIFromContext(ctx)
	if err := resolver.ResolveReferenceWithBaseURI(&targetSchema, rootSchema, r.reference, baseURI); err != nil {
		return nil, fmt.Errorf("failed to resolve reference %s: %w", r.reference, err)
	}

	// Compile the resolved schema into a validator
	return Compile(ctx, &targetSchema)
}

// Context keys for passing resolver and root schema
type resolverKeyType struct{}
type rootSchemaKeyType struct{}
type referenceStackKeyType struct{}
type baseURIKeyType struct{}

var resolverKey = resolverKeyType{}
var rootSchemaKey = rootSchemaKeyType{}
var referenceStackKey = referenceStackKeyType{}
var baseURIKey = baseURIKeyType{}

// WithResolver adds a resolver to the context
func WithResolver(ctx context.Context, resolver *schema.Resolver) context.Context {
	return context.WithValue(ctx, resolverKey, resolver)
}

// ResolverFromContext extracts the resolver from context, returns nil if not present
func ResolverFromContext(ctx context.Context) *schema.Resolver {
	if resolver, ok := ctx.Value(resolverKey).(*schema.Resolver); ok {
		return resolver
	}
	return nil
}

// WithRootSchema adds the root schema to the context
func WithRootSchema(ctx context.Context, rootSchema *schema.Schema) context.Context {
	return context.WithValue(ctx, rootSchemaKey, rootSchema)
}

// RootSchemaFromContext extracts the root schema from context, returns nil if not present
func RootSchemaFromContext(ctx context.Context) *schema.Schema {
	if rootSchema, ok := ctx.Value(rootSchemaKey).(*schema.Schema); ok {
		return rootSchema
	}
	return nil
}

// WithBaseURI adds a base URI to the context for reference resolution
func WithBaseURI(ctx context.Context, baseURI string) context.Context {
	return context.WithValue(ctx, baseURIKey, baseURI)
}

// BaseURIFromContext extracts the base URI from context, returns empty string if not present
func BaseURIFromContext(ctx context.Context) string {
	if baseURI, ok := ctx.Value(baseURIKey).(string); ok {
		return baseURI
	}
	return ""
}

// extractBaseURI extracts the base URI from a reference for context resolution
func extractBaseURI(reference string) string {
	// Handle absolute URIs
	if strings.HasPrefix(reference, "http://") || strings.HasPrefix(reference, "https://") {
		// Split on '#' to get the URI part without fragment
		parts := strings.Split(reference, "#")
		uri := parts[0]

		// Find the last '/' to get the directory path
		lastSlash := strings.LastIndex(uri, "/")
		if lastSlash != -1 {
			return uri[:lastSlash+1] // Include the trailing slash
		}
		return uri + "/" // Add trailing slash if not present
	}

	// For relative references, we can't determine base URI without context
	return ""
}

type MultiValidator struct {
	and        bool
	oneOf      bool
	validators []Interface
}

type MultiValidatorMode int

const (
	OrMode MultiValidatorMode = iota
	AndMode
	OneOfMode
	InvalidMode
)

func NewMultiValidator(mode MultiValidatorMode) *MultiValidator {
	mv := &MultiValidator{}
	if mode == AndMode {
		mv.and = true
	} else if mode == OneOfMode {
		mv.and = false
		mv.oneOf = true
	}
	return mv
}

func (v *MultiValidator) Append(in Interface) *MultiValidator {
	v.validators = append(v.validators, in)
	return v
}

// UnevaluatedPropertiesCompositionValidator handles complex unevaluatedProperties with allOf
type UnevaluatedPropertiesCompositionValidator struct {
	allOfValidators []Interface
	baseValidator   Interface
	schema          *schema.Schema
}

func NewUnevaluatedPropertiesCompositionValidator(s *schema.Schema) *UnevaluatedPropertiesCompositionValidator {
	v, err := NewUnevaluatedPropertiesCompositionValidatorWithResolver(context.Background(), s, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create composition validator: %v", err))
	}
	return v
}

func NewUnevaluatedPropertiesCompositionValidatorWithResolver(ctx context.Context, s *schema.Schema, resolver *schema.Resolver) (*UnevaluatedPropertiesCompositionValidator, error) {
	v := &UnevaluatedPropertiesCompositionValidator{
		schema: s,
	}

	// Compile allOf validators
	for _, subSchema := range s.AllOf() {
		subValidator, err := Compile(ctx, convertSchemaOrBool(subSchema))
		if err != nil {
			return nil, fmt.Errorf("failed to compile allOf validator: %w", err)
		}
		v.allOfValidators = append(v.allOfValidators, subValidator)
	}

	// Compile base validator (everything except allOf)
	baseSchema := createBaseSchema(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to compile base schema: %w", err)
	}
	v.baseValidator = baseValidator

	return v, nil
}

func (v *UnevaluatedPropertiesCompositionValidator) Validate(ctx context.Context, in any) (Result, error) {
	// First, validate all allOf subschemas and collect their annotations
	var mergedResult *ObjectResult
	for i, subValidator := range v.allOfValidators {
		result, err := subValidator.Validate(ctx, in)
		if err != nil {
			return nil, fmt.Errorf(`allOf validation failed: validator #%d failed: %w`, i, err)
		}

		// Merge object results for property evaluation tracking
		if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
			if mergedResult == nil {
				mergedResult = NewObjectResult()
			}
			for prop := range objResult.EvaluatedProperties() {
				mergedResult.SetEvaluatedProperty(prop)
			}
		}
	}

	// Now validate base constraints, passing the evaluated properties from allOf
	baseResult, err := v.validateBaseWithContext(ctx, in, mergedResult)
	if err != nil {
		return nil, err
	}

	// Merge the base result with allOf result
	if baseObjResult, ok := baseResult.(*ObjectResult); ok && baseObjResult != nil {
		if mergedResult == nil {
			mergedResult = NewObjectResult()
		}
		for prop := range baseObjResult.EvaluatedProperties() {
			mergedResult.SetEvaluatedProperty(prop)
		}
	}

	return mergedResult, nil
}

// validateBaseWithContext validates the base schema with annotation context
func (v *UnevaluatedPropertiesCompositionValidator) validateBaseWithContext(ctx context.Context, in any, previousResult *ObjectResult) (Result, error) {
	// Create context with stash if we have previous evaluation results
	var currentCtx context.Context
	if previousResult != nil {
		evalProps := previousResult.EvaluatedProperties()
		if len(evalProps) > 0 {
			stash := &Stash{EvaluatedProperties: evalProps}
			currentCtx = WithStash(ctx, stash)
		} else {
			currentCtx = ctx
		}
	} else {
		currentCtx = ctx
	}

	return v.baseValidator.Validate(currentCtx, in)
}

// AnyOfUnevaluatedPropertiesCompositionValidator handles complex unevaluatedProperties with anyOf
type AnyOfUnevaluatedPropertiesCompositionValidator struct {
	anyOfValidators []Interface
	baseValidator   Interface
	schema          *schema.Schema
}

func NewAnyOfUnevaluatedPropertiesCompositionValidator(s *schema.Schema) *AnyOfUnevaluatedPropertiesCompositionValidator {
	v, err := NewAnyOfUnevaluatedPropertiesCompositionValidatorWithResolver(context.Background(), s, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create anyOf composition validator: %v", err))
	}
	return v
}

func NewAnyOfUnevaluatedPropertiesCompositionValidatorWithResolver(ctx context.Context, s *schema.Schema, anyOfValidators []Interface, resolver *schema.Resolver) (*AnyOfUnevaluatedPropertiesCompositionValidator, error) {
	v := &AnyOfUnevaluatedPropertiesCompositionValidator{
		schema: s,
	}

	// Use provided validators or compile them if not provided
	if anyOfValidators != nil {
		v.anyOfValidators = anyOfValidators
	} else {
		// Compile anyOf validators
		for _, subSchema := range s.AnyOf() {
			subValidator, err := Compile(ctx, convertSchemaOrBool(subSchema))
			if err != nil {
				return nil, fmt.Errorf("failed to compile anyOf validator: %w", err)
			}
			v.anyOfValidators = append(v.anyOfValidators, subValidator)
		}
	}

	// Compile base validator (everything except anyOf)
	baseSchema := createBaseSchema(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to compile base schema: %w", err)
	}
	v.baseValidator = baseValidator

	return v, nil
}

func (v *AnyOfUnevaluatedPropertiesCompositionValidator) Validate(ctx context.Context, in any) (Result, error) {
	// For anyOf, we need at least one subschema to pass and collect its annotations
	var validResult *ObjectResult
	anyOfPassed := false

	for _, subValidator := range v.anyOfValidators {
		result, err := subValidator.Validate(ctx, in)
		if err == nil {
			anyOfPassed = true
			// Collect annotations from ALL passing validators (not just the first)
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				if validResult == nil {
					validResult = NewObjectResult()
				}
				for prop := range objResult.EvaluatedProperties() {
					validResult.SetEvaluatedProperty(prop)
				}
			}
			// Continue to check other validators for annotation collection
		}
	}

	if !anyOfPassed {
		return nil, fmt.Errorf(`anyOf validation failed: none of the validators passed`)
	}

	// Now validate base constraints, passing the evaluated properties from anyOf
	baseResult, err := v.validateBaseWithContext(ctx, in, validResult)
	if err != nil {
		return nil, err
	}

	// Merge the base result with anyOf result
	if baseObjResult, ok := baseResult.(*ObjectResult); ok && baseObjResult != nil {
		if validResult == nil {
			validResult = NewObjectResult()
		}
		for prop := range baseObjResult.EvaluatedProperties() {
			validResult.SetEvaluatedProperty(prop)
		}
	}

	return validResult, nil
}

// validateBaseWithContext for AnyOf
func (v *AnyOfUnevaluatedPropertiesCompositionValidator) validateBaseWithContext(ctx context.Context, in any, previousResult *ObjectResult) (Result, error) {
	if objValidator, ok := v.baseValidator.(*objectValidator); ok {
		var previouslyEvaluated map[string]bool
		if previousResult != nil {
			previouslyEvaluated = previousResult.EvaluatedProperties()
		}
		var currentCtx context.Context
		if previouslyEvaluated != nil && len(previouslyEvaluated) > 0 {
			stash := &Stash{EvaluatedProperties: previouslyEvaluated}
			currentCtx = WithStash(ctx, stash)
		} else {
			currentCtx = ctx
		}
		return objValidator.Validate(currentCtx, in)
	} else if multiValidator, ok := v.baseValidator.(*MultiValidator); ok {
		// If the base validator is a MultiValidator, we need to handle it specially
		return v.validateMultiValidatorWithContext(ctx, multiValidator, in, previousResult)
	} else {
		// For other validator types, just validate normally without annotation context
		return v.baseValidator.Validate(ctx, in)
	}
}

// validateMultiValidatorWithContext for AnyOf
func (v *AnyOfUnevaluatedPropertiesCompositionValidator) validateMultiValidatorWithContext(ctx context.Context, mv *MultiValidator, in any, previousResult *ObjectResult) (Result, error) {
	if mv.and {
		// For AND mode (allOf), validate each sub-validator independently (cousins cannot see each other)
		var mergedResult *ObjectResult
		if previousResult != nil {
			mergedResult = NewObjectResult()
			for prop := range previousResult.EvaluatedProperties() {
				mergedResult.SetEvaluatedProperty(prop)
			}
		}

		for i, subValidator := range mv.validators {
			var result Result
			var err error

			// Each cousin validator should be validated independently
			// without seeing evaluated properties from other cousins
			// Only pass the original previousResult context, not accumulated cousin results
			if objValidator, ok := subValidator.(*objectValidator); ok {
				var previouslyEvaluated map[string]bool
				if previousResult != nil {
					previouslyEvaluated = previousResult.EvaluatedProperties()
				}
				var currentCtx context.Context
				if previouslyEvaluated != nil && len(previouslyEvaluated) > 0 {
					stash := &Stash{EvaluatedProperties: previouslyEvaluated}
					currentCtx = WithStash(ctx, stash)
				} else {
					currentCtx = ctx
				}
				result, err = objValidator.Validate(currentCtx, in)
			} else {
				result, err = subValidator.Validate(ctx, in)
			}

			if err != nil {
				return nil, fmt.Errorf(`allOf validation failed: validator #%d failed: %w`, i, err)
			}

			// Merge object results
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				if mergedResult == nil {
					mergedResult = NewObjectResult()
				}
				for prop := range objResult.EvaluatedProperties() {
					mergedResult.SetEvaluatedProperty(prop)
				}
			}
		}
		return mergedResult, nil
	} else {
		// For OR mode, just validate normally
		return mv.Validate(ctx, in)
	}
}

// OneOfUnevaluatedPropertiesCompositionValidator handles complex unevaluatedProperties with oneOf
type OneOfUnevaluatedPropertiesCompositionValidator struct {
	oneOfValidators []Interface
	baseValidator   Interface
	schema          *schema.Schema
}

func NewOneOfUnevaluatedPropertiesCompositionValidator(s *schema.Schema) *OneOfUnevaluatedPropertiesCompositionValidator {
	v, err := NewOneOfUnevaluatedPropertiesCompositionValidatorWithResolver(context.Background(), s, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create oneOf composition validator: %v", err))
	}
	return v
}

func NewOneOfUnevaluatedPropertiesCompositionValidatorWithResolver(ctx context.Context, s *schema.Schema, oneOfValidators []Interface, resolver *schema.Resolver) (*OneOfUnevaluatedPropertiesCompositionValidator, error) {
	v := &OneOfUnevaluatedPropertiesCompositionValidator{
		schema: s,
	}

	// Use provided validators or compile them if not provided
	if oneOfValidators != nil {
		v.oneOfValidators = oneOfValidators
	} else {
		// Compile oneOf validators
		for _, subSchema := range s.OneOf() {
			subValidator, err := Compile(ctx, convertSchemaOrBool(subSchema))
			if err != nil {
				return nil, fmt.Errorf("failed to compile oneOf validator: %w", err)
			}
			v.oneOfValidators = append(v.oneOfValidators, subValidator)
		}
	}

	// Compile base validator (everything except oneOf)
	baseSchema := createBaseSchema(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to compile base schema: %w", err)
	}
	v.baseValidator = baseValidator

	return v, nil
}

func (v *OneOfUnevaluatedPropertiesCompositionValidator) Validate(ctx context.Context, in any) (Result, error) {
	// For oneOf, exactly one subschema must pass and we collect its annotations
	var validResult *ObjectResult
	passedCount := 0

	for _, subValidator := range v.oneOfValidators {
		result, err := subValidator.Validate(ctx, in)
		if err == nil {
			passedCount++
			// Collect annotations from the passing validator
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				validResult = NewObjectResult()
				for prop := range objResult.EvaluatedProperties() {
					validResult.SetEvaluatedProperty(prop)
				}
			}
		}
	}

	if passedCount == 0 {
		return nil, fmt.Errorf(`oneOf validation failed: none of the validators passed`)
	}
	if passedCount > 1 {
		return nil, fmt.Errorf(`oneOf validation failed: more than one validator passed (%d), expected exactly one`, passedCount)
	}

	// Now validate base constraints, passing the evaluated properties from oneOf
	baseResult, err := v.validateBaseWithContext(ctx, in, validResult)
	if err != nil {
		return nil, err
	}

	// Merge the base result with oneOf result
	if baseObjResult, ok := baseResult.(*ObjectResult); ok && baseObjResult != nil {
		if validResult == nil {
			validResult = NewObjectResult()
		}
		for prop := range baseObjResult.EvaluatedProperties() {
			validResult.SetEvaluatedProperty(prop)
		}
	}

	return validResult, nil
}

// validateBaseWithContext for OneOf
func (v *OneOfUnevaluatedPropertiesCompositionValidator) validateBaseWithContext(ctx context.Context, in any, previousResult *ObjectResult) (Result, error) {
	if objValidator, ok := v.baseValidator.(*objectValidator); ok {
		var previouslyEvaluated map[string]bool
		if previousResult != nil {
			previouslyEvaluated = previousResult.EvaluatedProperties()
		}
		var currentCtx context.Context
		if previouslyEvaluated != nil && len(previouslyEvaluated) > 0 {
			stash := &Stash{EvaluatedProperties: previouslyEvaluated}
			currentCtx = WithStash(ctx, stash)
		} else {
			currentCtx = ctx
		}
		return objValidator.Validate(currentCtx, in)
	} else if multiValidator, ok := v.baseValidator.(*MultiValidator); ok {
		// If the base validator is a MultiValidator, we need to handle it specially
		return v.validateMultiValidatorWithContext(ctx, multiValidator, in, previousResult)
	} else {
		// For other validator types, just validate normally without annotation context
		return v.baseValidator.Validate(ctx, in)
	}
}

// validateMultiValidatorWithContext for OneOf
func (v *OneOfUnevaluatedPropertiesCompositionValidator) validateMultiValidatorWithContext(ctx context.Context, mv *MultiValidator, in any, previousResult *ObjectResult) (Result, error) {
	if mv.and {
		// For AND mode (allOf), validate each sub-validator independently (cousins cannot see each other)
		var mergedResult *ObjectResult
		if previousResult != nil {
			mergedResult = NewObjectResult()
			for prop := range previousResult.EvaluatedProperties() {
				mergedResult.SetEvaluatedProperty(prop)
			}
		}

		for i, subValidator := range mv.validators {
			var result Result
			var err error

			// Each cousin validator should be validated independently
			// without seeing evaluated properties from other cousins
			// Only pass the original previousResult context, not accumulated cousin results
			if objValidator, ok := subValidator.(*objectValidator); ok {
				var previouslyEvaluated map[string]bool
				if previousResult != nil {
					previouslyEvaluated = previousResult.EvaluatedProperties()
				}
				var currentCtx context.Context
				if previouslyEvaluated != nil && len(previouslyEvaluated) > 0 {
					stash := &Stash{EvaluatedProperties: previouslyEvaluated}
					currentCtx = WithStash(ctx, stash)
				} else {
					currentCtx = ctx
				}
				result, err = objValidator.Validate(currentCtx, in)
			} else {
				result, err = subValidator.Validate(ctx, in)
			}

			if err != nil {
				return nil, fmt.Errorf(`allOf validation failed: validator #%d failed: %w`, i, err)
			}

			// Merge object results
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				if mergedResult == nil {
					mergedResult = NewObjectResult()
				}
				for prop := range objResult.EvaluatedProperties() {
					mergedResult.SetEvaluatedProperty(prop)
				}
			}
		}
		return mergedResult, nil
	} else {
		// For OR mode, just validate normally
		return mv.Validate(ctx, in)
	}
}

// RefUnevaluatedPropertiesCompositionValidator handles complex unevaluatedProperties with $ref
type RefUnevaluatedPropertiesCompositionValidator struct {
	refValidator  Interface
	baseValidator Interface
	schema        *schema.Schema
}

func NewRefUnevaluatedPropertiesCompositionValidator(ctx context.Context, s *schema.Schema, refValidator Interface) *RefUnevaluatedPropertiesCompositionValidator {
	v := &RefUnevaluatedPropertiesCompositionValidator{
		schema:       s,
		refValidator: refValidator,
	}

	// Compile base validator (everything except $ref)
	baseSchema := createSchemaWithoutRef(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		panic(fmt.Sprintf("failed to compile base schema: %v", err))
	}
	v.baseValidator = baseValidator

	return v
}

func (v *RefUnevaluatedPropertiesCompositionValidator) Validate(ctx context.Context, in any) (Result, error) {
	// First, validate the $ref and collect its annotations
	refResult, err := v.refValidator.Validate(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("$ref validation failed: %w", err)
	}

	// Now validate base constraints, passing the evaluated properties from $ref
	baseResult, err := v.validateBaseWithContext(ctx, in, refResult)
	if err != nil {
		return nil, err
	}

	// Merge the base result with $ref result
	var finalResult *ObjectResult
	if err := MergeResults(&finalResult, refResult, baseResult); err != nil {
		// Fall back to simple merging if MergeResults fails
		if objRefResult, ok := refResult.(*ObjectResult); ok {
			if objBaseResult, ok := baseResult.(*ObjectResult); ok {
				finalResult = mergeObjectResults(objRefResult, objBaseResult)
			} else {
				finalResult = objRefResult
			}
		} else if objBaseResult, ok := baseResult.(*ObjectResult); ok {
			finalResult = objBaseResult
		}
	}
	return finalResult, nil
}

// validateBaseWithContext validates the base schema with annotation context from $ref
func (v *RefUnevaluatedPropertiesCompositionValidator) validateBaseWithContext(ctx context.Context, in any, refResult Result) (Result, error) {
	// Create context with stash if we have evaluation results from $ref
	var currentCtx context.Context
	if objResult, ok := refResult.(*ObjectResult); ok && objResult != nil {
		evalProps := objResult.EvaluatedProperties()
		if len(evalProps) > 0 {
			stash := &Stash{EvaluatedProperties: evalProps}
			currentCtx = WithStash(ctx, stash)
		} else {
			currentCtx = ctx
		}
	} else {
		currentCtx = ctx
	}

	return v.baseValidator.Validate(currentCtx, in)
}

func (v *MultiValidator) Validate(ctx context.Context, in any) (Result, error) {
	if v.and {
		// For allOf, collect all results and merge them while passing context between validators
		var mergedObjectResult *ObjectResult
		var mergedArrayResult *ArrayResult

		for i, subv := range v.validators {
			// Create stash context with accumulated annotations for this validator
			var currentCtx context.Context
			stash := &Stash{}

			// Add evaluated items if we have them (items annotations flow between allOf subschemas)
			if mergedArrayResult != nil {
				evalItems := mergedArrayResult.EvaluatedItems()
				if len(evalItems) > 0 {
					stash.EvaluatedItems = evalItems
				}
			}

			// NOTE: We do NOT pass evaluated properties between allOf subschemas
			// This implements the "cousin" semantics where properties evaluated by one
			// subschema are not visible to other subschemas in the same allOf

			// Only create stash context if we have something to pass
			if len(stash.EvaluatedItems) > 0 {
				currentCtx = WithStash(ctx, stash)
			} else {
				currentCtx = ctx
			}

			result, err := subv.Validate(currentCtx, in)
			if err != nil {
				return nil, fmt.Errorf(`allOf validation failed: validator #%d failed: %w`, i, err)
			}
			// Merge object results for property evaluation tracking
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				if mergedObjectResult == nil {
					mergedObjectResult = NewObjectResult()
				}
				for prop := range objResult.EvaluatedProperties() {
					mergedObjectResult.SetEvaluatedProperty(prop)
				}
			}

			// Merge array results for item evaluation tracking
			if arrResult, ok := result.(*ArrayResult); ok && arrResult != nil {
				if mergedArrayResult == nil {
					mergedArrayResult = NewArrayResult()
				}
				arrItems := arrResult.EvaluatedItems()
				for i, evaluated := range arrItems {
					if evaluated {
						mergedArrayResult.SetEvaluatedItem(i)
					}
				}
			}
		}

		// Return appropriate result type based on what we merged
		if mergedObjectResult != nil && mergedArrayResult != nil {
			// Both object and array results - this shouldn't happen in normal validation
			// but prioritize object result for now
			return mergedObjectResult, nil
		} else if mergedObjectResult != nil {
			return mergedObjectResult, nil
		} else if mergedArrayResult != nil {
			return mergedArrayResult, nil
		}

		return nil, nil
	}

	if v.oneOf {
		passedCount := 0
		var validResult Result
		for _, subv := range v.validators {
			result, err := subv.Validate(ctx, in)
			if err == nil {
				passedCount++
				validResult = result
			}
		}
		if passedCount == 0 {
			return nil, fmt.Errorf(`oneOf validation failed: none of the validators passed`)
		}
		if passedCount > 1 {
			return nil, fmt.Errorf(`oneOf validation failed: more than one validator passed (%d), expected exactly one`, passedCount)
		}
		return validResult, nil
	}

	// This is for anyOf (OrMode)
	for _, subv := range v.validators {
		result, err := subv.Validate(ctx, in)
		if err == nil {
			return result, nil
		}
	}
	return nil, fmt.Errorf(`anyOf validation failed: none of the validators passed`)
}

// hasBaseConstraints checks if a schema has base-level constraints that need validation
// when used with allOf/anyOf/oneOf
func hasBaseConstraints(s *schema.Schema) bool {
	return len(s.Types()) > 0 ||
		s.HasMinLength() || s.HasMaxLength() || s.HasPattern() ||
		s.HasMinimum() || s.HasMaximum() || s.HasExclusiveMinimum() || s.HasExclusiveMaximum() || s.HasMultipleOf() ||
		s.HasMinItems() || s.HasMaxItems() || s.HasUniqueItems() || s.HasItems() || s.HasContains() ||
		s.HasMinProperties() || s.HasMaxProperties() || s.HasRequired() || s.HasProperties() || s.HasPatternProperties() || s.HasAdditionalProperties() || s.HasUnevaluatedProperties() || s.HasDependentSchemas() || s.HasPropertyNames() ||
		s.HasEnum() || s.HasConst()
}

// createBaseSchema creates a new schema with only the base constraints (no composition keywords).
// This function excludes ALL composition and control flow keywords:
//   - allOf, anyOf, oneOf (composition keywords)
//   - not (negation keyword)
//   - if/then/else (conditional keywords)
//   - $ref, $dynamicRef (reference keywords)
//
// Only basic validation constraints are copied (types, string/number/array/object constraints, enum/const).
func createBaseSchema(s *schema.Schema) *schema.Schema {
	builder := schema.NewBuilder()

	// Copy types
	if len(s.Types()) > 0 {
		builder.Types(s.Types()...)
	}

	// Copy string constraints
	if s.HasMinLength() {
		builder.MinLength(s.MinLength())
	}
	if s.HasMaxLength() {
		builder.MaxLength(s.MaxLength())
	}
	if s.HasPattern() {
		builder.Pattern(s.Pattern())
	}

	// Copy number constraints
	if s.HasMinimum() {
		builder.Minimum(s.Minimum())
	}
	if s.HasMaximum() {
		builder.Maximum(s.Maximum())
	}
	if s.HasExclusiveMinimum() {
		builder.ExclusiveMinimum(s.ExclusiveMinimum())
	}
	if s.HasExclusiveMaximum() {
		builder.ExclusiveMaximum(s.ExclusiveMaximum())
	}
	if s.HasMultipleOf() {
		builder.MultipleOf(s.MultipleOf())
	}

	// Copy array constraints
	if s.HasMinItems() {
		builder.MinItems(s.MinItems())
	}
	if s.HasMaxItems() {
		builder.MaxItems(s.MaxItems())
	}
	if s.HasUniqueItems() {
		builder.UniqueItems(s.UniqueItems())
	}
	if s.HasItems() {
		builder.Items(s.Items())
	}
	if s.HasContains() {
		builder.Contains(s.Contains())
	}

	// Copy object constraints
	if s.HasMinProperties() {
		builder.MinProperties(s.MinProperties())
	}
	if s.HasMaxProperties() {
		builder.MaxProperties(s.MaxProperties())
	}
	if s.HasRequired() {
		for _, req := range s.Required() {
			builder.Required(req)
		}
	}
	if s.HasProperties() {
		for name, prop := range s.Properties() {
			builder.Property(name, prop)
		}
	}
	if s.HasPatternProperties() {
		for pattern, prop := range s.PatternProperties() {
			builder.PatternProperty(pattern, prop)
		}
	}
	if s.HasAdditionalProperties() {
		builder.AdditionalProperties(s.AdditionalProperties())
	}
	if s.HasUnevaluatedProperties() {
		builder.UnevaluatedProperties(s.UnevaluatedProperties())
	}
	if s.HasDependentSchemas() {
		builder.DependentSchemas(s.DependentSchemas())
	}
	if s.HasPropertyNames() {
		builder.PropertyNames(s.PropertyNames())
	}

	// Copy enum/const
	if s.HasEnum() {
		builder.Enum(s.Enum()...)
	}
	if s.HasConst() {
		builder.Const(s.Const())
	}

	return builder.MustBuild()
}

// IfThenElseValidator handles if/then/else conditional validation
type IfThenElseValidator struct {
	ifValidator   Interface
	thenValidator Interface
	elseValidator Interface
}

func compileIfThenElseValidator(ctx context.Context, s *schema.Schema) (Interface, error) {
	v := &IfThenElseValidator{}

	// Compile 'if' validator (required)
	ifValidator, err := Compile(ctx, s.IfSchema())
	if err != nil {
		return nil, fmt.Errorf(`failed to compile if validator: %w`, err)
	}
	v.ifValidator = ifValidator

	// Compile 'then' validator (optional)
	if s.HasThenSchema() {
		thenValidator, err := Compile(ctx, s.ThenSchema())
		if err != nil {
			return nil, fmt.Errorf(`failed to compile then validator: %w`, err)
		}
		v.thenValidator = thenValidator
	}

	// Compile 'else' validator (optional)
	if s.HasElseSchema() {
		elseValidator, err := Compile(ctx, s.ElseSchema())
		if err != nil {
			return nil, fmt.Errorf(`failed to compile else validator: %w`, err)
		}
		v.elseValidator = elseValidator
	}

	return v, nil
}

func (v *IfThenElseValidator) Validate(ctx context.Context, in any) (Result, error) {
	// First, check the 'if' condition and collect its annotations
	ifResult, ifErr := v.ifValidator.Validate(ctx, in)

	// The 'if' schema contributes annotations regardless of whether it passes or fails
	var conditionalResult Result

	if ifErr == nil {
		// 'if' condition passed, validate against 'then' if it exists
		if v.thenValidator != nil {
			thenResult, err := v.thenValidator.Validate(ctx, in)
			if err != nil {
				return nil, err
			}
			// Merge 'if' and 'then' results
			var mergedResult *ObjectResult
			if err := MergeResults(&mergedResult, ifResult, thenResult); err != nil {
				// Fall back to simple merging
				if objIfResult, ok := ifResult.(*ObjectResult); ok {
					if objThenResult, ok := thenResult.(*ObjectResult); ok {
						mergedResult = mergeObjectResults(objIfResult, objThenResult)
					} else {
						mergedResult = objIfResult
					}
				} else if objThenResult, ok := thenResult.(*ObjectResult); ok {
					mergedResult = objThenResult
				}
			}
			conditionalResult = mergedResult
		} else {
			// Only 'if' result
			conditionalResult = ifResult
		}
	} else {
		// 'if' condition failed, validate against 'else' if it exists
		if v.elseValidator != nil {
			elseResult, err := v.elseValidator.Validate(ctx, in)
			if err != nil {
				return nil, err
			}
			// Merge 'if' and 'else' results
			var mergedResult *ObjectResult
			if err := MergeResults(&mergedResult, ifResult, elseResult); err != nil {
				// Fall back to simple merging
				if objIfResult, ok := ifResult.(*ObjectResult); ok {
					if objElseResult, ok := elseResult.(*ObjectResult); ok {
						mergedResult = mergeObjectResults(objIfResult, objElseResult)
					} else {
						mergedResult = objIfResult
					}
				} else if objElseResult, ok := elseResult.(*ObjectResult); ok {
					mergedResult = objElseResult
				}
			}
			conditionalResult = mergedResult
		} else {
			// Only 'if' result (even though it failed validation, it may have annotations)
			conditionalResult = ifResult
		}
	}

	return conditionalResult, nil
}

// IfThenElseUnevaluatedPropertiesCompositionValidator handles complex unevaluatedProperties with if/then/else
type IfThenElseUnevaluatedPropertiesCompositionValidator struct {
	ifValidator   Interface
	thenValidator Interface
	elseValidator Interface
	baseValidator Interface
	schema        *schema.Schema
}

func NewIfThenElseUnevaluatedPropertiesCompositionValidator(ctx context.Context, s *schema.Schema) *IfThenElseUnevaluatedPropertiesCompositionValidator {
	v := &IfThenElseUnevaluatedPropertiesCompositionValidator{
		schema: s,
	}

	// Compile if validator
	ifValidator, err := Compile(ctx, s.IfSchema())
	if err != nil {
		panic(fmt.Sprintf("failed to compile if validator: %v", err))
	}
	v.ifValidator = ifValidator

	// Compile then validator if it exists
	if s.HasThenSchema() {
		thenValidator, err := Compile(ctx, s.ThenSchema())
		if err != nil {
			panic(fmt.Sprintf("failed to compile then validator: %v", err))
		}
		v.thenValidator = thenValidator
	}

	// Compile else validator if it exists
	if s.HasElseSchema() {
		elseValidator, err := Compile(ctx, s.ElseSchema())
		if err != nil {
			panic(fmt.Sprintf("failed to compile else validator: %v", err))
		}
		v.elseValidator = elseValidator
	}

	// Compile base validator (everything except if/then/else)
	baseSchema := createBaseSchema(s)
	baseValidator, err := Compile(ctx, baseSchema)
	if err != nil {
		panic(fmt.Sprintf("failed to compile base schema: %v", err))
	}
	v.baseValidator = baseValidator

	return v
}

func (v *IfThenElseUnevaluatedPropertiesCompositionValidator) Validate(ctx context.Context, in any) (Result, error) {
	// First, evaluate if/then/else and collect annotations
	var conditionalResult *ObjectResult

	// Check the 'if' condition and collect its annotations
	ifResult, ifErr := v.ifValidator.Validate(ctx, in)

	// Collect annotations from 'if' schema (contributes regardless of outcome)
	if ifObjResult, ok := ifResult.(*ObjectResult); ok && ifObjResult != nil {
		conditionalResult = NewObjectResult()
		for prop := range ifObjResult.EvaluatedProperties() {
			conditionalResult.SetEvaluatedProperty(prop)
		}
	}

	if ifErr == nil {
		// 'if' condition passed, validate against 'then' if it exists
		if v.thenValidator != nil {
			result, err := v.thenValidator.Validate(ctx, in)
			if err != nil {
				return nil, fmt.Errorf(`if/then validation failed: %w`, err)
			}
			// Merge annotations from 'then' with 'if' annotations
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				if conditionalResult == nil {
					conditionalResult = NewObjectResult()
				}
				for prop := range objResult.EvaluatedProperties() {
					conditionalResult.SetEvaluatedProperty(prop)
				}
			}
		}
	} else {
		// 'if' condition failed, validate against 'else' if it exists
		if v.elseValidator != nil {
			result, err := v.elseValidator.Validate(ctx, in)
			if err != nil {
				return nil, fmt.Errorf(`if/else validation failed: %w`, err)
			}
			// Merge annotations from 'else' with 'if' annotations
			if objResult, ok := result.(*ObjectResult); ok && objResult != nil {
				if conditionalResult == nil {
					conditionalResult = NewObjectResult()
				}
				for prop := range objResult.EvaluatedProperties() {
					conditionalResult.SetEvaluatedProperty(prop)
				}
			}
		}
	}

	// Now validate base constraints, passing the evaluated properties from if/then/else
	baseResult, err := v.validateBaseWithContext(ctx, in, conditionalResult)
	if err != nil {
		return nil, err
	}

	// Merge the base result with if/then/else result
	if baseObjResult, ok := baseResult.(*ObjectResult); ok && baseObjResult != nil {
		if conditionalResult == nil {
			conditionalResult = NewObjectResult()
		}
		for prop := range baseObjResult.EvaluatedProperties() {
			conditionalResult.SetEvaluatedProperty(prop)
		}
	}

	return conditionalResult, nil
}

// validateBaseWithContext for if/then/else
func (v *IfThenElseUnevaluatedPropertiesCompositionValidator) validateBaseWithContext(ctx context.Context, in any, previousResult *ObjectResult) (Result, error) {
	// Create context with stash if we have previous evaluation results
	var currentCtx context.Context
	if previousResult != nil {
		evalProps := previousResult.EvaluatedProperties()
		if len(evalProps) > 0 {
			stash := &Stash{EvaluatedProperties: evalProps}
			currentCtx = WithStash(ctx, stash)
		} else {
			currentCtx = ctx
		}
	} else {
		currentCtx = ctx
	}

	return v.baseValidator.Validate(currentCtx, in)
}
