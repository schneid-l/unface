package unface_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/schneid-l/unface"
)

// Deep mode: ensure src nil ladder terminates correctly and short-circuits.
func TestPointerResolveDeepNilSrc(t *testing.T) {
	var p *int
	f := unface.New(unface.WithPointerResolve(unface.PointerResolveDeep))
	err := f.Unface(nil, &p)
	// Neither an Unfacer nor a plugin is registered; Deep mode walks every
	// combination but no handler fires, so we get ErrNoCoercion.
	if err == nil {
		t.Fatal("expected failure with no handlers")
	}
}

// Deep mode must propagate hard errors without retrying other levels.
type hardErrDest struct{}

var errHard = errors.New("hard failure")

func (*hardErrDest) Unface(any) error { return errHard }

func TestPointerResolveDeepStopsOnHardError(t *testing.T) {
	var d hardErrDest
	p := &d
	f := unface.New(unface.WithPointerResolve(unface.PointerResolveDeep))
	err := f.Unface("x", &p)
	if !errors.Is(err, errHard) {
		t.Fatalf("err=%v", err)
	}
}

// None mode must never dereference src either.
func TestPointerResolveNoneExactSrcMatch(t *testing.T) {
	// With None, the dispatcher treats src verbatim. If dest is *int and
	// src is int, assignDirect should still succeed (one layer deep).
	var i int
	f := unface.New(unface.WithPointerResolve(unface.PointerResolveNone))
	if err := f.Unface(42, &i); err != nil {
		t.Fatal(err)
	}
	if i != 42 {
		t.Fatalf("i=%d", i)
	}
}

// Flat mode: deeply nested *****int dest must receive the value.
func TestPointerResolveFlatVeryDeep(t *testing.T) {
	var p *****int
	f := unface.New(unface.With(unface.IntPlugin))
	if err := f.Unface(7, &p); err != nil {
		t.Fatal(err)
	}
	if p == nil || *p == nil || **p == nil || ***p == nil || ****p == nil {
		t.Fatal("intermediate nil pointer not allocated")
	}
	if *****p != 7 {
		t.Fatalf("got %d", *****p)
	}
}

// Package-level Unface with nil dest must return ErrInvalidDest.
func TestPackageUnfaceNilDest(t *testing.T) {
	err := unface.Unface("x", nil)
	if !errors.Is(err, unface.ErrInvalidDest) {
		t.Fatalf("err=%v", err)
	}
}

// Package-level Unface with nil-pointer dest must return ErrInvalidDest.
func TestPackageUnfaceNilPointerDest(t *testing.T) {
	var p *int
	err := unface.Unface("x", p)
	if !errors.Is(err, unface.ErrInvalidDest) {
		t.Fatalf("err=%v", err)
	}
}

// Plugin hard error must short-circuit further plugin attempts.
type hardAdapter struct{}

func (hardAdapter) Unface(any) error { return errHard }

func TestPluginHardErrorStopsFallback(t *testing.T) {
	stringType := reflect.TypeOf("")
	hard := unface.NewPlugin("hard",
		unface.FactoryFunc(
			func(t reflect.Type) bool { return t == stringType },
			func(_ any) unface.Adapter { return hardAdapter{} },
		),
	)
	// Use int → *string (no assignDirect shortcut). Hard plugin is queried
	// before StringPlugin and its hard error must short-circuit.
	var s string
	f := unface.New(unface.With(hard, unface.StringPlugin))
	err := f.Unface(42, &s)
	if !errors.Is(err, errHard) {
		t.Fatalf("err=%v", err)
	}
}

// --- Struct walker error wrapping ---

type badPort struct{}

func (*badPort) Unstring(string) error {
	return errors.New("bad port format")
}

func TestStructWalkErrorHasFieldPath(t *testing.T) {
	type cfg struct {
		Port badPort `unface:"port"`
	}
	var c cfg
	f := unface.New(unface.With(unface.StandardPlugin))
	err := f.Unface(map[string]any{"port": "bogus"}, &c)
	if err == nil {
		t.Fatal("expected error")
	}
	// Must be wrapped with the field name in the path.
	var e *unface.Error
	if !errors.As(err, &e) {
		t.Fatalf("err not an *unface.Error: %T", err)
	}
	if len(e.Path) == 0 || e.Path[0] != "port" {
		t.Fatalf("path=%v", e.Path)
	}
}

