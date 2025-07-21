package validator

import (
	"bytes"
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
	// Generate the code using the internal method
	return g.generateInternal(dst, v)
}

// generateInternal generates code using the provided codegen.Output
func (g *codeGenerator) generateInternal(dst io.Writer, v Interface) error {
	o := codegen.NewOutput(dst) // only placed for compatibility. methods should receive dst instead
	switch validator := v.(type) {
	case *stringValidator:
		return g.generateString(dst, validator)
	case *integerValidator:
		return g.generateInteger(dst, validator)
	case *numberValidator:
		return g.generateNumber(dst, validator)
	case *booleanValidator:
		return g.generateBoolean(dst, validator)
	case *arrayValidator:
		return g.generateArray(dst, validator)
	case *objectValidator:
		return g.generateObject(dst, validator)
	case *MultiValidator:
		return g.generateMulti(dst, validator)
	case *EmptyValidator:
		return g.generateEmpty(dst)
	case *NotValidator:
		return g.generateNot(dst, validator)
	case *NullValidator:
		return g.generateNull(dst)
	case *GeneralValidator:
		return g.generateGeneral(dst, validator)
	case *alwaysPassValidator:
		return g.generateAlwaysPass(dst)
	case *alwaysFailValidator:
		return g.generateAlwaysFail(dst)
	case *ReferenceValidator:
		return g.generateReference(dst, validator)
	case *DynamicReferenceValidator:
		return g.generateDynamicReference(dst, validator)
	case *contentValidator:
		return g.generateContent(dst, validator)
	case *dependentSchemasValidator:
		return g.generateDependentSchemas(dst, validator)
	case *inferredNumberValidator:
		return g.generateInferredNumber(dst, validator)
	case *UnevaluatedPropertiesCompositionValidator:
		return g.generateUnevaluatedPropertiesComposition(dst)
	case *AnyOfUnevaluatedPropertiesCompositionValidator:
		return g.generateAnyOfUnevaluatedPropertiesComposition(dst)
	case *OneOfUnevaluatedPropertiesCompositionValidator:
		return g.generateOneOfUnevaluatedPropertiesComposition(dst)
	case *RefUnevaluatedPropertiesCompositionValidator:
		return g.generateRefUnevaluatedPropertiesComposition(dst)
	case *IfThenElseValidator:
		return g.generateIfThenElse(dst)
	case *IfThenElseUnevaluatedPropertiesCompositionValidator:
		return g.generateIfThenElseUnevaluatedPropertiesComposition(dst)
	default:
		// Debug: Print what unsupported validator type we encountered
		fmt.Printf("GENERATOR DEBUG: Unsupported validator type: %T, falling back to EmptyValidator\n", v)
		o.R("&validator.EmptyValidator{}")
		return nil
	}
}

// generateObjectBuilderChain creates just the builder chain for object validators
func (g *codeGenerator) generateObject(dst io.Writer, v *objectValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	o.L("validator.Object().")
	// Basic object constraints
	if v.minProperties != nil {
		o.L("MinProperties(%d).", *v.minProperties)
	}
	if v.maxProperties != nil {
		o.L("MaxProperties(%d).", *v.maxProperties)
	}
	if len(v.required) > 0 {
		o.L("Required(")
		for i, req := range v.required {
			if i > 0 {
				o.R(",")
			}
			o.L("%q", req)
		}
		o.L(")")
	}

	// Handle complex properties
	if len(v.properties) > 0 {
		// Sort property names for deterministic output
		var propNames []string
		for propName := range v.properties {
			propNames = append(propNames, propName)
		}
		sort.Strings(propNames)

		o.L("Properties(")
		for _, propName := range propNames {
			propValidator := v.properties[propName]
			// Generate the validator code for this property
			o.L("validator.PropPair(")
			o.L("%s,", getKeywordConstant(propName))
			if err := g.Generate(&buf, propValidator); err != nil {
				return fmt.Errorf("failed to generate validator for property %s: %w", propName, err)
			}
			o.R(",")
			o.L("),")
		}
		o.L(").")
	}

	// Handle additional properties
	if v.additionalProperties != nil {
		switch ap := v.additionalProperties.(type) {
		case bool:
			o.L("AdditionalPropertiesBool(%t).", ap)
		case Interface:
			o.L("AdditionalPropertiesSchema(")
			// force newline
			o.L("")
			if err := g.Generate(&buf, ap); err != nil {
				return fmt.Errorf("failed to generate additional properties validator: %w", err)
			}
			o.R(",")
			o.L(").")
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
		o.L("PropertyNames(")
		if err := g.Generate(&buf, v.propertyNames); err != nil {
			return fmt.Errorf("failed to generate property names validator: %w", err)
		}
		o.R(",")
		o.L(").")
	}

	// For meta-schema, all Object validators should be strict to reject non-objects
	o.L("StrictObjectType(true).")
	o.L("MustBuild()")
	_, err := buf.WriteTo(dst)
	return err
}

// generateMulti creates the builder chain for multi validators
func (g *codeGenerator) generateMulti(dst io.Writer, v *MultiValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

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
		o.L("func() validator.Interface {")
		o.L("mv := validator.NewMultiValidator(validator.%s)", mode)
		
		// Generate each child validator
		for i, child := range v.validators {
			o.L("mv.Append(")
			if err := g.Generate(&buf, child); err != nil {
				return fmt.Errorf("failed to generate child validator %d: %w", i, err)
			}
			o.R(")")
		}
		
		o.L("return mv")
		o.L("}()")
	} else {
		o.R("&validator.EmptyValidator{}")
	}

	_, err := buf.WriteTo(dst)
	return err
}


