package unface_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/schneid-l/unface"
)

// exampleURL is a small URL-shaped destination used across the godoc
// examples. It implements Unstringer (for shorthand string sources) and
// uses unface tags for the structured map form.
type exampleURL struct {
	Scheme string `unface:"scheme"`
	Host   string `unface:"host"`
	Port   int    `unface:"port"`
	Path   string `unface:"path"`
}

// Unstring parses "scheme://host:port/path".
func (u *exampleURL) Unstring(s string) error {
	rest := s
	if i := strings.Index(rest, "://"); i > 0 {
		u.Scheme, rest = rest[:i], rest[i+3:]
	}
	if i := strings.Index(rest, "/"); i >= 0 {
		u.Path = rest[i:]
		rest = rest[:i]
	}
	if i := strings.LastIndex(rest, ":"); i > 0 {
		if _, err := fmt.Sscanf(rest[i+1:], "%d", &u.Port); err != nil {
			return err
		}
		rest = rest[:i]
	}
	u.Host = rest
	return nil
}

// Example shows the package-level idiom: a struct field accepts either a
// URL shorthand string or a structured map thanks to its Unstringer
// implementation. The root unface.Unface function uses the Default Facer
// (preloaded with StandardPlugin) so no setup is required.
func Example() {
	type config struct {
		Name string     `unface:"name"`
		API  exampleURL `unface:"api"`
		Web  exampleURL `unface:"web"`
	}
	raw := map[string]any{
		"name": "demo",
		"api":  "https://api.example.com:8080/v1",
		"web": map[string]any{
			"scheme": "https",
			"host":   "example.com",
			"path":   "/",
		},
	}
	var cfg config
	if err := unface.Unface(raw, &cfg); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(cfg.Name)
	fmt.Printf("API: %s %s:%d%s\n", cfg.API.Scheme, cfg.API.Host, cfg.API.Port, cfg.API.Path)
	fmt.Printf("Web: %s %s%s\n", cfg.Web.Scheme, cfg.Web.Host, cfg.Web.Path)
	// Output:
	// demo
	// API: https api.example.com:8080/v1
	// Web: https example.com/
}

// ExampleUnface shows the terse form: coerce a loose map into a struct.
func ExampleUnface() {
	type server struct {
		Host string `unface:"host"`
		Port int    `unface:"port"`
	}
	raw := map[string]any{"host": "localhost", "port": "8080"}
	var s server
	if err := unface.Unface(raw, &s); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%s:%d\n", s.Host, s.Port)
	// Output: localhost:8080
}

// ExampleNew constructs a Facer with a minimal, explicit plugin set.
// Only the String and Int plugins are loaded, so non-string/int targets
// return ErrNoCoercion.
func ExampleNew() {
	f := unface.New(unface.With(unface.StringPlugin, unface.IntPlugin))
	var n int
	if err := f.Unface("42", &n); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(n)
	// Output: 42
}

// ExampleFacer_Unface reuses a single Facer across multiple coercions.
func ExampleFacer_Unface() {
	f := unface.New(unface.With(unface.StandardPlugin))
	var a, b int
	_ = f.Unface("1", &a)
	_ = f.Unface(2.0, &b)
	fmt.Println(a, b)
	// Output: 1 2
}

// ExampleFacer_Unface_perCallOverrides demonstrates the per-call options
// that Facer.Unface accepts: here, the baseline Facer has the standard set
// but a single call narrows the allowed plugins via Only.
func ExampleFacer_Unface_perCallOverrides() {
	f := unface.New(unface.With(unface.StandardPlugin))
	var n int
	// Per-call override: only IntPlugin is available for this call.
	if err := f.Unface("42", &n, unface.Only(unface.IntPlugin)); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(n)
	// Output: 42
}

// ExampleWith adds a plugin to a Facer's set.
func ExampleWith() {
	// Start with just StringPlugin; With adds IntPlugin on top.
	f := unface.New(unface.With(unface.StringPlugin), unface.With(unface.IntPlugin))
	var n int
	_ = f.Unface("7", &n)
	fmt.Println(n)
	// Output: 7
}

// ExampleWithout drops a composite plugin (and, transitively, its named
// children) from the plugin set.
func ExampleWithout() {
	// StandardPlugin includes NumberPlugin. Without(NumberPlugin) removes
	// every numeric atomic plugin, so "42" cannot be coerced into an int.
	f := unface.New(unface.With(unface.StandardPlugin), unface.Without(unface.NumberPlugin))
	var n int
	err := f.Unface("42", &n)
	fmt.Println(errors.Is(err, unface.ErrNoCoercion))
	// Output: true
}

// ExampleOnly replaces the entire plugin set.
func ExampleOnly() {
	f := unface.New(unface.With(unface.StandardPlugin), unface.Only(unface.BoolPlugin))
	var b bool
	_ = f.Unface("true", &b)
	fmt.Println(b)
	// Output: true
}

// ExampleWithFieldMatch shows MatchExact rejecting a case-mismatched key.
// With the default MatchFold the key "host" would match the field Host.
func ExampleWithFieldMatch() {
	type server struct {
		Host string `unface:"Host"`
	}
	f := unface.New(
		unface.With(unface.StandardPlugin),
		unface.WithFieldMatch(unface.MatchExact),
	)
	var s server
	_ = f.Unface(map[string]any{"host": "localhost"}, &s)
	fmt.Printf("%q\n", s.Host)
	// Output: ""
}

// ExampleOnUnknown shows the UnknownError policy returning
// ErrUnknownField for a map key that has no destination field.
func ExampleOnUnknown() {
	type server struct {
		Host string `unface:"host"`
	}
	f := unface.New(
		unface.With(unface.StandardPlugin),
		unface.OnUnknown(unface.UnknownError),
	)
	var s server
	err := f.Unface(map[string]any{"host": "localhost", "extra": 1}, &s)
	fmt.Println(errors.Is(err, unface.ErrUnknownField))
	// Output: true
}

// ExampleWithPointerResolve contrasts PointerResolveNone (refuses to
// traverse extra pointer layers) with PointerResolveDeep (tries each
// level) for a **T destination. A plain Unfacer target (no plugin set) is
// used so the dispatch is level-sensitive.
func ExampleWithPointerResolve() {
	var inner1 exampleTarget
	p1 := &inner1
	none := unface.New(unface.WithPointerResolve(unface.PointerResolveNone))
	errNone := none.Unface("hi", &p1) // dest is **exampleTarget; None refuses.

	var inner2 exampleTarget
	p2 := &inner2
	deep := unface.New(unface.WithPointerResolve(unface.PointerResolveDeep))
	errDeep := deep.Unface("hi", &p2) // Deep walks to *exampleTarget, where Unface fires.

	fmt.Println("none:", errors.Is(errNone, unface.ErrNoCoercion))
	fmt.Println("deep:", errDeep, inner2.got)
	// Output:
	// none: true
	// deep: <nil> hi
}

type exampleTarget struct{ got string }

func (t *exampleTarget) Unface(src any) error {
	if s, ok := src.(string); ok {
		t.got = s
		return nil
	}
	return unface.ErrNotHandled
}
