# Reference: StructPlugin

**What this page covers.** The struct walker — the plugin that turns a map source into a typed struct destination using the tag grammar.

**Package:** `github.com/schneid-l/unface/unfacers`
**Dest kind:** any struct type

## Semantics

`StructPlugin` delegates to `internal/walker.StructWalk`, which:

1. Reads the target struct's fields once (cached) and resolves each to a source key using `unface` tags + the fallback list (default `unface`, `yaml`, `json`).
2. For each field present in the source, re-enters the dispatcher with the field's value and pointer, using the active plugin set — or the `Strict` facer (no plugins) for fields tagged `strict`.
3. Checks `required` after the walk; missing keys produce `ErrRequired` wrapped in `*unface.Error` with the field path.
4. Applies the unknown-key policy on the leftover source keys (ignore / warn / error).
5. Absorbs unknown keys into a `remainder`-tagged `map[string]any` field if one exists, bypassing the policy.

The walker is `CfgAware`, so every recursive dispatch uses the caller's match mode, tag fallback, and options.

## Tag grammar

Full reference lives in [tags.md](../../tags.md). Quick recap:

| Modifier | Effect |
|---|---|
| `<name>` | Source key; `-` skips the field. |
| `required` | Missing key → `ErrRequired`. |
| `alias=X` | Extra accepted name (repeat). |
| `remainder` | Catch-all `map[string]any`. |
| `strict` | Dispatch via the zero-plugin Facer. |
| `inline` | Walk fields into parent namespace (same as anonymous embed). |
| `match=...` | Per-field match mode override. |

Struct-wide marker options:

```go
type Cfg struct {
    _    struct{} `unface:",match=exact,unknown=error,tags=unface+json"`
    ...
}
```

## Example

```go
type Server struct {
    Host     string            `unface:"host,required,alias=hostname"`
    Port     int               `unface:"port"`
    Tags     map[string]string `unface:"tags"`
    Extras   map[string]any    `unface:",remainder"`
}

raw := map[string]any{
    "hostname": "api.example.com", // alias
    "port":     "8080",            // string → int
    "tags":     map[string]any{"env": "prod"},
    "note":     "spare",            // absorbed by Extras
}
var s Server
_ = unface.Unface(raw, &s)
```

## Error modes

| Cause | Error |
|---|---|
| `required` field missing | `*unface.Error` wrapping `ErrRequired` |
| Field coercion failure | `*unface.Error` wrapping the inner plugin/user error |
| Unknown key under `UnknownError` policy | `*unface.Error` wrapping `ErrUnknownField` |

All path-traced — `err.Error()` reads like `unface: server.port: ...`.

## See also

- [tags.md](../../tags.md) — complete tag grammar.
- [guides/http-binding.md](../../guides/http-binding.md) — practical example.
- [guides/k8s-resources.md](../../guides/k8s-resources.md) — overriding the walker with `Unmapper`.
