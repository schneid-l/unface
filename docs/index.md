# unface Documentation

Welcome. unface is a composable coercion library for Go: transform any `interface{}` into any typed value, with destination types declaring how they want to be populated.

## Pick your path

| You are... | Start here |
|---|---|
| A first-time user who wants a working example in five minutes | [quickstart.md](./quickstart.md) |
| Building something custom — plugin, enum, polymorphic dispatch | [guides/authoring-plugins.md](./guides/authoring-plugins.md) and [guides/custom-types.md](./guides/custom-types.md) |
| Looking up a specific option, sentinel, or plugin | The [reference](#reference) section below |

## Concepts

| Topic | Page |
|---|---|
| Package layout and runtime pipeline | [concepts/architecture.md](./concepts/architecture.md) |
| Hard vs. soft errors, `*Error` path tracing | [concepts/errors.md](./concepts/errors.md) |
| Step-by-step dispatch order | [dispatch-order.md](./dispatch-order.md) |
| Full tag grammar, struct marker, fallbacks | [tags.md](./tags.md) |
| Built-in plugin catalogue, composition, per-call overrides | [plugins.md](./plugins.md) |

## Guides

| Scenario | Page |
|---|---|
| JSON request body → typed struct | [guides/http-binding.md](./guides/http-binding.md) |
| Polymorphic Kind/Spec via `Unmapper` | [guides/k8s-resources.md](./guides/k8s-resources.md) |
| Teach your own types via `Un*er` | [guides/custom-types.md](./guides/custom-types.md) |
| Write and register a plugin | [guides/authoring-plugins.md](./guides/authoring-plugins.md) |
| Generate `Unface` methods with `unfacegen` | [guides/codegen.md](./guides/codegen.md) |

## Reference

| Topic | Page |
|---|---|
| Every `Option` with signature and example | [reference/options.md](./reference/options.md) |
| Every sentinel + the `Error` type | [reference/errors.md](./reference/errors.md) |
| `unfacegen` CLI flags | [reference/codegen-cli.md](./reference/codegen-cli.md) |
| `StringPlugin`, `BoolPlugin`, `BytesPlugin`, `RunePlugin` | [reference/plugins/scalar.md](./reference/plugins/scalar.md) |
| `Int*Plugin`, `Uint*Plugin` | [reference/plugins/integer.md](./reference/plugins/integer.md) |
| `Float*Plugin`, `Complex*Plugin` | [reference/plugins/float-complex.md](./reference/plugins/float-complex.md) |
| `BigIntPlugin`, `BigFloatPlugin` | [reference/plugins/bigint.md](./reference/plugins/bigint.md) |
| `TimePlugin` (time.Time + time.Duration) | [reference/plugins/time.md](./reference/plugins/time.md) |
| `PointerPlugin` | [reference/plugins/pointer.md](./reference/plugins/pointer.md) |
| `JSONRawPlugin` | [reference/plugins/jsonraw.md](./reference/plugins/jsonraw.md) |
| `ListPlugin` (+ scalar promotion) | [reference/plugins/list.md](./reference/plugins/list.md) |
| `MapPlugin` (+ key coercion) | [reference/plugins/map.md](./reference/plugins/map.md) |
| `StructPlugin` | [reference/plugins/struct.md](./reference/plugins/struct.md) |
| `IntPluginBundle` / `NumberPlugin` / `PrimitivesPlugin` / `StandardPlugin` | [reference/plugins/composites.md](./reference/plugins/composites.md) |

## Also

- [faq.md](./faq.md) — common questions.
- Runnable examples in [`../examples/`](../examples/): [`url`](../examples/url), [`httpbind`](../examples/httpbind), [`k8sresource`](../examples/k8sresource).
- Godoc: <https://pkg.go.dev/github.com/schneid-l/unface>
- Changelog: [`../CHANGELOG.md`](../CHANGELOG.md)
- Contributing: [`../CONTRIBUTING.md`](../CONTRIBUTING.md)
