package unface_test

import (
	"encoding/json"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/schneid-l/unface"
)

// --- StringPlugin edges ---

func TestStringPluginFromNilZeros(t *testing.T) {
	s := "prior"
	if err := unface.New(unface.With(unface.StringPlugin)).Unface(nil, &s); err != nil {
		t.Fatal(err)
	}
	if s != "" {
		t.Fatalf("s=%q", s)
	}
}

func TestStringPluginUnsupportedSrc(t *testing.T) {
	var s string
	err := unface.New(unface.With(unface.StringPlugin)).Unface([]int{1, 2}, &s)
	if err == nil {
		t.Fatal("expected failure on unsupported source")
	}
}

// --- BoolPlugin edges ---

func TestBoolPluginAllPositiveStrings(t *testing.T) {
	for _, in := range []string{" true ", " YES ", "Y", "On", "1", "enabled"} {
		var b bool
		if err := unface.New(unface.With(unface.BoolPlugin)).Unface(in, &b); err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		if !b {
			t.Fatalf("%q → false", in)
		}
	}
}

func TestBoolPluginFromFloat(t *testing.T) {
	for _, in := range []any{1.5, -0.1} {
		var b bool
		if err := unface.New(unface.With(unface.BoolPlugin)).Unface(in, &b); err != nil {
			t.Fatal(err)
		}
		if !b {
			t.Fatalf("non-zero float %v → false", in)
		}
	}
	var b bool
	if err := unface.New(unface.With(unface.BoolPlugin)).Unface(0.0, &b); err != nil {
		t.Fatal(err)
	}
	if b {
		t.Fatal("0.0 → true")
	}
}

func TestBoolPluginUnsupported(t *testing.T) {
	var b bool
	if err := unface.New(unface.With(unface.BoolPlugin)).Unface([]byte{1}, &b); err == nil {
		t.Fatal("expected failure")
	}
}

// --- BytesPlugin edges ---

func TestBytesPluginFromNil(t *testing.T) {
	b := []byte{1, 2}
	if err := unface.New(unface.With(unface.BytesPlugin)).Unface(nil, &b); err != nil {
		t.Fatal(err)
	}
	if b != nil {
		t.Fatalf("b=%v", b)
	}
}

func TestBytesPluginUnsupported(t *testing.T) {
	var b []byte
	if err := unface.New(unface.With(unface.BytesPlugin)).Unface(42, &b); err == nil {
		t.Fatal("expected failure")
	}
}

func TestBytesPluginCopyIsIndependent(t *testing.T) {
	src := []byte{1, 2, 3}
	var dst []byte
	if err := unface.New(unface.With(unface.BytesPlugin)).Unface(src, &dst); err != nil {
		t.Fatal(err)
	}
	src[0] = 99
	if dst[0] == 99 {
		t.Fatal("BytesPlugin must copy, not alias")
	}
}

// --- RunePlugin edges ---

func TestRunePluginEmptyString(t *testing.T) {
	var r rune
	if err := unface.New(unface.With(unface.RunePlugin)).Unface("", &r); err == nil {
		t.Fatal("empty string should fail")
	}
}

func TestRunePluginMultiChar(t *testing.T) {
	var r rune
	if err := unface.New(unface.With(unface.RunePlugin)).Unface("abc", &r); err == nil {
		t.Fatal("multi-char string should fail")
	}
}

func TestRunePluginInvalidUTF8(t *testing.T) {
	var r rune
	// Deliberately invalid UTF-8 single byte.
	if err := unface.New(unface.With(unface.RunePlugin)).Unface(string([]byte{0xff}), &r); err == nil {
		t.Fatal("invalid utf-8 should fail")
	}
}

func TestRunePluginUnicode(t *testing.T) {
	var r rune
	if err := unface.New(unface.With(unface.RunePlugin)).Unface("é", &r); err != nil {
		t.Fatal(err)
	}
	if r != 'é' {
		t.Fatalf("r=%q", r)
	}
}

// --- Integer plugin edges ---

func TestIntPluginFromFloatNonInteger(t *testing.T) {
	var x int64
	err := unface.New(unface.With(unface.Int64Plugin)).Unface(1.5, &x)
	if err == nil {
		t.Fatal("non-integer float should fail")
	}
}

