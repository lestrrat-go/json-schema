package validator

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/lestrrat-go/codegen"
	"github.com/lestrrat-go/json-schema/keywords"
)

// Ensure keywords package is imported (referenced in string maps)
var _ = keywords.Type

// keywordConstantMap maps JSON Schema keywords to their keywords package constant names
var keywordConstantMap = map[string]string{
	"$id":                   "keywords.ID",
	"$schema":               "keywords.Schema",
	"$anchor":               "keywords.Anchor",
	"$dynamicAnchor":        "keywords.DynamicAnchor",
	"$dynamicRef":           "keywords.DynamicReference",
	"$ref":                  "keywords.Reference",
	"$comment":              "keywords.Comment",
	"$defs":                 "keywords.Definitions",
	"$vocabulary":           "keywords.Vocabulary",
	"type":                  "keywords.Type",
	"enum":                  "keywords.Enum",
	"const":                 "keywords.Const",
	"multipleOf":            "keywords.MultipleOf",
	"maximum":               "keywords.Maximum",
	"exclusiveMaximum":      "keywords.ExclusiveMaximum",
	"minimum":               "keywords.Minimum",
	"exclusiveMinimum":      "keywords.ExclusiveMinimum",
	"maxLength":             "keywords.MaxLength",
	"minLength":             "keywords.MinLength",
	"pattern":               "keywords.Pattern",
	"additionalItems":       "keywords.AdditionalItems",
	"items":                 "keywords.Items",
	"maxItems":              "keywords.MaxItems",
	"minItems":              "keywords.MinItems",
	"uniqueItems":           "keywords.UniqueItems",
	"contains":              "keywords.Contains",
	"maxContains":           "keywords.MaxContains",
	"minContains":           "keywords.MinContains",
	"maxProperties":         "keywords.MaxProperties",
	"minProperties":         "keywords.MinProperties",
	"required":              "keywords.Required",
	"additionalProperties":  "keywords.AdditionalProperties",
	"definitions":           "keywords.Definitions",
	"properties":            "keywords.Properties",
	"patternProperties":     "keywords.PatternProperties",
	"dependencies":          "keywords.DependentSchemas", // Note: "dependencies" maps to DependentSchemas in 2020-12
	"dependentSchemas":      "keywords.DependentSchemas",
	"dependentRequired":     "keywords.DependentRequired",
	"propertyNames":         "keywords.PropertyNames",
	"allOf":                 "keywords.AllOf",
	"anyOf":                 "keywords.AnyOf",
	"oneOf":                 "keywords.OneOf",
	"not":                   "keywords.Not",
	"if":                    "keywords.If",
	"then":                  "keywords.Then",
	"else":                  "keywords.Else",
	"format":                "keywords.Format",
	"contentEncoding":       "keywords.ContentEncoding",
	"contentMediaType":      "keywords.ContentMediaType",
	"contentSchema":         "keywords.ContentSchema",
	"title":                 "keywords.Title",
	"description":           "keywords.Description",
	"default":               "keywords.Default",
	"deprecated":            "keywords.Deprecated",
	"readOnly":              "keywords.ReadOnly",
	"writeOnly":             "keywords.WriteOnly",
	"examples":              "keywords.Examples",
	"prefixItems":           "keywords.PrefixItems",
	"unevaluatedItems":      "keywords.UnevaluatedItems",
	"unevaluatedProperties": "keywords.UnevaluatedProperties",
	// Legacy keywords for backward compatibility
	"$recursiveRef":    "keywords.RecursiveRef",    // Deprecated in 2020-12 but still in some schemas
	"$recursiveAnchor": "keywords.RecursiveAnchor", // Deprecated in 2020-12 but still in some schemas
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
	switch validator := v.(type) {
	case *stringValidator:
		return g.generateStringBuilderChain(dst, validator)
	case *integerValidator:
		return g.generateIntegerBuilderChain(dst, validator)
	case *numberValidator:
		return g.generateNumberBuilderChain(dst, validator)
	case *booleanValidator:
		return g.generateBooleanBuilderChain(dst, validator)
	case *arrayValidator:
		return g.generateArrayBuilderChain(dst, validator)
	case *objectValidator:
		return g.generateObjectBuilderChain(dst, validator)
	case *MultiValidator:
		return g.generateMultiBuilderChain(dst, validator)
	case *EmptyValidator:
		return g.generateEmptyBuilderChain(dst)
	case *NotValidator:
		return g.generateNotBuilderChain(dst, validator)
	case *NullValidator:
		return g.generateNullBuilderChain(dst)
	case *GeneralValidator:
		return g.generateGeneralBuilderChain(dst, validator)
	case *alwaysPassValidator:
		return g.generateAlwaysPassBuilderChain(dst)
	case *alwaysFailValidator:
		return g.generateAlwaysFailBuilderChain(dst)
	case *ReferenceValidator:
		return g.generateReferenceBuilderChain(dst, validator)
	case *DynamicReferenceValidator:
		return g.generateDynamicReferenceBuilderChain(dst, validator)
	case *contentValidator:
		return g.generateContentBuilderChain(dst, validator)
	case *dependentSchemasValidator:
		return g.generateDependentSchemasBuilderChain(dst, validator)
	case *inferredNumberValidator:
		return g.generateInferredNumberBuilderChain(dst, validator)
	case *UnevaluatedPropertiesCompositionValidator:
		return g.generateUnevaluatedPropertiesCompositionBuilderChain(dst, validator)
	case *AnyOfUnevaluatedPropertiesCompositionValidator:
		return g.generateAnyOfUnevaluatedPropertiesCompositionBuilderChain(dst, validator)
	case *OneOfUnevaluatedPropertiesCompositionValidator:
		return g.generateOneOfUnevaluatedPropertiesCompositionBuilderChain(dst, validator)
	case *RefUnevaluatedPropertiesCompositionValidator:
		return g.generateRefUnevaluatedPropertiesCompositionBuilderChain(dst, validator)
	case *IfThenElseValidator:
		return g.generateIfThenElseBuilderChain(dst, validator)
	case *IfThenElseUnevaluatedPropertiesCompositionValidator:
		return g.generateIfThenElseUnevaluatedPropertiesCompositionBuilderChain(dst, validator)
	default:
		// Debug: Print what unsupported validator type we encountered
		fmt.Printf("GENERATOR DEBUG: Unsupported validator type: %T, falling back to EmptyValidator\n", v)
		_, err := fmt.Fprint(dst, "&validator.EmptyValidator{}")
		return err
	}
}

