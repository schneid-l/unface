package strcase_test

import (
	"testing"
	"testing/quick"
	"unicode"

	"github.com/schneid-l/unface/internal/strcase"
)

// Property: Fold is symmetric.
func TestFoldSymmetryProperty(t *testing.T) {
	f := func(a, b string) bool {
		return strcase.Fold(a, b) == strcase.Fold(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

// Property: Normalize is idempotent.
func TestNormalizeIdempotentProperty(t *testing.T) {
	f := func(s string) bool {
		once := strcase.Normalize(s)
		twice := strcase.Normalize(once)
		return once == twice
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

// Property: Normalize output contains no "_" / "-" and every rune equals
// its own unicode.ToLower (fully case-folded).
func TestNormalizeOutputShape(t *testing.T) {
	f := func(s string) bool {
		n := strcase.Normalize(s)
		for _, r := range n {
			if r == '_' || r == '-' {
				return false
			}
			if r != unicode.ToLower(r) {
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

// Property: two strings that normalize identically always Fold equal.
func TestFoldConsistentWithNormalize(t *testing.T) {
	f := func(a, b string) bool {
		return (strcase.Normalize(a) == strcase.Normalize(b)) == strcase.Fold(a, b)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeHandlesUnicode(t *testing.T) {
	// Non-ASCII lowercase letters pass through untouched (no case-folding
	// for non-ASCII in the current implementation).
	cases := map[string]string{
		"héllo":       "héllo",
		"héllo_world": "hélloworld",
	}
	for in, want := range cases {
		if got := strcase.Normalize(in); got != want {
			t.Errorf("Normalize(%q)=%q want %q", in, got, want)
		}
	}
}

func TestEqualFoldCovers(t *testing.T) {
	if !strcase.EqualFold("", "") {
		t.Fatal("empty/empty")
	}
	if strcase.EqualFold("a", "") {
		t.Fatal("a vs empty")
	}
}