func TestStructWalkNestedPathTrace(t *testing.T) {
	type inner struct {
		X badPort `unface:"x"`
	}
	type outer struct {
		In inner `unface:"in"`
	}
	var o outer
	f := unface.New(unface.With(unface.StandardPlugin))
	err := f.Unface(map[string]any{"in": map[string]any{"x": "bogus"}}, &o)
	var e *unface.Error
	if !errors.As(err, &e) {
		t.Fatalf("err=%T %v", err, err)
	}
	if len(e.Path) < 2 || e.Path[0] != "in" || e.Path[1] != "x" {
		t.Fatalf("path=%v", e.Path)
	}
}

// Unknown policy with a warn handler must be invoked per unknown key.
func TestStructUnknownWarnCallsHandler(t *testing.T) {
	type T struct {
		A int `unface:"a"`
	}
	var seen []string
	var v T
	err := unface.New(unface.With(unface.StandardPlugin)).Unface(
		map[string]any{"a": 1, "b": 2, "c": 3},
		&v,
		unface.OnUnknown(unface.UnknownWarn, func(k string, _ any) {
			seen = append(seen, k)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(seen) != 2 {
		t.Fatalf("seen=%v (expected 2)", seen)
	}
}

func TestStructUnknownWarnNoHandlerIsSilent(t *testing.T) {
	type T struct {
		A int `unface:"a"`
	}
	var v T
	err := unface.New(unface.With(unface.StandardPlugin)).Unface(
		map[string]any{"a": 1, "unknown": 42},
		&v,
		unface.OnUnknown(unface.UnknownWarn),
	)
	if err != nil {
		t.Fatal(err)
	}
}

// Remainder absorbs unknowns so the UnknownError policy never fires.
func TestStructRemainderSuppressesUnknownError(t *testing.T) {
	type T struct {
		A    int            `unface:"a"`
		Rest map[string]any `unface:",remainder"`
	}
	var v T
	err := unface.New(unface.With(unface.StandardPlugin)).Unface(
		map[string]any{"a": 1, "b": 2},
		&v,
		unface.OnUnknown(unface.UnknownError),
	)
	if err != nil {
		t.Fatal(err)
	}
	if v.Rest["b"] != 2 {
		t.Fatalf("rest=%v", v.Rest)
	}
}

// Strict-tagged field must NOT fall through to plugins; user Un*er only.
type noStringSupport struct{}

func (*noStringSupport) Unnumber(unface.Number) error { return nil }

func TestStructStrictFieldBypassesPlugins(t *testing.T) {
	// noStringSupport dest refuses string sources on Unnumberer-only impl.
	// With strict, the IntPlugin can't fire → the string "42" fails.
	type cfg struct {
		N noStringSupport `unface:"n,strict"`
	}
	var c cfg
	f := unface.New(unface.With(unface.StandardPlugin))
	err := f.Unface(map[string]any{"n": "42"}, &c)
	if err == nil {
		t.Fatal("strict field should reject string → noStringSupport")
	}
}

// Map key must not be a non-string with unknown policy (only string keys
// participate in the walker).
func TestStructIntKeyStringified(t *testing.T) {
	type T struct {
		A int `unface:"1"`
	}
	var v T
	// Source has numeric keys; walker stringifies via fmt.Sprint.
	err := unface.New(unface.With(unface.StandardPlugin)).Unface(
		map[any]any{1: 42},
		&v,
	)
	if err != nil {
		t.Fatal(err)
	}
	if v.A != 42 {
		t.Fatalf("A=%d", v.A)
	}
}

// ---- Options chaining ----

func TestMultipleWithAccumulate(t *testing.T) {
	f := unface.New(
		unface.With(unface.StringPlugin),
		unface.With(unface.IntPlugin),
	)
	var s string
	if err := f.Unface("hi", &s); err != nil {
		t.Fatal(err)
	}
	var i int
	if err := f.Unface(42, &i); err != nil {
		t.Fatal(err)
	}
}

func TestWithoutNamedDropsAtomicByName(t *testing.T) {
	// WithoutNamed matches flattened plugin names in the instance set. After
	// With(StandardPlugin), every atomic leaf is present by its own name;
	// users can drop specific ones (e.g. "int"). Use a src type that forces
	// the plugin path (string → int goes through IntPlugin's Unstring).
	f := unface.New(unface.With(unface.StandardPlugin), unface.WithoutNamed("int"))
	var i int
	if err := f.Unface("42", &i); err == nil {
		t.Fatal("removing 'int' by name should drop string→int coercion")
	}
}
