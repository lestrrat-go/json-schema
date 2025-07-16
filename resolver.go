package schema

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/lestrrat-go/jsref/v2"
)

// Resolver provides JSON Schema reference resolution capabilities.
// It uses jsref for resolving JSON references within and across schemas.
type Resolver struct {
	resolver *jsref.StackedResolver
}

// NewResolver creates a new schema resolver with HTTP and filesystem support.
func NewResolver() *Resolver {
	resolver := jsref.New()

	// Add HTTP resolver for remote schema references
	resolver.AddResolver(jsref.NewHTTPResolver())

	// Add filesystem resolver rooted at current directory for local files
	if fsResolver, err := jsref.NewFSResolver("."); err == nil {
		resolver.AddResolver(fsResolver)
	}

	return &Resolver{resolver: resolver}
}

// ResolveJSONReference resolves JSON pointer references against a base schema.
// This method supports local JSON pointer references such as "#/$defs/person", relative references such as "person.json#/$defs/person", and absolute references such as "https://example.com/schemas/person.json#/$defs/person".
// This method only handles JSON pointer references, not anchor references.
func (r *Resolver) ResolveJSONReference(dst *Schema, baseSchema *Schema, reference string) error {
	// If the reference is a pure local JSON pointer reference (starts with #/)
	if len(reference) > 1 && reference[0] == '#' && reference[1] == '/' {
		// Convert base schema to interface{} for jsref resolution
		schemaData, err := r.schemaToData(baseSchema)
		if err != nil {
			return fmt.Errorf("failed to convert base schema to data: %w", err)
		}

		var resolved any
		if err := r.resolver.Resolve(&resolved, schemaData, reference); err != nil {
			return fmt.Errorf("failed to resolve local JSON pointer reference %s: %w", reference, err)
		}

		return r.dataToSchema(resolved, dst)
	}

	// For external references, split the reference and use our resolver
	external, local, err := jsref.Split(reference)
	if err != nil {
		return fmt.Errorf("failed to split reference %s: %w", reference, err)
	}

	// If no local reference is provided, default to root reference
	if local == "" {
		local = "#"
	}

	var resolved any
	if err := r.resolver.Resolve(&resolved, external, local); err != nil {
		return fmt.Errorf("failed to resolve external JSON pointer reference %s: %w", reference, err)
	}

	return r.dataToSchema(resolved, dst)
}

// ResolveAnchor resolves anchor references against a base schema.
// It searches for schemas with the specified $anchor value within the provided base schema.
// The anchorName parameter should not include the # prefix.
func (r *Resolver) ResolveAnchor(dst *Schema, baseSchema *Schema, anchorName string) error {
	anchorSchema, err := r.findSchemaByAnchor(baseSchema, anchorName)
	if err != nil {
		return fmt.Errorf("failed to find anchor %s: %w", anchorName, err)
	}
	*dst = *anchorSchema
	return nil
}

// ResolveReference resolves any type of JSON Schema reference against a base schema.
// It automatically dispatches to the appropriate resolver based on the reference format.
// Anchor references such as "#person" are handled by ResolveAnchor, while JSON pointer references such as "#/$defs/person" and external references such as "https://example.com/schema.json#..." are handled by ResolveJSONReference.
func (r *Resolver) ResolveReference(dst *Schema, baseSchema *Schema, reference string) error {
	return r.ResolveReferenceWithBaseURI(dst, baseSchema, reference, "")
}

// ResolveReferenceWithBaseURI resolves a reference with an optional base URI for relative reference resolution
func (r *Resolver) ResolveReferenceWithBaseURI(dst *Schema, baseSchema *Schema, reference string, baseURI string) error {
	// Check if this is an anchor reference (starts with # but no slash after)
	if len(reference) > 1 && reference[0] == '#' && reference[1] != '/' {
		anchorName := reference[1:] // Remove the '#' prefix
		return r.ResolveAnchor(dst, baseSchema, anchorName)
	}

	// Handle relative references with base URI
	resolvedReference := reference
	if baseURI != "" && !strings.HasPrefix(reference, "http://") && !strings.HasPrefix(reference, "https://") && !strings.HasPrefix(reference, "#") {
		// This is a relative reference that should be resolved against base URI
		if strings.HasSuffix(baseURI, "/") {
			resolvedReference = baseURI + reference
		} else {
			resolvedReference = baseURI + "/" + reference
		}
	}

	// Otherwise, treat as JSON pointer reference
	err := r.ResolveJSONReference(dst, baseSchema, resolvedReference)
	if err != nil {
		// Wrap the error with appropriate context based on reference type
		if strings.HasPrefix(resolvedReference, "#") {
			return fmt.Errorf("failed to resolve local reference %s: %w", resolvedReference, err)
		} else {
			return fmt.Errorf("failed to resolve external reference %s: %w", resolvedReference, err)
		}
	}
	return nil
}

// schemaToData converts a Schema to interface{} for jsref processing
func (r *Resolver) schemaToData(s *Schema) (any, error) {
	// Marshal schema to JSON, then unmarshal to interface{}
	jsonData, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	var data any
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema data: %w", err)
	}

	return data, nil
}

// dataToSchema converts resolved data back to a Schema
func (r *Resolver) dataToSchema(data any, dst *Schema) error {
	// Marshal the resolved data to JSON, then use Schema's UnmarshalJSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal resolved data: %w", err)
	}

	// Use Schema's UnmarshalJSON method which handles type field properly
	if err := dst.UnmarshalJSON(jsonData); err != nil {
		return fmt.Errorf("failed to unmarshal resolved data to schema: %w", err)
	}

	return nil
}

