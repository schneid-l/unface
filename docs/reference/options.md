# Reference: options

**What this page covers.** Every `Option` exported by `unface` (re-exported from the `engine` package), its signature, default, and a short example. Options can be passed to `unface.New(...)` (instance-wide) or to `Facer.Unface(src, dest, opts...)` (per call). Per-call options clone and apply on top of the instance's config without mutating it.

Godoc: <https://pkg.go.dev/github.com/schneid-l/unface#Option>

## Plugin set

### `With(plugins ...Plugin) Option`

Append plugins. Composites flatten to their atomic leaves on insertion.

```go
f := unface.New(unface.With(unface.StandardPlugin, mypkg.ColorPlugin))
```

### `Without(plugins ...Plugin) Option`

Remove by identity. Composites are expanded — removing a composite removes every named child.

```go
f := unface.New(
    unface.With(unface.StandardPlugin),
    unface.Without(unface.TimePlugin),
)
```

### `WithoutNamed(names ...string) Option`

Remove by name. Matches atomic leaves (e.g. `"int64"`) and composite names (e.g. `"primitives"`).

```go
f := unface.New(
    unface.With(unface.StandardPlugin),
    unface.WithoutNamed("int", "int64"),
)
```

### `Only(plugins ...Plugin) Option`

Replace the current set wholesale. Empty = strict mode for this call.

```go
f.Unface(src, &dst, unface.Only()) // no plugins for this call
```

## Struct walker

### `WithFieldMatch(m MatchMode) Option`

Default: `MatchFold`. See [tags.md](../tags.md#match-modes) for the full table.

```go
f := unface.New(
    unface.With(unface.StandardPlugin),
    unface.WithFieldMatch(unface.MatchExact),
)
```

### `OnUnknown(p UnknownPolicy, handler ...UnknownHandler) Option`

Default policy: `UnknownIgnore`. For `UnknownWarn`, supply a handler.

```go
f := unface.New(
    unface.With(unface.StandardPlugin),
    unface.OnUnknown(unface.UnknownWarn, func(field string, value any) {
        log.Printf("unknown key %q = %v", field, value)
    }),
)
```

### `WithTagFallback(tags ...string) Option`

Default: `["unface", "yaml", "json"]`. Fallback names are tried in order; only the positional name is consulted.

```go
f := unface.New(unface.WithTagFallback("unface", "toml"))
```

### `WithoutTagFallback() Option`

Disables fallback. Only the `unface` tag is read.

```go
f := unface.New(unface.WithoutTagFallback())
```

## Error observability

### `OnSoftError(h SoftErrorHandler) Option`

Observe every soft error the dispatcher sees. Read-only; cannot alter behavior.

```go
f := unface.New(
    unface.With(unface.StandardPlugin),
    unface.OnSoftError(func(src, dest any, err error) {
        log.Printf("soft: %T → %T: %v", src, dest, err)
    }),
)
```

## Pointer resolution

### `WithPointerResolve(mode PointerResolution) Option`

Default: `PointerResolveFlat`. See [dispatch-order.md](../dispatch-order.md#step-2--pointer-resolution) for the per-mode semantics.

```go
f := unface.New(
    unface.With(unface.StandardPlugin),
    unface.WithPointerResolve(unface.PointerResolveDeep),
)
```

## Strict mode

`unface.Strict` is a package-level `*Facer` constructed with no plugins. It's not an option, but worth noting here: the struct walker uses it to enforce a `strict` field tag, and it's what you reach for when testing `Un*er` implementations in isolation.

```go
var v MyType
_ = unface.Strict.Unface(src, &v) // only MyType's own Un*er methods run
```
