package unface_test

import (
	"math/big"
	"testing"

	"github.com/schneid-l/unface"
)

func TestBigIntFromString(t *testing.T) {
	var x big.Int
	f := unface.New(unface.With(unface.BigIntPlugin))
	if err := f.Unface("12345678901234567890", &x); err != nil {
		t.Fatal(err)
	}
	if x.String() != "12345678901234567890" {
		t.Fatalf("x=%v", &x)
	}
}

func TestBigIntFromInt(t *testing.T) {
	var x big.Int
	f := unface.New(unface.With(unface.BigIntPlugin))
	if err := f.Unface(42, &x); err != nil {
		t.Fatal(err)
	}
	if x.Int64() != 42 {
		t.Fatalf("x=%v", &x)
	}
}

func TestBigFloatFromFloat(t *testing.T) {
	var x big.Float
	f := unface.New(unface.With(unface.BigFloatPlugin))
	if err := f.Unface(1.5, &x); err != nil {
		t.Fatal(err)
	}
	if v, _ := x.Float64(); v != 1.5 {
		t.Fatalf("x=%v", &x)
	}
}

func TestBigFloatFromString(t *testing.T) {
	var x big.Float
	f := unface.New(unface.With(unface.BigFloatPlugin))
	if err := f.Unface("3.14159", &x); err != nil {
		t.Fatal(err)
	}
	if v, _ := x.Float64(); v < 3.14 || v > 3.15 {
		t.Fatalf("x=%v", &x)
	}
}

func TestBigIntInvalidString(t *testing.T) {
	var x big.Int
	f := unface.New(unface.With(unface.BigIntPlugin))
	if err := f.Unface("not-a-number", &x); err == nil {
		t.Fatal("expected parse error")
	}
}
