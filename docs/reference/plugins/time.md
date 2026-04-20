# Reference: time plugin

**What this page covers.** `TimePlugin` — the single plugin that handles both `*time.Time` and `*time.Duration` destinations (it registers two factories).

**Package:** `github.com/schneid-l/unface/unfacers`

## `*time.Time`

| Source | Result |
|---|---|
| `time.Time` | as-is |
| `nil` | `time.Time{}` |
| `string` | parsed against the layout list below; first match wins |
| integer (via `Number.Int64`) | `time.Unix(v, 0).UTC()` — seconds since epoch |
| other | `ErrNotHandled` |

### Layouts tried (in order)

```
time.RFC3339Nano           // 2006-01-02T15:04:05.999999999Z07:00
time.RFC3339               // 2006-01-02T15:04:05Z07:00
"2006-01-02T15:04:05"      // local-ish, no zone
"2006-01-02 15:04:05"      // space-separated
"2006-01-02"               // date-only
time.RFC1123               // Mon, 02 Jan 2006 15:04:05 MST
```

Unparseable strings return `unface/time: cannot parse %q as time.Time` (hard error). For custom layouts, implement `Untimer` on your own wrapper type.

## `*time.Duration`

| Source | Result |
|---|---|
| `time.Duration` | as-is |
| `nil` | `0` |
| `string` | `time.ParseDuration(v)` (`"1h30m"`, `"250ms"`, ...) |
| integer (`Number.Int64`) | `time.Duration(v) * time.Second` |
| float (`Number.Float64`) | `time.Duration(v * float64(time.Second))` — fractional seconds |
| other | `ErrNotHandled` |

### Example

```go
type Config struct {
    Deadline time.Time     `unface:"deadline"`
    Backoff  time.Duration `unface:"backoff"`
}

raw := map[string]any{
    "deadline": "2026-01-01T00:00:00Z",
    "backoff":  "2h15m",
}
var cfg Config
_ = unface.Unface(raw, &cfg)
```

Or with numeric sources:

```go
raw := map[string]any{
    "deadline": 1735689600, // unix seconds
    "backoff":  0.5,         // 500ms
}
```
