package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/url"
	"strings"
	"sync"

	"github.com/lestrrat-go/jsref/v2"
	"github.com/lestrrat-go/option/v3"
)

// Resolver provides JSON Schema reference resolution capabilities.
// It uses jsref for resolving JSON references within and across schemas.
type Resolver struct {
	resolver *jsref.StackedResolver
	index    *resourceIndex

	mu         sync.Mutex
	registered map[*Schema]struct{} // roots already indexed, to avoid re-indexing
}

// NewResolver creates a new schema resolver. By default it resolves references
// only from memory: in-document $id resources/anchors and documents preloaded
// via RegisterRoot/RegisterDocument/RegisterFS.
//
// Access to system resources is opt-in. To let references reach the network or
// the filesystem, pass resolvers explicitly with WithResolver — for example
// WithResolver(HTTPResolver()) for HTTP/HTTPS or WithResolver(DirResolver("."))
// for local files. Without those, an external $ref that is not preloaded fails
// to resolve rather than silently fetching it.
func NewResolver(options ...ResolverOption) *Resolver {
	resolver := jsref.New()

	index := newResourceIndex()

	// Add the in-document resolver FIRST so references whose base URI names a
	// known $id resource are resolved from memory before any opt-in network/
	// filesystem access is attempted. It declines unknown URIs, falling through
	// to the resolvers below.
	resolver.AddResolver(&registryResolver{idx: index, obj: jsref.NewObjectResolver()})

	// Add caller-supplied resolvers (HTTP, filesystem, custom) in order. These
	// sit between the in-document registry and the final JSON-pointer fallback.
	for _, o := range options {
		if o.Ident() == (identResolver{}) {
			resolver.AddResolver(option.MustGet[jsref.Resolver](o))
		}
	}

	// Add object resolver LAST for JSON pointer resolution within map data
	// structures. Its CanResolve accepts everything, so it must stay last or it
	// would shadow the resolvers above.
	resolver.AddResolver(jsref.NewObjectResolver())

	return &Resolver{resolver: resolver, index: index}
}

// RegisterRoot indexes the $id resources and anchors reachable from root so
// that in-document references resolve without external retrieval. The root's
// own $id (if any) establishes the base URI for the document. It is safe to
// call multiple times and across multiple documents; entries accumulate.
func (r *Resolver) RegisterRoot(root *Schema) {
	if r.index == nil || root == nil {
		return
	}
	// Index each root at most once. Compilation happens before validation, so
	// once indexed there are no further writes to the index; this also prevents
	// a data race where validation-time compilation of a target that happens to
	// be the root would re-index concurrently with registry lookups.
	r.mu.Lock()
	if r.registered == nil {
		r.registered = make(map[*Schema]struct{})
	}
	if _, done := r.registered[root]; done {
		r.mu.Unlock()
		return
	}
	r.registered[root] = struct{}{}
	base := ""
	if root.HasID() && root.ID() != "" {
		base, _, _ = splitFragment(root.ID())
	}
	r.index.index(root, base, make(map[*Schema]struct{}))
	r.mu.Unlock()
}

// RegisterDocument registers a document retrieved from (or identified by) an
// explicit URI, so references to that URI resolve from memory instead of being
// fetched. The document is addressable both by uri and by its own canonical $id
// (resolved against uri), and its nested $id resources and anchors are indexed.
// This is how callers preload remote/bundled schemas without network access.
func (r *Resolver) RegisterDocument(uri string, root *Schema) {
	if r.index == nil || root == nil || uri == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.registered == nil {
		r.registered = make(map[*Schema]struct{})
	}
	r.registered[root] = struct{}{}

	retrieval, _, _ := splitFragment(uri)
	base := retrieval
	if root.HasID() && root.ID() != "" {
		base, _, _ = splitFragment(resolveURI(retrieval, root.ID()))
	}
	// Address the document by its retrieval URI even when it has no $id of its own.
	r.index.byURI[retrieval] = root
	r.index.index(root, base, make(map[*Schema]struct{}))
}

// RegisterFS walks fsys and registers every ".json" file as a document via
// RegisterDocument, addressed at baseURI joined with the file's (slash-separated)
// path. It lets a whole tree of schemas — an embed.FS, os.DirFS, zip, etc. — be
// preloaded for offline resolution in one call. Files that do not parse as an
// object schema (e.g. a boolean schema document) are skipped; read/walk errors
// are returned.
func (r *Resolver) RegisterFS(baseURI string, fsys fs.FS) error {
	prefix := strings.TrimSuffix(baseURI, "/")
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		var s Schema
		if s.UnmarshalJSON(data) != nil {
			return nil //nolint:nilerr // not an object schema (e.g. boolean document): skip it
		}
		r.RegisterDocument(prefix+"/"+path, &s)
		return nil
	})
}

// ResourceFor returns the schema resource registered under the given absolute
// base URI (no fragment), or nil if none. It lets reference resolution record
// which resource an instance enters for $dynamicRef dynamic-scope tracking.
func (r *Resolver) ResourceFor(uri string) *Schema {
	if r.index == nil || uri == "" {
		return nil
	}
	base, _, _ := splitFragment(uri)
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.index.byURI[base]
}

