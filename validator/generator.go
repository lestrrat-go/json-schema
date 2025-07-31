package validator

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/lestrrat-go/codegen"
)

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
			o.L("AdditionalProperties(%t).", ap)
		case Interface:
			o.L("AdditionalProperties(")
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
		o.L("PatternProperties(")
		o.L("func() map[*regexp.Regexp]validator.Interface {")
		o.L("patternProps := make(map[*regexp.Regexp]validator.Interface)")

		patternIndex := 0
		for pattern, patternValidator := range v.patternProperties {
			// Generate unique variable names for each pattern
			validatorVar := fmt.Sprintf("patternValidator%d", patternIndex)
			regexVar := fmt.Sprintf("patternRegex%d", patternIndex)

			// Generate the validator for this pattern
			o.L("%s := ", validatorVar)
			if err := g.Generate(&buf, patternValidator); err != nil {
				return fmt.Errorf("failed to generate pattern property validator for %s: %w", pattern.String(), err)
			}
			o.R("")

			// Generate the regex compilation
			patternStr := pattern.String()
			o.L("%s, _ := regexp.Compile(%q)", regexVar, patternStr)
			o.L("patternProps[%s] = %s", regexVar, validatorVar)

			patternIndex++
		}

		o.L("return patternProps")
		o.L("}(),")
		o.L(").")
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

func (g *codeGenerator) generateAllOf(dst io.Writer, v *allOfValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	o.L("validator.AllOf(")
	for i, child := range v.validators {
		if err := g.Generate(&buf, child); err != nil {
			return fmt.Errorf("failed to generate child validator %d: %w", i, err)
		}
		o.R(",")
	}
	o.L(")")

	_, err := buf.WriteTo(dst)
	return err
}

func (g *codeGenerator) generateAnyOf(dst io.Writer, v *anyOfValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	o.L("validator.AnyOf(")
	for i, child := range v.validators {
		if err := g.Generate(&buf, child); err != nil {
			return fmt.Errorf("failed to generate child validator %d: %w", i, err)
		}
		o.R(",")
	}
	o.L(")")

	_, err := buf.WriteTo(dst)
	return err
}

func (g *codeGenerator) generateOneOf(dst io.Writer, v *oneOfValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	o.L("validator.OneOf(")
	for i, child := range v.validators {
		if err := g.Generate(&buf, child); err != nil {
			return fmt.Errorf("failed to generate child validator %d: %w", i, err)
		}
		o.R(",")
	}
	o.L(")")

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

// generateUntyped creates the builder chain for untyped validators
func (g *codeGenerator) generateUntyped(dst io.Writer, v *untypedValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	if v.constantValue != nil {
		// For const validation, use the public builder API
		o.L("validator.Untyped().Const(%#v).MustBuild()", *v.constantValue)
	} else if len(v.enum) > 0 {
		// For enum validation, use the public builder API
		if len(v.enum) == 1 {
			o.L("validator.Untyped().Enum(%#v).MustBuild()", v.enum[0])
		} else {
			o.L("validator.Untyped().Enum(")
			for _, e := range v.enum {
				o.L("\t%#v,", e)
			}
			o.L(").MustBuild()")
		}
	} else {
		o.L("&validator.EmptyValidator{}")
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
			o.L("Enum(%#v).", v.enum[0])
		} else {
			o.L("Enum(")
			for _, s := range v.enum {
				o.L("%#v,", s)
			}
			o.L(").")
		}
	}
	if v.constantValue != nil {
		o.L("Const(%#v).", v.constantValue)
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
			o.L("Enum(%#v).", v.enum[0])
		} else {
			o.L("Enum(")
			for _, b := range v.enum {
				o.L("%#v,", b)
			}
			o.L(").")
		}
	}
	if v.constantValue != nil {
		o.L("Const(%#v).", v.constantValue)
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

	// Handle prefixItems
	if len(v.prefixItems) > 0 {
		o.L("PrefixItems(")
		for i, prefixItem := range v.prefixItems {
			if i > 0 {
				o.R(",")
			}
			if err := g.Generate(&buf, prefixItem); err != nil {
				return fmt.Errorf("failed to generate prefixItem %d: %w", i, err)
			}
		}
		o.L(").")
	}

	// Handle items validator
	if v.items != nil {
		o.L("Items(")
		if err := g.Generate(&buf, v.items); err != nil {
			return fmt.Errorf("failed to generate items validator: %w", err)
		}
		o.R(").")
	}

	// Handle additionalItems
	if v.additionalItems != nil {
		o.L("AdditionalItems(")
		if err := g.Generate(&buf, v.additionalItems); err != nil {
			return fmt.Errorf("failed to generate additionalItems validator: %w", err)
		}
		o.R(").")
	}

	// Handle contains validator
	if v.contains != nil {
		o.L("Contains(")
		if err := g.Generate(&buf, v.contains); err != nil {
			return fmt.Errorf("failed to generate contains validator: %w", err)
		}
		o.R(").")
	}

	// Handle unevaluatedItems (can be bool or Interface)
	if v.unevaluatedItems != nil {
		switch ui := v.unevaluatedItems.(type) {
		case bool:
			o.L("UnevaluatedItemsBool(%t).", ui)
		case Interface:
			o.L("UnevaluatedItemsSchema(")
			if err := g.Generate(&buf, ui); err != nil {
				return fmt.Errorf("failed to generate unevaluatedItems validator: %w", err)
			}
			o.R(").")
		default:
			// Fallback for unexpected type
			o.L("UnevaluatedItemsBool(true).")
		}
	}

	// Handle strict array type
	if v.strictArrayType {
		o.L("StrictArrayType(true).")
	}

	o.L("MustBuild()")
	_, err := buf.WriteTo(dst)
	return err
}

