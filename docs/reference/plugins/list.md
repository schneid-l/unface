# Reference: ListPlugin

**What this page covers.** Recursive list coercion with scalar-to-singleton promotion.

**Package:** `github.com/schneid-l/unface/unfacers`
**Dest kind:** any slice type *except* `[]byte` (which is owned by `BytesPlugin`)

## Semantics

1. If the source is `nil`, the destination is set to a zero slice.
2. If the source implements `List` (any slice/array — see `plugin.ListOf`), the adapter allocates `make([]T, n, n)` and re-dispatches each element through the active plugin set.
3. If the source is **not** a list, the adapter performs **scalar-to-singleton promotion**: wraps the source in a one-element slice and re-dispatches.

Element dispatch uses the same `*plugin.Config` the outer call had (via `CfgAware`), so match modes, tag fallback, and custom plugins all propagate.

## Scalar promotion

```go
type Config struct {
    Hosts []string `unface:"hosts"`
}

raw := map[string]any{"hosts": "single.example"}
var cfg Config
_ = unface.Unface(raw, &cfg)
// cfg.Hosts == []string{"single.example"}
```

Useful when YAML / JSON authors sometimes write a single scalar and sometimes a list.

## List → list

```go
raw := map[string]any{"hosts": []any{"a", "b", 42}}
// cfg.Hosts == []string{"a", "b", "42"}  — each element coerced through StringPlugin
```

## Nested

Lists of structs, lists of lists, lists of pointers — all re-enter the dispatcher:

```go
type Pair struct { X, Y int }
var pairs []Pair
src := []any{
    map[string]any{"x": 1, "y": 2},
    map[string]any{"x": 3, "y": 4},
}
_ = unface.Unface(src, &pairs)
```

## Error modes

- Per-element error: wrapped as `unface: list[i]: <inner>` (hard error). Propagates up.
- `[]byte` dest: this plugin's factory does **not** match; `BytesPlugin` handles it instead.

## Why `[]byte` is excluded

`[]byte` is semantically a buffer, not a list of octets to coerce one-by-one. `BytesPlugin` copies the buffer verbatim.
