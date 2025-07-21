package validator

import (
	"fmt"
	"go/format"
	"io"
	"strings"

	"github.com/lestrrat-go/codegen"
	"github.com/lestrrat-go/json-schema/keywords"
)

// Ensure keywords package is imported (referenced in string maps)
var _ = keywords.Type

// keywordConstantMap maps JSON Schema keywords to their keywords package constant names
var keywordConstantMap = map[string]string{
	"$id":                    "keywords.ID",
	"$schema":                "keywords.Schema", 
	"$anchor":                "keywords.Anchor",
	"$dynamicAnchor":         "keywords.DynamicAnchor",
	"$dynamicRef":            "keywords.DynamicReference",
	"$ref":                   "keywords.Reference",
	"$comment":               "keywords.Comment",
	"$defs":                  "keywords.Definitions",
	"$vocabulary":            "keywords.Vocabulary",
	"type":                   "keywords.Type",
	"enum":                   "keywords.Enum",
	"const":                  "keywords.Const",
	"multipleOf":             "keywords.MultipleOf",
	"maximum":                "keywords.Maximum",
	"exclusiveMaximum":       "keywords.ExclusiveMaximum",
	"minimum":                "keywords.Minimum",
	"exclusiveMinimum":       "keywords.ExclusiveMinimum",
	"maxLength":              "keywords.MaxLength",
	"minLength":              "keywords.MinLength",
	"pattern":                "keywords.Pattern",
	"additionalItems":        "keywords.AdditionalItems",
	"items":                  "keywords.Items",
	"maxItems":               "keywords.MaxItems",
	"minItems":               "keywords.MinItems",
	"uniqueItems":            "keywords.UniqueItems",
	"contains":               "keywords.Contains",
	"maxContains":            "keywords.MaxContains",
	"minContains":            "keywords.MinContains",
	"maxProperties":          "keywords.MaxProperties",
	"minProperties":          "keywords.MinProperties", 
	"required":               "keywords.Required",
	"additionalProperties":   "keywords.AdditionalProperties",
	"definitions":            "keywords.Definitions",
	"properties":             "keywords.Properties",
	"patternProperties":      "keywords.PatternProperties",
	"dependencies":           "keywords.DependentSchemas", // Note: "dependencies" maps to DependentSchemas in 2020-12
	"dependentSchemas":       "keywords.DependentSchemas",
	"dependentRequired":      "keywords.DependentRequired",
	"propertyNames":          "keywords.PropertyNames",
	"allOf":                  "keywords.AllOf",
	"anyOf":                  "keywords.AnyOf",
	"oneOf":                  "keywords.OneOf",
	"not":                    "keywords.Not",
	"if":                     "keywords.If",
	"then":                   "keywords.Then",
	"else":                   "keywords.Else",
	"format":                 "keywords.Format",
	"contentEncoding":        "keywords.ContentEncoding",
	"contentMediaType":       "keywords.ContentMediaType",
	"contentSchema":          "keywords.ContentSchema",
	"title":                  "keywords.Title",
	"description":            "keywords.Description",
	"default":                "keywords.Default",
	"deprecated":             "keywords.Deprecated",
	"readOnly":               "keywords.ReadOnly",
	"writeOnly":              "keywords.WriteOnly",
	"examples":               "keywords.Examples",
	"prefixItems":            "keywords.PrefixItems",
	"unevaluatedItems":       "keywords.UnevaluatedItems",
	"unevaluatedProperties":  "keywords.UnevaluatedProperties",
	// Legacy keywords for backward compatibility
	"$recursiveRef":          "keywords.RecursiveRef",    // Deprecated in 2020-12 but still in some schemas
	"$recursiveAnchor":       "keywords.RecursiveAnchor", // Deprecated in 2020-12 but still in some schemas
}

// getKeywordConstant returns the keywords package constant reference for a JSON Schema keyword,
// or returns the quoted string if it's not a standard keyword
func getKeywordConstant(propName string) string {
	if constant, exists := keywordConstantMap[propName]; exists {
		return constant
	}
	
	// For non-standard keywords, return the quoted string
	return fmt.Sprintf("%q", propName)
}

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
		// For multiline arguments, format differently
		if len(v.enum) > 1 {
			parts = append(parts, fmt.Sprintf("Enum(\n\t%s\n)", enumStr))
		} else {
			parts = append(parts, fmt.Sprintf("Enum(%s)", enumStr))
		}
	}
	if v.constantValue != nil {
		parts = append(parts, fmt.Sprintf("Const(%g)", *v.constantValue))
	}
	
	return buildMethodChain("validator.Number()", parts), nil
}

