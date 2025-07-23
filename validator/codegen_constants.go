package validator

import "github.com/lestrrat-go/json-schema/keywords"

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
	// Return quoted string for non-standard properties
	return `"` + propName + `"`
}