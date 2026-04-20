# FAQ

**What this page covers.** The questions that come up most often when people encounter unface for the first time.

## Why not just use `encoding/json`?

`encoding/json` is excellent when the wire format matches your Go types exactly. It struggles when:

- A field should accept either a scalar shorthand (`"https://..."`) or a structured object (`{scheme, host, ...}`).
- Numeric types arrive as strings (`"8080"` into `int`).
- You want `required`, `alias`, `remainder`, or a `strict` mode per field.
- The concrete shape of a field depends on a sibling field (polymorphic Kind/Spec).

unface composes on top of any decoder: JSON → `map[string]any` → `unface.Unface(raw, &typed)`. Use both.

## How is unface different from mapstructure?

Similar surface (map → struct), different philosophy:

- **Interface-first.** Coercion is opt-in via `Un*er` methods on your types. No reflection-heavy "DecodeHook" registry.
- **Two-tier errors.** A soft `ErrNotHandled` lets methods decline without aborting, enabling layered defaults.
- **Pointer resolution modes.** Explicit `Flat` / `None` / `Deep` control over multi-level pointer dispatch.
- **Codegen path.** Types that implement `Unface(src any) error` skip reflection; `unfacegen` emits this for you.
- **Rich source views.** `Number`, `List`, `Map` give you typed accessors instead of raw `reflect`.

Mapstructure is battle-tested and widely deployed. Reach for unface when you want the scalar-or-object flexibility, compile-time fast path, or `Un*er`-style method dispatch.

## How do I add a plugin for my custom type?

Two options:

1. **Implement `Un*er` on the type itself.** See [guides/custom-types.md](./guides/custom-types.md). Zero registration.
2. **Write a plugin.** See [guides/authoring-plugins.md](./guides/authoring-plugins.md). Needed when you don't own the type.

## What's the performance cost of reflection?

Roughly comparable to `encoding/json`'s reflective path. For hot paths:

- Implement `Unface(src any) error` directly (or generate it with `unfacegen`) — the dispatcher skips all reflection for that type.
- Reuse a shared `*Facer`; construction interns the plugin set.
- Prefer `PointerResolveFlat` (the default); `Deep` does more work.

The bench suite lives in [`bench_test.go`](../bench_test.go); run `go test -bench=.` for your workload.

## How do I turn off all default plugins?

Use `unface.Strict` (a preconfigured zero-plugin Facer) or construct one:

```go
f := unface.New() // no With(...) = no plugins
```

Only `Un*er` methods on the destination will run. A `strict` field tag has the same effect for one field.

## Does unface support recursive types?

Yes — the struct walker descends into nested structs, slices, maps, and pointers through the same dispatcher. There's no recursion limit beyond the Go stack. Adapters that need per-call config (pointer/list/map/struct) participate via `plugin.CfgAware` so the effective options propagate the whole way down.

## Can I use unface with YAML / TOML?

Yes. Decode to `any` with your YAML / TOML library, then pass the result to `unface.Unface(raw, &typed)`. The default tag fallback is `unface, yaml, json`; for TOML use `unface.WithTagFallback("unface", "toml")` to read `toml:"..."` names.

## How do I debug why a field wasn't populated?

Three tools:

1. **Soft-error observer**: `unface.OnSoftError(func(src, dest any, err error) { log.Printf(...) })` sees every dispatcher decline.
2. **Path-traced hard errors**: check for `*unface.Error` via `errors.As`; its `Path` field tells you the exact field.
3. **Strict probe**: run the value through `unface.Strict` — if it still succeeds, your type's `Un*er` is handling it; if not, a plugin is.

For unknown keys specifically: `unface.OnUnknown(unface.UnknownWarn, handler)` or `UnknownError` surfaces them.
