package unface_test

import (
	"math"
	"math/big"
	"reflect"
	"testing"

	"github.com/schneid-l/unface"
)

func TestNumberOfAcceptsAllIntegerKinds(t *testing.T) {
	for _, in := range []any{
		int(1), int8(1), int16(1), int32(1), int64(1),
		uint(1), uint8(1), uint16(1), uint32(1), uint64(1),
	} {
		n, ok := unface.NumberOf(in)
		if !ok {
			t.Fatalf("NumberOf(%T) not ok", in)
		}
		if v, ok := n.Int64(); !ok || v != 1 {
			t.Fatalf("Int64 got %v ok=%v", v, ok)
		}
	}
}

func TestNumberOfAcceptsFloats(t *testing.T) {
	for _, in := range []any{float32(1.5), float64(1.5)} {
		n, _ := unface.NumberOf(in)
		v, ok := n.Float64()
		if !ok || math.Abs(v-1.5) > 1e-6 {
			t.Fatalf("Float64 got %v ok=%v", v, ok)
		}
	}
}

func TestNumberOfRejectsNonNumeric(t *testing.T) {
	for _, in := range []any{"1", true, nil, []int{1}} {
		if _, ok := unface.NumberOf(in); ok {
			t.Fatalf("NumberOf(%T) should be !ok", in)
		}
	}
}

func TestNumberInt64OverflowReportsFalse(t *testing.T) {
	n, _ := unface.NumberOf(uint64(math.MaxUint64))
	if _, ok := n.Int64(); ok {
		t.Fatal("Int64 should fail on uint64 max")
	}
}

func TestNumberUint64RejectsNegative(t *testing.T) {
	n, _ := unface.NumberOf(int64(-1))
	if _, ok := n.Uint64(); ok {
		t.Fatal("Uint64 should fail on negative")
	}
}

func TestNumberBigInt(t *testing.T) {
	n, _ := unface.NumberOf(int64(42))
	bi, ok := n.BigInt()
	if !ok || bi.Int64() != 42 {
		t.Fatalf("BigInt got %v ok=%v", bi, ok)
	}
}

func TestNumberKindPreserved(t *testing.T) {
	n, _ := unface.NumberOf(int32(7))
	if n.Kind() != reflect.Int32 {
		t.Fatalf("Kind=%v want int32", n.Kind())
	}
}

func TestNumberStringCanonical(t *testing.T) {
	n, _ := unface.NumberOf(int64(42))
	if n.String() != "42" {
		t.Fatalf("String=%q", n.String())
	}
	n2, _ := unface.NumberOf(float64(1.5))
	if n2.String() != "1.5" {
		t.Fatalf("String=%q", n2.String())
	}
}

func TestNumberFromBigInt(t *testing.T) {
	b := big.NewInt(123)
	n, ok := unface.NumberOf(b)
	if !ok {
		t.Fatal("NumberOf(*big.Int) must succeed")
	}
	if v, ok := n.Int64(); !ok || v != 123 {
		t.Fatalf("Int64 got %v ok=%v", v, ok)
	}
}

func TestNumberFromBigFloat(t *testing.T) {
	b := big.NewFloat(1.5)
	n, ok := unface.NumberOf(b)
	if !ok {
		t.Fatal("NumberOf(*big.Float) must succeed")
	}
	if v, ok := n.Float64(); !ok || v != 1.5 {
		t.Fatalf("Float64 got %v ok=%v", v, ok)
	}
}

func TestNumberComplex(t *testing.T) {
	n, _ := unface.NumberOf(complex128(complex(2, 3)))
	c, ok := n.Complex128()
	if !ok || c != complex(2, 3) {
		t.Fatalf("Complex128 got %v ok=%v", c, ok)
	}
	if _, ok := n.Float64(); ok {
		t.Fatal("Float64 must fail when imag != 0")
	}
}

func TestNumberIsZero(t *testing.T) {
	n, _ := unface.NumberOf(int64(0))
	if !n.IsZero() {
		t.Fatal("int64(0) should be IsZero")
	}
	n2, _ := unface.NumberOf(1.5)
	if n2.IsZero() {
		t.Fatal("1.5 should not be IsZero")
	}
}

func TestNumberFloatNonInteger(t *testing.T) {
	n, _ := unface.NumberOf(1.5)
	if _, ok := n.Int64(); ok {
		t.Fatal("Int64(1.5) must fail")
	}
}
