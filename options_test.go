package unface

import (
	"testing"

	"github.com/schneid-l/unface/plugin"
)

func newTestCfg(opts ...Option) *plugin.Config {
	c := plugin.NewDefaultConfig()
	for _, o := range opts {
		o(c)
	}
	return c
}

func TestWithAddsPlugins(t *testing.T) {
	p := NewPlugin("p")
	c := newTestCfg(With(p))
	if len(c.Plugins) != 1 || c.Plugins[0].Name() != "p" {
		t.Fatalf("plugins=%v", c.Plugins)
	}
}

func TestWithoutRemovesPlugin(t *testing.T) {
	p1, p2 := NewPlugin("a"), NewPlugin("b")
	c := newTestCfg(With(p1, p2), Without(p1))
	if len(c.Plugins) != 1 || c.Plugins[0].Name() != "b" {
		t.Fatalf("plugins=%v", c.Plugins)
	}
}

func TestWithoutComposite(t *testing.T) {
	p1, p2, p3 := NewPlugin("a"), NewPlugin("b"), NewPlugin("c")
	combo := Compose("combo", p1, p2)
	c := newTestCfg(With(combo, p3), Without(combo))
	if len(c.Plugins) != 1 || c.Plugins[0].Name() != "c" {
		t.Fatalf("plugins=%v", c.Plugins)
	}
}

func TestWithoutNamed(t *testing.T) {
	p1, p2 := NewPlugin("a"), NewPlugin("b")
	c := newTestCfg(With(p1, p2), WithoutNamed("a"))
	if len(c.Plugins) != 1 || c.Plugins[0].Name() != "b" {
		t.Fatalf("plugins=%v", c.Plugins)
	}
}

func TestOnlyReplaces(t *testing.T) {
	p1, p2 := NewPlugin("a"), NewPlugin("b")
	c := newTestCfg(With(p1), Only(p2))
	if len(c.Plugins) != 1 || c.Plugins[0].Name() != "b" {
		t.Fatalf("plugins=%v", c.Plugins)
	}
}

func TestWithFieldMatch(t *testing.T) {
	c := newTestCfg(WithFieldMatch(MatchExact))
	if c.Match != MatchExact {
		t.Fatalf("match=%v", c.Match)
	}
}

func TestOnUnknownErrorMode(t *testing.T) {
	c := newTestCfg(OnUnknown(UnknownError))
	if c.OnUnknown != UnknownError {
		t.Fatalf("onUnknown=%v", c.OnUnknown)
	}
}

func TestOnUnknownWarnHandler(t *testing.T) {
	var called bool
	c := newTestCfg(OnUnknown(UnknownWarn, func(_ string, _ any) { called = true }))
	if c.UnknownHandler == nil {
		t.Fatal("handler not set")
	}
	c.UnknownHandler("x", 1)
	if !called {
		t.Fatal("handler not called")
	}
}

func TestWithTagFallback(t *testing.T) {
	c := newTestCfg(WithTagFallback("unface", "toml"))
	if len(c.TagFallback) != 2 || c.TagFallback[1] != "toml" {
		t.Fatalf("tagFallback=%v", c.TagFallback)
	}
}

func TestWithoutTagFallback(t *testing.T) {
	c := newTestCfg(WithoutTagFallback())
	if len(c.TagFallback) != 1 || c.TagFallback[0] != "unface" {
		t.Fatalf("tagFallback=%v", c.TagFallback)
	}
}

func TestDefaultConfig(t *testing.T) {
	c := plugin.NewDefaultConfig()
	if c.Match != MatchFold {
		t.Fatalf("default match=%v want MatchFold", c.Match)
	}
	if c.OnUnknown != UnknownIgnore {
		t.Fatalf("default onUnknown=%v", c.OnUnknown)
	}
	if want := []string{"unface", "yaml", "json"}; !equalStrings(c.TagFallback, want) {
		t.Fatalf("default tagFallback=%v", c.TagFallback)
	}
	if c.PointerMode != PointerResolveFlat {
		t.Fatalf("default pointerMode=%v want Flat", c.PointerMode)
	}
}

func TestWithPointerResolve(t *testing.T) {
	for _, mode := range []PointerResolution{PointerResolveNone, PointerResolveFlat, PointerResolveDeep} {
		c := newTestCfg(WithPointerResolve(mode))
		if c.PointerMode != mode {
			t.Fatalf("got %v want %v", c.PointerMode, mode)
		}
	}
}

func TestOnSoftError(t *testing.T) {
	var called bool
	c := newTestCfg(OnSoftError(func(_, _ any, _ error) { called = true }))
	if c.SoftHandler == nil {
		t.Fatal("handler not set")
	}
	c.SoftHandler(nil, nil, nil)
	if !called {
		t.Fatal("handler not called")
	}
}

func TestConfigClone(t *testing.T) {
	c := newTestCfg(With(NewPlugin("a")))
	c2 := c.Clone()
	c2.Plugins = append(c2.Plugins, NewPlugin("b"))
	if len(c.Plugins) != 1 {
		t.Fatalf("clone must not mutate original: %v", c.Plugins)
	}
	if len(c2.Plugins) != 2 {
		t.Fatalf("clone mutation lost: %v", c2.Plugins)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
