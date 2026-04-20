# Architecture

**What this page covers.** The four packages that make up unface, what lives where, and why the split exists.

## The four layers

```
+-----------------------------------------------------------+
|  github.com/schneid-l/unface          (root convenience)  |
|  - Unface(src, dest, opts...) top-level function          |
|  - Type aliases (Facer, Option, Unfacer, MatchMode, ...)  |
|  - Built-in plugin variables (StringPlugin, ...)          |
+-----------------------------------------------------------+
                         |                 |
                         v                 v
+-----------------------+  +------------------------------+
|  .../engine           |  |  .../unfacers                |
|  - Facer (dispatcher) |  |  - Every built-in plugin     |
|  - Option builders    |  |  - StandardPlugin bundle     |
|                       |  |  - Default *Facer            |
+-----------------------+  +------------------------------+
           |                              |
           +-------------+----------------+
                         v
+-----------------------------------------------------------+
|  .../plugin            (contract)                         |
|  - Un*er interfaces (Unfacer, Unstringer, ...)            |
|  - Plugin + AdapterFactory + Adapter                      |
|  - Config + CfgAware                                      |
|  - Errors (ErrNotHandled, ErrInvalidDest, ...)            |
|  - Number / List / Map rich abstractions                  |
+-----------------------------------------------------------+
```

## Why the split

| Package | Role | Depends on |
|---|---|---|
| `plugin` | Contract-only. Defines the interfaces and errors everyone agrees on. | std only |
| `engine` | Runtime. Holds the `Facer` type and the dispatch loop. | `plugin` |
| `unfacers` | Out-of-the-box plugin catalogue plus the preconfigured `Default` Facer. | `engine`, `plugin` |
| `unface` (root) | Convenience surface. Re-exports the common names so simple code doesn't need three imports. | all three |

User code typically imports only the root package. Plugin authors import `plugin` for the contract; rarely `engine`.

## Dispatch pipeline at a glance

A single `Unface` call walks this seven-step sequence — [dispatch-order.md](../dispatch-order.md) has the full reference.

```
dest validation -> pointer resolve -> Unfacer on dest -> assignDirect
                -> Un<srcKind>er on dest -> plugin fallback -> ErrNoCoercion
```

## Source abstractions

The `plugin` package provides three rich views that let adapters avoid raw `reflect`:

- `Number` — typed numeric accessors with overflow-safe conversion (`Int64() (int64, bool)`, `BigInt()`, ...). See `plugin.NumberOf`.
- `List` — `Len`, `At`, `Iter`, `Slice`, `Map`, `Filter`. See `plugin.ListOf`.
- `Map` — `Get`, `GetString`, `GetInt64`, `GetMap`, `GetList`, `Iter`. See `plugin.MapOf`.

These are what `Unnumberer.Unnumber(n Number)`, `Unlister.Unlist(l List)`, `Unmapper.Unmap(m Map)` receive.

## Config propagation

Each `Unface` call may supply `Option`s. Options mutate a cloned `*plugin.Config`; the original Facer is never touched, which is what makes `Facer` safe to share across goroutines. Recursive plugins (pointer, list, map, struct) opt into the live config by implementing `plugin.CfgAware`:

```go
type CfgAware interface {
    WithConfig(*Config) Unfacer
}
```

The dispatcher rebinds any adapter that satisfies this interface before calling `Unface` on it.

## Codegen fast path

When a destination type implements `Unface(src any) error` directly, the dispatcher skips reflection entirely. The `cmd/unfacegen` tool emits such methods from a type's existing `Un*er` implementations — see [guides/codegen.md](../guides/codegen.md).

## Further reading

- Godoc: <https://pkg.go.dev/github.com/schneid-l/unface>
- [dispatch-order.md](../dispatch-order.md) — the dispatch sequence in detail.
- [errors.md](./errors.md) — soft vs. hard error protocol.