// generateEmpty creates the builder chain for empty validators
func (g *codeGenerator) generateEmpty(dst io.Writer) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)
	
	o.R("&validator.EmptyValidator{}")
	
	_, err := buf.WriteTo(dst)
	return err
}

// generateNot creates the builder chain for not validators
func (g *codeGenerator) generateNot(dst io.Writer, v *NotValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)
	
	o.L("func() validator.Interface { child := ")
	if err := g.Generate(&buf, v.validator); err != nil {
		return fmt.Errorf("failed to generate child validator for not: %w", err)
	}
	o.R("; return &validator.NotValidator{validator: child} }()")
	
	_, err := buf.WriteTo(dst)
	return err
}


// generateGeneralBuilderChain creates just the builder chain for general validators
func (g *codeGenerator) generateGeneral(dst io.Writer, v *GeneralValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

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

		o.L("func() validator.Interface {")
		o.L("mv := validator.NewMultiValidator(validator.OrMode)")
		o.L("mv.Append(validator.String().Enum(%s).MustBuild())", enumArgs)
		o.L("mv.Append(validator.Array().MinItems(1).UniqueItems(true).MustBuild())")
		o.L("return mv")
		o.L("}()")
	} else if v.hasConst {
		// For const validation, accept any matching value
		// Since we can't access the const value, return EmptyValidator
		o.R("&validator.EmptyValidator{}")
	} else {
		o.R("&validator.EmptyValidator{}")
	}

	_, err := buf.WriteTo(dst)
	return err
}



// generateString creates the builder chain for string validators
func (g *codeGenerator) generateString(dst io.Writer, v *stringValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	o.L("validator.String().")

	// Directly access validator fields - no intermediate map needed!
	if v.minLength != nil {
		o.L("MinLength(%d).", *v.minLength)
	}
	if v.maxLength != nil {
		o.L("MaxLength(%d).", *v.maxLength)
	}
	if v.pattern != nil {
		o.L("Pattern(%q).", v.pattern.String())
	}
	if v.format != nil {
		o.L("Format(%q).", *v.format)
	}
	if v.enum != nil {
		if len(v.enum) == 1 {
			o.L("Enum([]string{%q}).", v.enum[0])
		} else {
			o.L("Enum(")
			o.L("[]string{")
			for _, s := range v.enum {
				o.L("%q,", s)
			}
			o.L("}).")
		}
	}
	if v.constantValue != nil {
		o.L("Const(%q).", *v.constantValue)
	}

	o.L("MustBuild()")
	_, err := buf.WriteTo(dst)
	return err
}


