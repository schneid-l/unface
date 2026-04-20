# Reference: MapPlugin

**What this page covers.** Recursive map-to-map coercion with key coercion.

**Package:** `github.com/schneid-l/unface/unfacers`
**Dest kind:** any `map[K]V`

## Semantics

1. If the source is `nil`, the destination is set to a zero map.
2. If the source is a map (`plugin.MapOf`), the adapter allocates `make(map[K]V, n)` and re-dispatches every key and value through the active plugin set.
3. If the source isn't a map, the adapter hard-fails with `unface: cannot coerce %T to <dest map type>`. (Unlike `ListPlugin`, there is no scalar-to-map promotion.)

Both keys and values re-enter the full dispatcher, so any plugin or `Un*er` implementation can handle them.

## Key coercion

If the source map's key type is `any` (typical of JSON-decoded data, `map[string]any`, etc.), the adapter coerces each key into the destination key type `K` via `Facer.Dispatch`. This means `map[string]any` with string keys can bind to `map[int]V` if the strings are numeric:

```go
src := map[string]any{"1": "a", "2": "b"}
var dst map[int]string
_ = unface.Unface(src, &dst)
// dst == map[int]string{1: "a", 2: "b"}
```

Failing key coercion yields `unface: map key %v: <inner>`.

## Value coercion

Same mechanism: per-entry `Facer.Dispatch`. Value errors surface as `unface: map value for %v: <inner>`.

## Struct vs. Map

- `StructPlugin` handles **struct destinations** from map sources using the tag grammar.
- `MapPlugin` handles **map destinations** from map sources — recursive, type-coerced.

The factory predicates are mutually exclusive (`t.Kind() == reflect.Map` vs. `reflect.Struct`), so registering both is safe.

## Example

```go
type Metrics map[string]float64

raw := map[string]any{
    "p50": "0.12",
    "p95": 0.37,
    "p99": 1.0,
}
var m Metrics
_ = unface.Unface(raw, &m)
// m == Metrics{"p50":0.12,"p95":0.37,"p99":1.0}
```
