# Reference: composite plugin bundles

**What this page covers.** The named groups that make it easy to add or remove a whole family of plugins with a single option.

**Package:** `github.com/schneid-l/unface/unfacers`

Composites are built with `plugin.Compose(name, children...)`. On insertion, `With(composite)` flattens the tree to atomic leaves; on removal, `Without(composite)` recursively drops every child by name. This means you can add `StandardPlugin` and then `Without(TimePlugin)` to disable only the time factories without re-listing every other plugin.

## Bundles

| Bundle | Name | Contents |
|---|---|---|
| `IntPluginBundle` | `int.bundle` | `Int8Plugin`, `Int16Plugin`, `Int32Plugin`, `Int64Plugin`, `IntPlugin` |
| `UintPluginBundle` | `uint.bundle` | `Uint8Plugin`, `Uint16Plugin`, `Uint32Plugin`, `Uint64Plugin`, `UintPlugin` |
| `FloatPluginBundle` | `float.bundle` | `Float32Plugin`, `Float64Plugin` |
| `ComplexPluginBundle` | `complex.bundle` | `Complex64Plugin`, `Complex128Plugin` |
| `NumberPlugin` | `number` | `IntPluginBundle` + `UintPluginBundle` + `FloatPluginBundle` + `ComplexPluginBundle` + `BigIntPlugin` + `BigFloatPlugin` |
| `PrimitivesPlugin` | `primitives` | `NumberPlugin` + `BoolPlugin` + `StringPlugin` + `BytesPlugin` + `RunePlugin` |
| `StandardPlugin` | `standard` | `PrimitivesPlugin` + `TimePlugin` + `ListPlugin` + `MapPlugin` + `StructPlugin` + `PointerPlugin` + `JSONRawPlugin` |

`unface.Default` is a `*Facer` preloaded with `StandardPlugin`; the top-level `unface.Unface(...)` uses it.

## Use cases

### Start from everything, drop a family

```go
f := unface.New(
    unface.With(unface.StandardPlugin),
    unface.Without(unface.ComplexPluginBundle), // no complex coercion
)
```

### Build a minimal Facer

```go
f := unface.New(
    unface.With(unface.IntPluginBundle, unface.StringPlugin),
)
```

### Remove a specific atomic leaf

`Without(StandardPlugin)` removes the entire bundle; use `WithoutNamed` to pick off a leaf by name:

```go
f := unface.New(
    unface.With(unface.StandardPlugin),
    unface.WithoutNamed("int64"), // drop Int64Plugin only
)
```

### Replace wholesale

```go
f := unface.New(unface.Only(unface.StringPlugin, mypkg.ColorPlugin))
```

## Implementation note

Every `Compose` call records its children; `Without` walks `ChildNames()` recursively to build the removal set. Atomic plugins return `nil` for both `Children()` and `ChildNames()`, so `With(atomic)` is a no-op flatten.

See [`unfacers/composite.go`](../../../unfacers/composite.go) for the actual composition tree.
