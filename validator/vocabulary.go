package validator

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	schema "github.com/lestrrat-go/json-schema"
)

// VocabularyRegistry maps vocabulary URIs to their enabled keywords
type VocabularyRegistry struct {
	vocabularies map[string][]string
}

// DefaultVocabularyRegistry contains the standard JSON Schema 2020-12 vocabularies
var DefaultVocabularyRegistry = &VocabularyRegistry{
	vocabularies: map[string][]string{
		"https://json-schema.org/draft/2020-12/vocab/core": {
			"$schema", "$id", "$ref", "$dynamicRef", "$dynamicAnchor", "$vocabulary", "$comment", "$defs",
		},
		"https://json-schema.org/draft/2020-12/vocab/applicator": {
			"allOf", "anyOf", "oneOf", "not", "if", "then", "else", "dependentSchemas",
			"prefixItems", "items", "contains", "properties", "patternProperties", "additionalProperties",
			"propertyNames",
		},
		"https://json-schema.org/draft/2020-12/vocab/unevaluated": {
			"unevaluatedItems", "unevaluatedProperties",
		},
		"https://json-schema.org/draft/2020-12/vocab/validation": {
			"type", "enum", "const", "multipleOf", "maximum", "exclusiveMaximum", "minimum", "exclusiveMinimum",
			"maxLength", "minLength", "pattern", "maxItems", "minItems", "uniqueItems", "maxContains", "minContains",
			"maxProperties", "minProperties", "required", "dependentRequired",
		},
		"https://json-schema.org/draft/2020-12/vocab/format-annotation": {
			"format",
		},
		"https://json-schema.org/draft/2020-12/vocab/format-assertion": {
			"format",
		},
		"https://json-schema.org/draft/2020-12/vocab/content": {
			"contentEncoding", "contentMediaType", "contentSchema",
		},
		"https://json-schema.org/draft/2020-12/vocab/meta-data": {
			"title", "description", "default", "deprecated", "readOnly", "writeOnly", "examples",
		},
	},
}

// GetKeywords returns the keywords for a vocabulary URI
func (vr *VocabularyRegistry) GetKeywords(vocabularyURI string) []string {
	if vr == nil || vr.vocabularies == nil {
		return nil
	}
	return vr.vocabularies[vocabularyURI]
}

// IsKeywordInVocabulary checks if a keyword belongs to a vocabulary
func (vr *VocabularyRegistry) IsKeywordInVocabulary(vocabularyURI string, keyword string) bool {
	keywords := vr.GetKeywords(vocabularyURI)
	for _, k := range keywords {
		if k == keyword {
			return true
		}
	}
	return false
}

// GetVocabularyForKeyword returns the vocabulary URI that contains the given keyword
func (vr *VocabularyRegistry) GetVocabularyForKeyword(keyword string) string {
	if vr == nil || vr.vocabularies == nil {
		return ""
	}
	for vocabURI, keywords := range vr.vocabularies {
		for _, k := range keywords {
			if k == keyword {
				return vocabURI
			}
		}
	}
	return ""
}

// VocabularySet represents a set of enabled vocabularies
type VocabularySet map[string]bool

// IsEnabled checks if a vocabulary is enabled
func (vs VocabularySet) IsEnabled(vocabularyURI string) bool {
	if vs == nil {
		return true // Default to enabled if no vocabulary set
	}
	enabled, exists := vs[vocabularyURI]
	return exists && enabled
}

// IsKeywordEnabled checks if a keyword is enabled based on vocabulary enablement
func (vs VocabularySet) IsKeywordEnabled(keyword string) bool {
	if vs == nil {
		return true // Default to enabled if no vocabulary set
	}
	
	// Find which vocabulary contains this keyword
	vocabularyURI := DefaultVocabularyRegistry.GetVocabularyForKeyword(keyword)
	if vocabularyURI == "" {
		return true // Unknown keywords are allowed by default
	}
	
	return vs.IsEnabled(vocabularyURI)
}

// AllEnabled returns a vocabulary set where all standard vocabularies are enabled
func AllEnabled() VocabularySet {
	return VocabularySet{
		"https://json-schema.org/draft/2020-12/vocab/core":             true,
		"https://json-schema.org/draft/2020-12/vocab/applicator":       true,
		"https://json-schema.org/draft/2020-12/vocab/unevaluated":      true,
		"https://json-schema.org/draft/2020-12/vocab/validation":       true,
		"https://json-schema.org/draft/2020-12/vocab/format-annotation": true,
		"https://json-schema.org/draft/2020-12/vocab/format-assertion":  true,
		"https://json-schema.org/draft/2020-12/vocab/content":          true,
		"https://json-schema.org/draft/2020-12/vocab/meta-data":        true,
	}
}

// ExtractVocabularySet extracts the vocabulary set from a schema's $vocabulary declaration
func ExtractVocabularySet(s *schema.Schema) VocabularySet {
	if s == nil || !s.HasVocabulary() {
		return AllEnabled() // Default to all enabled if no vocabulary declaration
	}
	
	vocabMap := s.Vocabulary()
	if vocabMap == nil {
		return AllEnabled()
	}
	
	result := make(VocabularySet)
	for uri, enabled := range vocabMap {
		result[uri] = enabled
	}
	
	return result
}

// ResolveVocabularyFromMetaschema resolves the vocabulary set from a metaschema
func ResolveVocabularyFromMetaschema(ctx context.Context, metaschemaURI string) (VocabularySet, error) {
	if metaschemaURI == "" {
		return AllEnabled(), nil
	}
	
	resolver := ResolverFromContext(ctx)
	if resolver == nil {
		resolver = schema.NewResolver()
	}
	
	rootSchema := RootSchemaFromContext(ctx)
	if rootSchema == nil {
		return AllEnabled(), nil
	}
	
	// Try to resolve the metaschema
	var metaschema schema.Schema
	if err := resolver.ResolveReference(&metaschema, rootSchema, metaschemaURI); err != nil {
		// If we can't resolve the metaschema, default to all enabled
		return AllEnabled(), nil
	}
	
	return ExtractVocabularySet(&metaschema), nil
}

// Context keys for vocabulary support
type contextKey string

const (
	vocabularySetKey contextKey = "vocabularySet"
)

// WithVocabularySet adds a vocabulary set to the context
func WithVocabularySet(ctx context.Context, vocabSet VocabularySet) context.Context {
	return context.WithValue(ctx, vocabularySetKey, vocabSet)
}

// VocabularySetFromContext extracts the vocabulary set from the context
func VocabularySetFromContext(ctx context.Context) VocabularySet {
	if vs, ok := ctx.Value(vocabularySetKey).(VocabularySet); ok {
		return vs
	}
	return AllEnabled() // Default to all enabled
}

// IsKeywordEnabledInContext checks if a keyword is enabled in the current context
func IsKeywordEnabledInContext(ctx context.Context, keyword string) bool {
	vocabSet := VocabularySetFromContext(ctx)
	return vocabSet.IsKeywordEnabled(keyword)
}

// normalizeVocabularyURI normalizes a vocabulary URI for comparison
func normalizeVocabularyURI(uri string) string {
	// Remove trailing slash if present
	return strings.TrimSuffix(uri, "/")
}

// ValidateVocabularyURI validates that a vocabulary URI is well-formed
func ValidateVocabularyURI(uri string) error {
	if uri == "" {
		return fmt.Errorf("vocabulary URI cannot be empty")
	}
	
	_, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("invalid vocabulary URI: %w", err)
	}
	
	return nil
}