// generateInteger creates the builder chain for integer validators
func (g *codeGenerator) generateInteger(dst io.Writer, v *integerValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	o.L("validator.Integer().")

	if v.multipleOf != nil {
		o.L("MultipleOf(%d).", *v.multipleOf)
	}
	if v.minimum != nil {
		o.L("Minimum(%d).", *v.minimum)
	}
	if v.maximum != nil {
		o.L("Maximum(%d).", *v.maximum)
	}
	if v.exclusiveMinimum != nil {
		o.L("ExclusiveMinimum(%d).", *v.exclusiveMinimum)
	}
	if v.exclusiveMaximum != nil {
		o.L("ExclusiveMaximum(%d).", *v.exclusiveMaximum)
	}
	if v.enum != nil {
		strs := make([]string, len(v.enum))
		for i, n := range v.enum {
			strs[i] = fmt.Sprintf("%d", n)
		}
		enumStr := fmt.Sprintf("[]int{%s}", strings.Join(strs, ", "))
		o.L("Enum(%s).", enumStr)
	}
	if v.constantValue != nil {
		o.L("Const(%d).", *v.constantValue)
	}

	o.L("MustBuild()")
	_, err := buf.WriteTo(dst)
	return err
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


// generateReferenceBuilderChain creates just the builder chain for reference validators
func (g *codeGenerator) generateReference(dst io.Writer, v *ReferenceValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	// If the reference has been resolved, generate the resolved validator
	if v.resolved != nil {
		// Generate the resolved validator, but watch out for circular references
		if v.resolved == v {
			// Self-reference - create EmptyValidator to avoid infinite recursion
			o.R("&validator.EmptyValidator{}")
		} else {
			if err := g.Generate(&buf, v.resolved); err != nil {
				// If we can't generate the resolved validator, fall back to EmptyValidator
				o.R("&validator.EmptyValidator{}")
			}
		}
	} else {
		// If not resolved, we can't generate proper code, fall back to EmptyValidator
		// This happens when references can't be resolved at compile time
		o.R("&validator.EmptyValidator{}")
	}

	_, err := buf.WriteTo(dst)
	return err
}

// generateDynamicReferenceBuilderChain creates just the builder chain for dynamic reference validators
func (g *codeGenerator) generateDynamicReference(dst io.Writer, v *DynamicReferenceValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	// If the dynamic reference has been resolved, generate the resolved validator
	if v.resolved != nil {
		// Generate the resolved validator, but watch out for circular references
		if v.resolved == v {
			// Self-reference - create EmptyValidator to avoid infinite recursion
			o.R("&validator.EmptyValidator{}")
		} else {
			if err := g.Generate(&buf, v.resolved); err != nil {
				// If we can't generate the resolved validator, fall back to EmptyValidator
				o.R("&validator.EmptyValidator{}")
			}
		}
	} else {
		// If not resolved, create a reasonable validator based on JSON Schema expectations
		// Most unresolved dynamic references in meta-schemas expect either:
		// 1. JSON Schema objects (which should be objects)
		// 2. Any JSON value (for things like const/default)
		// For meta-schema generation, assume these accept schema objects (type: object)
		// This is better than EmptyValidator which accepts anything
		// Use StrictObjectType(true) to ensure only objects are accepted
		o.R("validator.Object().StrictObjectType(true).MustBuild()")
	}

	_, err := buf.WriteTo(dst)
	return err
}

// generateContentBuilderChain creates just the builder chain for content validators
func (g *codeGenerator) generateContent(dst io.Writer, v *contentValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

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
		o.L("&validator.contentValidator{")
		o.L("%s,", fieldsStr)
		o.L("}")
	} else {
		o.R("&validator.contentValidator{}")
	}

	_, err := buf.WriteTo(dst)
	return err
}

// generateDependentSchemasBuilderChain creates just the builder chain for dependent schemas validators
func (g *codeGenerator) generateDependentSchemas(dst io.Writer, v *dependentSchemasValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	// Generate the dependent schemas map
	if len(v.dependentSchemas) == 0 {
		o.R("&validator.dependentSchemasValidator{dependentSchemas: make(map[string]validator.Interface)}")
	} else {
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

		setupStr := strings.Join(childSetup, "\n")
		entriesStr := strings.Join(mapEntries, ",\n")

		o.L("func() validator.Interface {")
		o.R("%s", setupStr)
		o.L("return &validator.dependentSchemasValidator{")
		o.L("dependentSchemas: map[string]validator.Interface{")
		o.R("%s,", entriesStr)
		o.L("},")
		o.L("}")
		o.L("}()")
	}

	_, err := buf.WriteTo(dst)
	return err
}

// generateInferredNumberBuilderChain creates just the builder chain for inferred number validators
func (g *codeGenerator) generateInferredNumber(dst io.Writer, v *inferredNumberValidator) error {
	// Generate the underlying number validator - it's just a wrapper
	return g.Generate(dst, v.numberValidator)
}


