package unface_test

import (
	"testing"

	"github.com/schneid-l/unface"
)

func TestPointerPluginAllocatesAndDelegates(t *testing.T) {
	var p *int
	f := unface.New(unface.With(unface.PointerPlugin, unface.IntPlugin))
	if err := f.Unface(7, &p); err != nil {
		t.Fatal(err)
	}
	if p == nil || *p != 7 {
		t.Fatalf("p=%v", p)
	}
}

func TestPointerPluginAllocatesIntoNil(t *testing.T) {
	// With Flat (default) resolution, a **int dest whose *int is nil is
	// auto-allocated by DerefToAddressable; the value coercion then fills
	// the inner int.
	var p *int
	f := unface.New(unface.With(unface.PointerPlugin, unface.IntPlugin))
	if err := f.Unface("42", &p); err != nil {
		t.Fatal(err)
	}
	if p == nil || *p != 42 {
		t.Fatalf("p=%v", p)
	}
}
