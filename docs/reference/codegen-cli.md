# Reference: `unfacegen` CLI

**What this page covers.** Every flag `unfacegen` accepts today, with concrete invocations. For the conceptual walkthrough see [guides/codegen.md](../guides/codegen.md).

## Install

```bash
go install github.com/schneid-l/unface/cmd/unfacegen@latest
```

## Synopsis

```
unfacegen -type=T1[,T2,...] [-output=<file>] [-dir=<path>]
          [-mode=dispatch|walker] [-strict]
          [-tags=<buildtags>] [-matchmode=exact|fold|insensitive]
```

## Flags

| Flag | Required | Default | Purpose |
|---|---|---|---|
| `-type` | yes | — | Comma-separated list of type names to generate `Unface` methods for. |
| `-output` | no | `<first_type>_unface.go` in the inspected directory (lowercased) | Output file path. |
| `-dir` | no | `.` | Directory of the Go package to inspect. |
| `-mode` | no | `dispatch` | `dispatch` emits a type switch over the type's existing `Un*er` methods. `walker` emits an inlined struct walker driven by `unface:"..."` tags. |
| `-strict` | no | `false` | When set, the generated code calls `unface.Strict.Unface(...)` in per-field fallback paths instead of the default `unface.Unface(...)`. Use this to opt into strict normalization (no plugin fallback) for the target type. |
| `-tags` | no | — | Build constraint tags to emit as `//go:build <tags>` at the top of the generated file. Example: `-tags='!unfacegen_skip'`. |
| `-matchmode` | no | `exact` | Walker-mode key match: `exact`, `fold`, `insensitive`. **Only `exact` is honored today**; `fold` and `insensitive` are accepted for CLI symmetry but have no effect on the generated walker — fall back to the runtime `StructPlugin` if you need case-insensitive keys. |

## Generation modes

### Dispatch mode (default)

Scans the method set of `*T` and emits a type switch that forwards `src` to whichever `Un*er` method matches. Good for hand-authored scalar types that already implement `Unstring` / `Unnumber` / etc. — the generated `Unface` replaces the dispatcher's reflective fast-path.

### Walker mode

Scans the fields and tags of a struct type and emits an inline equivalent of `unfacers.StructPlugin` with zero runtime reflection.

**Supported tag modifiers in v1:**

- Positional `name` (e.g. `unface:"host"`)
- `-` (skip the field)
- `required` (emit a post-pass `%w: <name>` error when missing)
- `alias=X` (additional lookup keys, tried after the primary)

**Deferred — NOT supported by the walker codegen yet:**

- `remainder`
- `inline`
- `strict` (field-level; use `-strict` CLI flag for the file-wide fallback)
- `match=fold|insensitive|exact` (always `exact`; CLI `-matchmode` accepts them for forward-compat but generates `exact`-mode lookups)

If your struct needs any of the deferred features, use the runtime `StructPlugin` — the generator will return an error rather than silently produce wrong code.

The generated file header also documents these limitations inline.

## Examples

Generate dispatch-mode for a single type in the current directory:

```bash
unfacegen -type=URL
# writes url_unface.go
```

Multiple types, custom output:

```bash
unfacegen -type=URL,Endpoint,Route -output=unface_gen.go
```

Walker mode:

```bash
unfacegen -mode=walker -type=Config
# writes config_unface.go with an inline struct walker
```

Gate the generated file behind a build tag (so you can regenerate without the file interfering):

```bash
unfacegen -type=URL -tags='!unfacegen_skip'
```

Strict fallback:

```bash
unfacegen -mode=walker -type=Config -strict
# generated per-field recursion uses unface.Strict.Unface
```

Typical `go generate` directive:

```go
//go:generate unfacegen -mode=walker -type=Config
```

## Dogfood goldens

The repository checks in reference outputs at:

- [`cmd/unfacegen/testdata/basic/basic.golden`](../../cmd/unfacegen/testdata/basic/basic.golden) — dispatch mode
- [`cmd/unfacegen/testdata/walker/walker.golden`](../../cmd/unfacegen/testdata/walker/walker.golden) — walker mode

`TestGenerateBasicGolden` and `TestGenerateWalkerGolden` in `cmd/unfacegen/generator_test.go` diff the generator output against these files on every run. Regenerate with:

```bash
UNFACEGEN_UPDATE=1 go test ./cmd/unfacegen/...
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Wrote the output file. |
| `1` | Load failed, type not found, unsupported tag modifier, or formatting failed. Unformatted source is written to stderr on a format failure to aid debugging. |
| `2` | `-type` missing or empty. Usage printed. |

## Version

```bash
unfacegen       # prints usage with the version banner
```

The current binary reports `v0.1.0`.
