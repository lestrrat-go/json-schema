package validator

import (
	"fmt"
	"go/format"
	"io"
	"strings"
)


// Generate writes Go code that constructs the given validator to the provided Writer
// The output is just the builder chain, e.g.: validator.String().MinLength(5).MaxLength(100)
func (g *codeGenerator) Generate(dst io.Writer, v Interface) error {
	builderChain, err := g.generateBuilderChain(v)
	if err != nil {
		return err
	}
	_, err = dst.Write([]byte(builderChain))
	return err
}

// generateBuilderChain creates just the builder chain without function wrapper
func (g *codeGenerator) generateBuilderChain(v Interface) (string, error) {
	return g.generateBuilderChainInternal(v)
}

// generateCompleteValidator creates a complete function wrapper around a validator (for package generation)
func (g *codeGenerator) generateCompleteValidator(name string, v Interface) (string, error) {
	builderChain, err := g.generateBuilderChain(v)
	if err != nil {
		return "", err
	}
	
	template := `func New%s() validator.Interface {
	return %s
}`
	
	code := fmt.Sprintf(template, name, builderChain)
	return g.formatCode(code)
}

func (g *codeGenerator) generateBuilderChainInternal(v Interface) (string, error) {
	switch validator := v.(type) {
	case *stringValidator:
		return g.generateStringBuilderChain(validator)
	case *integerValidator:
		return g.generateIntegerBuilderChain(validator)
	case *numberValidator:
		return g.generateNumberBuilderChain(validator)
	case *booleanValidator:
		return g.generateBooleanBuilderChain(validator)
	case *arrayValidator:
		return g.generateArrayBuilderChain(validator)
	case *objectValidator:
		return g.generateObjectBuilderChain(validator)
	case *MultiValidator:
		return g.generateMultiBuilderChain(validator)
	case *EmptyValidator:
		return g.generateEmptyBuilderChain()
	case *NotValidator:
		return g.generateNotBuilderChain(validator)
	case *NullValidator:
		return g.generateNullBuilderChain()
	case *GeneralValidator:
		return g.generateGeneralBuilderChain(validator)
	case *alwaysPassValidator:
		return g.generateAlwaysPassBuilderChain()
	case *alwaysFailValidator:
		return g.generateAlwaysFailBuilderChain()
	case *ReferenceValidator:
		return g.generateReferenceBuilderChain(validator)
	case *DynamicReferenceValidator:
		return g.generateDynamicReferenceBuilderChain(validator)
	case *contentValidator:
		return g.generateContentBuilderChain(validator)
	case *dependentSchemasValidator:
		return g.generateDependentSchemasBuilderChain(validator)
	case *inferredNumberValidator:
		return g.generateInferredNumberBuilderChain(validator)
	case *UnevaluatedPropertiesCompositionValidator:
		return g.generateUnevaluatedPropertiesCompositionBuilderChain(validator)
	case *AnyOfUnevaluatedPropertiesCompositionValidator:
		return g.generateAnyOfUnevaluatedPropertiesCompositionBuilderChain(validator)
	case *OneOfUnevaluatedPropertiesCompositionValidator:
		return g.generateOneOfUnevaluatedPropertiesCompositionBuilderChain(validator)
	case *RefUnevaluatedPropertiesCompositionValidator:
		return g.generateRefUnevaluatedPropertiesCompositionBuilderChain(validator)
	case *IfThenElseValidator:
		return g.generateIfThenElseBuilderChain(validator)
	case *IfThenElseUnevaluatedPropertiesCompositionValidator:
		return g.generateIfThenElseUnevaluatedPropertiesCompositionBuilderChain(validator)
	default:
		// Debug: Print what unsupported validator type we encountered
		fmt.Printf("GENERATOR DEBUG: Unsupported validator type: %T, falling back to EmptyValidator\n", v)
		return "&validator.EmptyValidator{}", nil
	}
}


// generateNumberBuilderChain creates just the builder chain for number validators
func (g *codeGenerator) generateNumberBuilderChain(v *numberValidator) (string, error) {
	var parts []string
	
	if v.multipleOf != nil {
		parts = append(parts, fmt.Sprintf("MultipleOf(%g)", *v.multipleOf))
	}
	if v.minimum != nil {
		parts = append(parts, fmt.Sprintf("Minimum(%g)", *v.minimum))
	}
	if v.maximum != nil {
		parts = append(parts, fmt.Sprintf("Maximum(%g)", *v.maximum))
	}
	if v.exclusiveMinimum != nil {
		parts = append(parts, fmt.Sprintf("ExclusiveMinimum(%g)", *v.exclusiveMinimum))
	}
	if v.exclusiveMaximum != nil {
		parts = append(parts, fmt.Sprintf("ExclusiveMaximum(%g)", *v.exclusiveMaximum))
	}
	if v.enum != nil {
		enumStr := formatFloat64Slice(v.enum)
		parts = append(parts, fmt.Sprintf("Enum(%s)", enumStr))
	}
	if v.constantValue != nil {
		parts = append(parts, fmt.Sprintf("Const(%g)", *v.constantValue))
	}
	
	builderCalls := strings.Join(parts, ".")
	if builderCalls != "" {
		builderCalls = "." + builderCalls
	}
	
	return fmt.Sprintf("validator.Number()%s.MustBuild()", builderCalls), nil
}

// generateBooleanBuilderChain creates just the builder chain for boolean validators
func (g *codeGenerator) generateBooleanBuilderChain(v *booleanValidator) (string, error) {
	var parts []string
	
	if v.enum != nil {
		enumStr := formatBoolSlice(v.enum)
		parts = append(parts, fmt.Sprintf("Enum(%s)", enumStr))
	}
	if v.constantValue != nil {
		parts = append(parts, fmt.Sprintf("Const(%t)", *v.constantValue))
	}
	
	builderCalls := strings.Join(parts, ".")
	if builderCalls != "" {
		builderCalls = "." + builderCalls
	}
	
	return fmt.Sprintf("validator.Boolean()%s.MustBuild()", builderCalls), nil
}

