package schema

import (
	"net/url"
	"strings"
)

// resolveURI resolves a (possibly relative) reference against a base URI using
// RFC 3986 reference resolution. It correctly handles hierarchical URIs
// (http/https/relative paths) as well as opaque URIs such as URNs, where only
// the fragment is replaced. If either input cannot be parsed, it falls back to
// returning the reference unchanged so callers degrade to the legacy external
// resolution path rather than failing.
func resolveURI(base, ref string) string {
	if ref == "" {
		return base
	}
	if base == "" {
		return ref
	}

	b, err := url.Parse(base)
	if err != nil {
		return ref
	}
	r, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return b.ResolveReference(r).String()
}

// ResolveURI resolves a (possibly relative) reference against a base URI using
// RFC 3986 reference resolution, returning the resulting absolute URI. It is the
// exported entry point for callers outside this package (e.g. the validator's
// compiler) that need to compute the absolute base URI established by an $id.
func ResolveURI(base, ref string) string {
	return resolveURI(base, ref)
}

// splitFragment splits a URI into its base (everything before the first '#')
// and its fragment (everything after, excluding the '#'). A URI with no '#'
// yields an empty fragment; the returned hasFragment reports whether a '#' was
// present so callers can distinguish "no fragment" from "empty fragment" (a
// bare "#", which denotes the document root).
func splitFragment(uri string) (base, fragment string, hasFragment bool) {
	if base, fragment, found := strings.Cut(uri, "#"); found {
		return base, fragment, true
	}
	return uri, "", false
}

// unescapeFragment percent-decodes a URI fragment so that JSON Pointer and
// anchor lookups operate on the decoded value (e.g. "%25" -> "%", "%22" -> '"').
// JSON Pointer "~0"/"~1" escaping is left intact for the pointer evaluator. On
// decode failure the original fragment is returned unchanged.
func unescapeFragment(fragment string) string {
	decoded, err := url.PathUnescape(fragment)
	if err != nil {
		return fragment
	}
	return decoded
}