// generateNull creates the builder chain for null validators
func (g *codeGenerator) generateNull(dst io.Writer) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	o.R("validator.Null()")

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

func (g *codeGenerator) generateUnevaluatedPropertiesComposition(dst io.Writer, v *unevaluatedPropertiesValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	// UnevaluatedPropertiesComposition validators have allOf validators and a base validator
	o.L("func() validator.Interface {")

	// Generate allOf validators
	if len(v.allOfValidators) > 0 {
		o.L("allOfValidators := make([]validator.Interface, %d)", len(v.allOfValidators))
		for i, allOfValidator := range v.allOfValidators {
			o.L("allOfValidators[%d] = ", i)
			if err := g.Generate(&buf, allOfValidator); err != nil {
				return fmt.Errorf("failed to generate allOf validator %d: %w", i, err)
			}
			o.R("")
		}
	} else {
		o.L("allOfValidators := make([]validator.Interface, 0)")
	}

	// Generate base validator
	if v.baseValidator != nil {
		o.L("baseValidator := ")
		if err := g.Generate(&buf, v.baseValidator); err != nil {
			return fmt.Errorf("failed to generate base validator: %w", err)
		}
		o.R("")
	} else {
		o.L("baseValidator := &validator.EmptyValidator{}")
	}

	// Create the UnevaluatedPropertiesCompositionValidator
	o.L("return &validator.UnevaluatedPropertiesCompositionValidator{")
	o.L("allOfValidators: allOfValidators,")
	o.L("baseValidator: baseValidator,")
	o.L("schema: nil, // Schema not available during generation")
	o.L("}")
	o.L("}()")

	_, err := buf.WriteTo(dst)
	return err
}

