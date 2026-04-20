# Errors

**What this page covers.** The two-tier error protocol (hard vs. soft), the `*Error` path-tracing type, and how to check for specific failure modes.

## Two tiers

| Tier | Example | What the dispatcher does |
|---|---|---|
| Hard | `fmt.Errorf("parse int: %w", err)`, `ErrInvalidDest`, `ErrRequired`, `ErrNoCoercion` | Abort the pipeline immediately and return the error up. |
| Soft | `ErrNotHandled`, anything wrapping it, any error implementing `Unhandled() bool` | "I decline" — dispatcher moves to the next candidate. If nothing handles the call it finally returns `ErrNoCoercion`. |

Rule of thumb when writing an `Un*er`: return soft when *you don't know* how to map the given source; return hard when *you know but the data is bad*.

## Soft-error idioms

```go
// Sentinel:
return unface.ErrNotHandled

// Helper (same thing, reads nicer at call sites):
return unface.Skip()

// Wrapping with context:
return fmt.Errorf("myplugin: %s: %w", s, unface.ErrNotHandled)

// Custom type (no wrapping needed):
type silentSkip struct{}
func (silentSkip) Error() string   { return "skipped" }
func (silentSkip) Unhandled() bool { return true }
```

`plugin.IsUnhandled(err)` is the canonical check; it handles all three styles.

## Sentinels

| Sentinel | Returned when |
|---|---|
| `ErrNotHandled` | Soft signal — try the next candidate. |
| `ErrInvalidDest` | `dest` is `nil` or not a non-nil pointer. |
| `ErrNoCoercion` | Pipeline exhausted every step and nothing matched. |
| `ErrSrcNil` | Reserved for plugins that want to reject nil src explicitly. |
| `ErrUnknownField` | `OnUnknown(UnknownError)` saw an unknown source key. |
| `ErrRequired` | Struct walker found no source value for a `required` field. |

All are `errors.Is`-compatible:

```go
if errors.Is(err, unface.ErrRequired) { ... }
```

## The `Error` path-tracing type

The struct walker wraps child-field errors in `*plugin.Error` (aliased as `unface.Error`) with a dotted path:

```go
type Error struct {
    Path []string
    Src  any
    Dest any
    Err  error
}
```

Example message:

```
unface: server.endpoints.0.port: strconv.ParseInt: parsing "abc": invalid syntax
```

Unwrap it for the underlying cause:

```go
var e *unface.Error
if errors.As(err, &e) {
    log.Printf("field=%s cause=%v", strings.Join(e.Path, "."), e.Err)
}
```

`(*Error).Unwrap` returns `e.Err`, so `errors.Is(err, ErrRequired)` still works even when the error is wrapped.

## Soft-error observability

Install a handler to see every soft return as the dispatcher traverses:

```go
f := unface.New(
    unface.With(unface.StandardPlugin),
    unface.OnSoftError(func(src, dest any, err error) {
        log.Printf("soft: src=%T dest=%T err=%v", src, dest, err)
    }),
)
```

Useful for "why wasn't my field populated?" debugging. The handler can't change behavior — it observes only.

## `ErrNoCoercion` vs. missing field

- `ErrRequired` — the **map entry** was absent.
- `ErrNoCoercion` — the map entry existed but nothing in the plugin pipeline knew how to turn it into the destination type.

These are distinct. Build your error messages around the distinction.
