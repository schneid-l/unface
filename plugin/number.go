package plugin

import (
	"math"
	"math/big"
	"reflect"
	"strconv"
)

// Number is a type-neutral view over a numeric source. It gives implementers
// of Unnumberer lossless access to the original value via typed accessors.
// Each accessor returns ok=false when the requested representation would
// lose information (overflow, negative-to-unsigned, non-integer-to-integer).
type Number interface {
	// Kind returns the original reflect.Kind of the value.
	Kind() reflect.Kind
	// Int64 returns the value as int64 if representable.
	Int64() (int64, bool)
	// Uint64 returns the value as uint64 if representable.
	Uint64() (uint64, bool)
	// Float64 returns the value as float64 if representable.
	Float64() (float64, bool)
	// Complex128 returns the value as complex128.
	Complex128() (complex128, bool)
	// BigInt returns the value as *big.Int if it has an integer value.
	BigInt() (*big.Int, bool)
	// BigFloat returns the value as *big.Float if representable.
	BigFloat() (*big.Float, bool)
	// IsZero reports whether the value equals its type's zero.
	IsZero() bool
	// Raw returns the original any value.
	Raw() any
	// String returns a canonical string form suitable for error messages.
	String() string
}

// Unnumberer consumes any numeric source via the Number abstraction.
type Unnumberer interface {
	Unnumber(n Number) error
}

// NumberOf wraps v in a Number if v is of a recognized numeric kind.
// Accepted kinds: int*, uint*, float*, complex*, *big.Int, *big.Float.
func NumberOf(v any) (Number, bool) {
	switch x := v.(type) {
	case int:
		return numInt{k: reflect.Int, v: int64(x), raw: x}, true
	case int8:
		return numInt{k: reflect.Int8, v: int64(x), raw: x}, true
	case int16:
		return numInt{k: reflect.Int16, v: int64(x), raw: x}, true
	case int32:
		return numInt{k: reflect.Int32, v: int64(x), raw: x}, true
	case int64:
		return numInt{k: reflect.Int64, v: x, raw: x}, true
	case uint:
		return numUint{k: reflect.Uint, v: uint64(x), raw: x}, true
	case uint8:
		return numUint{k: reflect.Uint8, v: uint64(x), raw: x}, true
	case uint16:
		return numUint{k: reflect.Uint16, v: uint64(x), raw: x}, true
	case uint32:
		return numUint{k: reflect.Uint32, v: uint64(x), raw: x}, true
	case uint64:
		return numUint{k: reflect.Uint64, v: x, raw: x}, true
	case float32:
		return numFloat{k: reflect.Float32, v: float64(x), raw: x}, true
	case float64:
		return numFloat{k: reflect.Float64, v: x, raw: x}, true
	case complex64:
		return numComplex{k: reflect.Complex64, v: complex128(x), raw: x}, true
	case complex128:
		return numComplex{k: reflect.Complex128, v: x, raw: x}, true
	case *big.Int:
		if x == nil {
			return nil, false
		}
		return numBigInt{v: new(big.Int).Set(x)}, true
	case *big.Float:
		if x == nil {
			return nil, false
		}
		return numBigFloat{v: new(big.Float).Set(x)}, true
	default:
		return nil, false
	}
}

// --- numInt ---

type numInt struct {
	k   reflect.Kind
	v   int64
	raw any // original typed value (int, int8, ..., int64)
}

func (n numInt) Kind() reflect.Kind   { return n.k }
func (n numInt) Int64() (int64, bool) { return n.v, true }
func (n numInt) Uint64() (uint64, bool) {
	if n.v < 0 {
		return 0, false
	}
	return uint64(n.v), true
}
func (n numInt) Float64() (float64, bool)       { return float64(n.v), true }
func (n numInt) Complex128() (complex128, bool) { return complex(float64(n.v), 0), true }
func (n numInt) BigInt() (*big.Int, bool)       { return big.NewInt(n.v), true }
func (n numInt) BigFloat() (*big.Float, bool)   { return new(big.Float).SetInt64(n.v), true }
func (n numInt) IsZero() bool                   { return n.v == 0 }
func (n numInt) Raw() any                       { return n.raw }
func (n numInt) String() string                 { return strconv.FormatInt(n.v, 10) }

// --- numUint ---

type numUint struct {
	k   reflect.Kind
	v   uint64
	raw any
}

func (n numUint) Kind() reflect.Kind { return n.k }
func (n numUint) Int64() (int64, bool) {
	if n.v > math.MaxInt64 {
		return 0, false
	}
	return int64(n.v), true
}
func (n numUint) Uint64() (uint64, bool)         { return n.v, true }
func (n numUint) Float64() (float64, bool)       { return float64(n.v), true }
func (n numUint) Complex128() (complex128, bool) { return complex(float64(n.v), 0), true }
func (n numUint) BigInt() (*big.Int, bool)       { return new(big.Int).SetUint64(n.v), true }
func (n numUint) BigFloat() (*big.Float, bool)   { return new(big.Float).SetUint64(n.v), true }
func (n numUint) IsZero() bool                   { return n.v == 0 }
func (n numUint) Raw() any                       { return n.raw }
func (n numUint) String() string                 { return strconv.FormatUint(n.v, 10) }

// --- numFloat ---

type numFloat struct {
	k   reflect.Kind
	v   float64
	raw any
}

