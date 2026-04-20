package strcase_test

import (
	"testing"

	"github.com/schneid-l/unface/internal/strcase"
)

// FuzzNormalize ensures Normalize never panics and always produces a
// string whose Normalize is itself (idempotence property).
func FuzzNormalize(f *testing.F) {
	for _, s := range []string{
		"", "a", "ABC", "http_port", "http-port",
		"HTTPPort", "héllo", "é_B-c",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		once := strcase.Normalize(s)
		twice := strcase.Normalize(once)
		if once != twice {
			t.Fatalf("not idempotent: %q → %q → %q", s, once, twice)
		}
	})
}

// FuzzFold ensures Fold is symmetric for arbitrary inputs.
func FuzzFold(f *testing.F) {
	f.Add("Port", "port")
	f.Add("", "")
	f.Add("a", "b")
	f.Fuzz(func(t *testing.T, a, b string) {
		ab := strcase.Fold(a, b)
		ba := strcase.Fold(b, a)
		if ab != ba {
			t.Fatalf("asymmetric: Fold(%q,%q)=%v vs Fold(%q,%q)=%v",
				a, b, ab, b, a, ba)
		}
	})
}
