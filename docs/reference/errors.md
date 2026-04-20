# Reference: errors

**What this page covers.** The error types, sentinels, and helpers exposed by `unface`. For the conceptual model (hard vs. soft, why there are two tiers) see [concepts/errors.md](../concepts/errors.md).

Godoc: <https://pkg.go.dev/github.com/schneid-l/unface/plugin>

## Sentinels

| Symbol | Check | Fires when |
|---|---|---|
| `unface.ErrNotHandled` | `plugin.IsUnhandled(err)` | Soft signal from any `Un*er` / adapter meaning "try the next candidate". Don't use `errors.Is` alone — there's also the `Unhandled() bool` interface form. |
| `unface.ErrInvalidDest` | `errors.Is(err, unface.ErrInvalidDest)` | `dest` is `nil` or not a non-nil pointer. Returned from `Unface` / `Facer.Unface`. |
| `unface.ErrNoCoercion` | `errors.Is(err, unface.ErrNoCoercion)` | Dispatcher exhausted every step (Unfacer, assignDirect, Un*er, plugins) without a hit. |
| `unface.ErrSrcNil` | `errors.Is(err, unface.ErrSrcNil)` | Reserved sentinel for plugins that want to reject nil src with a distinct error. |
| `unface.ErrUnknownField` | `errors.Is(err, unface.ErrUnknownField)` | `OnUnknown(UnknownError)` set and a source key doesn't match any field. Wrapped with the key name. |
| `unface.ErrRequired` | `errors.Is(err, unface.ErrRequired)` | Struct walker saw no source value for a field tagged `required`. |

## Helpers

### `unface.Skip() error`

Returns `ErrNotHandled`. Prefer this over referencing the sentinel directly when writing an `Un*er`:

```go
func (u *URL) Unstring(s string) error {
    if !strings.Contains(s, "://") {
        return unface.Skip()
    }
    return u.parse(s)
}
```

### `unface.IsUnhandled(err error) bool`

Single check covering all soft-error forms: plain `ErrNotHandled`, `fmt.Errorf("...: %w", ErrNotHandled)`, and custom types implementing `Unhandled() bool`. Use this instead of `errors.Is(err, ErrNotHandled)` when authoring code that consumes user errors.

### The `Unhandled` interface

```go
type Unhandled interface {
    Unhandled() bool
}
```

Lets custom error types opt into the soft-error protocol without wrapping. Rarely needed; wrapping `ErrNotHandled` is simpler.

## The `Error` type

Aliased from `plugin.Error`:

```go
type Error struct {
    Path []string // dotted path from the root dest to the failure site
    Src  any
    Dest any
    Err  error
}
```

- `Error.Error() string` — formatted as `"unface: <path>: <inner error>"`.
- `Error.Unwrap() error` — returns `Err`, so `errors.Is` / `errors.As` see through it.
- `Error.WithPath(seg ...string) *Error` — returns a new `Error` with extra path segments prepended (used internally by the struct walker).

Use `errors.As` to pull one out:

```go
var e *unface.Error
if errors.As(err, &e) {
    log.Printf("path=%v cause=%v", e.Path, e.Err)
}
```

## Reading a message

```
unface: server.listeners.0.port: strconv.ParseInt: parsing "x": invalid syntax
```

- `server.listeners.0.port` is the path from the root dest down to the failing field.
- After the final colon is the underlying error.

## Related options

- [`OnSoftError`](./options.md#onsofterrorh-softerrorhandler-option) — observe soft errors as they unwind.
- [`OnUnknown`](./options.md#onunknownp-unknownpolicy-handler-unknownhandler-option) — choose behavior for unknown source keys (drives `ErrUnknownField`).