// generateBooleanBuilderChain creates just the builder chain for boolean validators
func (g *codeGenerator) generateBooleanBuilderChain(v *booleanValidator) (string, error) {
	var parts []string
	
	if v.enum != nil {
		enumStr := formatBoolSlice(v.enum)
		// For multiline arguments, format differently
		if len(v.enum) > 1 {
			parts = append(parts, fmt.Sprintf("Enum(\n\t%s\n)", enumStr))
		} else {
			parts = append(parts, fmt.Sprintf("Enum(%s)", enumStr))
		}
	}
	if v.constantValue != nil {
		parts = append(parts, fmt.Sprintf("Const(%t)", *v.constantValue))
	}
	
	return buildMethodChain("validator.Boolean()", parts), nil
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
		return buildMethodChain("validator.Array()", parts), nil
	}
	
	// Simple case - just constraints
	return buildMethodChain("validator.Array()", parts), nil
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
		// Generate properties as PropertyPair arguments
		propertiesPairs, err := g.generatePropertyPairs(v.properties)
		if err != nil {
			return "", fmt.Errorf("failed to generate property pairs: %w", err)
		}
		parts = append(parts, fmt.Sprintf("Properties(%s)", propertiesPairs))
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
	// For meta-schema, all Object validators should be strict to reject non-objects
	parts = append(parts, "StrictObjectType(true)")
	return buildMethodChain("validator.Object()", parts), nil
}


// generatePropertyPairs generates Go code for PropertyPair arguments
func (g *codeGenerator) generatePropertyPairs(properties map[string]Interface) (string, error) {
	if len(properties) == 0 {
		return "", nil
	}
	
	var propPairs []string
	
	for propName, propValidator := range properties {
		// Generate the validator code for this property
		propCode, err := g.generateBuilderChain(propValidator)
		if err != nil {
			return "", fmt.Errorf("failed to generate validator for property %s: %w", propName, err)
		}
		
		// Create the PropertyPair
		propPairs = append(propPairs, fmt.Sprintf("validator.PropPair(%s, %s)", getKeywordConstant(propName), propCode))
	}
	
	// Format with newlines if multiple properties
	if len(propPairs) > 1 {
		return "\n\t\t" + strings.Join(propPairs, ",\n\t\t") + ",\n\t", nil
	}
	return strings.Join(propPairs, ", "), nil
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
		// Format enum arguments with newlines if there are multiple
		var enumArgs string
		if len(enumStrs) > 1 {
			enumArgs = "\n\t\t" + strings.Join(enumStrs, ",\n\t\t") + ",\n\t"
		} else {
			enumArgs = strings.Join(enumStrs, ", ")
		}
		
		return fmt.Sprintf(`func() validator.Interface {
			mv := validator.NewMultiValidator(validator.OrMode)
			mv.Append(validator.String().Enum(%s).MustBuild())
			mv.Append(validator.Array().MinItems(1).UniqueItems(true).MustBuild())
			return mv
		}()`, enumArgs), nil
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
		// For multiline arguments, format differently
		if len(v.enum) > 1 {
			parts = append(parts, fmt.Sprintf("Enum(\n\t%s\n)", enumStr))
		} else {
			parts = append(parts, fmt.Sprintf("Enum(%s)", enumStr))
		}
	}
	if v.constantValue != nil {
		parts = append(parts, fmt.Sprintf("Const(%q)", *v.constantValue))
	}
	
	return buildMethodChain("validator.String()", parts), nil
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
		// For multiline arguments, format differently
		if len(v.enum) > 1 {
			parts = append(parts, fmt.Sprintf("Enum(\n\t%s\n)", enumStr))
		} else {
			parts = append(parts, fmt.Sprintf("Enum(%s)", enumStr))
		}
	}
	if v.constantValue != nil {
		parts = append(parts, fmt.Sprintf("Const(%d)", *v.constantValue))
	}
	
	return buildMethodChain("validator.Integer()", parts), nil
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

// buildMethodChain creates a method chain with proper line formatting using codegen
func buildMethodChain(baseMethod string, parts []string) string {
	var buf strings.Builder
	o := codegen.NewOutput(&buf)
	
	if len(parts) == 0 {
		o.R("%s.MustBuild()", baseMethod)
		return buf.String()
	}
	
	// Write the base method call
	o.L("%s.", baseMethod)
	
	// Write each method call on its own line with proper indentation
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - add MustBuild and close
			o.L("\t%s.", part)
			o.L("\tMustBuild()")
		} else {
			o.L("\t%s.", part)
		}
	}
	
	return buf.String()
}