// generateArrayBuilderChain creates just the builder chain for array validators
func (g *codeGenerator) generateArrayBuilderChain(v *arrayValidator) (string, error) {
	var parts []string
	
	// Basic array constraints
	if v.minItems != nil {
		parts = append(parts, fmt.Sprintf("MinItems(%d)", *v.minItems))
	}
	if v.maxItems != nil {
		parts = append(parts, fmt.Sprintf("MaxItems(%d)", *v.maxItems))
	}
	if v.uniqueItems {
		parts = append(parts, "UniqueItems(true)")
	}
	if v.minContains != nil {
		parts = append(parts, fmt.Sprintf("MinContains(%d)", *v.minContains))
	}
	if v.maxContains != nil {
		parts = append(parts, fmt.Sprintf("MaxContains(%d)", *v.maxContains))
	}
	
	// For complex items, prefixItems etc, we'll need more complex generation
	hasComplexItems := v.items != nil || v.prefixItems != nil || v.contains != nil
	
	if hasComplexItems {
		// For now, create a basic array validator without complex items
		// TODO: Add support for items/prefixItems generation
		builderCalls := strings.Join(parts, ".")
		if builderCalls != "" {
			builderCalls = "." + builderCalls
		}
		return fmt.Sprintf("validator.Array()%s.MustBuild()", builderCalls), nil
	}
	
	// Simple case - just constraints
	builderCalls := strings.Join(parts, ".")
	if builderCalls != "" {
		builderCalls = "." + builderCalls
	}
	
	return fmt.Sprintf("validator.Array()%s.MustBuild()", builderCalls), nil
}

// generateObjectBuilderChain creates just the builder chain for object validators
func (g *codeGenerator) generateObjectBuilderChain(v *objectValidator) (string, error) {
	var parts []string
	
	// Basic object constraints
	if v.minProperties != nil {
		parts = append(parts, fmt.Sprintf("MinProperties(%d)", *v.minProperties))
	}
	if v.maxProperties != nil {
		parts = append(parts, fmt.Sprintf("MaxProperties(%d)", *v.maxProperties))
	}
	if v.required != nil && len(v.required) > 0 {
		requiredStr := formatStringSlice(v.required)
		parts = append(parts, fmt.Sprintf("Required(%s)", requiredStr))
	}
	
	// Handle complex properties
	if v.properties != nil && len(v.properties) > 0 {
		// Generate properties map
		propertiesCode, err := g.generatePropertiesMap(v.properties)
		if err != nil {
			return "", fmt.Errorf("failed to generate properties map: %w", err)
		}
		parts = append(parts, fmt.Sprintf("Properties(%s)", propertiesCode))
	}
	
	// Handle additional properties
	if v.additionalProperties != nil {
		switch ap := v.additionalProperties.(type) {
		case bool:
			parts = append(parts, fmt.Sprintf("AdditionalPropertiesBool(%t)", ap))
		case Interface:
			apCode, err := g.generateBuilderChain(ap)
			if err != nil {
				return "", fmt.Errorf("failed to generate additional properties validator: %w", err)
			}
			parts = append(parts, fmt.Sprintf("AdditionalPropertiesSchema(%s)", apCode))
		}
	}
	
	// Handle pattern properties
	if v.patternProperties != nil && len(v.patternProperties) > 0 {
		// For pattern properties, we need to generate a more complex structure
		// For now, we'll skip this as it's complex to generate regexp objects
		fmt.Printf("GENERATOR DEBUG: Skipping pattern properties generation (not implemented)\n")
	}
	
	// Handle property names validator
	if v.propertyNames != nil {
		pnCode, err := g.generateBuilderChain(v.propertyNames)
		if err != nil {
			return "", fmt.Errorf("failed to generate property names validator: %w", err)
		}
		parts = append(parts, fmt.Sprintf("PropertyNames(%s)", pnCode))
	}
	
	// Simple case - just constraints
	builderCalls := strings.Join(parts, ".")
	if builderCalls != "" {
		builderCalls = "." + builderCalls
	}
	
	return fmt.Sprintf("validator.Object()%s.MustBuild()", builderCalls), nil
}

// generatePropertiesMap generates Go code for a properties map
func (g *codeGenerator) generatePropertiesMap(properties map[string]Interface) (string, error) {
	if len(properties) == 0 {
		return "nil", nil
	}
	
	var mapEntries []string
	
	for propName, propValidator := range properties {
		// Generate the validator code for this property
		propCode, err := g.generateBuilderChain(propValidator)
		if err != nil {
			return "", fmt.Errorf("failed to generate validator for property %s: %w", propName, err)
		}
		
		// Escape the property name as a Go string literal
		escapedName := fmt.Sprintf("%q", propName)
		mapEntries = append(mapEntries, fmt.Sprintf("\t%s: %s", escapedName, propCode))
	}
	
	// Generate the complete map literal
	mapCode := fmt.Sprintf("map[string]validator.Interface{\n%s,\n}", strings.Join(mapEntries, ",\n"))
	return mapCode, nil
}

// generateMultiBuilderChain creates just the builder chain for multi validators
func (g *codeGenerator) generateMultiBuilderChain(v *MultiValidator) (string, error) {
	var parts []string
	
	// Generate each child validator
	for i, child := range v.validators {
		childChain, err := g.generateBuilderChain(child)
		if err != nil {
			return "", fmt.Errorf("failed to generate child validator %d: %w", i, err)
		}
		parts = append(parts, childChain)
	}
	
	// Determine the mode
	var mode string
	if v.and {
		mode = "AndMode"
	} else if v.oneOf {
		mode = "OneOfMode"
	} else {
		mode = "OrMode"
	}
	
	// If we have child validators, create a proper MultiValidator
	if len(parts) > 0 {
		// For complex nested cases, we'll need to create a function that builds this
		// For now, let's create a simple multi-validator structure
		return fmt.Sprintf(`func() validator.Interface {
			mv := validator.NewMultiValidator(validator.%s)
			%s
			return mv
		}()`, mode, generateAppendCalls(parts)), nil
	}
	
	return "&validator.EmptyValidator{}", nil
}

// generateAppendCalls creates the mv.Append() calls for child validators
func generateAppendCalls(parts []string) string {
	var calls []string
	for _, part := range parts {
		calls = append(calls, fmt.Sprintf("mv.Append(%s)", part))
	}
	return strings.Join(calls, "\n\t\t\t")
}

