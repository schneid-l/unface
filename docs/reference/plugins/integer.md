# Reference: integer plugins

**What this page covers.** The signed and unsigned integer plugins, their overflow semantics, and accepted source types.

**Package:** `github.com/schneid-l/unface/unfacers`

## Signed

| Plugin | Dest type |
|---|---|
| `Int8Plugin` | `*int8` |
| `Int16Plugin` | `*int16` |
| `Int32Plugin` | `*int32` |
| `Int64Plugin` | `*int64` |
| `IntPlugin` | `*int` |

Bundled as `IntPluginBundle`.

## Unsigned

| Plugin | Dest type |
|---|---|
| `Uint8Plugin` | `*uint8` |
| `Uint16Plugin` | `*uint16` |
| `Uint32Plugin` | `*uint32` |
| `Uint64Plugin` | `*uint64` |
| `UintPlugin` | `*uint` |

Bundled as `UintPluginBundle`.

Note: `*uint8` is the same underlying type as `*byte`. `BytesPlugin` (`*[]byte`) is a separate plugin that deals with the slice form.

## Accepted sources

All integer plugins share one adapter with identical rules:

| Source | Result | Notes |
|---|---|---|
| `nil` | zero value | — |
| `bool` | `1` / `0` | — |
| any numeric (`Number`) | coerced with overflow check | unsigned plugins reject negative and non-representable values |
| `string` | parsed via `strconv.ParseInt` (signed) / `ParseUint` (unsigned), base `0` | `"0x"`, `"0b"`, `"0o"` prefixes accepted |
| other | `ErrNotHandled` | — |

## Overflow

Overflow returns a hard error of the form:

```
unface: 300 overflows int8
```

Rules:

- Signed: use `Number.Int64()` first. If the source is a `uint64` > `math.MaxInt64`, hard error.
- Unsigned: use `Number.Uint64()`. A negative source fails (`Uint64` returns `ok=false`, which becomes `ErrNotHandled` → ends up as `ErrNoCoercion`).
- Narrowing: after widening to 64-bit, the adapter casts to `T` and verifies round-trip (`int64(T(v)) == v`). A mismatch is a hard error.

## Error modes

| Cause | Kind |
|---|---|
| `strconv.ParseInt/ParseUint` failure | hard error (wrapped) |
| Overflow on numeric source | hard error |
| Negative into unsigned | soft (`ErrNotHandled`) — lets other plugins try |
| Unsupported source kind | soft (`ErrNotHandled`) |

## Example

```go
var n int16
_ = unface.Unface("0x7f", &n) // n == 127

var m uint32
_ = unface.Unface(3.0, &m)    // m == 3 (integer-valued float ok)

var b byte
_ = unface.Unface("255", &b)  // b == 255; "256" would be an overflow error
```