func TestIntPluginInvalidString(t *testing.T) {
	var x int64
	err := unface.New(unface.With(unface.Int64Plugin)).Unface("not-a-number", &x)
	if err == nil {
		t.Fatal("invalid string should fail")
	}
}

func TestIntPluginHexString(t *testing.T) {
	var x int64
	if err := unface.New(unface.With(unface.Int64Plugin)).Unface("0xff", &x); err != nil {
		t.Fatal(err)
	}
	if x != 255 {
		t.Fatalf("x=%d", x)
	}
}

func TestUintPluginFromLargeUint64(t *testing.T) {
	var x uint64
	if err := unface.New(unface.With(unface.Uint64Plugin)).Unface(uint64(math.MaxUint64), &x); err != nil {
		t.Fatal(err)
	}
	if x != math.MaxUint64 {
		t.Fatalf("x=%d", x)
	}
}

func TestUintPluginOverflowFromString(t *testing.T) {
	var x uint8
	err := unface.New(unface.With(unface.Uint8Plugin)).Unface("300", &x)
	if err == nil {
		t.Fatal("overflow should fail")
	}
}

func TestUintPluginInvalidString(t *testing.T) {
	var x uint32
	if err := unface.New(unface.With(unface.Uint32Plugin)).Unface("abc", &x); err == nil {
		t.Fatal("invalid string should fail")
	}
}

func TestUintPluginFromBoolTrue(t *testing.T) {
	var x uint8
	if err := unface.New(unface.With(unface.Uint8Plugin)).Unface(true, &x); err != nil {
		t.Fatal(err)
	}
	if x != 1 {
		t.Fatalf("x=%d", x)
	}
}

func TestUintPluginFromBoolFalse(t *testing.T) {
	x := uint8(99)
	if err := unface.New(unface.With(unface.Uint8Plugin)).Unface(false, &x); err != nil {
		t.Fatal(err)
	}
	if x != 0 {
		t.Fatalf("x=%d", x)
	}
}

func TestUintPluginFromNilZeroes(t *testing.T) {
	x := uint32(99)
	if err := unface.New(unface.With(unface.Uint32Plugin)).Unface(nil, &x); err != nil {
		t.Fatal(err)
	}
	if x != 0 {
		t.Fatalf("x=%d", x)
	}
}

func TestUintPluginUnsupportedSrc(t *testing.T) {
	var x uint32
	err := unface.New(unface.With(unface.Uint32Plugin)).Unface([]int{1}, &x)
	if err == nil {
		t.Fatal("unsupported source should fail")
	}
}

// --- Float plugin edges ---

func TestFloatPluginInvalidString(t *testing.T) {
	var x float64
	if err := unface.New(unface.With(unface.Float64Plugin)).Unface("nope", &x); err == nil {
		t.Fatal("invalid string should fail")
	}
}

func TestFloatPluginUnsupported(t *testing.T) {
	var x float64
	if err := unface.New(unface.With(unface.Float64Plugin)).Unface(true, &x); err == nil {
		t.Fatal("bool should fail for float plugin (no conversion defined)")
	}
}

func TestFloatPluginComplexRejected(t *testing.T) {
	var x float64
	if err := unface.New(unface.With(unface.Float64Plugin)).Unface(complex(1, 2), &x); err == nil {
		t.Fatal("complex with imag≠0 should fail")
	}
}

// --- Complex plugin edges ---

func TestComplexPluginInvalidString(t *testing.T) {
	var x complex128
	if err := unface.New(unface.With(unface.Complex128Plugin)).Unface("not-complex", &x); err == nil {
		t.Fatal("invalid string should fail")
	}
}

func TestComplexPluginUnsupported(t *testing.T) {
	var x complex64
	if err := unface.New(unface.With(unface.Complex64Plugin)).Unface(true, &x); err == nil {
		t.Fatal("bool should fail for complex plugin")
	}
}

func TestComplexPluginFromNil(t *testing.T) {
	x := complex(1, 2)
	if err := unface.New(unface.With(unface.Complex128Plugin)).Unface(nil, &x); err != nil {
		t.Fatal(err)
	}
	if x != 0 {
		t.Fatalf("x=%v", x)
	}
}

