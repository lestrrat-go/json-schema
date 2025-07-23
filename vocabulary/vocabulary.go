package vocabulary

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/internal/schemactx"
	"github.com/lestrrat-go/json-schema/keywords"
)

// Vocabulary URI constants for JSON Schema 2020-12
const (
	CoreURL             = "https://json-schema.org/draft/2020-12/vocab/core"
	ApplicatorURL       = "https://json-schema.org/draft/2020-12/vocab/applicator"
	UnevaluatedURL      = "https://json-schema.org/draft/2020-12/vocab/unevaluated"
	ValidationURL       = "https://json-schema.org/draft/2020-12/vocab/validation"
	FormatAnnotationURL = "https://json-schema.org/draft/2020-12/vocab/format-annotation"
	FormatAssertionURL  = "https://json-schema.org/draft/2020-12/vocab/format-assertion"
	ContentURL          = "https://json-schema.org/draft/2020-12/vocab/content"
	MetaDataURL         = "https://json-schema.org/draft/2020-12/vocab/meta-data"
)

type Set2 struct {
	mu       sync.RWMutex
	uri      string
	keywords map[string]struct{}
	list     []string
}

func NewSet(uri string) *Set2 {
	return &Set2{
		uri: uri,
	}
}

func (s *Set2) URI() string {
	return s.uri
}

func (s *Set2) Keywords() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.list
}

func (s *Set2) KeywordExists(keyword string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.keywords[keyword]
	return exists
}

func (s *Set2) Add(keywords ...string) *Set2 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.keywords == nil {
		s.keywords = make(map[string]struct{})
	}

	for _, keyword := range keywords {
		if _, exists := s.keywords[keyword]; !exists {
			s.keywords[keyword] = struct{}{}
			s.list = append(s.list, keyword)
		}
	}

	return s
}

// Registry maps vocabulary URIs to their enabled keywords
type Registry struct {
	mu           sync.RWMutex
	vocabularies map[string]*Set2
}

func (r *Registry) Add(set *Set2) *Registry {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.vocabularies == nil {
		r.vocabularies = make(map[string]*Set2)
	}
	r.vocabularies[set.URI()] = set
	return r
}

func (r *Registry) Get(vocabularyURI string) *Set2 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.vocabularies == nil {
		return nil
	}
	return r.vocabularies[vocabularyURI]
}

var defaultRegistry *Registry

// DefaultRegistry contains the standard JSON Schema 2020-12 vocabularies
func DefaultRegistry() *Registry {
	return defaultRegistry
}

func init() {
	defaultRegistry = &Registry{}

	vocabularies := []*Set2{
		NewSet(CoreURL).Add(
			keywords.Schema, keywords.ID, keywords.Reference, keywords.DynamicReference, keywords.DynamicAnchor, keywords.Vocabulary, keywords.Comment, keywords.Definitions,
		),
		NewSet(ApplicatorURL).Add(
			keywords.AllOf, keywords.AnyOf, keywords.OneOf, keywords.Not, keywords.If, keywords.Then, keywords.Else, keywords.DependentSchemas,
			keywords.PrefixItems, keywords.Items, keywords.Contains, keywords.Properties, keywords.PatternProperties, keywords.AdditionalProperties,
			keywords.PropertyNames,
		),
		NewSet(UnevaluatedURL).Add(
			keywords.UnevaluatedItems, keywords.UnevaluatedProperties,
		),
		NewSet(ValidationURL).Add(
			keywords.Type, keywords.Enum, keywords.Const, keywords.MultipleOf, keywords.Maximum, keywords.ExclusiveMaximum, keywords.Minimum, keywords.ExclusiveMinimum,
			keywords.MaxLength, keywords.MinLength, keywords.Pattern, keywords.MaxItems, keywords.MinItems, keywords.UniqueItems, keywords.MaxContains, keywords.MinContains,
			keywords.MaxProperties, keywords.MinProperties, keywords.Required, keywords.DependentRequired,
		),
		NewSet(FormatAnnotationURL).Add(
			keywords.Format,
		),
		NewSet(FormatAssertionURL).Add(
			keywords.Format,
		),
		NewSet(ContentURL).Add(
			keywords.ContentEncoding, keywords.ContentMediaType, keywords.ContentSchema,
		),
		NewSet(MetaDataURL).Add(
			keywords.Title, keywords.Description, keywords.Default, keywords.Deprecated, keywords.ReadOnly, keywords.WriteOnly, keywords.Examples,
		),
	}
	for _, vocab := range vocabularies {
		defaultRegistry.Add(vocab)
	}
}

// GetKeywords returns the keywords for a vocabulary URI
func (vr *Registry) GetKeywords(vocabularyURI string) []string {
	if vr == nil || vr.vocabularies == nil {
		return nil
	}
	return vr.vocabularies[vocabularyURI].Keywords()
}

// IsKeywordInVocabulary checks if a keyword belongs to a vocabulary
func (vr *Registry) IsKeywordInVocabulary(vocabularyURI string, keyword string) bool {
	return vr.Get(vocabularyURI).KeywordExists(keyword)
}

// GetVocabularyForKeyword returns the vocabulary URI that contains the given keyword
func (vr *Registry) GetVocabularyForKeyword(keyword string) string {
	for vocabURI, set := range vr.vocabularies {
		if set.KeywordExists(keyword) {
			return vocabURI
		}
	}
	return ""
}

// VocabularySet represents a set of enabled vocabularies using Set2 objects
// This replaces the old Set map[string]bool approach with a structured approach
type VocabularySet struct {
	mu           sync.RWMutex
	enabled      map[string]bool  // vocabulary URI -> enabled status
	vocabularies map[string]*Set2 // vocabulary URI -> Set2 object (for keyword lookup)
}

