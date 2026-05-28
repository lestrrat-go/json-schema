package schema

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/lestrrat-go/jsref/v2"
	"github.com/lestrrat-go/option/v3"
)

// ResolverOption configures NewResolver.
type ResolverOption interface {
	option.Interface
	resolverOption()
}

type resolverOption struct{ option.Interface }

func (resolverOption) resolverOption() {}

type identResolver struct{}

// WithResolver appends an external reference resolver to the resolver stack.
//
// By default NewResolver resolves references only from memory (in-document $id
// resources and preloaded documents). To allow a resolver to reach outside the
// process — over the network or to the filesystem — supply it explicitly:
// HTTPResolver for HTTP/HTTPS, FSResolver/DirResolver for files,
// or any custom jsref.Resolver.
//
// WithResolver may be supplied multiple times; the resolvers are consulted in
// the order given, after the in-document registry and before the final
// JSON-pointer fallback.
func WithResolver(r jsref.Resolver) ResolverOption {
	return resolverOption{option.New(identResolver{}, r)}
}

// HTTPResolver returns a resolver that fetches remote references over HTTP/HTTPS.
// Pass it to WithResolver to enable network access:
//
//	r := schema.NewResolver(schema.WithResolver(schema.HTTPResolver()))
func HTTPResolver() jsref.Resolver {
	return jsref.NewHTTPResolver()
}

// FSResolver returns a resolver that reads references from fsys. It works
// with any io/fs source — an embed.FS, os.DirFS, zip archive, etc. Pass it to
// WithResolver to enable filesystem access:
//
//	r := schema.NewResolver(schema.WithResolver(schema.FSResolver(embedded)))
//
// References are looked up as slash-separated paths relative to the root of fsys
// (a leading "/" and a "file://" scheme are stripped). JSON and YAML documents
// are supported.
func FSResolver(fsys fs.FS) jsref.Resolver {
	return &fsResolver{fsys: fsys}
}

// DirResolver is shorthand for FSResolver(os.DirFS(dir)). It reads references
// from the local directory tree rooted at dir (e.g. "." for the current working
// directory):
//
//	r := schema.NewResolver(schema.WithResolver(schema.DirResolver(".")))
func DirResolver(dir string) jsref.Resolver {
	return FSResolver(os.DirFS(dir))
}

// fsResolver resolves references against an io/fs filesystem. jsref's own
// filesystem resolver is path-only (built on os.OpenRoot) and cannot take an
// fs.FS, so this small resolver bridges the gap, mirroring jsref's load → parse
// → resolve-fragment flow.
type fsResolver struct {
	fsys fs.FS
}

func (r *fsResolver) CanResolve(resource any) bool {
	s, ok := resource.(string)
	if !ok || strings.HasPrefix(s, "#") {
		return false
	}
	// Decline URLs with a network scheme so HTTPResolver handles those instead.
	if u, err := url.Parse(s); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		return false
	}
	return true
}

func (r *fsResolver) Resolve(dst any, resource any, localRef string) error {
	p, ok := resource.(string)
	if !ok {
		return fmt.Errorf("fsResolver requires string resource, got %T", resource)
	}

	// Accept "file://" URLs by extracting their path component.
	if u, err := url.Parse(p); err == nil && u.Scheme == "file" {
		p = u.Path
	}

	// fs.FS uses unrooted, slash-separated, cleaned paths.
	p = path.Clean(strings.TrimPrefix(p, "/"))

	data, err := fs.ReadFile(r.fsys, p)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", p, err)
	}

	parsed, err := parseDocument(data)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", p, err)
	}

	if localRef == "" {
		localRef = "#"
	}
	return jsref.NewObjectResolver().Resolve(dst, parsed, localRef)
}

// parseDocument decodes a schema document, accepting both JSON and YAML (YAML is
// a JSON superset, and the path-based jsref resolver this replaces accepted
// both). JSON is tried first to preserve JSON number semantics; YAML is the
// fallback.
func parseDocument(data []byte) (any, error) {
	var parsed any
	if err := json.Unmarshal(data, &parsed); err == nil {
		return parsed, nil
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("not valid JSON or YAML: %w", err)
	}
	return parsed, nil
}
