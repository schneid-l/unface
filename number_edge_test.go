package unface_test

import (
	"math"
	"math/big"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/schneid-l/unface"
)

func TestNumberInt64MinAndMax(t *testing.T) {
	for _, v := range []int64{math.MinInt64, math.MaxInt64, 0, -1, 1} {
		n, _ := unface.NumberOf(v)
		got, ok := n.Int64()
		if !ok || got != v {
			t.Fatalf("round-trip int64(%d): got %d ok=%v", v, got, ok)
		}
	}
}

func TestNumberFloatNaN(t *testing.T) {
	n, _ := unface.NumberOf(math.NaN())
	if _, ok := n.Int64(); ok {
		t.Fatal("NaN must not convert to int64")
	}
	if _, ok := n.Uint64(); ok {
		t.Fatal("NaN must not convert to uint64")
	}
	// Float64 round-trip preserves NaN; ok=true.
	f, ok := n.Float64()
	if !ok || !math.IsNaN(f) {
		t.Fatalf("Float64=%v ok=%v", f, ok)
	}
}

func TestNumberFloatInfinity(t *testing.T) {
	n, _ := unface.NumberOf(math.Inf(1))
	if _, ok := n.Int64(); ok {
		t.Fatal("+Inf must not convert to int64")
	}
	n2, _ := unface.NumberOf(math.Inf(-1))
	if _, ok := n2.Int64(); ok {
		t.Fatal("-Inf must not convert to int64")
	}
}

func TestNumberFloatOverflowInt64(t *testing.T) {
	// A float larger than MaxInt64 — conversion must fail.
	n, _ := unface.NumberOf(float64(math.MaxInt64) * 2.0)
	if _, ok := n.Int64(); ok {
		t.Fatal("float overflowing int64 must fail")
	}
}

func TestNumberComplexImaginaryRejects(t *testing.T) {
	n, _ := unface.NumberOf(complex(1, 2))
	if _, ok := n.Int64(); ok {
		t.Fatal("complex with imag≠0 must not convert to int64")
	}
	if _, ok := n.Uint64(); ok {
		t.Fatal("complex with imag≠0 must not convert to uint64")
	}
	if _, ok := n.Float64(); ok {
		t.Fatal("complex with imag≠0 must not convert to float64")
	}
	if _, ok := n.BigFloat(); ok {
		t.Fatal("complex with imag≠0 must not convert to big.Float")
	}
}

func TestNumberBigIntOverflow(t *testing.T) {
	// Big int beyond int64 range.
	b, _ := new(big.Int).SetString("99999999999999999999999", 10)
	n, _ := unface.NumberOf(b)
	if _, ok := n.Int64(); ok {
		t.Fatal("big.Int beyond int64 must fail")
	}
	if _, ok := n.Uint64(); ok {
		t.Fatal("big.Int beyond uint64 must fail")
	}
	// BigInt accessor always works.
	got, ok := n.BigInt()
	if !ok || got.Cmp(b) != 0 {
		t.Fatalf("BigInt round-trip failed: %v ok=%v", got, ok)
	}
}

func TestNumberOfNilBigInt(t *testing.T) {
	var b *big.Int
	if _, ok := unface.NumberOf(b); ok {
		t.Fatal("NumberOf(nil *big.Int) must be !ok")
	}
	var f *big.Float
	if _, ok := unface.NumberOf(f); ok {
		t.Fatal("NumberOf(nil *big.Float) must be !ok")
	}
}

func TestNumberRawReturnsOriginal(t *testing.T) {
	cases := []any{
		int(1), int8(1), int16(1), int32(1), int64(1),
		uint(1), uint8(1), uint16(1), uint32(1), uint64(1),
		float32(1.0), float64(1.0), complex64(1), complex128(1),
	}
	for _, v := range cases {
		n, _ := unface.NumberOf(v)
		if reflect.TypeOf(n.Raw()) != reflect.TypeOf(v) {
			t.Errorf("Raw type for %T got %T", v, n.Raw())
		}
	}
}

func TestNumberKindForAllTypes(t *testing.T) {
	cases := map[reflect.Kind]any{
		reflect.Int:        int(1),
		reflect.Int8:       int8(1),
		reflect.Int16:      int16(1),
		reflect.Int32:      int32(1),
		reflect.Int64:      int64(1),
		reflect.Uint:       uint(1),
		reflect.Uint8:      uint8(1),
		reflect.Uint16:     uint16(1),
		reflect.Uint32:     uint32(1),
		reflect.Uint64:     uint64(1),
		reflect.Float32:    float32(1),
		reflect.Float64:    float64(1),
		reflect.Complex64:  complex64(1),
		reflect.Complex128: complex128(1),
	}
	for want, v := range cases {
		n, _ := unface.NumberOf(v)
		if n.Kind() != want {
			t.Errorf("%T: Kind=%v want %v", v, n.Kind(), want)
		}
	}
}

// Property: int64 → NumberOf → Int64 round-trips for all int64 values.
func TestNumberInt64RoundTripProperty(t *testing.T) {
	f := func(v int64) bool {
		n, ok := unface.NumberOf(v)
		if !ok {
			return false
		}
		got, ok := n.Int64()
		return ok && got == v
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

// Property: non-negative int64 round-trips through Uint64.
func TestNumberUint64RoundTripProperty(t *testing.T) {
	f := func(v uint32) bool {
		n, ok := unface.NumberOf(uint64(v))
		if !ok {
			return false
		}
		got, ok := n.Uint64()
		return ok && got == uint64(v)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

// Property: float64 → NumberOf → Float64 round-trips (modulo NaN).
func TestNumberFloat64RoundTripProperty(t *testing.T) {
	f := func(v float64) bool {
		if math.IsNaN(v) {
			return true // excluded; tested separately
		}
		n, ok := unface.NumberOf(v)
		if !ok {
			return false
		}
		got, ok := n.Float64()
		return ok && got == v
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

func TestNumberIsZeroForAll(t *testing.T) {
	zeroes := []any{
		int(0), int8(0), int16(0), int32(0), int64(0),
		uint(0), uint8(0), uint16(0), uint32(0), uint64(0),
		float32(0), float64(0), complex64(0), complex128(0),
		new(big.Int), new(big.Float),
	}
	for _, v := range zeroes {
		n, _ := unface.NumberOf(v)
		if !n.IsZero() {
			t.Errorf("IsZero for %T should be true", v)
		}
	}
}

func TestNumberStringVariants(t *testing.T) {
	cases := []struct {
		v    any
		want string
	}{
		{int64(-42), "-42"},
		{uint64(42), "42"},
		{float64(1.5), "1.5"},
	}
	for _, tc := range cases {
		n, _ := unface.NumberOf(tc.v)
		if got := n.String(); got != tc.want {
			t.Errorf("%T(%v): String=%q want %q", tc.v, tc.v, got, tc.want)
		}
	}
}