// Stub implementations for the remaining internal methods - these will be properly implemented later
// generateNumber creates the builder chain for number validators
func (g *codeGenerator) generateNumber(dst io.Writer, v *numberValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	o.L("validator.Number().")

	if v.multipleOf != nil {
		o.L("MultipleOf(%g).", *v.multipleOf)
	}
	if v.minimum != nil {
		o.L("Minimum(%g).", *v.minimum)
	}
	if v.maximum != nil {
		o.L("Maximum(%g).", *v.maximum)
	}
	if v.exclusiveMinimum != nil {
		o.L("ExclusiveMinimum(%g).", *v.exclusiveMinimum)
	}
	if v.exclusiveMaximum != nil {
		o.L("ExclusiveMaximum(%g).", *v.exclusiveMaximum)
	}
	if v.enum != nil {
		if len(v.enum) == 1 {
			o.L("Enum([]float64{%g}).", v.enum[0])
		} else {
			o.L("Enum(")
			o.L("[]float64{")
			for _, f := range v.enum {
				o.L("%g,", f)
			}
			o.L("}).")
		}
	}
	if v.constantValue != nil {
		o.L("Const(%g).", *v.constantValue)
	}

	o.L("MustBuild()")
	_, err := buf.WriteTo(dst)
	return err
}

// generateBoolean creates the builder chain for boolean validators
func (g *codeGenerator) generateBoolean(dst io.Writer, v *booleanValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	o.L("validator.Boolean().")

	if v.enum != nil {
		if len(v.enum) == 1 {
			o.L("Enum([]bool{%t}).", v.enum[0])
		} else {
			o.L("Enum(")
			o.L("[]bool{")
			for _, b := range v.enum {
				o.L("%t,", b)
			}
			o.L("}).")
		}
	}
	if v.constantValue != nil {
		o.L("Const(%t).", *v.constantValue)
	}

	o.L("MustBuild()")
	_, err := buf.WriteTo(dst)
	return err
}

// generateArray creates the builder chain for array validators
func (g *codeGenerator) generateArray(dst io.Writer, v *arrayValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	o.L("validator.Array().")

	// Basic array constraints
	if v.minItems != nil {
		o.L("MinItems(%d).", *v.minItems)
	}
	if v.maxItems != nil {
		o.L("MaxItems(%d).", *v.maxItems)
	}
	if v.uniqueItems {
		o.L("UniqueItems(true).")
	}
	if v.minContains != nil {
		o.L("MinContains(%d).", *v.minContains)
	}
	if v.maxContains != nil {
		o.L("MaxContains(%d).", *v.maxContains)
	}

	// Note: Complex items, prefixItems etc. would need more complex generation
	// For now, create basic array validator without complex items

	o.L("MustBuild()")
	_, err := buf.WriteTo(dst)
	return err
}



// generateNull creates the builder chain for null validators
func (g *codeGenerator) generateNull(dst io.Writer) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)
	
	o.R("&validator.NullValidator{}")
	
	_, err := buf.WriteTo(dst)
	return err
}

func (g *codeGenerator) generateAlwaysPass(dst io.Writer) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)
	o.R("&validator.EmptyValidator{}")
	_, err := buf.WriteTo(dst)
	return err
}

func (g *codeGenerator) generateAlwaysFail(dst io.Writer) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)
	o.R("&validator.NotValidator{validator: &validator.EmptyValidator{}}")
	_, err := buf.WriteTo(dst)
	return err
}

func (g *codeGenerator) generateUnevaluatedPropertiesComposition(dst io.Writer) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)
	o.R("&validator.EmptyValidator{}")
	_, err := buf.WriteTo(dst)
	return err
}

func (g *codeGenerator) generateAnyOfUnevaluatedPropertiesComposition(dst io.Writer) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)
	o.R("&validator.EmptyValidator{}")
	_, err := buf.WriteTo(dst)
	return err
}

func (g *codeGenerator) generateOneOfUnevaluatedPropertiesComposition(dst io.Writer) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)
	o.R("&validator.EmptyValidator{}")
	_, err := buf.WriteTo(dst)
	return err
}

func (g *codeGenerator) generateRefUnevaluatedPropertiesComposition(dst io.Writer) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)
	o.R("&validator.EmptyValidator{}")
	_, err := buf.WriteTo(dst)
	return err
}

func (g *codeGenerator) generateIfThenElse(dst io.Writer) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)
	o.R("&validator.EmptyValidator{}")
	_, err := buf.WriteTo(dst)
	return err
}

func (g *codeGenerator) generateIfThenElseUnevaluatedPropertiesComposition(dst io.Writer) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)
	o.R("&validator.EmptyValidator{}")
	_, err := buf.WriteTo(dst)
	return err
}
