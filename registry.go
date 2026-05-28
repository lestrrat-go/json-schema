package schema

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lestrrat-go/jsref/v2"
)

// resourceIndex maps absolute URIs to the schemas they identify within a single
// in-memory document tree. It is the in-document counterpart to the external
// (HTTP/file) resolvers: every $id establishes a resource whose absolute URI is
// recorded in byURI, and every $anchor/$dynamicAnchor is recorded in anchors
// keyed by "<resourceBaseURI>#<anchorName>". Reference resolution consults this
// index before falling back to network/filesystem retrieval.
type resourceIndex struct {
	byURI   map[string]*Schema
	anchors map[string]*Schema
}

func newResourceIndex() *resourceIndex {
	return &resourceIndex{
		byURI:   make(map[string]*Schema),
		anchors: make(map[string]*Schema),
	}
}

// index walks the schema tree rooted at s, registering every $id-bearing
// resource and every anchor against the base URI in effect at that node. A node
// with its own $id re-bases the subtree beneath it (resolved against the
// enclosing base per RFC 3986), which is how nested/bundled resources are made
// addressable by their canonical absolute URIs.
func (idx *resourceIndex) index(s *Schema, baseURI string, visited map[*Schema]struct{}) {
	if s == nil {
		return
	}
	if _, seen := visited[s]; seen {
		return
	}
	visited[s] = struct{}{}

	current := baseURI
	if s.HasID() && s.ID() != "" {
		resolved := resolveURI(baseURI, s.ID())
		base, _, _ := splitFragment(resolved)
		idx.byURI[base] = s
		current = base
	}
	if s.HasAnchor() && s.Anchor() != "" {
		idx.anchors[current+"#"+s.Anchor()] = s
	}
	if s.HasDynamicAnchor() && s.DynamicAnchor() != "" {
		idx.anchors[current+"#"+s.DynamicAnchor()] = s
	}

	for _, child := range childSchemas(s) {
		idx.index(child, current, visited)
	}
}

// FindDynamicAnchor searches a schema resource for a subschema declaring
// $dynamicAnchor == name, without crossing into nested $id resources (which are
// distinct resources with their own dynamic scope). It returns nil if not found.
func FindDynamicAnchor(resource *Schema, name string) *Schema {
	if resource == nil {
		return nil
	}
	if resource.HasDynamicAnchor() && resource.DynamicAnchor() == name {
		return resource
	}
	for _, child := range childSchemas(resource) {
		if child.HasID() && child.ID() != "" {
			continue // a nested $id starts a separate resource
		}
		if found := FindDynamicAnchor(child, name); found != nil {
			return found
		}
	}
	return nil
}

// childSchemas returns the immediate subschemas of s across every keyword that
// can hold one. It mirrors the traversal in findSchemaByAnchor (plus
// prefixItems and list-valued items), so the resource index sees the same nodes
// anchor resolution does.
func childSchemas(s *Schema) []*Schema {
	var out []*Schema
	add := func(v any) {
		if sub, ok := v.(*Schema); ok && sub != nil {
			out = append(out, sub)
		}
	}

	if s.HasDefinitions() {
		for _, def := range s.Definitions() {
			out = append(out, def)
		}
	}
	if s.HasProperties() {
		for _, p := range s.Properties() {
			out = append(out, p)
		}
	}
	if s.HasPatternProperties() {
		for _, p := range s.PatternProperties() {
			out = append(out, p)
		}
	}
	if s.HasPrefixItems() {
		for _, it := range s.PrefixItems() {
			add(it)
		}
	}
	if s.HasItems() {
		add(s.Items())
	}
	if s.HasAdditionalProperties() {
		add(s.AdditionalProperties())
	}
	if s.HasUnevaluatedProperties() {
		add(s.UnevaluatedProperties())
	}
	if s.HasUnevaluatedItems() {
		add(s.UnevaluatedItems())
	}
	if s.HasAllOf() {
		for _, sub := range s.AllOf() {
			add(sub)
		}
	}
	if s.HasAnyOf() {
		for _, sub := range s.AnyOf() {
			add(sub)
		}
	}
	if s.HasOneOf() {
		for _, sub := range s.OneOf() {
			add(sub)
		}
	}
	if s.HasNot() {
		out = append(out, s.Not())
	}
	if s.HasIfSchema() {
		add(s.IfSchema())
	}
	if s.HasThenSchema() {
		add(s.ThenSchema())
	}
	if s.HasElseSchema() {
		add(s.ElseSchema())
	}
	if s.HasContains() {
		add(s.Contains())
	}
	if s.HasPropertyNames() {
		out = append(out, s.PropertyNames())
	}
	if s.HasContentSchema() {
		out = append(out, s.ContentSchema())
	}
	return out
}

// registryResolver is a jsref.Resolver backed by a resourceIndex. Registered
// ahead of the HTTP/filesystem resolvers, it intercepts references whose base
// URI names an in-document resource and resolves them without any external
// retrieval. References it does not recognize are declined (CanResolve returns
// false) so the stacked resolver falls through to the external resolvers.
type registryResolver struct {
	idx *resourceIndex
	obj jsref.Resolver
}

func (rr *registryResolver) CanResolve(resource any) bool {
	uri, ok := resource.(string)
	if !ok {
		return false
	}
	_, ok = rr.idx.byURI[uri]
	return ok
}

func (rr *registryResolver) Resolve(dst any, resource any, localRef string) error {
	uri, ok := resource.(string)
	if !ok {
		return fmt.Errorf("registry resolver: non-string resource %T", resource)
	}
	root, ok := rr.idx.byURI[uri]
	if !ok {
		return fmt.Errorf("registry resolver: no resource registered for %q", uri)
	}

	// localRef arrives already percent-decoded from ResolveJSONReference.
	fragment := strings.TrimPrefix(localRef, "#")

	// Determine the target schema and the JSON pointer to apply within it.
	target := root
	pointer := "#"
	switch {
	case fragment == "":
		// Whole-document reference.
	case fragment[0] == '/':
		pointer = "#" + fragment
	default:
		// Plain anchor name.
		anchor, ok := rr.idx.anchors[uri+"#"+fragment]
		if !ok {
			return fmt.Errorf("registry resolver: anchor %q not found in resource %q", fragment, uri)
		}
		target = anchor
	}

	data, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("registry resolver: failed to marshal target schema: %w", err)
	}
	var doc any
	if err := json.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("registry resolver: failed to decode target schema: %w", err)
	}
	return rr.obj.Resolve(dst, doc, pointer)
}
