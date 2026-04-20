# Plugins

Plugins supply default coercions. They are **opt-in**: a fresh `unface.New()` has zero plugins (equivalent to `unface.Strict`); the package-level `Default` and the top-level `unface.Unface(...)` call load the `StandardPlugin` bundle so the 80 % use case works with no setup.

This page covers: the built-in plugin catalogue, composition semantics, per-call overrides, and how to write your own plugin.

## Built-in plugins

### Scalars

| Plugin | Dest type | Accepted sources |
|---|---|---|
| `StringPlugin` | `*string` | `string`, `[]byte`, `bool`, any number, `nil` |
| `BoolPlugin` | `*bool` | `bool`, case-insensitive `"true"/"false"/"yes"/"no"/"y"/"n"/"on"/"off"/"1"/"0"/"enabled"/"disabled"`, any number (non-zero → true) |
| `BytesPlugin` | `*[]byte` | `string`, `[]byte` (always copied), `nil` |
| `RunePlugin` | `*rune` | `rune`, single-codepoint `string`, any number |

### Integers (atomic)

| Plugin | Dest type |
|---|---|
| `Int8Plugin` … `Int64Plugin`, `IntPlugin` | `*int8` … `*int64`, `*int` |
| `Uint8Plugin` … `Uint64Plugin`, `UintPlugin` | `*uint8` … `*uint64`, `*uint` |

Each accepts: any number (with overflow check), string (via `strconv.Parse*` — also accepts `0x`, `0b`, `0o` prefixes), `bool` (true → 1, false → 0), `nil` (→ zero).

### Floats & complex

| Plugin | Dest type |
|---|---|
| `Float32Plugin`, `Float64Plugin` | `*float32`, `*float64` |
| `Complex64Plugin`, `Complex128Plugin` | `*complex64`, `*complex128` |

Accept: numbers, strings (stdlib-parsed), `nil`. Complex plugins also accept strings like `"1+2i"`.

### Big math

| Plugin | Dest type |
|---|---|
| `BigIntPlugin` | `*big.Int` |
| `BigFloatPlugin` | `*big.Float` |

Accept numbers and strings (via `big.{Int,Float}.SetString`).

### Special

| Plugin | Dest type | Accepted sources |
|---|---|---|
| `TimePlugin` | `*time.Time` | RFC3339 and RFC3339Nano, `2006-01-02T15:04:05`, `2006-01-02 15:04:05`, `2006-01-02`, RFC1123 strings; `time.Time`; any number (interpreted as Unix seconds) |
| `TimePlugin` *(same)* | `*time.Duration` | `time.ParseDuration` strings; `time.Duration`; integer (seconds); float (fractional seconds) |
| `JSONRawPlugin` | `*json.RawMessage` | `json.RawMessage`, `[]byte`, or any JSON-marshalable value (re-encoded) |
| `PointerPlugin` | `*T` (any pointer) | Auto-allocates nil dests and forwards the coercion to the element type. Mostly redundant under Flat pointer resolution. |

### Structural

| Plugin | Dest type | Source | Notes |
|---|---|---|---|
| `ListPlugin` | any non-`[]byte` slice | array / slice | Recursively unfaces each element via the active plugin set. Scalar-to-singleton promotion: a non-list source is wrapped in a single-element slice. |
| `MapPlugin` | any map | map | Recursively unfaces each key and value. Keys are coerced into the destination key type. |
| `StructPlugin` | any struct | map | Full struct walker: tags, marker options, matching, aliases, required, remainder, inline, strict — see [tags.md](./tags.md). |

### Composite bundles

Bundles name a group so you can drop it wholesale via `Without(bundle)`. They flatten into the atomic children on insertion, and `Without(bundle)` uses the bundle's recorded `ChildNames()` to filter.

| Bundle | Contents |
|---|---|
| `IntPluginBundle` | all signed integer plugins |
| `UintPluginBundle` | all unsigned integer plugins |
| `FloatPluginBundle` | `Float32Plugin`, `Float64Plugin` |
| `ComplexPluginBundle` | `Complex64Plugin`, `Complex128Plugin` |
| `NumberPlugin` | every integer, float, complex, BigInt, BigFloat |
| `PrimitivesPlugin` | `NumberPlugin` + BoolPlugin + StringPlugin + BytesPlugin + RunePlugin |
| `StandardPlugin` | `PrimitivesPlugin` + TimePlugin + ListPlugin + MapPlugin + StructPlugin + PointerPlugin + JSONRawPlugin |

## Composition

```go
// Instance scope:
f := unface.New(unface.With(unface.NumberPlugin, unface.StringPlugin))

// Add more later (returns a new Option, not a mutation):
f2 := unface.New(
    unface.With(unface.StandardPlugin),
    unface.With(MyCustomPlugin),
)

// Drop a composite (propagates to all its children):
f3 := unface.New(
    unface.With(unface.StandardPlugin),
    unface.Without(unface.TimePlugin),
)

// Drop by name (flattened atomic leaves — matches leaf names, not composite names):
f4 := unface.New(
    unface.With(unface.StandardPlugin),
    unface.WithoutNamed("int", "int64"),
)

// Replace the set entirely:
f5 := unface.New(unface.Only(MyPlugin))
```