func (n numFloat) Kind() reflect.Kind { return n.k }
func (n numFloat) Int64() (int64, bool) {
	if math.IsNaN(n.v) || math.IsInf(n.v, 0) {
		return 0, false
	}
	t, frac := math.Modf(n.v)
	if frac != 0 {
		return 0, false
	}
	if t < math.MinInt64 || t > math.MaxInt64 {
		return 0, false
	}
	return int64(t), true
}

func (n numFloat) Uint64() (uint64, bool) {
	i, ok := n.Int64()
	if !ok || i < 0 {
		return 0, false
	}
	return uint64(i), true
}
func (n numFloat) Float64() (float64, bool)       { return n.v, true }
func (n numFloat) Complex128() (complex128, bool) { return complex(n.v, 0), true }
func (n numFloat) BigInt() (*big.Int, bool) {
	i, ok := n.Int64()
	if !ok {
		return nil, false
	}
	return big.NewInt(i), true
}
func (n numFloat) BigFloat() (*big.Float, bool) { return big.NewFloat(n.v), true }
func (n numFloat) IsZero() bool                 { return n.v == 0 }
func (n numFloat) Raw() any                     { return n.raw }
func (n numFloat) String() string               { return strconv.FormatFloat(n.v, 'g', -1, 64) }

// --- numComplex ---

type numComplex struct {
	k   reflect.Kind
	v   complex128
	raw any
}

func (n numComplex) Kind() reflect.Kind { return n.k }
func (n numComplex) Int64() (int64, bool) {
	if imag(n.v) != 0 {
		return 0, false
	}
	return numFloat{v: real(n.v)}.Int64()
}

func (n numComplex) Uint64() (uint64, bool) {
	if imag(n.v) != 0 {
		return 0, false
	}
	return numFloat{v: real(n.v)}.Uint64()
}

func (n numComplex) Float64() (float64, bool) {
	if imag(n.v) != 0 {
		return 0, false
	}
	return real(n.v), true
}
func (n numComplex) Complex128() (complex128, bool) { return n.v, true }
func (n numComplex) BigInt() (*big.Int, bool) {
	i, ok := n.Int64()
	if !ok {
		return nil, false
	}
	return big.NewInt(i), true
}

func (n numComplex) BigFloat() (*big.Float, bool) {
	if imag(n.v) != 0 {
		return nil, false
	}
	return big.NewFloat(real(n.v)), true
}
func (n numComplex) IsZero() bool { return n.v == 0 }
func (n numComplex) Raw() any     { return n.raw }
func (n numComplex) String() string {
	return strconv.FormatComplex(n.v, 'g', -1, 128)
}

// --- numBigInt ---

type numBigInt struct {
	v *big.Int
}

func (n numBigInt) Kind() reflect.Kind { return reflect.Invalid }
func (n numBigInt) Int64() (int64, bool) {
	if !n.v.IsInt64() {
		return 0, false
	}
	return n.v.Int64(), true
}

func (n numBigInt) Uint64() (uint64, bool) {
	if !n.v.IsUint64() {
		return 0, false
	}
	return n.v.Uint64(), true
}

func (n numBigInt) Float64() (float64, bool) {
	f, acc := new(big.Float).SetInt(n.v).Float64()
	return f, acc == big.Exact
}

func (n numBigInt) Complex128() (complex128, bool) {
	f, ok := n.Float64()
	if !ok {
		return 0, false
	}
	return complex(f, 0), true
}
func (n numBigInt) BigInt() (*big.Int, bool)     { return new(big.Int).Set(n.v), true }
func (n numBigInt) BigFloat() (*big.Float, bool) { return new(big.Float).SetInt(n.v), true }
func (n numBigInt) IsZero() bool                 { return n.v.Sign() == 0 }
func (n numBigInt) Raw() any                     { return new(big.Int).Set(n.v) }
func (n numBigInt) String() string               { return n.v.String() }

// --- numBigFloat ---

type numBigFloat struct {
	v *big.Float
}

func (n numBigFloat) Kind() reflect.Kind { return reflect.Invalid }
func (n numBigFloat) Int64() (int64, bool) {
	if !n.v.IsInt() {
		return 0, false
	}
	i, acc := n.v.Int64()
	return i, acc == big.Exact
}

func (n numBigFloat) Uint64() (uint64, bool) {
	if !n.v.IsInt() || n.v.Sign() < 0 {
		return 0, false
	}
	u, acc := n.v.Uint64()
	return u, acc == big.Exact
}

func (n numBigFloat) Float64() (float64, bool) {
	f, acc := n.v.Float64()
	return f, acc == big.Exact
}

func (n numBigFloat) Complex128() (complex128, bool) {
	f, ok := n.Float64()
	if !ok {
		return 0, false
	}
	return complex(f, 0), true
}

func (n numBigFloat) BigInt() (*big.Int, bool) {
	if !n.v.IsInt() {
		return nil, false
	}
	i, _ := n.v.Int(nil)
	return i, true
}
func (n numBigFloat) BigFloat() (*big.Float, bool) { return new(big.Float).Set(n.v), true }
func (n numBigFloat) IsZero() bool                 { return n.v.Sign() == 0 }
func (n numBigFloat) Raw() any                     { return new(big.Float).Set(n.v) }
func (n numBigFloat) String() string               { return n.v.Text('g', -1) }
