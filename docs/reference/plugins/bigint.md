# Reference: big number plugins

**What this page covers.** Arbitrary-precision numeric coercion for `*big.Int` and `*big.Float`.

**Package:** `github.com/schneid-l/unface/unfacers`

## BigIntPlugin

**Dest type:** `*big.Int`

| Source | Result |
|---|---|
| `nil` | set to 0 |
| any numeric | `Number.BigInt()` → `dst.Set(...)`. Soft-fails if source isn't integer-valued (e.g. `3.14`) |
| `string` | `dst.SetString(s, 0)` — accepts base-10, hex (`0x`), binary (`0b`), octal (`0o`) |
| other | `ErrNotHandled` |

```go
var n big.Int
_ = unface.Unface("0xdeadbeef", &n)
```

## BigFloatPlugin

**Dest type:** `*big.Float`

| Source | Result |
|---|---|
| `nil` | set to 0 |
| any numeric | `Number.BigFloat()` → `dst.Set(...)` |
| `string` | `dst.SetString(s)` — accepts decimal, exponent notation |
| other | `ErrNotHandled` |

```go
var f big.Float
_ = unface.Unface("3.1415926535897932384626433832795", &f)
```

## Error modes

| Cause | Kind |
|---|---|
| `SetString` rejects input | hard error (`unface: cannot parse big.Int from %q`) |
| `Number.BigInt()` soft-fails because source is fractional | soft → `ErrNoCoercion` if nothing else matches |
| Unsupported source | soft |

## Matching

Both plugins match the pointed-to type exactly (`t == reflect.TypeOf(big.Int{})` / `big.Float{}`). Named types wrapping `big.Int` won't be auto-handled; implement `Unnumberer` / `Unstringer` on them, or register your own plugin.