// buildMultilineCall creates a function call with arguments formatted across multiple lines
func buildMultilineCall(funcName string, args []string) string {
	var buf strings.Builder
	o := codegen.NewOutput(&buf)
	
	if len(args) <= 1 {
		// Single argument - keep on one line
		if len(args) == 1 {
			o.R("%s{%s}", funcName, args[0])
		} else {
			o.R("%s{}", funcName)
		}
		return buf.String()
	}
	
	// Multiple arguments - format across multiple lines
	o.L("%s{", funcName)
	for _, arg := range args {
		o.L("\t%s,", arg)
	}
	o.L("}")
	
	return buf.String()
}

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
	return buildMultilineCall("[]string", quoted)
}

// formatIntSlice formats an int slice for Go code
func formatIntSlice(ints []int) string {
	strs := make([]string, len(ints))
	for i, n := range ints {
		strs[i] = fmt.Sprintf("%d", n)
	}
	return buildMultilineCall("[]int", strs)
}

// formatFloat64Slice formats a float64 slice for Go code
func formatFloat64Slice(floats []float64) string {
	strs := make([]string, len(floats))
	for i, f := range floats {
		strs[i] = fmt.Sprintf("%g", f)
	}
	return buildMultilineCall("[]float64", strs)
}

// formatBoolSlice formats a bool slice for Go code
func formatBoolSlice(bools []bool) string {
	strs := make([]string, len(bools))
	for i, b := range bools {
		strs[i] = fmt.Sprintf("%t", b)
	}
	return buildMultilineCall("[]bool", strs)
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
	// Use StrictObjectType(true) to ensure only objects are accepted
	return "validator.Object().StrictObjectType(true).MustBuild()", nil
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
func (g *codeGenerator) generateUnevaluatedPropertiesCompositionBuilderChain(_ *UnevaluatedPropertiesCompositionValidator) (string, error) {
	// Unevaluated properties validators are very complex
	// For now, create a basic validator that accepts everything
	// TODO: Implement proper unevaluated properties generation
	return "&validator.EmptyValidator{}", nil
}

// generateAnyOfUnevaluatedPropertiesCompositionBuilderChain creates just the builder chain for anyOf unevaluated properties composition validators
func (g *codeGenerator) generateAnyOfUnevaluatedPropertiesCompositionBuilderChain(_ *AnyOfUnevaluatedPropertiesCompositionValidator) (string, error) {
	// AnyOf unevaluated properties validators are very complex
	// For now, create a basic validator that accepts everything
	// TODO: Implement proper anyOf unevaluated properties generation
	return "&validator.EmptyValidator{}", nil
}

// generateOneOfUnevaluatedPropertiesCompositionBuilderChain creates just the builder chain for oneOf unevaluated properties composition validators
func (g *codeGenerator) generateOneOfUnevaluatedPropertiesCompositionBuilderChain(_ *OneOfUnevaluatedPropertiesCompositionValidator) (string, error) {
	// OneOf unevaluated properties validators are very complex
	// For now, create a basic validator that accepts everything
	// TODO: Implement proper oneOf unevaluated properties generation
	return "&validator.EmptyValidator{}", nil
}

// generateRefUnevaluatedPropertiesCompositionBuilderChain creates just the builder chain for ref unevaluated properties composition validators
func (g *codeGenerator) generateRefUnevaluatedPropertiesCompositionBuilderChain(_ *RefUnevaluatedPropertiesCompositionValidator) (string, error) {
	// Ref unevaluated properties validators are very complex
	// For now, create a basic validator that accepts everything
	// TODO: Implement proper ref unevaluated properties generation
	return "&validator.EmptyValidator{}", nil
}

// generateIfThenElseBuilderChain creates just the builder chain for if-then-else validators
func (g *codeGenerator) generateIfThenElseBuilderChain(_ *IfThenElseValidator) (string, error) {
	// If-then-else validators are complex conditional validators
	// For now, create a basic validator that accepts everything
	// TODO: Implement proper if-then-else generation
	return "&validator.EmptyValidator{}", nil
}

// generateIfThenElseUnevaluatedPropertiesCompositionBuilderChain creates just the builder chain for if-then-else unevaluated properties composition validators
func (g *codeGenerator) generateIfThenElseUnevaluatedPropertiesCompositionBuilderChain(_ *IfThenElseUnevaluatedPropertiesCompositionValidator) (string, error) {
	// If-then-else unevaluated properties validators are very complex
	// For now, create a basic validator that accepts everything
	// TODO: Implement proper if-then-else unevaluated properties generation
	return "&validator.EmptyValidator{}", nil
}