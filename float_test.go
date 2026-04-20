package unface_test

import (
	"math"
	"testing"

	"github.com/schneid-l/unface"
)

func TestFloat64PluginFromInt(t *testing.T) {
	var x float64
	f := unface.New(unface.With(unface.Float64Plugin))
	if err := f.Unface(3, &x); err != nil {
		t.Fatal(err)
	}
	if x != 3.0 {
		t.Fatalf("x=%v", x)
	}
}

func TestFloat32PluginFromString(t *testing.T) {
	var x float32
	f := unface.New(unface.With(unface.Float32Plugin))
	if err := f.Unface("1.5", &x); err != nil {
		t.Fatal(err)
	}
	if math.Abs(float64(x)-1.5) > 1e-6 {
		t.Fatalf("x=%v", x)
	}
}

func TestFloat64PluginFromNil(t *testing.T) {
	x := 99.9
	f := unface.New(unface.With(unface.Float64Plugin))
	if err := f.Unface(nil, &x); err != nil {
		t.Fatal(err)
	}
	if x != 0 {
		t.Fatalf("x=%v", x)
	}
}

func TestComplex128PluginFromFloat(t *testing.T) {
	var x complex128
	f := unface.New(unface.With(unface.Complex128Plugin))
	if err := f.Unface(1.5, &x); err != nil {
		t.Fatal(err)
	}
	if real(x) != 1.5 || imag(x) != 0 {
		t.Fatalf("x=%v", x)
	}
}

func TestComplex128PluginFromString(t *testing.T) {
	var x complex128
	f := unface.New(unface.With(unface.Complex128Plugin))
	if err := f.Unface("1+2i", &x); err != nil {
		t.Fatal(err)
	}
	if real(x) != 1 || imag(x) != 2 {
		t.Fatalf("x=%v", x)
	}
}

func TestComplex64PluginFromComplex(t *testing.T) {
	var x complex64
	f := unface.New(unface.With(unface.Complex64Plugin))
	if err := f.Unface(complex(3.0, 4.0), &x); err != nil {
		t.Fatal(err)
	}
	if real(x) != 3 || imag(x) != 4 {
		t.Fatalf("x=%v", x)
	}
}
