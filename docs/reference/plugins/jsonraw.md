# Reference: JSONRawPlugin

**What this page covers.** The plugin that fills `*json.RawMessage` destinations, round-tripping arbitrary source values through `encoding/json`.

**Package:** `github.com/schneid-l/unface/unfacers`
**Dest type:** `*json.RawMessage`

## Semantics

| Source | Result |
|---|---|
| `nil` | `json.RawMessage("null")` |
| `json.RawMessage` | defensive copy |
| `[]byte` | defensive copy |
| anything else | `json.Marshal(src)` — whatever `encoding/json` produces |

## Example

```go
type Event struct {
    Kind    string          `unface:"kind"`
    Payload json.RawMessage `unface:"payload"`
}

raw := map[string]any{
    "kind":    "click",
    "payload": map[string]any{"x": 10, "y": 20},
}

var e Event
_ = unface.Unface(raw, &e)
// e.Payload == []byte(`{"x":10,"y":20}`)
```

## When to reach for this

- You want a lazy-parsed body — defer decoding until you know the `kind`.
- You need to forward a nested object verbatim to another JSON consumer without flattening.

Prefer an `Unmapper` implementation (see [k8s-resources.md](../../guides/k8s-resources.md)) when you have a bounded set of polymorphic shapes you'd rather decode strongly-typed.

## Error modes

- `json.Marshal` failure on an unsupported source (channel, function, recursive type) → hard error from the marshaler.
- No `ErrNotHandled` path: the plugin always tries to marshal anything that isn't `nil` / `RawMessage` / `[]byte`.