// ResolveJSONReference resolves JSON pointer references against the given base schema.
// This method supports local JSON pointer references such as "#/$defs/person", relative references such as "person.json#/$defs/person", and absolute references such as "https://example.com/schemas/person.json#/$defs/person".
// This method only handles JSON pointer references, not anchor references.
func (r *Resolver) ResolveJSONReference(_ context.Context, dst *Schema, reference string, baseSchema *Schema) error {
	// If the reference is a pure local reference into the current document:
	// either a JSON pointer ("#/...") or a bare "#" denoting the document root.
	if reference == "#" || (len(reference) > 1 && reference[0] == '#' && reference[1] == '/') {
		if baseSchema == nil {
			return fmt.Errorf("no base schema provided in context for resolving local reference %s", reference)
		}

		// Convert base schema to any for jsref resolution
		schemaData, err := r.schemaToData(baseSchema)
		if err != nil {
			return fmt.Errorf("failed to convert base schema to data: %w", err)
		}

		// Percent-decode the fragment so JSON Pointer lookups operate on the
		// decoded key (e.g. "%25" -> "%"). JSON Pointer "~0"/"~1" escaping is
		// left intact for the pointer evaluator.
		localRef := "#" + unescapeFragment(reference[1:])

		var resolved any
		if err := r.resolver.Resolve(&resolved, schemaData, localRef); err != nil {
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
	// Percent-decode the fragment so JSON Pointer / anchor lookups operate on the
	// decoded value (e.g. "%25" -> "%"). JSON Pointer "~0"/"~1" escaping is left
	// intact for the pointer evaluator.
	local = "#" + unescapeFragment(strings.TrimPrefix(local, "#"))

	var resolved any
	if err := r.resolver.Resolve(&resolved, external, local); err != nil {
		return fmt.Errorf("failed to resolve external JSON pointer reference %s: %w", reference, err)
	}

	return r.dataToSchema(resolved, dst)
}

// ResolveAnchor resolves anchor references against the given base schema.
// It searches for schemas with the specified $anchor value within the base schema.
// The anchorName parameter should not include the # prefix.
func (r *Resolver) ResolveAnchor(_ context.Context, dst *Schema, anchorName string, baseSchema *Schema) error {
	if baseSchema == nil {
		return fmt.Errorf("no base schema provided for resolving anchor %s", anchorName)
	}

	anchorSchema, err := r.findSchemaByAnchor(baseSchema, anchorName)
	if err != nil {
		return fmt.Errorf("failed to find anchor %s: %w", anchorName, err)
	}
	*dst = *anchorSchema
	return nil
}

// ResolveReference resolves any type of JSON Schema reference against the given base schema.
// It automatically dispatches to the appropriate resolver based on the reference format.
// Anchor references such as "#person" are handled by ResolveAnchor, while JSON pointer references such as "#/$defs/person" and external references such as "https://example.com/schema.json#..." are handled by ResolveJSONReference.
// baseURI is used to resolve relative references to an absolute URI.
func (r *Resolver) ResolveReference(ctx context.Context, dst *Schema, reference string, baseSchema *Schema, baseURI string) error {
	// Check if this is an anchor reference (starts with # but no slash after)
	if len(reference) > 1 && reference[0] == '#' && reference[1] != '/' {
		anchorName := reference[1:] // Remove the '#' prefix
		return r.ResolveAnchor(ctx, dst, anchorName, baseSchema)
	}

	// Pure local references ("#" or "#/...") resolve against the base schema;
	// leave them untouched.
	resolvedReference := reference
	if reference != "#" && !strings.HasPrefix(reference, "#/") {
		// A reference with a URI part: resolve it against the current base URI
		// (RFC 3986) to obtain an absolute URI. The in-document registry resolver
		// can then recognize it as a known $id resource before any external
		// retrieval is attempted; if it is genuinely external the absolute form
		// is what HTTP/filesystem resolution needs anyway.
		if baseURI != "" {
			resolvedReference = resolveURI(baseURI, reference)
		}
	}

	// Otherwise, treat as JSON pointer reference
	err := r.ResolveJSONReference(ctx, dst, resolvedReference, baseSchema)
	if err != nil {
		// Wrap the error with appropriate context based on reference type
		if strings.HasPrefix(resolvedReference, "#") {
			return fmt.Errorf("failed to resolve local reference %s: %w", resolvedReference, err)
		}
		return fmt.Errorf("failed to resolve external reference %s: %w", resolvedReference, err)
	}
	return nil
}

// schemaToData converts a Schema to any for jsref processing
func (r *Resolver) schemaToData(s *Schema) (any, error) {
	// Marshal schema to JSON, then unmarshal to any
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

	// Check if current schema has the dynamic anchor
	if schema.HasDynamicAnchor() && schema.DynamicAnchor() == anchorName {
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
		if ifSchema, ok := schema.IfSchema().(*Schema); ok {
			if found, err := r.findSchemaByAnchor(ifSchema, anchorName); err == nil {
				return found, nil
			}
		}
	}

	if schema.HasThenSchema() {
		if thenSchema, ok := schema.ThenSchema().(*Schema); ok {
			if found, err := r.findSchemaByAnchor(thenSchema, anchorName); err == nil {
				return found, nil
			}
		}
	}

	if schema.HasElseSchema() {
		if elseSchema, ok := schema.ElseSchema().(*Schema); ok {
			if found, err := r.findSchemaByAnchor(elseSchema, anchorName); err == nil {
				return found, nil
			}
		}
	}

	// Search in contains schema
	if schema.HasContains() {
		if containsSchema, ok := schema.Contains().(*Schema); ok {
			if found, err := r.findSchemaByAnchor(containsSchema, anchorName); err == nil {
				return found, nil
			}
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