// --- BigInt/BigFloat edges ---

func TestBigIntFromFloatNonInteger(t *testing.T) {
	var x big.Int
	err := unface.New(unface.With(unface.BigIntPlugin)).Unface(1.5, &x)
	// Float → big.Int is only defined when the float has integer value.
	if err == nil {
		t.Fatal("non-integer float should not convert to big.Int")
	}
}

func TestBigFloatInvalidString(t *testing.T) {
	var x big.Float
	if err := unface.New(unface.With(unface.BigFloatPlugin)).Unface("not-a-float", &x); err == nil {
		t.Fatal("invalid string should fail")
	}
}

func TestBigIntUnsupportedSrc(t *testing.T) {
	var x big.Int
	err := unface.New(unface.With(unface.BigIntPlugin)).Unface([]int{1, 2}, &x)
	if err == nil {
		t.Fatal("unsupported source should fail")
	}
}

func TestBigFloatUnsupportedSrc(t *testing.T) {
	var x big.Float
	err := unface.New(unface.With(unface.BigFloatPlugin)).Unface([]int{1}, &x)
	if err == nil {
		t.Fatal("unsupported source should fail")
	}
}

// --- Time / Duration edges ---

func TestTimePluginUnsupportedSrc(t *testing.T) {
	var tm time.Time
	err := unface.New(unface.With(unface.TimePlugin)).Unface([]int{1}, &tm)
	if err == nil {
		t.Fatal("unsupported src should fail for TimePlugin")
	}
}

func TestTimePluginFromNilZeroes(t *testing.T) {
	tm := time.Unix(42, 0)
	if err := unface.New(unface.With(unface.TimePlugin)).Unface(nil, &tm); err != nil {
		t.Fatal(err)
	}
	if !tm.IsZero() {
		t.Fatalf("tm=%v", tm)
	}
}

func TestDurationPluginUnsupportedSrc(t *testing.T) {
	var d time.Duration
	err := unface.New(unface.With(unface.TimePlugin)).Unface([]int{1}, &d)
	if err == nil {
		t.Fatal("unsupported src should fail for duration")
	}
}

func TestDurationFromNilZeroes(t *testing.T) {
	d := time.Minute
	if err := unface.New(unface.With(unface.TimePlugin)).Unface(nil, &d); err != nil {
		t.Fatal(err)
	}
	if d != 0 {
		t.Fatalf("d=%v", d)
	}
}

func TestDurationFromDurationDirect(t *testing.T) {
	var d time.Duration
	if err := unface.New(unface.With(unface.TimePlugin)).Unface(42*time.Second, &d); err != nil {
		t.Fatal(err)
	}
	if d != 42*time.Second {
		t.Fatalf("d=%v", d)
	}
}

// --- JSONRaw edges ---

func TestJSONRawFromBytesNonJSON(t *testing.T) {
	// []byte path copies the bytes verbatim; it does NOT validate JSON.
	var raw json.RawMessage
	in := []byte("not-json")
	if err := unface.New(unface.With(unface.JSONRawPlugin)).Unface(in, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw) != "not-json" {
		t.Fatalf("raw=%s", raw)
	}
}

func TestJSONRawFromUnmarshalableFails(t *testing.T) {
	var raw json.RawMessage
	bad := make(chan int)
	err := unface.New(unface.With(unface.JSONRawPlugin)).Unface(bad, &raw)
	if err == nil {
		t.Fatal("chan is not JSON-marshalable; expected error")
	}
}

// --- Plugin soft-error propagation ---

func TestSoftErrorHandlerIsInvoked(t *testing.T) {
	var observed []error
	handler := func(_, _ any, err error) { observed = append(observed, err) }

	// dest implements Unstringer and returns ErrNotHandled; soft-error
	// handler must observe the decline.
	f := unface.New(
		unface.With(unface.StandardPlugin),
		unface.OnSoftError(handler),
	)
	var d decliningString
	_ = f.Unface("x", &d)
	if len(observed) == 0 {
		t.Fatal("soft-error handler not invoked")
	}
}