// generateEmptyBuilderChain creates just the builder chain for empty validators
func (g *codeGenerator) generateEmptyBuilderChain() (string, error) {
	return "&validator.EmptyValidator{}", nil
}

// generateNotBuilderChain creates just the builder chain for not validators
func (g *codeGenerator) generateNotBuilderChain(v *NotValidator) (string, error) {
	childChain, err := g.generateBuilderChain(v.validator)
	if err != nil {
		return "", fmt.Errorf("failed to generate child validator for not: %w", err)
	}
	
	return fmt.Sprintf("func() validator.Interface { child := %s; return &validator.NotValidator{validator: child} }()", childChain), nil
}

// generateNullBuilderChain creates just the builder chain for null validators
func (g *codeGenerator) generateNullBuilderChain() (string, error) {
	return "&validator.NullValidator{}", nil
}

// generateGeneralBuilderChain creates just the builder chain for general validators
func (g *codeGenerator) generateGeneralBuilderChain(v *GeneralValidator) (string, error) {
	// GeneralValidator with enum is typically used for "type" property validation
	// which can be either a single string or array of strings
	// Since we can't access unexported fields, create a MultiValidator with proper constraints
	
	if v.enum != nil {
		// For enum validation, create a MultiValidator that accepts either:
		// 1. String with enum values, or 2. Array of unique enum values
		enumStrs := make([]string, len(v.enum))
		for i, e := range v.enum {
			if s, ok := e.(string); ok {
				enumStrs[i] = fmt.Sprintf("%q", s)
			} else {
				enumStrs[i] = fmt.Sprintf("%#v", e)
			}
		}
		enumSlice := fmt.Sprintf("[]string{%s}", strings.Join(enumStrs, ", "))
		
		return fmt.Sprintf(`func() validator.Interface {
			mv := validator.NewMultiValidator(validator.OrMode)
			mv.Append(validator.String().Enum(%s).MustBuild())
			mv.Append(validator.Array().MinItems(1).UniqueItems(true).MustBuild())
			return mv
		}()`, enumSlice), nil
	}
	
	if v.hasConst {
		// For const validation, accept any matching value
		// Since we can't access the const value, return EmptyValidator
		return "&validator.EmptyValidator{}", nil
	}
	
	return "&validator.EmptyValidator{}", nil
}

// generateAlwaysPassBuilderChain creates just the builder chain for always pass validators
func (g *codeGenerator) generateAlwaysPassBuilderChain() (string, error) {
	return "&validator.EmptyValidator{}", nil
}

// generateAlwaysFailBuilderChain creates just the builder chain for always fail validators
func (g *codeGenerator) generateAlwaysFailBuilderChain() (string, error) {
	return "&validator.NotValidator{validator: &validator.EmptyValidator{}}", nil
}

// generateStringBuilderChain creates just the builder chain for string validators
func (g *codeGenerator) generateStringBuilderChain(v *stringValidator) (string, error) {
	var parts []string
	
	// Directly access validator fields - no intermediate map needed!
	if v.minLength != nil {
		parts = append(parts, fmt.Sprintf("MinLength(%d)", *v.minLength))
	}
	if v.maxLength != nil {
		parts = append(parts, fmt.Sprintf("MaxLength(%d)", *v.maxLength))
	}
	if v.pattern != nil {
		parts = append(parts, fmt.Sprintf("Pattern(%q)", v.pattern.String()))
	}
	if v.format != nil {
		parts = append(parts, fmt.Sprintf("Format(%q)", *v.format))
	}
	if v.enum != nil {
		enumStr := formatStringSlice(v.enum)
		parts = append(parts, fmt.Sprintf("Enum(%s)", enumStr))
	}
	if v.constantValue != nil {
		parts = append(parts, fmt.Sprintf("Const(%q)", *v.constantValue))
	}
	
	builderCalls := strings.Join(parts, ".")
	if builderCalls != "" {
		builderCalls = "." + builderCalls
	}
	
	return fmt.Sprintf("validator.String()%s.MustBuild()", builderCalls), nil
}

// generateStringValidator creates code for string validators by directly accessing fields
func (g *codeGenerator) generateStringValidator(name string, v *stringValidator) (string, error) {
	builderChain, err := g.generateStringBuilderChain(v)
	if err != nil {
		return "", err
	}
	
	template := `func New%s() validator.Interface {
	return %s
}`
	
	code := fmt.Sprintf(template, name, builderChain)
	return g.formatCode(code)
}

// generateIntegerBuilderChain creates just the builder chain for integer validators
func (g *codeGenerator) generateIntegerBuilderChain(v *integerValidator) (string, error) {
	var parts []string
	
	if v.multipleOf != nil {
		parts = append(parts, fmt.Sprintf("MultipleOf(%d)", *v.multipleOf))
	}
	if v.minimum != nil {
		parts = append(parts, fmt.Sprintf("Minimum(%d)", *v.minimum))
	}
	if v.maximum != nil {
		parts = append(parts, fmt.Sprintf("Maximum(%d)", *v.maximum))
	}
	if v.exclusiveMinimum != nil {
		parts = append(parts, fmt.Sprintf("ExclusiveMinimum(%d)", *v.exclusiveMinimum))
	}
	if v.exclusiveMaximum != nil {
		parts = append(parts, fmt.Sprintf("ExclusiveMaximum(%d)", *v.exclusiveMaximum))
	}
	if v.enum != nil {
		enumStr := formatIntSlice(v.enum)
		parts = append(parts, fmt.Sprintf("Enum(%s)", enumStr))
	}
	if v.constantValue != nil {
		parts = append(parts, fmt.Sprintf("Const(%d)", *v.constantValue))
	}
	
	builderCalls := strings.Join(parts, ".")
	if builderCalls != "" {
		builderCalls = "." + builderCalls
	}
	
	return fmt.Sprintf("validator.Integer()%s.MustBuild()", builderCalls), nil
}