func (g *codeGenerator) generateAnyOfUnevaluatedPropertiesComposition(dst io.Writer, v *AnyOfUnevaluatedPropertiesCompositionValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	// AnyOfUnevaluatedPropertiesComposition validators have anyOf validators and a base validator
	o.L("func() validator.Interface {")

	// Generate anyOf validators
	if len(v.anyOfValidators) > 0 {
		o.L("anyOfValidators := make([]validator.Interface, %d)", len(v.anyOfValidators))
		for i, anyOfValidator := range v.anyOfValidators {
			o.L("anyOfValidators[%d] = ", i)
			if err := g.Generate(&buf, anyOfValidator); err != nil {
				return fmt.Errorf("failed to generate anyOf validator %d: %w", i, err)
			}
			o.R("")
		}
	} else {
		o.L("anyOfValidators := make([]validator.Interface, 0)")
	}

	// Generate base validator
	if v.baseValidator != nil {
		o.L("baseValidator := ")
		if err := g.Generate(&buf, v.baseValidator); err != nil {
			return fmt.Errorf("failed to generate base validator: %w", err)
		}
		o.R("")
	} else {
		o.L("baseValidator := &validator.EmptyValidator{}")
	}

	// Create the AnyOfUnevaluatedPropertiesCompositionValidator
	o.L("return &validator.AnyOfUnevaluatedPropertiesCompositionValidator{")
	o.L("anyOfValidators: anyOfValidators,")
	o.L("baseValidator: baseValidator,")
	o.L("schema: nil, // Schema not available during generation")
	o.L("}")
	o.L("}()")

	_, err := buf.WriteTo(dst)
	return err
}

func (g *codeGenerator) generateOneOfUnevaluatedPropertiesComposition(dst io.Writer, v *OneOfUnevaluatedPropertiesCompositionValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	// OneOfUnevaluatedPropertiesComposition validators have oneOf validators and a base validator
	o.L("func() validator.Interface {")

	// Generate oneOf validators
	if len(v.oneOfValidators) > 0 {
		o.L("oneOfValidators := make([]validator.Interface, %d)", len(v.oneOfValidators))
		for i, oneOfValidator := range v.oneOfValidators {
			o.L("oneOfValidators[%d] = ", i)
			if err := g.Generate(&buf, oneOfValidator); err != nil {
				return fmt.Errorf("failed to generate oneOf validator %d: %w", i, err)
			}
			o.R("")
		}
	} else {
		o.L("oneOfValidators := make([]validator.Interface, 0)")
	}

	// Generate base validator
	if v.baseValidator != nil {
		o.L("baseValidator := ")
		if err := g.Generate(&buf, v.baseValidator); err != nil {
			return fmt.Errorf("failed to generate base validator: %w", err)
		}
		o.R("")
	} else {
		o.L("baseValidator := &validator.EmptyValidator{}")
	}

	// Create the OneOfUnevaluatedPropertiesCompositionValidator
	o.L("return &validator.OneOfUnevaluatedPropertiesCompositionValidator{")
	o.L("oneOfValidators: oneOfValidators,")
	o.L("baseValidator: baseValidator,")
	o.L("schema: nil, // Schema not available during generation")
	o.L("}")
	o.L("}()")

	_, err := buf.WriteTo(dst)
	return err
}

func (g *codeGenerator) generateRefUnevaluatedPropertiesComposition(dst io.Writer, v *RefUnevaluatedPropertiesCompositionValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	// RefUnevaluatedPropertiesComposition validators have a ref validator and a base validator
	o.L("func() validator.Interface {")

	// Generate ref validator
	if v.refValidator != nil {
		o.L("refValidator := ")
		if err := g.Generate(&buf, v.refValidator); err != nil {
			return fmt.Errorf("failed to generate ref validator: %w", err)
		}
		o.R("")
	} else {
		o.L("refValidator := &validator.EmptyValidator{}")
	}

	// Generate base validator
	if v.baseValidator != nil {
		o.L("baseValidator := ")
		if err := g.Generate(&buf, v.baseValidator); err != nil {
			return fmt.Errorf("failed to generate base validator: %w", err)
		}
		o.R("")
	} else {
		o.L("baseValidator := &validator.EmptyValidator{}")
	}

	// Create the RefUnevaluatedPropertiesCompositionValidator
	o.L("return &validator.RefUnevaluatedPropertiesCompositionValidator{")
	o.L("refValidator: refValidator,")
	o.L("baseValidator: baseValidator,")
	o.L("schema: nil, // Schema not available during generation")
	o.L("}")
	o.L("}()")

	_, err := buf.WriteTo(dst)
	return err
}

