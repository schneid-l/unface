package unface_test

import (
	"math"
	"testing"

	"github.com/schneid-l/unface"
)

func TestInt64PluginFromInt(t *testing.T) {
	var x int64
	f := unface.New(unface.With(unface.Int64Plugin))
	if err := f.Unface(42, &x); err != nil {
		t.Fatal(err)
	}
	if x != 42 {
		t.Fatalf("x=%d", x)
	}
}

func TestInt64PluginFromString(t *testing.T) {
	var x int64
	f := unface.New(unface.With(unface.Int64Plugin))
	if err := f.Unface("-123", &x); err != nil {
		t.Fatal(err)
	}
	if x != -123 {
		t.Fatalf("x=%d", x)
	}
}

func TestInt8PluginOverflow(t *testing.T) {
	var x int8
	f := unface.New(unface.With(unface.Int8Plugin))
	err := f.Unface(1000, &x)
	if err == nil {
		t.Fatal("expected overflow error")
	}
}

func TestInt8PluginOverflowFromString(t *testing.T) {
	var x int8
	f := unface.New(unface.With(unface.Int8Plugin))
	err := f.Unface("1000", &x)
	if err == nil {
		t.Fatal("expected overflow error")
	}
}

func TestUint32PluginRejectsNegative(t *testing.T) {
	var x uint32
	f := unface.New(unface.With(unface.Uint32Plugin))
	err := f.Unface(-1, &x)
	if err == nil {
		t.Fatal("expected error on negative -> uint")
	}
}

func TestIntPluginBool(t *testing.T) {
	var x int
	f := unface.New(unface.With(unface.IntPlugin))
	if err := f.Unface(true, &x); err != nil {
		t.Fatal(err)
	}
	if x != 1 {
		t.Fatalf("x=%d", x)
	}
}

func TestUint64MaxPreserved(t *testing.T) {
	var x uint64
	f := unface.New(unface.With(unface.Uint64Plugin))
	m := uint64(math.MaxUint64)
	if err := f.Unface(m, &x); err != nil {
		t.Fatal(err)
	}
	if x != m {
		t.Fatalf("x=%d", x)
	}
}

func TestIntPluginFromNilZeroes(t *testing.T) {
	x := int64(99)
	f := unface.New(unface.With(unface.Int64Plugin))
	if err := f.Unface(nil, &x); err != nil {
		t.Fatal(err)
	}
	if x != 0 {
		t.Fatalf("x=%d", x)
	}
}