// generateNumberBuilderChain creates just the builder chain for number validators
func (g *codeGenerator) generateNumberBuilderChain(dst io.Writer, v *numberValidator) error {
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

	return buildMethodChain(dst, "validator.Number()", parts)
}

// generateBooleanBuilderChain creates just the builder chain for boolean validators
func (g *codeGenerator) generateBooleanBuilderChain(dst io.Writer, v *booleanValidator) error {
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

	return buildMethodChain(dst, "validator.Boolean()", parts)
}

// generateArrayBuilderChain creates just the builder chain for array validators
func (g *codeGenerator) generateArrayBuilderChain(dst io.Writer, v *arrayValidator) error {
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
		return buildMethodChain(dst, "validator.Array()", parts)
	}

	// Simple case - just constraints
	return buildMethodChain(dst, "validator.Array()", parts)
}

// generateObjectBuilderChain creates just the builder chain for object validators
func (g *codeGenerator) generateObjectBuilderChain(dst io.Writer, v *objectValidator) error {
	var parts []string

	// Basic object constraints
	if v.minProperties != nil {
		parts = append(parts, fmt.Sprintf("MinProperties(%d)", *v.minProperties))
	}
	if v.maxProperties != nil {
		parts = append(parts, fmt.Sprintf("MaxProperties(%d)", *v.maxProperties))
	}
	if len(v.required) > 0 {
		requiredStr := formatStringSlice(v.required)
		parts = append(parts, fmt.Sprintf("Required(%s)", requiredStr))
	}

	// Handle complex properties
	if len(v.properties) > 0 {
		var propPairs []string

		// Sort property names for deterministic output
		var propNames []string
		for propName := range v.properties {
			propNames = append(propNames, propName)
		}
		sort.Strings(propNames)

		for _, propName := range propNames {
			propValidator := v.properties[propName]
			// Generate the validator code for this property
			var propBuf strings.Builder
			if err := g.Generate(&propBuf, propValidator); err != nil {
				return fmt.Errorf("failed to generate validator for property %s: %w", propName, err)
			}
			propCode := propBuf.String()

			// Create the PropertyPair
			propPairs = append(propPairs, fmt.Sprintf("validator.PropPair(%s, %s)", getKeywordConstant(propName), propCode))
		}

		// Format with newlines if multiple properties
		var propertiesArg string
		if len(propPairs) > 1 {
			propertiesArg = "\n\t\t" + strings.Join(propPairs, ",\n\t\t") + ",\n\t"
		} else {
			propertiesArg = strings.Join(propPairs, ", ")
		}
		parts = append(parts, fmt.Sprintf("Properties(%s)", propertiesArg))
	}

	// Handle additional properties
	if v.additionalProperties != nil {
		switch ap := v.additionalProperties.(type) {
		case bool:
			parts = append(parts, fmt.Sprintf("AdditionalPropertiesBool(%t)", ap))
		case Interface:
			var apBuf strings.Builder
			if err := g.Generate(&apBuf, ap); err != nil {
				return fmt.Errorf("failed to generate additional properties validator: %w", err)
			}
			apCode := apBuf.String()
			parts = append(parts, fmt.Sprintf("AdditionalPropertiesSchema(%s)", apCode))
		}
	}

	// Handle pattern properties
	if len(v.patternProperties) > 0 {
		// For pattern properties, we need to generate a more complex structure
		// For now, we'll skip this as it's complex to generate regexp objects
		fmt.Printf("GENERATOR DEBUG: Skipping pattern properties generation (not implemented)\n")
	}

	// Handle property names validator
	if v.propertyNames != nil {
		var pnBuf strings.Builder
		if err := g.Generate(&pnBuf, v.propertyNames); err != nil {
			return fmt.Errorf("failed to generate property names validator: %w", err)
		}
		pnCode := pnBuf.String()
		parts = append(parts, fmt.Sprintf("PropertyNames(%s)", pnCode))
	}

	// For meta-schema, all Object validators should be strict to reject non-objects
	parts = append(parts, "StrictObjectType(true)")
	return buildMethodChain(dst, "validator.Object()", parts)
}

