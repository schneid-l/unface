// Package strcase implements the identifier-folding rules used by unface's
// struct walker. Fold reports whether two identifiers match under the
// library's permissive match mode: case-insensitive, with snake_case,
// kebab-case, and CamelCase treated as equivalent.
package strcase

import (
	"strings"
	"unicode"
)

// Normalize returns the canonical form of s: lowercased (full Unicode) with
// underscores and hyphens removed. Two strings match under the permissive
// mode iff they share the same Normalize output.
func Normalize(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '_' || r == '-' {
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// Fold reports whether a and b are equal under permissive folding.
func Fold(a, b string) bool {
	return Normalize(a) == Normalize(b)
}

// EqualFold reports whether two strings are equal case-insensitively (no
// separator folding). Provided for the MatchInsensitive mode.
func EqualFold(a, b string) bool {
	return strings.EqualFold(a, b)
}
