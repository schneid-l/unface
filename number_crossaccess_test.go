package unface_test

import (
	"math/big"
	"testing"

	"github.com/schneid-l/unface"
)

// TestNumberCrossAccessors exercises every accessor on every underlying
// Number variant. Catches any accessor whose code path was never run by
// other tests.
func TestNumberCrossAccessors(t *testing.T) {
	cases := []struct {
		name string
		v    any
	}{
		{"int", int(5)},
		{"int8", int8(5)},
		{"int16", int16(5)},
		{"int32", int32(5)},
		{"int64", int64(5)},
		{"uint", uint(5)},
		{"uint8", uint8(5)},
		{"uint16", uint16(5)},
		{"uint32", uint32(5)},
		{"uint64", uint64(5)},
		{"float32", float32(5)},
		{"float64", float64(5)},
		{"complex64", complex64(complex(5, 0))},
		{"complex128", complex128(complex(5, 0))},
		{"big.Int", big.NewInt(5)},
		{"big.Float", big.NewFloat(5)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n, ok := unface.NumberOf(tc.v)
			if !ok {
				t.Fatalf("NumberOf(%T) not ok", tc.v)
			}
			// Int64
			if v, ok := n.Int64(); !ok || v != 5 {
				t.Errorf("Int64=%d ok=%v", v, ok)
			}
			// Uint64
			if v, ok := n.Uint64(); !ok || v != 5 {
				t.Errorf("Uint64=%d ok=%v", v, ok)
			}
			// Float64
			if v, ok := n.Float64(); !ok || v != 5 {
				t.Errorf("Float64=%v ok=%v", v, ok)
			}
			// Complex128
			if v, ok := n.Complex128(); !ok || real(v) != 5 {
				t.Errorf("Complex128=%v ok=%v", v, ok)
			}
			// BigInt
			if v, ok := n.BigInt(); !ok || v.Int64() != 5 {
				t.Errorf("BigInt=%v ok=%v", v, ok)
			}
			// BigFloat
			if v, ok := n.BigFloat(); !ok {
				t.Errorf("BigFloat ok=%v", ok)
			} else if f, _ := v.Float64(); f != 5 {
				t.Errorf("BigFloat value=%v", f)
			}
			// Kind is non-nil via the Number interface (concrete types
			// may return Invalid for big.* which is fine).
			_ = n.Kind()
			_ = n.Raw()
			_ = n.String()
			_ = n.IsZero()
		})
	}
}

// BoolPlugin matches the exact bool type only (not named types) —
// consistent with every other atomic scalar plugin.
func TestBoolPluginRejectsNamedBool(t *testing.T) {
	type Enabled bool
	var e Enabled
	f := unface.New(unface.With(unface.BoolPlugin))
	err := f.Unface(true, &e)
	if err == nil {
		t.Fatal("BoolPlugin should not match named bool types")
	}
}

// Ensure the complex accessor rejects for negative-sqrt-like edge cases
// is also exercised by constructing non-real complex numbers.
func TestNumberComplexWithImaginary(t *testing.T) {
	n, _ := unface.NumberOf(complex(1, 1))
	if _, ok := n.BigInt(); ok {
		t.Fatal("complex with imag should not convert to BigInt")
	}
	// Complex128 accessor always works.
	if v, ok := n.Complex128(); !ok || v != complex(1, 1) {
		t.Fatalf("Complex128=%v ok=%v", v, ok)
	}
	// Raw round-trips.
	if n.Raw().(complex128) != complex(1, 1) {
		t.Fatalf("Raw=%v", n.Raw())
	}
	// String and IsZero don't panic.
	_ = n.String()
	if n.IsZero() {
		t.Fatal("1+i is not zero")
	}
}

// Exercise numBigInt and numBigFloat paths that aren't on the happy path.
func TestNumberBigIntAccessors(t *testing.T) {
	b := big.NewInt(42)
	n, _ := unface.NumberOf(b)
	if n.Kind() != 0 { // reflect.Invalid
		t.Fatalf("big.Int Kind=%v want Invalid (0)", n.Kind())
	}
	// Raw returns a copy.
	r := n.Raw().(*big.Int)
	if r.Cmp(b) != 0 {
		t.Fatalf("Raw=%v want %v", r, b)
	}
	if n.String() != "42" {
		t.Fatalf("String=%q", n.String())
	}
	// Complex128 works for integer big values.
	if c, ok := n.Complex128(); !ok || real(c) != 42 {
		t.Fatalf("Complex128=%v ok=%v", c, ok)
	}
}

func TestNumberBigFloatAccessors(t *testing.T) {
	b := big.NewFloat(42.5)
	n, _ := unface.NumberOf(b)
	if n.Kind() != 0 {
		t.Fatalf("big.Float Kind=%v", n.Kind())
	}
	// Int64 fails for non-integer.
	if _, ok := n.Int64(); ok {
		t.Fatal("42.5 should not convert to Int64")
	}
	// Uint64 fails for non-integer.
	if _, ok := n.Uint64(); ok {
		t.Fatal("42.5 should not convert to Uint64")
	}
	// BigInt fails for non-integer.
	if _, ok := n.BigInt(); ok {
		t.Fatal("42.5 should not convert to BigInt")
	}
	// Complex128 works (real part).
	if c, ok := n.Complex128(); !ok || real(c) != 42.5 {
		t.Fatalf("Complex128=%v ok=%v", c, ok)
	}
	// Raw.
	r := n.Raw().(*big.Float)
	if f, _ := r.Float64(); f != 42.5 {
		t.Fatalf("Raw=%v", r)
	}
	_ = n.String()
	if n.IsZero() {
		t.Fatal("42.5 is not zero")
	}
}

func TestNumberBigFloatNegativeRejectsUint(t *testing.T) {
	n, _ := unface.NumberOf(big.NewFloat(-5))
	if _, ok := n.Uint64(); ok {
		t.Fatal("negative big.Float must not convert to Uint64")
	}
}
