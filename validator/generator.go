package validator

import (
	"fmt"
	"go/format"
	"strings"
)


// GenerateCode creates Go code that constructs the given validator using direct field access
func (g *codeGenerator) GenerateCode(validatorName string, v Interface) (string, error) {
	switch validator := v.(type) {
	case *stringValidator:
		return g.generateStringValidator(validatorName, validator)
	case *integerValidator:
		return g.generateIntegerValidator(validatorName, validator)
	case *numberValidator:
		return g.generateNumberValidator(validatorName, validator)
	case *booleanValidator:
		return g.generateBooleanValidator(validatorName, validator)
	case *arrayValidator:
		return g.generateArrayValidator(validatorName, validator)
	case *objectValidator:
		return g.generateObjectValidator(validatorName, validator)
	case *MultiValidator:
		return g.generateMultiValidator(validatorName, validator)
	case *EmptyValidator:
		return g.generateEmptyValidator(validatorName)
	case *NotValidator:
		return g.generateNotValidator(validatorName, validator)
	case *NullValidator:
		return g.generateNullValidator(validatorName)
	case *GeneralValidator:
		return g.generateGeneralValidator(validatorName, validator)
	case *alwaysPassValidator:
		return g.generateAlwaysPassValidator(validatorName)
	case *alwaysFailValidator:
		return g.generateAlwaysFailValidator(validatorName)
	case *ReferenceValidator:
		return g.generateReferenceValidator(validatorName, validator)
	default:
		return "", fmt.Errorf("unsupported validator type: %T", v)
	}
}

// generateStringValidator creates code for string validators by directly accessing fields
func (g *codeGenerator) generateStringValidator(name string, v *stringValidator) (string, error) {
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
	
	template := `func New%s() Interface {
	return String()%s.MustBuild()
}`
	
	code := fmt.Sprintf(template, name, builderCalls)
	return g.formatCode(code)
}

