# Reference: float and complex plugins

**What this page covers.** Floating-point (`float32`, `float64`) and complex (`complex64`, `complex128`) coercion.

**Package:** `github.com/schneid-l/unface/unfacers`

## Plugins

| Plugin | Dest type | Bundle |
|---|---|---|
| `Float32Plugin` | `*float32` | `FloatPluginBundle` |
| `Float64Plugin` | `*float64` | `FloatPluginBundle` |
| `Complex64Plugin` | `*complex64` | `ComplexPluginBundle` |
| `Complex128Plugin` | `*complex128` | `ComplexPluginBundle` |

## Float

| Source | Result | Notes |
|---|---|---|
| `nil` | zero | — |
| any numeric | `Number.Float64()` → cast to `T` | integer sources widen without loss; big.Int / big.Float may round |
| `string` | `strconv.ParseFloat(s, 64)` → cast | accepts `"NaN"`, `"+Inf"`, exponent notation |
| other | `ErrNotHandled` | — |

```go
var f float64
_ = unface.Unface("1.5e-3", &f)  // f == 0.0015

var g float32
_ = unface.Unface(42, &g)        // g == 42.0
```

## Complex

| Source | Result | Notes |
|---|---|---|
| `nil` | zero | — |
| any numeric | `Number.Complex128()` → cast | real-only sources produce `imag=0` |
| `string` | `strconv.ParseComplex(s, 128)` | formats like `"1+2i"`, `"(3-4i)"`, `"5"` accepted |
| other | `ErrNotHandled` | — |

```go
var c complex128
_ = unface.Unface("2+3i", &c)  // c == complex(2, 3)
```

## Error modes

| Cause | Kind |
|---|---|
| `ParseFloat` / `ParseComplex` failure | hard error (wrapped) |
| `Number.Float64()` / `Number.Complex128()` returns `ok=false` (e.g. big.Float not exactly representable) | soft |
| Unsupported source | soft |

## Note on precision

- Float plugins accept integer-valued floats round-trip-safely.
- `Number.Float64()` on a very large `big.Int` may return `ok=false` (precision loss); the plugin then soft-fails.
- Complex plugins do **not** accept non-zero-imaginary when the source is an ordinary Go number — only a literal `complex*` src or a parseable string yields an imaginary component.