// generateMultiBuilderChain creates just the builder chain for multi validators
func (g *codeGenerator) generateMultiBuilderChain(dst io.Writer, v *MultiValidator) error {
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
	if len(v.validators) > 0 {
		var childParts []string

		// Generate each child validator
		for i, child := range v.validators {
			var childBuf strings.Builder
			if err := g.Generate(&childBuf, child); err != nil {
				return fmt.Errorf("failed to generate child validator %d: %w", i, err)
			}
			childParts = append(childParts, childBuf.String())
		}

		// Generate append calls
		var appendCalls []string
		for _, childPart := range childParts {
			appendCalls = append(appendCalls, fmt.Sprintf("\t\tmv.Append(%s)", childPart))
		}

		appendCallsStr := strings.Join(appendCalls, "\n")
		_, err := fmt.Fprintf(dst, "func() validator.Interface {\n\t\tmv := validator.NewMultiValidator(validator.%s)\n%s\n\t\treturn mv\n\t}()", mode, appendCallsStr)
		return err
	}

	_, err := fmt.Fprint(dst, "&validator.EmptyValidator{}")
	return err
}

// generateEmptyBuilderChain creates just the builder chain for empty validators
func (g *codeGenerator) generateEmptyBuilderChain(dst io.Writer) error {
	_, err := fmt.Fprint(dst, "&validator.EmptyValidator{}")
	return err
}