## Per-call overrides

Every `Unface` call accepts the same `Option`s:

```go
// Add a plugin just for this call:
f.Unface(src, &dst, unface.With(MyTemporaryPlugin))

// Strip plugins for a strict one-off:
f.Unface(src, &dst, unface.Only())

// Change match mode for a single call:
f.Unface(src, &dst, unface.WithFieldMatch(unface.MatchExact))
```

Per-call options clone the instance's configuration and apply on top of the clone — the instance itself is never mutated. This is what makes `Facer` safe to share across goroutines.

## Authoring your own plugin

Every plugin is a name plus one or more `AdapterFactory`s. A factory says "I can build an `Adapter` for any destination whose type matches this predicate". An `Adapter` is just `Unfacer` — it receives the source value and writes through to the destination pointer the factory closed over.

### Minimal example

Parse `"red"` / `"green"` / `"blue"` into an enum type:

```go
package mypkg

import (
    "fmt"
    "reflect"

    "github.com/schneid-l/unface"
)

type Color int
const (
    Red Color = iota
    Green
    Blue
)

type colorAdapter struct{ ptr *Color }

func (a colorAdapter) Unface(src any) error {
    if s, ok := src.(string); ok {
        return a.Unstring(s)
    }
    return unface.ErrNotHandled
}

func (a colorAdapter) Unstring(s string) error {
    switch s {
    case "red":
        *a.ptr = Red
    case "green":
        *a.ptr = Green
    case "blue":
        *a.ptr = Blue
    default:
        return fmt.Errorf("unknown color %q", s)
    }
    return nil
}

var ColorPlugin = unface.NewPlugin("color", unface.FactoryFunc(
    func(t reflect.Type) bool { return t == reflect.TypeOf(Color(0)) },
    func(ptr any) unface.Adapter { return colorAdapter{ptr: ptr.(*Color)} },
))
```

Usage:

```go
f := unface.New(unface.With(unface.StandardPlugin, mypkg.ColorPlugin))
var c mypkg.Color
f.Unface("green", &c) // c == mypkg.Green
```

### Design tips

- **Match exact types.** Every built-in atomic scalar plugin uses `t == reflect.TypeOf(zero)` rather than `t.Kind() == …`. This prevents factories from claiming named types their adapter can't actually downcast (which otherwise panics at `For(ptr)` time).
- **Implement specific interfaces too.** If your adapter implements `Unstringer`, the dispatcher skips your `Unface` master switch entirely when the src is a string — saves a type assertion and makes the intent explicit.
- **Return `ErrNotHandled` generously.** Let the pipeline try other plugins. Hard-failing on every unexpected source prevents users from layering your plugin with others.
- **Deep-copy reference values.** If you accept a slice / map / chan, copy it into your destination. `assignDirect` deliberately doesn't handle those kinds so plugins own the safety guarantee.
- **Honor the `cfgAware` contract** if your adapter needs per-call config (match mode, tag fallback, etc.). Implement `withConfig(*config) Unfacer`. The dispatcher rebinds adapters that satisfy this interface to the effective per-call config.

### Composing plugins

Bundle several related plugins under a single name:

```go
var HardwarePlugin = unface.Compose("hardware",
    TemperaturePlugin,
    VoltagePlugin,
    ColorPlugin,
)
```

Users who `With(HardwarePlugin)` get all three. Users who `Without(HardwarePlugin)` drop all three (composition tracks children by name).

## Strict vs Default

`unface.Strict` is a `*Facer` with zero plugins. Useful for:

- Testing user `Un*er` implementations in isolation (no coercion magic).
- Fields tagged `strict` — the struct walker uses `Strict` for those.

`unface.Default` is the implicitly-used Facer behind `unface.Unface(src, dest)`. It is preloaded with `StandardPlugin`. Replace it if you want global non-standard defaults, but prefer passing a custom `*Facer` explicitly.

## See also

Per-plugin reference pages with accepted-source tables and error modes:

- Scalars: [reference/plugins/scalar.md](./reference/plugins/scalar.md)
- Integers: [reference/plugins/integer.md](./reference/plugins/integer.md)
- Floats & complex: [reference/plugins/float-complex.md](./reference/plugins/float-complex.md)
- Big math: [reference/plugins/bigint.md](./reference/plugins/bigint.md)
- Time & duration: [reference/plugins/time.md](./reference/plugins/time.md)
- Pointer: [reference/plugins/pointer.md](./reference/plugins/pointer.md)
- `json.RawMessage`: [reference/plugins/jsonraw.md](./reference/plugins/jsonraw.md)
- Lists: [reference/plugins/list.md](./reference/plugins/list.md)
- Maps: [reference/plugins/map.md](./reference/plugins/map.md)
- Structs: [reference/plugins/struct.md](./reference/plugins/struct.md)
- Composite bundles: [reference/plugins/composites.md](./reference/plugins/composites.md)

Authoring walkthrough: [guides/authoring-plugins.md](./guides/authoring-plugins.md).