func (g *codeGenerator) generateIfThenElse(dst io.Writer, v *IfThenElseValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	// IfThenElse validators are complex composites with three child validators
	o.L("func() validator.Interface {")

	// Generate if validator
	if v.ifValidator != nil {
		o.L("ifValidator := ")
		if err := g.Generate(&buf, v.ifValidator); err != nil {
			return fmt.Errorf("failed to generate if validator: %w", err)
		}
		o.R("")
	} else {
		o.L("ifValidator := &validator.EmptyValidator{}")
	}

	// Generate then validator
	if v.thenValidator != nil {
		o.L("thenValidator := ")
		if err := g.Generate(&buf, v.thenValidator); err != nil {
			return fmt.Errorf("failed to generate then validator: %w", err)
		}
		o.R("")
	} else {
		o.L("thenValidator := &validator.EmptyValidator{}")
	}

	// Generate else validator
	if v.elseValidator != nil {
		o.L("elseValidator := ")
		if err := g.Generate(&buf, v.elseValidator); err != nil {
			return fmt.Errorf("failed to generate else validator: %w", err)
		}
		o.R("")
	} else {
		o.L("elseValidator := &validator.EmptyValidator{}")
	}

	// Create the IfThenElse validator
	o.L("return &validator.IfThenElseValidator{")
	o.L("ifValidator: ifValidator,")
	o.L("thenValidator: thenValidator,")
	o.L("elseValidator: elseValidator,")
	o.L("}")
	o.L("}()")

	_, err := buf.WriteTo(dst)
	return err
}

func (g *codeGenerator) generateIfThenElseUnevaluatedPropertiesComposition(dst io.Writer, v *IfThenElseUnevaluatedPropertiesCompositionValidator) error {
	var buf bytes.Buffer
	o := codegen.NewOutput(&buf)

	// IfThenElseUnevaluatedPropertiesComposition validators have if/then/else validators and a base validator
	o.L("func() validator.Interface {")

	// Generate if validator
	if v.ifValidator != nil {
		o.L("ifValidator := ")
		if err := g.Generate(&buf, v.ifValidator); err != nil {
			return fmt.Errorf("failed to generate if validator: %w", err)
		}
		o.R("")
	} else {
		o.L("ifValidator := &validator.EmptyValidator{}")
	}

	// Generate then validator
	if v.thenValidator != nil {
		o.L("thenValidator := ")
		if err := g.Generate(&buf, v.thenValidator); err != nil {
			return fmt.Errorf("failed to generate then validator: %w", err)
		}
		o.R("")
	} else {
		o.L("thenValidator := &validator.EmptyValidator{}")
	}

	// Generate else validator
	if v.elseValidator != nil {
		o.L("elseValidator := ")
		if err := g.Generate(&buf, v.elseValidator); err != nil {
			return fmt.Errorf("failed to generate else validator: %w", err)
		}
		o.R("")
	} else {
		o.L("elseValidator := &validator.EmptyValidator{}")
	}

	// Generate base validator
	if v.baseValidator != nil {
		o.L("baseValidator := ")
		if err := g.Generate(&buf, v.baseValidator); err != nil {
			return fmt.Errorf("failed to generate base validator: %w", err)
		}
		o.R("")
	} else {
		o.L("baseValidator := &validator.EmptyValidator{}")
	}

	// Create the IfThenElseUnevaluatedPropertiesCompositionValidator
	o.L("return &validator.IfThenElseUnevaluatedPropertiesCompositionValidator{")
	o.L("ifValidator: ifValidator,")
	o.L("thenValidator: thenValidator,")
	o.L("elseValidator: elseValidator,")
	o.L("baseValidator: baseValidator,")
	o.L("schema: nil, // Schema not available during generation")
	o.L("}")
	o.L("}()")

	_, err := buf.WriteTo(dst)
	return err
}