// generateIntegerValidator creates code for integer validators
func (g *codeGenerator) generateIntegerValidator(name string, v *integerValidator) (string, error) {
	builderChain, err := g.generateIntegerBuilderChain(v)
	if err != nil {
		return "", err
	}
	
	template := `func New%s() validator.Interface {
	return %s
}`
	
	code := fmt.Sprintf(template, name, builderChain)
	return g.formatCode(code)
}

// generateNumberValidator creates code for number validators
func (g *codeGenerator) generateNumberValidator(name string, v *numberValidator) (string, error) {
	var parts []string
	
	if v.multipleOf != nil {
		parts = append(parts, fmt.Sprintf("MultipleOf(%g)", *v.multipleOf))
	}
	if v.minimum != nil {
		parts = append(parts, fmt.Sprintf("Minimum(%g)", *v.minimum))
	}
	if v.maximum != nil {
		parts = append(parts, fmt.Sprintf("Maximum(%g)", *v.maximum))
	}
	if v.exclusiveMinimum != nil {
		parts = append(parts, fmt.Sprintf("ExclusiveMinimum(%g)", *v.exclusiveMinimum))
	}
	if v.exclusiveMaximum != nil {
		parts = append(parts, fmt.Sprintf("ExclusiveMaximum(%g)", *v.exclusiveMaximum))
	}
	if v.enum != nil {
		enumStr := formatFloat64Slice(v.enum)
		parts = append(parts, fmt.Sprintf("Enum(%s)", enumStr))
	}
	if v.constantValue != nil {
		parts = append(parts, fmt.Sprintf("Const(%g)", *v.constantValue))
	}
	
	builderCalls := strings.Join(parts, ".")
	if builderCalls != "" {
		builderCalls = "." + builderCalls
	}
	
	template := `func New%s() validator.Interface {
	return validator.Number()%s.MustBuild()
}`
	
	code := fmt.Sprintf(template, name, builderCalls)
	return g.formatCode(code)
}

// generateBooleanValidator creates code for boolean validators
func (g *codeGenerator) generateBooleanValidator(name string, v *booleanValidator) (string, error) {
	var parts []string
	
	if v.enum != nil {
		enumStr := formatBoolSlice(v.enum)
		parts = append(parts, fmt.Sprintf("Enum(%s)", enumStr))
	}
	if v.constantValue != nil {
		parts = append(parts, fmt.Sprintf("Const(%t)", *v.constantValue))
	}
	
	builderCalls := strings.Join(parts, ".")
	if builderCalls != "" {
		builderCalls = "." + builderCalls
	}
	
	template := `func New%s() validator.Interface {
	return validator.Boolean()%s.MustBuild()
}`
	
	code := fmt.Sprintf(template, name, builderCalls)
	return g.formatCode(code)
}

// generateMultiValidator creates code for multi validators with child validators
func (g *codeGenerator) generateMultiValidator(name string, v *MultiValidator) (string, error) {
	var childDefs []string
	var childVars []string
	
	// Check for recursive/circular references by detecting self-referential children
	hasCircularRef := false
	for _, child := range v.validators {
		if child == v {
			hasCircularRef = true
			break
		}
	}
	
	// Handle circular references with a different approach
	if hasCircularRef {
		var mode string
		if v.and {
			mode = "AndMode"
		} else if v.oneOf {
			mode = "OneOfMode" 
		} else {
			mode = "OrMode"
		}
		
		template := `func New%s() validator.Interface {
	// Note: This validator contains circular references, simplified to EmptyValidator
	// Original was a MultiValidator with %s mode
	return &validator.EmptyValidator{}
}`
		
		code := fmt.Sprintf(template, name, mode)
		return g.formatCode(code)
	}
	
	// Generate code for child validators recursively (non-circular case)
	for i, child := range v.validators {
		childVar := fmt.Sprintf("child%d", i)
		creation, err := g.generateBuilderChain(child)
		if err != nil {
			return "", fmt.Errorf("failed to generate child validator %d: %w", i, err)
		}
		if creation == "" {
			return "", fmt.Errorf("failed to extract validator creation from child %d", i)
		}
		
		// Check if we got a circular reference (undefined variable like "mv")
		if strings.Contains(creation, "mv") && !strings.Contains(creation, "validator.") {
			// This is a circular reference, fallback to EmptyValidator
			var mode string
			if v.and {
				mode = "AndMode"
			} else if v.oneOf {
				mode = "OneOfMode" 
			} else {
				mode = "OrMode"
			}
			
			template := `func New%s() validator.Interface {
	// Note: This validator had circular references, simplified to EmptyValidator
	// Original was a MultiValidator with %s mode
	return &validator.EmptyValidator{}
}`
			
			code := fmt.Sprintf(template, name, mode)
			return g.formatCode(code)
		}
		
		childDefs = append(childDefs, fmt.Sprintf("\t%s := %s", childVar, creation))
		childVars = append(childVars, childVar)
	}
	
	var mode string
	if v.and {
		mode = "AndMode"
	} else if v.oneOf {
		mode = "OneOfMode" 
	} else {
		mode = "OrMode"
	}
	
	template := `func New%s() validator.Interface {
%s
	
	mv := validator.NewMultiValidator(validator.%s)
%s
	return mv
}`
	
	childDefsStr := strings.Join(childDefs, "\n")
	appendCalls := func() []string {
		var calls []string
		for _, childVar := range childVars {
			calls = append(calls, fmt.Sprintf("\tmv.Append(%s)", childVar))
		}
		return calls
	}()
	appendCallsStr := strings.Join(appendCalls, "\n")
	
	code := fmt.Sprintf(template, name, childDefsStr, mode, appendCallsStr)
	return g.formatCode(code)
}

