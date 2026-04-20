# Reference: scalar plugins

**What this page covers.** The four non-numeric scalar plugins: `StringPlugin`, `BoolPlugin`, `BytesPlugin`, `RunePlugin`. All match their dest type exactly (e.g. `t == reflect.TypeOf("")`); named types with the same underlying kind are **not** auto-handled.

**Package:** `github.com/schneid-l/unface/unfacers` (re-exported from `github.com/schneid-l/unface`)

## StringPlugin

**Dest type:** `*string`

| Source | Result |
|---|---|
| `string` | as-is |
| `[]byte` | `string(v)` |
| `bool` | `"true"` / `"false"` (via `fmt.Sprintf("%t", v)`) |
| any numeric (int*, uint*, float*, complex*, big.Int, big.Float) | canonical base-10 representation from `plugin.Number.String()` |
| `nil` | `""` |
| anything else | `ErrNotHandled` |

```go
var s string
_ = unface.Unface(42, &s)       // "42"
_ = unface.Unface(true, &s)     // "true"
_ = unface.Unface([]byte("x"), &s) // "x"
```

## BoolPlugin

**Dest type:** `*bool`

| Source | Result |
|---|---|
| `bool` | as-is |
| `string` matching `true/false/yes/no/y/n/on/off/1/0/enabled/disabled` (case-insensitive, trimmed) | parsed |
| any numeric | `v != 0` |
| other string | hard error `unface/bool: cannot parse ... as bool` |

```go
var b bool
_ = unface.Unface("ENABLED", &b) // true
_ = unface.Unface(0, &b)          // false
```

## BytesPlugin

**Dest type:** `*[]byte`

| Source | Result |
|---|---|
| `[]byte` | **defensive copy** — the dest never aliases the caller's buffer |
| `string` | `[]byte(v)` |
| `nil` | `nil` |
| anything else | `ErrNotHandled` |

Because `assignDirect` deliberately skips reference kinds, you can rely on `BytesPlugin` to always deep-copy.

```go
var b []byte
src := []byte("hello")
_ = unface.Unface(src, &b)
src[0] = 'H' // dest unchanged
```

## RunePlugin

**Dest type:** `*rune` (i.e. `*int32` matched by exact type)

| Source | Result | Error cases |
|---|---|---|
| `rune` | as-is | — |
| `string` | single-codepoint string decoded via `utf8.DecodeRuneInString` | empty, multi-rune, or invalid UTF-8 → hard error |
| any integer (via `Number.Int64`) | `rune(v)` | — |
| other | `ErrNotHandled` | — |

```go
var r rune
_ = unface.Unface("🧭", &r) // r == '🧭'
_ = unface.Unface(65, &r)    // r == 'A'
```