// ValidateReference checks if a reference string is valid.
// It validates the URI syntax and fragment format.
func ValidateReference(reference string) error {
	if reference == "" {
		return fmt.Errorf("reference cannot be empty")
	}

	// Try to split the reference to validate format
	_, _, err := jsref.Split(reference)
	if err != nil {
		return fmt.Errorf("invalid reference format: %w", err)
	}

	// If it contains a URI part, validate the URI syntax
	if len(reference) > 0 && reference[0] != '#' {
		external, _, _ := jsref.Split(reference)
		if external != "" {
			if _, err := url.Parse(external); err != nil {
				return fmt.Errorf("invalid URI in reference: %w", err)
			}
		}
	}

	return nil
}

// findSchemaByAnchor recursively searches for a schema with the given anchor name
func (r *Resolver) findSchemaByAnchor(schema *Schema, anchorName string) (*Schema, error) {
	// Check if current schema has the anchor
	if schema.HasAnchor() && schema.Anchor() == anchorName {
		return schema, nil
	}

	// Search in definitions, but be scope-aware
	if schema.HasDefinitions() {
		// First pass: search definitions that don't have their own $id (same scope)
		for _, defSchema := range schema.Definitions() {
			if !defSchema.HasID() {
				if found, err := r.findSchemaByAnchor(defSchema, anchorName); err == nil {
					return found, nil
				}
			}
		}

		// Second pass: search definitions that have their own $id (different scope)
		for _, defSchema := range schema.Definitions() {
			if defSchema.HasID() {
				if found, err := r.findSchemaByAnchor(defSchema, anchorName); err == nil {
					return found, nil
				}
			}
		}
	}

	// Search in properties
	if schema.HasProperties() {
		for _, propSchema := range schema.Properties() {
			if found, err := r.findSchemaByAnchor(propSchema, anchorName); err == nil {
				return found, nil
			}
		}
	}

	// Search in pattern properties
	if schema.HasPatternProperties() {
		for _, propSchema := range schema.PatternProperties() {
			if found, err := r.findSchemaByAnchor(propSchema, anchorName); err == nil {
				return found, nil
			}
		}
	}

	// Search in items
	if schema.HasItems() {
		if itemSchema, ok := schema.Items().(*Schema); ok {
			if found, err := r.findSchemaByAnchor(itemSchema, anchorName); err == nil {
				return found, nil
			}
		}
	}

	// Search in additional properties
	if schema.HasAdditionalProperties() {
		if addlSchema, ok := schema.AdditionalProperties().(*Schema); ok {
			if found, err := r.findSchemaByAnchor(addlSchema, anchorName); err == nil {
				return found, nil
			}
		}
	}

	// Search in unevaluated properties
	if schema.HasUnevaluatedProperties() {
		if unevalSchema, ok := schema.UnevaluatedProperties().(*Schema); ok {
			if found, err := r.findSchemaByAnchor(unevalSchema, anchorName); err == nil {
				return found, nil
			}
		}
	}

	// Search in unevaluated items
	if schema.HasUnevaluatedItems() {
		if unevalSchema, ok := schema.UnevaluatedItems().(*Schema); ok {
			if found, err := r.findSchemaByAnchor(unevalSchema, anchorName); err == nil {
				return found, nil
			}
		}
	}

	// Search in composition schemas
	if schema.HasAllOf() {
		for _, subSchema := range schema.AllOf() {
			if subSchema, ok := subSchema.(*Schema); ok {
				if found, err := r.findSchemaByAnchor(subSchema, anchorName); err == nil {
					return found, nil
				}
			}
		}
	}

	if schema.HasAnyOf() {
		for _, subSchema := range schema.AnyOf() {
			if subSchema, ok := subSchema.(*Schema); ok {
				if found, err := r.findSchemaByAnchor(subSchema, anchorName); err == nil {
					return found, nil
				}
			}
		}
	}

	if schema.HasOneOf() {
		for _, subSchema := range schema.OneOf() {
			if subSchema, ok := subSchema.(*Schema); ok {
				if found, err := r.findSchemaByAnchor(subSchema, anchorName); err == nil {
					return found, nil
				}
			}
		}
	}

	// Search in not schema
	if schema.HasNot() {
		if found, err := r.findSchemaByAnchor(schema.Not(), anchorName); err == nil {
			return found, nil
		}
	}

	// Search in if/then/else schemas
	if schema.HasIfSchema() {
		if found, err := r.findSchemaByAnchor(schema.IfSchema(), anchorName); err == nil {
			return found, nil
		}
	}

	if schema.HasThenSchema() {
		if found, err := r.findSchemaByAnchor(schema.ThenSchema(), anchorName); err == nil {
			return found, nil
		}
	}

	if schema.HasElseSchema() {
		if found, err := r.findSchemaByAnchor(schema.ElseSchema(), anchorName); err == nil {
			return found, nil
		}
	}

	// Search in contains schema
	if schema.HasContains() {
		if found, err := r.findSchemaByAnchor(schema.Contains(), anchorName); err == nil {
			return found, nil
		}
	}

	// Search in property names schema
	if schema.HasPropertyNames() {
		if found, err := r.findSchemaByAnchor(schema.PropertyNames(), anchorName); err == nil {
			return found, nil
		}
	}

	// Search in content schema
	if schema.HasContentSchema() {
		if found, err := r.findSchemaByAnchor(schema.ContentSchema(), anchorName); err == nil {
			return found, nil
		}
	}

	return nil, fmt.Errorf("anchor %s not found", anchorName)
}
