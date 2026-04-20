# Guide: custom types via `Un*er`

**What this page covers.** Teaching the library how to populate your own types by implementing one or more `Un*er` interfaces — no plugin registration required.

## The rule

If your destination type implements any of these, the dispatcher calls your method **before** falling back to plugins.

| Interface | Method | Fires when src is |
|---|---|---|
| `Unfacer` | `Unface(src any) error` | *anything* (master override) |
| `Unstringer` | `Unstring(s string) error` | `string` |
| `Unbooler` | `Unbool(b bool) error` | `bool` |
| `Unbyteser` | `Unbytes(b []byte) error` | `[]byte` (falls back to `Unstringer` if absent) |
| `Unruner` | `Unrune(r rune) error` | `rune` |
| `Unnumberer` | `Unnumber(n Number) error` | any numeric kind |
| `Unlister` | `Unlist(l List) error` | slice / array |
| `Unmapper` | `Unmap(m Map) error` | map |
| `Untimer` | `Untime(t time.Time) error` | `time.Time` |
| `Undurationer` | `Unduration(d time.Duration) error` | `time.Duration` |
| `Unniler` | `Unnil() error` | explicit `nil` |

Implement as many as you want. A type that only knows how to be populated from strings implements `Unstringer` only; the dispatcher falls through to plugins for other source kinds.

## Example: a three-state enum

```go
package traffic

import (
    "fmt"

    "github.com/schneid-l/unface"
)

type Light int

const (
    Red Light = iota
    Yellow
    Green
)

func (l *Light) Unstring(s string) error {
    switch s {
    case "red":    *l = Red
    case "yellow": *l = Yellow
    case "green":  *l = Green
    default:
        return fmt.Errorf("traffic: unknown color %q", s)
    }
    return nil
}

func (l *Light) Unnumber(n unface.Number) error {
    v, ok := n.Int64()
    if !ok || v < 0 || v > 2 {
        return fmt.Errorf("traffic: out of range")
    }
    *l = Light(v)
    return nil
}
```

Usage:

```go
var l traffic.Light
_ = unface.Unface("green", &l) // Light == Green
_ = unface.Unface(1, &l)       // Light == Yellow
```

## Falling through

Return `unface.Skip()` (i.e. `ErrNotHandled`) if you want the default plugins to try. Useful when your method only handles a *subset* of the source kind:

```go
func (u *URL) Unstring(s string) error {
    if !strings.Contains(s, "://") {
        return unface.Skip() // defer to plugin / default
    }
    // ... parse ...
    return nil
}
```

## When to prefer `Unfacer`

If you want a single entry point (for example to hit a codegen fast path), implement `Unfacer` directly:

```go
func (u *URL) Unface(src any) error {
    switch v := src.(type) {
    case string:
        return u.parseString(v)
    case map[string]any:
        return unface.Unface(v, (*urlAlias)(u)) // delegate back to walker
    }
    return unface.ErrNotHandled
}
```

`Unfacer` takes precedence over the specific `Un*er` methods (step 3 in the [dispatch order](../dispatch-order.md)). When invoking the generator, `unfacegen` produces an `Unface` that fans out to your specific methods — you get the fast path for free.

## Errors

- Return a plain error for "bad data" — it propagates up and is wrapped with the field path by the struct walker.
- Return `unface.ErrNotHandled` / `unface.Skip()` to fall through to plugins (soft error).

See [../concepts/errors.md](../concepts/errors.md) for the full protocol.