// generateNotBuilderChain creates just the builder chain for not validators
func (g *codeGenerator) generateNotBuilderChain(dst io.Writer, v *NotValidator) error {
	if _, err := fmt.Fprint(dst, "func() validator.Interface { child := "); err != nil {
		return err
	}
	if err := g.Generate(dst, v.validator); err != nil {
		return fmt.Errorf("failed to generate child validator for not: %w", err)
	}
	_, err := fmt.Fprint(dst, "; return &validator.NotValidator{validator: child} }()")
	return err
}

// generateNullBuilderChain creates just the builder chain for null validators
func (g *codeGenerator) generateNullBuilderChain(dst io.Writer) error {
	_, err := fmt.Fprint(dst, "&validator.NullValidator{}")
	return err
}

// generateGeneralBuilderChain creates just the builder chain for general validators
func (g *codeGenerator) generateGeneralBuilderChain(dst io.Writer, v *GeneralValidator) error {
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

		_, err := fmt.Fprintf(dst, `func() validator.Interface {
			mv := validator.NewMultiValidator(validator.OrMode)
			mv.Append(validator.String().Enum(%s).MustBuild())
			mv.Append(validator.Array().MinItems(1).UniqueItems(true).MustBuild())
			return mv
		}()`, enumArgs)
		return err
	}

	if v.hasConst {
		// For const validation, accept any matching value
		// Since we can't access the const value, return EmptyValidator
		_, err := fmt.Fprint(dst, "&validator.EmptyValidator{}")
		return err
	}

	_, err := fmt.Fprint(dst, "&validator.EmptyValidator{}")
	return err
}

// generateAlwaysPassBuilderChain creates just the builder chain for always pass validators
func (g *codeGenerator) generateAlwaysPassBuilderChain(dst io.Writer) error {
	_, err := fmt.Fprint(dst, "&validator.EmptyValidator{}")
	return err
}

// generateAlwaysFailBuilderChain creates just the builder chain for always fail validators
func (g *codeGenerator) generateAlwaysFailBuilderChain(dst io.Writer) error {
	_, err := fmt.Fprint(dst, "&validator.NotValidator{validator: &validator.EmptyValidator{}}")
	return err
}

// generateStringBuilderChain creates just the builder chain for string validators
func (g *codeGenerator) generateStringBuilderChain(dst io.Writer, v *stringValidator) error {
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

	return buildMethodChain(dst, "validator.String()", parts)
}

