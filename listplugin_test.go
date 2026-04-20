package unface_test

import (
	"testing"

	"github.com/schneid-l/unface"
)

func TestListPluginIntSlice(t *testing.T) {
	var out []int
	f := unface.New(unface.With(unface.ListPlugin, unface.IntPlugin))
	if err := f.Unface([]any{1, 2, 3}, &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 3 || out[2] != 3 {
		t.Fatalf("out=%v", out)
	}
}

func TestListPluginStringSlice(t *testing.T) {
	var out []string
	f := unface.New(unface.With(unface.ListPlugin, unface.StringPlugin))
	if err := f.Unface([]any{"a", "b"}, &out); err != nil {
		t.Fatal(err)
	}
	if out[0] != "a" || out[1] != "b" {
		t.Fatalf("out=%v", out)
	}
}

func TestListPluginScalarToSingletonSlice(t *testing.T) {
	var out []int
	f := unface.New(unface.With(unface.ListPlugin, unface.IntPlugin))
	if err := f.Unface(7, &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0] != 7 {
		t.Fatalf("out=%v", out)
	}
}

func TestListPluginNilSrcZeroes(t *testing.T) {
	out := []int{1, 2, 3}
	f := unface.New(unface.With(unface.ListPlugin, unface.IntPlugin))
	if err := f.Unface(nil, &out); err != nil {
		t.Fatal(err)
	}
	if out != nil {
		t.Fatalf("out=%v want nil", out)
	}
}