// generateEmptyValidator creates code for empty validators (allow everything)
func (g *codeGenerator) generateEmptyValidator(name string) (string, error) {
	template := `func New%s() validator.Interface {
	return &validator.EmptyValidator{}
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// generateNotValidator creates code for not validators
func (g *codeGenerator) generateNotValidator(name string, v *NotValidator) (string, error) {
	creation, err := g.generateBuilderChain(v.validator)
	if err != nil {
		return "", fmt.Errorf("failed to generate child validator for not: %w", err)
	}
	if creation == "" {
		return "", fmt.Errorf("failed to extract validator creation from not child")
	}
	
	template := `func New%s() validator.Interface {
	child := %s
	return &validator.NotValidator{validator: child}
}`
	
	code := fmt.Sprintf(template, name, creation)
	return g.formatCode(code)
}

// generateNullValidator creates code for null validators
func (g *codeGenerator) generateNullValidator(name string) (string, error) {
	template := `func New%s() validator.Interface {
	return &validator.NullValidator{}
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// generateGeneralValidator creates code for general validators (enum/const)
func (g *codeGenerator) generateGeneralValidator(name string, v *GeneralValidator) (string, error) {
	var parts []string
	
	if v.hasConst {
		parts = append(parts, fmt.Sprintf("constValue: %#v", v.constValue))
		parts = append(parts, "hasConst: true")
	}
	if v.enum != nil {
		parts = append(parts, fmt.Sprintf("enum: %#v", v.enum))
	}
	
	template := `func New%s() validator.Interface {
	return &validator.GeneralValidator{
		%s,
	}
}`
	
	fieldsStr := strings.Join(parts, ",\n\t\t")
	code := fmt.Sprintf(template, name, fieldsStr)
	return g.formatCode(code)
}

// generateAlwaysPassValidator creates code for always pass validators
func (g *codeGenerator) generateAlwaysPassValidator(name string) (string, error) {
	template := `func New%s() validator.Interface {
	return &validator.EmptyValidator{} // alwaysPassValidator is not exported, use EmptyValidator instead
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// generateAlwaysFailValidator creates code for always fail validators
func (g *codeGenerator) generateAlwaysFailValidator(name string) (string, error) {
	template := `func New%s() validator.Interface {
	// alwaysFailValidator is not exported, create equivalent using not(empty)
	return &validator.NotValidator{validator: &validator.EmptyValidator{}}
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// generateArrayValidator creates code for array validators (complex - handles children)
func (g *codeGenerator) generateArrayValidator(name string, v *arrayValidator) (string, error) {
	var parts []string
	var childSetup []string
	
	// Basic array constraints
	if v.minItems != nil {
		parts = append(parts, fmt.Sprintf("MinItems(%d)", *v.minItems))
	}
	if v.maxItems != nil {
		parts = append(parts, fmt.Sprintf("MaxItems(%d)", *v.maxItems))
	}
	if v.uniqueItems {
		parts = append(parts, "UniqueItems(true)")
	}
	if v.minContains != nil {
		parts = append(parts, fmt.Sprintf("MinContains(%d)", *v.minContains))
	}
	if v.maxContains != nil {
		parts = append(parts, fmt.Sprintf("MaxContains(%d)", *v.maxContains))
	}
	
	// Handle complex child validators
	if v.prefixItems != nil {
		var prefixVars []string
		for i, item := range v.prefixItems {
			prefixVar := fmt.Sprintf("prefixItem%d", i)
			creation, err := g.generateBuilderChain(item)
			if err != nil {
				return "", fmt.Errorf("failed to generate prefixItems[%d]: %w", i, err)
			}
			childSetup = append(childSetup, fmt.Sprintf("\t%s := %s", prefixVar, creation))
			prefixVars = append(prefixVars, prefixVar)
		}
		if len(prefixVars) > 0 {
			parts = append(parts, fmt.Sprintf("PrefixItems([]validator.Interface{%s})", strings.Join(prefixVars, ", ")))
		}
	}
	
	if v.items != nil {
		creation, err := g.generateBuilderChain(v.items)
		if err != nil {
			return "", fmt.Errorf("failed to generate items validator: %w", err)
		}
		childSetup = append(childSetup, fmt.Sprintf("\titems := %s", creation))
		parts = append(parts, "Items(items)")
	}
	
	if v.contains != nil {
		creation, err := g.generateBuilderChain(v.contains)
		if err != nil {
			return "", fmt.Errorf("failed to generate contains validator: %w", err)
		}
		childSetup = append(childSetup, fmt.Sprintf("\tcontains := %s", creation))
		parts = append(parts, "Contains(contains)")
	}
	
	// Build the function
	builderCalls := strings.Join(parts, ".")
	if builderCalls != "" {
		builderCalls = "." + builderCalls
	}
	
	var template string
	if len(childSetup) > 0 {
		template = `func New%s() validator.Interface {
%s
	
	return validator.Array()%s.MustBuild()
}`
		childSetupStr := strings.Join(childSetup, "\n")
		code := fmt.Sprintf(template, name, childSetupStr, builderCalls)
		return g.formatCode(code)
	} else {
		template = `func New%s() validator.Interface {
	return validator.Array()%s.MustBuild()
}`
		code := fmt.Sprintf(template, name, builderCalls)
		return g.formatCode(code)
	}
}

// generateObjectValidator creates code for object validators
func (g *codeGenerator) generateObjectValidator(name string, v *objectValidator) (string, error) {
	var parts []string
	var childSetup []string

	// Basic object constraints
	if v.minProperties != nil {
		parts = append(parts, fmt.Sprintf("MinProperties(%d)", *v.minProperties))
	}
	if v.maxProperties != nil {
		parts = append(parts, fmt.Sprintf("MaxProperties(%d)", *v.maxProperties))
	}
	if v.required != nil && len(v.required) > 0 {
		requiredStr := formatStringSlice(v.required)
		parts = append(parts, fmt.Sprintf("Required(%s)", requiredStr))
	}

	// Handle complex child validators
	if v.properties != nil && len(v.properties) > 0 {
		var propertySetup []string
		for propName, propValidator := range v.properties {
			propVar := fmt.Sprintf("prop_%s", sanitizeVarName(propName))
			creation, err := g.generateBuilderChain(propValidator)
			if err != nil {
				return "", fmt.Errorf("failed to generate property validator for %q: %w", propName, err)
			}
			propertySetup = append(propertySetup, fmt.Sprintf("\t%s := %s", propVar, creation))
		}
		childSetup = append(childSetup, propertySetup...)
		
		// Generate the properties map
		var propertyPairs []string
		for propName := range v.properties {
			propVar := fmt.Sprintf("prop_%s", sanitizeVarName(propName))
			propertyPairs = append(propertyPairs, fmt.Sprintf("\t\t%q: %s", propName, propVar))
		}
		childSetup = append(childSetup, fmt.Sprintf("\tproperties := map[string]validator.Interface{\n%s,\n\t}", strings.Join(propertyPairs, ",\n")))
		parts = append(parts, "Properties(properties)")
	}

	// Handle additional properties (simplified - just check if it's a boolean)
	if v.additionalProperties != nil {
		if b, ok := v.additionalProperties.(bool); ok {
			parts = append(parts, fmt.Sprintf("AdditionalProperties(%t)", b))
		}
		// TODO: Handle additionalProperties as validator
	}

	// Build the function
	builderCalls := strings.Join(parts, ".")
	if builderCalls != "" {
		builderCalls = "." + builderCalls
	}

	var template string
	if len(childSetup) > 0 {
		template = `func New%s() validator.Interface {
%s

	return validator.Object()%s.MustBuild()
}`
		childSetupStr := strings.Join(childSetup, "\n")
		code := fmt.Sprintf(template, name, childSetupStr, builderCalls)
		return g.formatCode(code)
	} else {
		template = `func New%s() validator.Interface {
	return validator.Object()%s.MustBuild()
}`
		code := fmt.Sprintf(template, name, builderCalls)
		return g.formatCode(code)
	}
}

// sanitizeVarName sanitizes a property name to be a valid Go variable name
func sanitizeVarName(name string) string {
	// Simple sanitization - replace non-alphanumeric with underscore
	result := ""
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			result += string(char)
		} else {
			result += "_"
		}
	}
	if result == "" {
		result = "prop"
	}
	return result
}

// generateReferenceValidator creates code for reference validators
// For simplicity in code generation, we create an EmptyValidator as a placeholder
// In a real implementation, proper reference resolution would be needed
func (g *codeGenerator) generateReferenceValidator(name string, v *ReferenceValidator) (string, error) {
	// For meta-schema generation, we'll create a comment explaining this is a reference
	// and fall back to EmptyValidator to allow everything through
	template := `func New%s() validator.Interface {
	// Note: This was originally a reference validator for: %s
	// Code generation simplifies this to an EmptyValidator
	return &validator.EmptyValidator{}
}`
	
	reference := "unknown"
	// We can't access the reference field directly since it's not exported
	// For now, just use a generic comment
	code := fmt.Sprintf(template, name, reference)
	return g.formatCode(code)
}

// generateDynamicReferenceValidator creates code for dynamic reference validators
// This handles the circular/recursive reference problem by using a lazy initialization pattern
func (g *codeGenerator) generateDynamicReferenceValidator(name string, v *DynamicReferenceValidator) (string, error) {
	// Since DynamicReferenceValidator fields are not exported, we can't generate 
	// the actual validator. For meta-schema purposes, we'll use a simpler approach.
	// The real solution would require refactoring to export necessary fields or
	// provide a builder pattern for DynamicReferenceValidator
	
	// For now, return an EmptyValidator as a safe fallback
	template := `func New%s() validator.Interface {
	// Note: This was originally a dynamic reference validator (circular reference)
	// Code generation uses EmptyValidator as fallback for recursive references
	return &validator.EmptyValidator{}
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// GeneratePackage creates a complete Go package with multiple validators
func (g *codeGenerator) GeneratePackage(packageName string, validators map[string]Interface) (string, error) {
	var functions []string
	
	// Generate individual validator functions
	for name, validator := range validators {
		// For package generation, we need the old-style function wrapper
		// So we'll continue to use the internal method that creates complete functions
		code, err := g.generateCompleteValidator(name, validator)
		if err != nil {
			return "", fmt.Errorf("failed to generate validator %s: %w", name, err)
		}
		functions = append(functions, code)
	}
	
	// Build complete package
	imports := []string{
		`"github.com/lestrrat-go/json-schema/validator"`,
	}
	imports = append(imports, g.config.packageImports...)
	
	var importsSection string
	if len(imports) == 1 {
		importsSection = fmt.Sprintf("import %s", imports[0])
	} else {
		importsSection = fmt.Sprintf("import (\n\t%s\n)", strings.Join(imports, "\n\t"))
	}
	
	template := `package %s

%s

%s`
	
	functionsStr := strings.Join(functions, "\n\n")
	code := fmt.Sprintf(template, packageName, importsSection, functionsStr)
	
	return g.formatCode(code)
}

// Helper functions

// formatCode formats the generated Go code
func (g *codeGenerator) formatCode(code string) (string, error) {
	formatted, err := format.Source([]byte(code))
	if err != nil {
		// Return unformatted code if formatting fails
		return code, nil
	}
	return string(formatted), nil
}

// formatStringSlice formats a string slice for Go code
func formatStringSlice(strs []string) string {
	quoted := make([]string, len(strs))
	for i, s := range strs {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("[]string{%s}", strings.Join(quoted, ", "))
}

// formatIntSlice formats an int slice for Go code
func formatIntSlice(ints []int) string {
	strs := make([]string, len(ints))
	for i, n := range ints {
		strs[i] = fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("[]int{%s}", strings.Join(strs, ", "))
}

// formatFloat64Slice formats a float64 slice for Go code
func formatFloat64Slice(floats []float64) string {
	strs := make([]string, len(floats))
	for i, f := range floats {
		strs[i] = fmt.Sprintf("%g", f)
	}
	return fmt.Sprintf("[]float64{%s}", strings.Join(strs, ", "))
}

// formatBoolSlice formats a bool slice for Go code
func formatBoolSlice(bools []bool) string {
	strs := make([]string, len(bools))
	for i, b := range bools {
		strs[i] = fmt.Sprintf("%t", b)
	}
	return fmt.Sprintf("[]bool{%s}", strings.Join(strs, ", "))
}

// generateContentValidator creates code for content validators
func (g *codeGenerator) generateContentValidator(name string, v *contentValidator) (string, error) {
	// Content validators are complex - for now return a simplified EmptyValidator
	template := `func New%s() validator.Interface {
	// Note: This was originally a content validator, simplified to EmptyValidator
	return &validator.EmptyValidator{}
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// generateDependentSchemasValidator creates code for dependent schemas validators
func (g *codeGenerator) generateDependentSchemasValidator(name string, v *dependentSchemasValidator) (string, error) {
	// Dependent schemas validators are complex - for now return a simplified EmptyValidator
	template := `func New%s() validator.Interface {
	// Note: This was originally a dependent schemas validator, simplified to EmptyValidator
	return &validator.EmptyValidator{}
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// generateInferredNumberValidator creates code for inferred number validators
func (g *codeGenerator) generateInferredNumberValidator(name string, v *inferredNumberValidator) (string, error) {
	// Inferred number validators can be treated like number validators
	template := `func New%s() validator.Interface {
	// Note: This was originally an inferred number validator
	return validator.Number().MustBuild()
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// generateUnevaluatedPropertiesCompositionValidator creates code for unevaluated properties composition validators
func (g *codeGenerator) generateUnevaluatedPropertiesCompositionValidator(name string, v *UnevaluatedPropertiesCompositionValidator) (string, error) {
	// Unevaluated properties validators are complex - for now return a simplified EmptyValidator
	template := `func New%s() validator.Interface {
	// Note: This was originally an unevaluated properties composition validator, simplified to EmptyValidator
	return &validator.EmptyValidator{}
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// generateAnyOfUnevaluatedPropertiesCompositionValidator creates code for anyOf unevaluated properties composition validators
func (g *codeGenerator) generateAnyOfUnevaluatedPropertiesCompositionValidator(name string, v *AnyOfUnevaluatedPropertiesCompositionValidator) (string, error) {
	// AnyOf unevaluated properties validators are complex - for now return a simplified EmptyValidator
	template := `func New%s() validator.Interface {
	// Note: This was originally an anyOf unevaluated properties composition validator, simplified to EmptyValidator
	return &validator.EmptyValidator{}
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// generateOneOfUnevaluatedPropertiesCompositionValidator creates code for oneOf unevaluated properties composition validators
func (g *codeGenerator) generateOneOfUnevaluatedPropertiesCompositionValidator(name string, v *OneOfUnevaluatedPropertiesCompositionValidator) (string, error) {
	// OneOf unevaluated properties validators are complex - for now return a simplified EmptyValidator
	template := `func New%s() validator.Interface {
	// Note: This was originally a oneOf unevaluated properties composition validator, simplified to EmptyValidator
	return &validator.EmptyValidator{}
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// generateRefUnevaluatedPropertiesCompositionValidator creates code for ref unevaluated properties composition validators
func (g *codeGenerator) generateRefUnevaluatedPropertiesCompositionValidator(name string, v *RefUnevaluatedPropertiesCompositionValidator) (string, error) {
	// Ref unevaluated properties validators are complex - for now return a simplified EmptyValidator
	template := `func New%s() validator.Interface {
	// Note: This was originally a ref unevaluated properties composition validator, simplified to EmptyValidator
	return &validator.EmptyValidator{}
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// generateIfThenElseValidator creates code for if-then-else validators
func (g *codeGenerator) generateIfThenElseValidator(name string, v *IfThenElseValidator) (string, error) {
	// If-then-else validators are complex - for now return a simplified EmptyValidator
	template := `func New%s() validator.Interface {
	// Note: This was originally an if-then-else validator, simplified to EmptyValidator
	return &validator.EmptyValidator{}
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// generateIfThenElseUnevaluatedPropertiesCompositionValidator creates code for if-then-else unevaluated properties composition validators
func (g *codeGenerator) generateIfThenElseUnevaluatedPropertiesCompositionValidator(name string, v *IfThenElseUnevaluatedPropertiesCompositionValidator) (string, error) {
	// If-then-else unevaluated properties validators are complex - for now return a simplified EmptyValidator
	template := `func New%s() validator.Interface {
	// Note: This was originally an if-then-else unevaluated properties composition validator, simplified to EmptyValidator
	return &validator.EmptyValidator{}
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// extractValidatorCreation extracts the validator creation part from generated code
func extractValidatorCreation(code string) string {
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "return ") {
			return strings.TrimPrefix(line, "return ")
		}
	}
	return ""
}

// generateReferenceBuilderChain creates just the builder chain for reference validators
func (g *codeGenerator) generateReferenceBuilderChain(v *ReferenceValidator) (string, error) {
	// If the reference has been resolved, generate the resolved validator
	if v.resolved != nil {
		// Generate the resolved validator, but watch out for circular references
		if v.resolved == v {
			// Self-reference - create EmptyValidator to avoid infinite recursion
			return "&validator.EmptyValidator{}", nil
		}
		
		resolvedChain, err := g.generateBuilderChain(v.resolved)
		if err != nil {
			// If we can't generate the resolved validator, fall back to EmptyValidator
			return "&validator.EmptyValidator{}", nil
		}
		return resolvedChain, nil
	}
	
	// If not resolved, we can't generate proper code, fall back to EmptyValidator
	// This happens when references can't be resolved at compile time
	return "&validator.EmptyValidator{}", nil
}

// generateDynamicReferenceBuilderChain creates just the builder chain for dynamic reference validators
func (g *codeGenerator) generateDynamicReferenceBuilderChain(v *DynamicReferenceValidator) (string, error) {
	// If the dynamic reference has been resolved, generate the resolved validator
	if v.resolved != nil {
		// Generate the resolved validator, but watch out for circular references
		if v.resolved == v {
			// Self-reference - create EmptyValidator to avoid infinite recursion
			return "&validator.EmptyValidator{}", nil
		}
		
		resolvedChain, err := g.generateBuilderChain(v.resolved)
		if err != nil {
			// If we can't generate the resolved validator, fall back to EmptyValidator
			return "&validator.EmptyValidator{}", nil
		}
		return resolvedChain, nil
	}
	
	// If not resolved, create a reasonable validator based on JSON Schema expectations
	// Most unresolved dynamic references in meta-schemas expect either:
	// 1. JSON Schema objects (which should be objects)
	// 2. Any JSON value (for things like const/default)
	// For meta-schema generation, assume these accept schema objects (type: object)
	// This is better than EmptyValidator which accepts anything
	return "validator.Object().MustBuild()", nil
}

// generateContentBuilderChain creates just the builder chain for content validators
func (g *codeGenerator) generateContentBuilderChain(v *contentValidator) (string, error) {
	// Content validators are complex but we can create a basic structure
	var parts []string
	
	if v.contentEncoding != "" {
		parts = append(parts, fmt.Sprintf("contentEncoding: %q", v.contentEncoding))
	}
	if v.contentMediaType != "" {
		parts = append(parts, fmt.Sprintf("contentMediaType: %q", v.contentMediaType))
	}
	
	// Handle contentSchema if present
	var contentSchemaStr string
	if v.contentSchema != nil {
		childChain, err := g.generateBuilderChain(v.contentSchema)
		if err != nil {
			return "", fmt.Errorf("failed to generate content schema: %w", err)
		}
		contentSchemaStr = fmt.Sprintf("contentSchema: %s", childChain)
		parts = append(parts, contentSchemaStr)
	}
	
	if len(parts) > 0 {
		fieldsStr := strings.Join(parts, ",\n\t\t")
		return fmt.Sprintf("&validator.contentValidator{\n\t\t%s,\n\t}", fieldsStr), nil
	}
	
	return "&validator.contentValidator{}", nil
}

// generateDependentSchemasBuilderChain creates just the builder chain for dependent schemas validators
func (g *codeGenerator) generateDependentSchemasBuilderChain(v *dependentSchemasValidator) (string, error) {
	// Generate the dependent schemas map
	if len(v.dependentSchemas) == 0 {
		return "&validator.dependentSchemasValidator{dependentSchemas: make(map[string]validator.Interface)}", nil
	}
	
	var mapEntries []string
	var childSetup []string
	
	for propName, propValidator := range v.dependentSchemas {
		childVar := fmt.Sprintf("dep_%s", sanitizeVarName(propName))
		childChain, err := g.generateBuilderChain(propValidator)
		if err != nil {
			return "", fmt.Errorf("failed to generate dependent schema for %q: %w", propName, err)
		}
		
		childSetup = append(childSetup, fmt.Sprintf("\t%s := %s", childVar, childChain))
		mapEntries = append(mapEntries, fmt.Sprintf("\t\t%q: %s", propName, childVar))
	}
	
	if len(childSetup) > 0 {
		setupStr := strings.Join(childSetup, "\n")
		entriesStr := strings.Join(mapEntries, ",\n")
		
		return fmt.Sprintf(`func() validator.Interface {
%s
	return &validator.dependentSchemasValidator{
		dependentSchemas: map[string]validator.Interface{
%s,
		},
	}
}()`, setupStr, entriesStr), nil
	}
	
	return "&validator.dependentSchemasValidator{dependentSchemas: make(map[string]validator.Interface)}", nil
}

// generateInferredNumberBuilderChain creates just the builder chain for inferred number validators
func (g *codeGenerator) generateInferredNumberBuilderChain(v *inferredNumberValidator) (string, error) {
	// Generate the underlying number validator
	childChain, err := g.generateBuilderChain(v.numberValidator)
	if err != nil {
		return "", fmt.Errorf("failed to generate inferred number validator: %w", err)
	}
	
	// For now, just return the child validator since inferredNumberValidator is a wrapper
	// TODO: Create proper inferredNumberValidator constructor if needed
	return childChain, nil
}

// generateUnevaluatedPropertiesCompositionBuilderChain creates just the builder chain for unevaluated properties composition validators
func (g *codeGenerator) generateUnevaluatedPropertiesCompositionBuilderChain(v *UnevaluatedPropertiesCompositionValidator) (string, error) {
	// Unevaluated properties validators are very complex
	// For now, create a basic validator that accepts everything
	// TODO: Implement proper unevaluated properties generation
	return "&validator.EmptyValidator{}", nil
}

// generateAnyOfUnevaluatedPropertiesCompositionBuilderChain creates just the builder chain for anyOf unevaluated properties composition validators
func (g *codeGenerator) generateAnyOfUnevaluatedPropertiesCompositionBuilderChain(v *AnyOfUnevaluatedPropertiesCompositionValidator) (string, error) {
	// AnyOf unevaluated properties validators are very complex
	// For now, create a basic validator that accepts everything
	// TODO: Implement proper anyOf unevaluated properties generation
	return "&validator.EmptyValidator{}", nil
}

// generateOneOfUnevaluatedPropertiesCompositionBuilderChain creates just the builder chain for oneOf unevaluated properties composition validators
func (g *codeGenerator) generateOneOfUnevaluatedPropertiesCompositionBuilderChain(v *OneOfUnevaluatedPropertiesCompositionValidator) (string, error) {
	// OneOf unevaluated properties validators are very complex
	// For now, create a basic validator that accepts everything
	// TODO: Implement proper oneOf unevaluated properties generation
	return "&validator.EmptyValidator{}", nil
}

// generateRefUnevaluatedPropertiesCompositionBuilderChain creates just the builder chain for ref unevaluated properties composition validators
func (g *codeGenerator) generateRefUnevaluatedPropertiesCompositionBuilderChain(v *RefUnevaluatedPropertiesCompositionValidator) (string, error) {
	// Ref unevaluated properties validators are very complex
	// For now, create a basic validator that accepts everything
	// TODO: Implement proper ref unevaluated properties generation
	return "&validator.EmptyValidator{}", nil
}

// generateIfThenElseBuilderChain creates just the builder chain for if-then-else validators
func (g *codeGenerator) generateIfThenElseBuilderChain(v *IfThenElseValidator) (string, error) {
	// If-then-else validators are complex conditional validators
	// For now, create a basic validator that accepts everything
	// TODO: Implement proper if-then-else generation
	return "&validator.EmptyValidator{}", nil
}

// generateIfThenElseUnevaluatedPropertiesCompositionBuilderChain creates just the builder chain for if-then-else unevaluated properties composition validators
func (g *codeGenerator) generateIfThenElseUnevaluatedPropertiesCompositionBuilderChain(v *IfThenElseUnevaluatedPropertiesCompositionValidator) (string, error) {
	// If-then-else unevaluated properties validators are very complex
	// For now, create a basic validator that accepts everything
	// TODO: Implement proper if-then-else unevaluated properties generation
	return "&validator.EmptyValidator{}", nil
}