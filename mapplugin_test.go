package unface_test

import (
	"testing"

	"github.com/schneid-l/unface"
)

func TestMapPluginStringInt(t *testing.T) {
	var out map[string]int
	f := unface.New(unface.With(unface.MapPlugin, unface.IntPlugin, unface.StringPlugin))
	if err := f.Unface(map[string]any{"a": 1, "b": "2"}, &out); err != nil {
		t.Fatal(err)
	}
	if out["a"] != 1 || out["b"] != 2 {
		t.Fatalf("out=%v", out)
	}
}

func TestMapPluginRejectsNonMap(t *testing.T) {
	var out map[string]int
	f := unface.New(unface.With(unface.MapPlugin, unface.IntPlugin, unface.StringPlugin))
	err := f.Unface(42, &out)
	if err == nil {
		t.Fatal("expected error on non-map source")
	}
}

func TestMapPluginNilSrcZeroes(t *testing.T) {
	out := map[string]int{"a": 1}
	f := unface.New(unface.With(unface.MapPlugin, unface.StringPlugin, unface.IntPlugin))
	if err := f.Unface(nil, &out); err != nil {
		t.Fatal(err)
	}
	if out != nil {
		t.Fatalf("out=%v want nil", out)
	}
}