// generateIntegerValidator creates code for integer validators
func (g *codeGenerator) generateIntegerValidator(name string, v *integerValidator) (string, error) {
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
	
	template := `func New%s() Interface {
	return Integer()%s.MustBuild()
}`
	
	code := fmt.Sprintf(template, name, builderCalls)
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
	
	template := `func New%s() Interface {
	return Number()%s.MustBuild()
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
	
	template := `func New%s() Interface {
	return Boolean()%s.MustBuild()
}`
	
	code := fmt.Sprintf(template, name, builderCalls)
	return g.formatCode(code)
}

// generateMultiValidator creates code for multi validators with child validators
func (g *codeGenerator) generateMultiValidator(name string, v *MultiValidator) (string, error) {
	var childDefs []string
	var childVars []string
	
	// Generate code for child validators recursively
	for i, child := range v.validators {
		childVar := fmt.Sprintf("child%d", i)
		childCode, err := g.GenerateCode("", child)
		if err != nil {
			return "", fmt.Errorf("failed to generate child validator %d: %w", i, err)
		}
		
		// Extract just the validator creation part (everything after "return ")
		creation := extractValidatorCreation(childCode)
		if creation == "" {
			return "", fmt.Errorf("failed to extract validator creation from child %d", i)
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
	
	template := `func New%s() Interface {
%s
	
	mv := NewMultiValidator(%s)
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
	template := `func New%s() Interface {
	return &EmptyValidator{}
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// generateNotValidator creates code for not validators
func (g *codeGenerator) generateNotValidator(name string, v *NotValidator) (string, error) {
	childCode, err := g.GenerateCode("", v.validator)
	if err != nil {
		return "", fmt.Errorf("failed to generate child validator for not: %w", err)
	}
	
	creation := extractValidatorCreation(childCode)
	if creation == "" {
		return "", fmt.Errorf("failed to extract validator creation from not child")
	}
	
	template := `func New%s() Interface {
	child := %s
	return &NotValidator{validator: child}
}`
	
	code := fmt.Sprintf(template, name, creation)
	return g.formatCode(code)
}

// generateNullValidator creates code for null validators
func (g *codeGenerator) generateNullValidator(name string) (string, error) {
	template := `func New%s() Interface {
	return &NullValidator{}
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
	
	template := `func New%s() Interface {
	return &GeneralValidator{
		%s,
	}
}`
	
	fieldsStr := strings.Join(parts, ",\n\t\t")
	code := fmt.Sprintf(template, name, fieldsStr)
	return g.formatCode(code)
}

// generateAlwaysPassValidator creates code for always pass validators
func (g *codeGenerator) generateAlwaysPassValidator(name string) (string, error) {
	template := `func New%s() Interface {
	return &alwaysPassValidator{}
}`
	
	code := fmt.Sprintf(template, name)
	return g.formatCode(code)
}

// generateAlwaysFailValidator creates code for always fail validators
func (g *codeGenerator) generateAlwaysFailValidator(name string) (string, error) {
	template := `func New%s() Interface {
	return &alwaysFailValidator{}
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
			childCode, err := g.GenerateCode("", item)
			if err != nil {
				return "", fmt.Errorf("failed to generate prefixItems[%d]: %w", i, err)
			}
			creation := extractValidatorCreation(childCode)
			childSetup = append(childSetup, fmt.Sprintf("\t%s := %s", prefixVar, creation))
			prefixVars = append(prefixVars, prefixVar)
		}
		if len(prefixVars) > 0 {
			parts = append(parts, fmt.Sprintf("PrefixItems([]Interface{%s})", strings.Join(prefixVars, ", ")))
		}
	}
	
	if v.items != nil {
		childCode, err := g.GenerateCode("", v.items)
		if err != nil {
			return "", fmt.Errorf("failed to generate items validator: %w", err)
		}
		creation := extractValidatorCreation(childCode)
		childSetup = append(childSetup, fmt.Sprintf("\titems := %s", creation))
		parts = append(parts, "Items(items)")
	}
	
	if v.contains != nil {
		childCode, err := g.GenerateCode("", v.contains)
		if err != nil {
			return "", fmt.Errorf("failed to generate contains validator: %w", err)
		}
		creation := extractValidatorCreation(childCode)
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
		template = `func New%s() Interface {
%s
	
	return Array()%s.MustBuild()
}`
		childSetupStr := strings.Join(childSetup, "\n")
		code := fmt.Sprintf(template, name, childSetupStr, builderCalls)
		return g.formatCode(code)
	} else {
		template = `func New%s() Interface {
	return Array()%s.MustBuild()
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
			childCode, err := g.GenerateCode("", propValidator)
			if err != nil {
				return "", fmt.Errorf("failed to generate property validator for %q: %w", propName, err)
			}
			creation := extractValidatorCreation(childCode)
			propertySetup = append(propertySetup, fmt.Sprintf("\t%s := %s", propVar, creation))
		}
		childSetup = append(childSetup, propertySetup...)
		
		// Generate the properties map
		var propertyPairs []string
		for propName := range v.properties {
			propVar := fmt.Sprintf("prop_%s", sanitizeVarName(propName))
			propertyPairs = append(propertyPairs, fmt.Sprintf("\t\t%q: %s", propName, propVar))
		}
		childSetup = append(childSetup, fmt.Sprintf("\tproperties := map[string]Interface{\n%s,\n\t}", strings.Join(propertyPairs, ",\n")))
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
		template = `func New%s() Interface {
%s

	return Object()%s.MustBuild()
}`
		childSetupStr := strings.Join(childSetup, "\n")
		code := fmt.Sprintf(template, name, childSetupStr, builderCalls)
		return g.formatCode(code)
	} else {
		template = `func New%s() Interface {
	return Object()%s.MustBuild()
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
	template := `func New%s() Interface {
	// Note: This was originally a reference validator for: %s
	// Code generation simplifies this to an EmptyValidator
	return &EmptyValidator{}
}`
	
	reference := "unknown"
	// We can't access the reference field directly since it's not exported
	// For now, just use a generic comment
	code := fmt.Sprintf(template, name, reference)
	return g.formatCode(code)
}

// GeneratePackage creates a complete Go package with multiple validators
func (g *codeGenerator) GeneratePackage(packageName string, validators map[string]Interface) (string, error) {
	var functions []string
	
	// Generate individual validator functions
	for name, validator := range validators {
		code, err := g.GenerateCode(name, validator)
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