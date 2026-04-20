package unface_test

import (
	"errors"
	"testing"
	"time"

	"github.com/schneid-l/unface"
)

// URL implements Unfacer: accepts string or map[string]any.
type URL struct {
	Scheme, Host, Path string
}

func (u *URL) Unface(src any) error {
	switch v := src.(type) {
	case string:
		return u.Unstring(v)
	case map[string]any:
		if s, ok := v["scheme"].(string); ok {
			u.Scheme = s
		}
		if s, ok := v["host"].(string); ok {
			u.Host = s
		}
		if s, ok := v["path"].(string); ok {
			u.Path = s
		}
		return nil
	}
	return unface.ErrNotHandled
}

func (u *URL) Unstring(s string) error {
	u.Scheme = "http"
	u.Host = s
	return nil
}

func TestFacerCallsUnfacerFirst(t *testing.T) {
	f := unface.New()
	var u URL
	if err := f.Unface("example.com", &u); err != nil {
		t.Fatal(err)
	}
	if u.Host != "example.com" || u.Scheme != "http" {
		t.Fatalf("u=%+v", u)
	}
}

type stringOnly struct{ v string }

func (s *stringOnly) Unstring(x string) error { s.v = x; return nil }

func TestFacerCallsUnstringerWhenNoUnfacer(t *testing.T) {
	f := unface.New()
	var s stringOnly
	if err := f.Unface("hi", &s); err != nil {
		t.Fatal(err)
	}
	if s.v != "hi" {
		t.Fatalf("s=%+v", s)
	}
}

func TestFacerRejectsNonPointerDest(t *testing.T) {
	f := unface.New()
	var u URL
	err := f.Unface("x", u)
	if !errors.Is(err, unface.ErrInvalidDest) {
		t.Fatalf("err=%v", err)
	}
}

func TestFacerRejectsNilDest(t *testing.T) {
	f := unface.New()
	err := f.Unface("x", nil)
	if !errors.Is(err, unface.ErrInvalidDest) {
		t.Fatalf("err=%v", err)
	}
}

func TestFacerNoCoercionWithoutPlugins(t *testing.T) {
	f := unface.New()
	var i int
	err := f.Unface("42", &i)
	if !errors.Is(err, unface.ErrNoCoercion) {
		t.Fatalf("err=%v", err)
	}
}

type decliningString struct{}

func (*decliningString) Unstring(string) error { return unface.ErrNotHandled }

func TestFacerSoftErrorFallsThrough(t *testing.T) {
	var d decliningString
	f := unface.New()
	err := f.Unface("hi", &d)
	if !errors.Is(err, unface.ErrNoCoercion) {
		t.Fatalf("err=%v", err)
	}
}

type nilAware struct{ cleared bool }

func (n *nilAware) Unnil() error { n.cleared = true; return nil }

func TestFacerUnniler(t *testing.T) {
	var n nilAware
	f := unface.New()
	if err := f.Unface(nil, &n); err != nil {
		t.Fatal(err)
	}
	if !n.cleared {
		t.Fatal("Unnil not called")
	}
}

type gotTime struct{ got time.Time }

func (g *gotTime) Untime(t time.Time) error { g.got = t; return nil }

func TestFacerUntimer(t *testing.T) {
	var x gotTime
	f := unface.New()
	tm := time.Unix(1700000000, 0).UTC()
	if err := f.Unface(tm, &x); err != nil {
		t.Fatal(err)
	}
	if !x.got.Equal(tm) {
		t.Fatalf("got %v", x.got)
	}
}

// --- Pointer-resolution tests ---

type pointerTest struct{ got string }

func (p *pointerTest) Unface(src any) error {
	if s, ok := src.(string); ok {
		p.got = s
		return nil
	}
	return unface.ErrNotHandled
}

func TestPointerResolveFlatIsDefault(t *testing.T) {
	var inner pointerTest
	p := &inner
	if err := unface.New().Unface("hi", &p); err != nil {
		t.Fatal(err)
	}
	if inner.got != "hi" {
		t.Fatalf("got=%q", inner.got)
	}
}

func TestPointerResolveNoneRefusesDepth(t *testing.T) {
	var inner pointerTest
	p := &inner
	f := unface.New(unface.WithPointerResolve(unface.PointerResolveNone))
	err := f.Unface("hi", &p)
	if !errors.Is(err, unface.ErrNoCoercion) {
		t.Fatalf("err=%v", err)
	}
}

func TestPointerResolveDeepWalksLevels(t *testing.T) {
	var inner pointerTest
	p1 := &inner
	p2 := &p1
	f := unface.New(unface.WithPointerResolve(unface.PointerResolveDeep))
	if err := f.Unface("deep", &p2); err != nil {
		t.Fatal(err)
	}
	if inner.got != "deep" {
		t.Fatalf("got=%q", inner.got)
	}
}

func TestPointerResolveDeepHonorsSrcLadder(t *testing.T) {
	s := "hello"
	ps := &s
	pps := &ps
	var inner pointerTest
	p := &inner
	f := unface.New(unface.WithPointerResolve(unface.PointerResolveDeep))
	if err := f.Unface(pps, &p); err != nil {
		t.Fatal(err)
	}
	if inner.got != "hello" {
		t.Fatalf("got=%q", inner.got)
	}
}

func TestFacerPointerFlatAssignment(t *testing.T) {
	// Direct assignment: src is int, dest is *int (innermost after flatten)
	var i int
	f := unface.New()
	if err := f.Unface(42, &i); err != nil {
		t.Fatal(err)
	}
	if i != 42 {
		t.Fatalf("i=%d", i)
	}
}
