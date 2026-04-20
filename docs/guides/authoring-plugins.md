# Guide: authoring a plugin

**What this page covers.** Writing a plugin for types you don't own (or can't change), registering it, and composing it with the built-ins.

When you control the destination type, prefer implementing `Un*er` methods on the type itself ([guides/custom-types.md](./custom-types.md)). Plugins are the escape hatch for third-party types, wide-kind matches, or when you want to override library defaults for a specific type.

## What a plugin is

A `Plugin` is a name plus one or more `AdapterFactory`s. A factory says "I can build an `Adapter` for any destination matching this predicate." An `Adapter` is just an `Unfacer`: it receives the source value and writes to the destination pointer the factory closed over.

```go
type Plugin interface {
    Name() string
    Factories() []AdapterFactory
    ChildNames() []string // composites only
    Children() []Plugin   // composites only
}

type AdapterFactory interface {
    Matches(destType reflect.Type) bool
    For(destPtr any) Adapter  // Adapter = Unfacer
}
```

The helpers `plugin.NewPlugin`, `plugin.FactoryFunc`, and `plugin.Compose` cover the common cases. All are re-exported as `unface.NewPlugin`, `unface.FactoryFunc`, `unface.Compose`.

## A complete example

Parse `"red" | "green" | "blue"` strings into a custom `Color` enum.

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

// colorAdapter holds a pointer to the destination value.
type colorAdapter struct{ ptr *Color }

// Unface is required (Adapter = Unfacer).
func (a colorAdapter) Unface(src any) error {
    if s, ok := src.(string); ok {
        return a.Unstring(s)
    }
    return unface.ErrNotHandled
}

// Unstring is optional: implementing it lets the dispatcher skip the
// type-switch above for string sources.
func (a colorAdapter) Unstring(s string) error {
    switch s {
    case "red":   *a.ptr = Red
    case "green": *a.ptr = Green
    case "blue":  *a.ptr = Blue
    default:
        return fmt.Errorf("unknown color %q", s)
    }
    return nil
}

var colorType = reflect.TypeOf(Color(0))

var ColorPlugin = unface.NewPlugin("color", unface.FactoryFunc(
    func(t reflect.Type) bool { return t == colorType },
    func(ptr any) unface.Adapter {
        return colorAdapter{ptr: ptr.(*Color)}
    },
))
```

## Registering it

```go
f := unface.New(unface.With(unface.StandardPlugin, mypkg.ColorPlugin))

var c mypkg.Color
_ = f.Unface("green", &c) // c == Green
```

Or for a one-shot, scope it to a single call:

```go
_ = unface.Unface("blue", &c, unface.With(mypkg.ColorPlugin))
```

## Matching patterns

| Predicate | Use for |
|---|---|
| `t == reflect.TypeOf(MyType(...))` | Exact type match — safest, won't claim aliased named types. |
| `t.Kind() == reflect.Pointer` | All pointers (see `PointerPlugin`). Risky for anything narrower. |
| `t.Implements(myInterface)` | Any dest that satisfies an interface. |

Every built-in scalar plugin uses the exact-type predicate. This prevents a plugin from claiming a named type it can't actually downcast (which would panic in `For`).

## Per-call config (CfgAware)

If your adapter needs the active `*plugin.Config` (match mode, tag fallback, plugin set), implement `plugin.CfgAware`:

```go
func (a myAdapter) WithConfig(cfg *plugin.Config) plugin.Unfacer {
    return myAdapter{ptr: a.ptr, cfg: cfg}
}
```

The dispatcher calls `WithConfig` and uses the returned adapter for that call. The built-in recursive plugins (`PointerPlugin`, `ListPlugin`, `MapPlugin`, `StructPlugin`) all use this pattern — look at [`unfacers/list.go`](../../unfacers/list.go) for a reference.

## Composing

Group related plugins under a single name so users can add or drop them as a unit:

```go
var HardwarePlugin = unface.Compose("hardware",
    TemperaturePlugin,
    VoltagePlugin,
    ColorPlugin,
)
```

`With(HardwarePlugin)` pulls in all three; `Without(HardwarePlugin)` removes all three (composition tracks children by name).

## Checklist

- Match by exact type unless you're consciously wide.
- Implement the specific `Un*er` methods when it simplifies your code; the dispatcher will use them.
- Return `ErrNotHandled` when you can't handle a given source — don't hard-fail.
- Deep-copy any incoming slice/map/chan — the dispatcher leaves aliasing safety to plugins.
- Name your plugin once; composites rely on names for `Without`.