// NewVocabularySet creates a new VocabularySet
func NewVocabularySet() *VocabularySet {
	return &VocabularySet{
		enabled:      make(map[string]bool),
		vocabularies: make(map[string]*Set2),
	}
}

// Enable enables a vocabulary by URI
func (vs *VocabularySet) Enable(vocabularyURI string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	vs.enabled[vocabularyURI] = true
	// Also store the Set2 object if available from registry
	if set2 := DefaultRegistry().Get(vocabularyURI); set2 != nil {
		vs.vocabularies[vocabularyURI] = set2
	}
}

// Disable disables a vocabulary by URI
func (vs *VocabularySet) Disable(vocabularyURI string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	vs.enabled[vocabularyURI] = false
	// Keep the Set2 object for potential re-enabling
	if set2 := DefaultRegistry().Get(vocabularyURI); set2 != nil {
		vs.vocabularies[vocabularyURI] = set2
	}
}

// IsEnabled checks if a vocabulary is enabled
func (vs *VocabularySet) IsEnabled(vocabularyURI string) bool {
	if vs == nil {
		return true // Default to enabled if no vocabulary set
	}
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	enabled, exists := vs.enabled[vocabularyURI]
	return !exists || enabled // Default to enabled if not explicitly set
}

// IsKeywordEnabled checks if a keyword is enabled based on vocabulary enablement
func (vs *VocabularySet) IsKeywordEnabled(keyword string) bool {
	if vs == nil {
		return true // Default to enabled if no vocabulary set
	}

	// Find which vocabulary contains this keyword
	vocabularyURI := DefaultRegistry().GetVocabularyForKeyword(keyword)
	if vocabularyURI == "" {
		return true // Unknown keywords are allowed by default
	}

	return vs.IsEnabled(vocabularyURI)
}

// AllEnabled returns a vocabulary set where all standard vocabularies are enabled
func AllEnabled() *VocabularySet {
	vs := NewVocabularySet()
	vs.Enable(CoreURL)
	vs.Enable(ApplicatorURL)
	vs.Enable(UnevaluatedURL)
	vs.Enable(ValidationURL)
	vs.Enable(FormatAnnotationURL)
	vs.Enable(FormatAssertionURL)
	vs.Enable(ContentURL)
	vs.Enable(MetaDataURL)
	return vs
}

// DefaultSet returns the default vocabulary set for JSON Schema 2020-12
// This is the vocabulary set that should be used when no explicit vocabulary is specified
// It includes annotation vocabularies but excludes assertion vocabularies like format-assertion
func DefaultSet() *VocabularySet {
	vs := NewVocabularySet()
	vs.Enable(CoreURL)
	vs.Enable(ApplicatorURL)
	vs.Enable(UnevaluatedURL)
	vs.Enable(ValidationURL)
	vs.Enable(FormatAnnotationURL)
	vs.Disable(FormatAssertionURL) // Disabled by default - format is annotation-only unless explicitly enabled
	vs.Enable(ContentURL)
	vs.Enable(MetaDataURL)
	return vs
}

// ExtractVocabularySet extracts the vocabulary set from a schema's $vocabulary declaration
func ExtractVocabularySet(s *schema.Schema) *VocabularySet {
	if s == nil || !s.HasVocabulary() {
		return AllEnabled() // Default to all enabled if no vocabulary declaration
	}

	vocabMap := s.Vocabulary()
	if vocabMap == nil {
		return AllEnabled()
	}

	vs := NewVocabularySet()
	for uri, enabled := range vocabMap {
		if enabled {
			vs.Enable(uri)
		} else {
			vs.Disable(uri)
		}
	}

	return vs
}

// ResolveVocabularyFromMetaschema resolves the vocabulary set from a metaschema
func ResolveVocabularyFromMetaschema(ctx context.Context, metaschemaURI string) (*VocabularySet, error) {
	if metaschemaURI == "" {
		return AllEnabled(), nil
	}

	resolver := schema.ResolverFromContext(ctx)
	if resolver == nil {
		resolver = schema.NewResolver()
	}

	rootSchema := schema.RootSchemaFromContext(ctx)
	if rootSchema == nil {
		return AllEnabled(), nil
	}

	// Try to resolve the metaschema
	var metaschema schema.Schema
	// Create context with base schema for resolver
	resolverCtx := schema.WithBaseSchema(ctx, rootSchema)
	if err := resolver.ResolveReference(resolverCtx, &metaschema, metaschemaURI); err != nil {
		// If we can't resolve the metaschema, default to all enabled
		return AllEnabled(), nil //nolint:nilerr // Intentional: fallback to default behavior on resolve error
	}

	return ExtractVocabularySet(&metaschema), nil
}

// Context keys for vocabulary support
// WithSet adds a vocabulary set to the context
func WithSet(ctx context.Context, vocabSet *VocabularySet) context.Context {
	return schemactx.WithVocabularySet(ctx, vocabSet)
}

// SetFromContext extracts the vocabulary set from the context
func SetFromContext(ctx context.Context) *VocabularySet {
	var vocabSet *VocabularySet
	if err := schemactx.VocabularySetFromContext(ctx, &vocabSet); err != nil {
		return DefaultSet() // Use default vocabulary set with format-assertion disabled
	}
	return vocabSet
}

// IsKeywordEnabledInContext checks if a keyword is enabled in the current context
func IsKeywordEnabledInContext(ctx context.Context, keyword string) bool {
	vocabSet := SetFromContext(ctx)
	return vocabSet.IsKeywordEnabled(keyword)
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