// generateIntegerBuilderChain creates just the builder chain for integer validators
func (g *codeGenerator) generateIntegerBuilderChain(dst io.Writer, v *integerValidator) error {
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

	return buildMethodChain(dst, "validator.Integer()", parts)
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

// buildMethodChain creates a method chain with proper line formatting using codegen
func buildMethodChain(dst io.Writer, baseMethod string, parts []string) error {
	o := codegen.NewOutput(dst)

	if len(parts) == 0 {
		o.R("%s.MustBuild()", baseMethod)
		return nil
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

	return nil
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
func (g *codeGenerator) generateReferenceBuilderChain(dst io.Writer, v *ReferenceValidator) error {
	// If the reference has been resolved, generate the resolved validator
	if v.resolved != nil {
		// Generate the resolved validator, but watch out for circular references
		if v.resolved == v {
			// Self-reference - create EmptyValidator to avoid infinite recursion
			_, err := fmt.Fprint(dst, "&validator.EmptyValidator{}")
			return err
		}

		if err := g.Generate(dst, v.resolved); err != nil {
			// If we can't generate the resolved validator, fall back to EmptyValidator
			_, fallbackErr := fmt.Fprint(dst, "&validator.EmptyValidator{}")
			return fallbackErr
		}
		return nil
	}

	// If not resolved, we can't generate proper code, fall back to EmptyValidator
	// This happens when references can't be resolved at compile time
	_, err := fmt.Fprint(dst, "&validator.EmptyValidator{}")
	return err
}

// generateDynamicReferenceBuilderChain creates just the builder chain for dynamic reference validators
func (g *codeGenerator) generateDynamicReferenceBuilderChain(dst io.Writer, v *DynamicReferenceValidator) error {
	// If the dynamic reference has been resolved, generate the resolved validator
	if v.resolved != nil {
		// Generate the resolved validator, but watch out for circular references
		if v.resolved == v {
			// Self-reference - create EmptyValidator to avoid infinite recursion
			_, err := fmt.Fprint(dst, "&validator.EmptyValidator{}")
			return err
		}

		if err := g.Generate(dst, v.resolved); err != nil {
			// If we can't generate the resolved validator, fall back to EmptyValidator
			_, fallbackErr := fmt.Fprint(dst, "&validator.EmptyValidator{}")
			return fallbackErr
		}
		return nil
	}

	// If not resolved, create a reasonable validator based on JSON Schema expectations
	// Most unresolved dynamic references in meta-schemas expect either:
	// 1. JSON Schema objects (which should be objects)
	// 2. Any JSON value (for things like const/default)
	// For meta-schema generation, assume these accept schema objects (type: object)
	// This is better than EmptyValidator which accepts anything
	// Use StrictObjectType(true) to ensure only objects are accepted
	_, err := fmt.Fprint(dst, "validator.Object().StrictObjectType(true).MustBuild()")
	return err
}

// generateContentBuilderChain creates just the builder chain for content validators
func (g *codeGenerator) generateContentBuilderChain(dst io.Writer, v *contentValidator) error {
	// Content validators are complex but we can create a basic structure
	var parts []string

	if v.contentEncoding != "" {
		parts = append(parts, fmt.Sprintf("contentEncoding: %q", v.contentEncoding))
	}
	if v.contentMediaType != "" {
		parts = append(parts, fmt.Sprintf("contentMediaType: %q", v.contentMediaType))
	}

	// Handle contentSchema if present
	if v.contentSchema != nil {
		var childBuf strings.Builder
		if err := g.Generate(&childBuf, v.contentSchema); err != nil {
			return fmt.Errorf("failed to generate content schema: %w", err)
		}
		contentSchemaStr := fmt.Sprintf("contentSchema: %s", childBuf.String())
		parts = append(parts, contentSchemaStr)
	}

	if len(parts) > 0 {
		fieldsStr := strings.Join(parts, ",\n\t\t")
		_, err := fmt.Fprintf(dst, "&validator.contentValidator{\n\t\t%s,\n\t}", fieldsStr)
		return err
	}

	_, err := fmt.Fprint(dst, "&validator.contentValidator{}")
	return err
}

// generateDependentSchemasBuilderChain creates just the builder chain for dependent schemas validators
func (g *codeGenerator) generateDependentSchemasBuilderChain(dst io.Writer, v *dependentSchemasValidator) error {
	// Generate the dependent schemas map
	if len(v.dependentSchemas) == 0 {
		_, err := fmt.Fprint(dst, "&validator.dependentSchemasValidator{dependentSchemas: make(map[string]validator.Interface)}")
		return err
	}

	var mapEntries []string
	var childSetup []string

	// Sort property names for deterministic output
	var propNames []string
	for propName := range v.dependentSchemas {
		propNames = append(propNames, propName)
	}
	sort.Strings(propNames)

	for _, propName := range propNames {
		propValidator := v.dependentSchemas[propName]
		childVar := fmt.Sprintf("dep_%s", sanitizeVarName(propName))

		// Generate child validator code into a buffer
		var childBuf strings.Builder
		if err := g.Generate(&childBuf, propValidator); err != nil {
			return fmt.Errorf("failed to generate dependent schema for %q: %w", propName, err)
		}
		childChain := childBuf.String()

		childSetup = append(childSetup, fmt.Sprintf("\t%s := %s", childVar, childChain))
		mapEntries = append(mapEntries, fmt.Sprintf("\t\t%q: %s", propName, childVar))
	}

	if len(childSetup) > 0 {
		setupStr := strings.Join(childSetup, "\n")
		entriesStr := strings.Join(mapEntries, ",\n")

		_, err := fmt.Fprintf(dst, `func() validator.Interface {
%s
	return &validator.dependentSchemasValidator{
		dependentSchemas: map[string]validator.Interface{
%s,
		},
	}
}()`, setupStr, entriesStr)
		return err
	}

	_, err := fmt.Fprint(dst, "&validator.dependentSchemasValidator{dependentSchemas: make(map[string]validator.Interface)}")
	return err
}

// generateInferredNumberBuilderChain creates just the builder chain for inferred number validators
func (g *codeGenerator) generateInferredNumberBuilderChain(dst io.Writer, v *inferredNumberValidator) error {
	// Generate the underlying number validator
	if err := g.Generate(dst, v.numberValidator); err != nil {
		return fmt.Errorf("failed to generate inferred number validator: %w", err)
	}

	// inferredNumberValidator is a wrapper around numberValidator
	return nil
}

// generateUnevaluatedPropertiesCompositionBuilderChain creates just the builder chain for unevaluated properties composition validators
func (g *codeGenerator) generateUnevaluatedPropertiesCompositionBuilderChain(dst io.Writer, _ *UnevaluatedPropertiesCompositionValidator) error {
	// Unevaluated properties validators are very complex
	// Create a basic validator that accepts everything
	_, err := fmt.Fprint(dst, "&validator.EmptyValidator{}")
	return err
}

// generateAnyOfUnevaluatedPropertiesCompositionBuilderChain creates just the builder chain for anyOf unevaluated properties composition validators
func (g *codeGenerator) generateAnyOfUnevaluatedPropertiesCompositionBuilderChain(dst io.Writer, _ *AnyOfUnevaluatedPropertiesCompositionValidator) error {
	// AnyOf unevaluated properties validators are very complex
	// Create a basic validator that accepts everything
	_, err := fmt.Fprint(dst, "&validator.EmptyValidator{}")
	return err
}

// generateOneOfUnevaluatedPropertiesCompositionBuilderChain creates just the builder chain for oneOf unevaluated properties composition validators
func (g *codeGenerator) generateOneOfUnevaluatedPropertiesCompositionBuilderChain(dst io.Writer, _ *OneOfUnevaluatedPropertiesCompositionValidator) error {
	// OneOf unevaluated properties validators are very complex
	// Create a basic validator that accepts everything
	_, err := fmt.Fprint(dst, "&validator.EmptyValidator{}")
	return err
}

// generateRefUnevaluatedPropertiesCompositionBuilderChain creates just the builder chain for ref unevaluated properties composition validators
func (g *codeGenerator) generateRefUnevaluatedPropertiesCompositionBuilderChain(dst io.Writer, _ *RefUnevaluatedPropertiesCompositionValidator) error {
	// Ref unevaluated properties validators are very complex
	// Create a basic validator that accepts everything
	_, err := fmt.Fprint(dst, "&validator.EmptyValidator{}")
	return err
}

// generateIfThenElseBuilderChain creates just the builder chain for if-then-else validators
func (g *codeGenerator) generateIfThenElseBuilderChain(dst io.Writer, _ *IfThenElseValidator) error {
	// If-then-else validators are complex conditional validators
	// Create a basic validator that accepts everything
	_, err := fmt.Fprint(dst, "&validator.EmptyValidator{}")
	return err
}

// generateIfThenElseUnevaluatedPropertiesCompositionBuilderChain creates just the builder chain for if-then-else unevaluated properties composition validators
func (g *codeGenerator) generateIfThenElseUnevaluatedPropertiesCompositionBuilderChain(dst io.Writer, _ *IfThenElseUnevaluatedPropertiesCompositionValidator) error {
	// If-then-else unevaluated properties validators are very complex
	// Create a basic validator that accepts everything
	_, err := fmt.Fprint(dst, "&validator.EmptyValidator{}")
	return err
}
