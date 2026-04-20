package unface_test

import (
	"reflect"
	"testing"

	"github.com/schneid-l/unface"
)

type stubAdapter struct{ ptr *string }

func (a stubAdapter) Unface(src any) error {
	if s, ok := src.(string); ok {
		*a.ptr = s
		return nil
	}
	return unface.ErrNotHandled
}

func newStubPlugin(name string) unface.Plugin {
	f := unface.FactoryFunc(
		func(t reflect.Type) bool { return t == reflect.TypeOf("") },
		func(ptr any) unface.Adapter { return stubAdapter{ptr: ptr.(*string)} },
	)
	return unface.NewPlugin(name, f)
}

func TestNewPluginName(t *testing.T) {
	p := newStubPlugin("stub")
	if p.Name() != "stub" {
		t.Fatalf("Name=%q", p.Name())
	}
	if len(p.Factories()) != 1 {
		t.Fatalf("factories=%d", len(p.Factories()))
	}
	if p.ChildNames() != nil {
		t.Fatalf("atomic plugin must have nil ChildNames, got %v", p.ChildNames())
	}
}

func TestFactoryMatchesAndFor(t *testing.T) {
	p := newStubPlugin("stub")
	f := p.Factories()[0]
	if !f.Matches(reflect.TypeOf("")) {
		t.Fatal("factory must match string")
	}
	if f.Matches(reflect.TypeOf(0)) {
		t.Fatal("factory must not match int")
	}
	var s string
	a := f.For(&s)
	if err := a.Unface("hello"); err != nil {
		t.Fatal(err)
	}
	if s != "hello" {
		t.Fatalf("s=%q", s)
	}
}

func TestComposeFlattens(t *testing.T) {
	p1 := newStubPlugin("a")
	p2 := newStubPlugin("b")
	c := unface.Compose("combo", p1, p2)
	if c.Name() != "combo" {
		t.Fatalf("Name=%q", c.Name())
	}
	if len(c.Factories()) != 2 {
		t.Fatalf("factories=%d", len(c.Factories()))
	}
	// Nested compose still flattens:
	c2 := unface.Compose("outer", c, newStubPlugin("c"))
	if len(c2.Factories()) != 3 {
		t.Fatalf("nested factories=%d", len(c2.Factories()))
	}
}

func TestComposeChildNames(t *testing.T) {
	c := unface.Compose("combo", newStubPlugin("a"), newStubPlugin("b"))
	names := c.ChildNames()
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Fatalf("ChildNames=%v", names)
	}
}
