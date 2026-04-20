# Reference: PointerPlugin

**What this page covers.** How `PointerPlugin` auto-allocates nil pointer destinations and forwards to the element type.

**Package:** `github.com/schneid-l/unface/unfacers`
**Dest kind:** any `*T` where `T.Kind() == reflect.Pointer` (i.e. matches `**U`, `***U`, ...)

## Semantics

```go
var dst *int
_ = unface.Unface(42, &dst) // *dst == 42, dst allocated
```

1. If the source is `nil`, the adapter sets the pointer to its zero value (nil).
2. Otherwise it allocates a new element via `reflect.New(elemType)`.
3. It re-enters the dispatcher with the element as the new dest, using the Facer's active config (via `CfgAware.WithConfig`).
4. On success it writes the element back through the pointer.

## Interaction with pointer resolution

`PointerPlugin` is mostly redundant under `PointerResolveFlat` (the default), because the engine flattens multi-level pointers before dispatch and allocates intermediates itself. It becomes load-bearing under `PointerResolveNone`, where the engine refuses to deref and a pointer dest must be handled by a plugin.

See [dispatch-order.md](../../dispatch-order.md#step-2--pointer-resolution) for the full picture.

## Example

```go
type Server struct {
    Proxy *string `unface:"proxy"` // auto-allocated when source key present
}

raw := map[string]any{"proxy": "http://example"}
var s Server
_ = unface.Unface(raw, &s)
// s.Proxy != nil, *s.Proxy == "http://example"
```

## Error modes

- Source is nil: pointer set to nil, no error.
- Source can't be coerced to the element type: whatever the element plugin returns (hard error or soft).
