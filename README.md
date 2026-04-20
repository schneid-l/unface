# unface

[![Go Reference](https://pkg.go.dev/badge/github.com/schneid-l/unface.svg)](https://pkg.go.dev/github.com/schneid-l/unface)
[![CI](https://github.com/schneid-l/unface/actions/workflows/ci.yml/badge.svg)](https://github.com/schneid-l/unface/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/schneid-l/unface)](https://goreportcard.com/report/github.com/schneid-l/unface)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **Transform any `interface{}` into any Go type.**

`unface` is a composable, plugin-driven coercion library. It solves the common "I parsed a YAML/JSON blob and now I need a typed struct" problem — with the flexibility for a single field to accept either a rich object or a shorthand scalar, and sane defaults for every native Go type.

## Quickstart — one field, two shapes

The `URL` field below accepts **either** a shorthand string (`"https://example.com/api"`) **or** a structured object (`{scheme: https, host: ..., path: ...}`) — from the same config file. `URL` opts into the string form by implementing `Unstringer`; the object form is handled automatically by the default struct walker.

```go
package main

import (
    "fmt"
    "strings"

    "github.com/schneid-l/unface"
)

// URL is a config value that accepts either a string shorthand or a
// structured map. Implementing Unstringer teaches the library how to
// parse the shorthand; the map form falls through to the default struct
// walker (no code needed).
type URL struct {
    Scheme string `unface:"scheme"`
    Host   string `unface:"host,required"`
    Port   int    `unface:"port"`
    Path   string `unface:"path"`
}

// Unstring parses "https://example.com:8080/api" into the struct fields.
func (u *URL) Unstring(s string) error {
    rest := s
    if i := strings.Index(rest, "://"); i > 0 {
        u.Scheme, rest = rest[:i], rest[i+3:]
    }
    if i := strings.Index(rest, "/"); i >= 0 {
        u.Path = rest[i:]
        rest = rest[:i]
    }
    if i := strings.LastIndex(rest, ":"); i > 0 {
        fmt.Sscanf(rest[i+1:], "%d", &u.Port)
        rest = rest[:i]
    }
    u.Host = rest
    return nil
}

type Config struct {
    Name string `unface:"name"`
    API  URL    `unface:"api"`
    Web  URL    `unface:"web"`
}

func main() {
    raw := map[string]any{
        "name": "demo",
        // Shorthand: string → URL via Unstring
        "api": "https://api.example.com:8080/v1",
        // Structured: map → URL via default struct walker
        "web": map[string]any{
            "scheme": "https",
            "host":   "example.com",
            "path":   "/",
        },
    }

    var cfg Config
    if err := unface.Unface(raw, &cfg); err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", cfg)
    // {Name:demo
    //  API:{Scheme:https Host:api.example.com Port:8080 Path:/v1}
    //  Web:{Scheme:https Host:example.com Port:0 Path:/}}
}
```

**What happened.** For `api`, the library sees a `string` source and a `*URL` dest; it dispatches to `URL.Unstring`. For `web`, the source is `map[string]any`, so it falls through to the default struct walker which recurses into each field. Same field type, zero branching at the call site — just one `Unface` call.

That's the core idea: destination types declare which source shapes they accept, and the pipeline composes the rest.

## Features

- **Interface-first design.** Destination types opt into coercion by implementing `Un*er` interfaces. No magic.
- **Composable plugins.** Mix and match atomic type adapters, or use the `Standard` bundle.
- **Rich source abstractions.** `Number`, `Map`, `List` expose typed views so your methods don't touch `reflect`.
- **Soft-error protocol.** Methods can decline (`ErrNotHandled`) and let the pipeline try defaults.
- **Struct walker.** Full tag grammar (`required`, `alias=`, `remainder`, `inline`, `strict`, `-`, `match=`), YAML/JSON tag fallback, configurable match modes.
- **Pointer resolution modes.** `Flat` (default), `None`, `Deep` — for both src and dest.
- **Designed for codegen.** Types that implement `Unface(src any) error` skip reflection. A future `unfacegen` tool will generate these from struct tags.

## Install

```bash
go get github.com/schneid-l/unface
```

Requires Go 1.25+.

## The Un*er interfaces

| Interface    | Method                               | Source type     |
|--------------|--------------------------------------|-----------------|
| Unfacer      | `Unface(src any) error`              | any             |
| Unstringer   | `Unstring(s string) error`           | `string`        |
| Unbooler     | `Unbool(b bool) error`               | `bool`          |
| Unnumberer   | `Unnumber(n Number) error`           | numeric         |
| Unbyteser    | `Unbytes(b []byte) error`            | `[]byte`        |
| Unruner      | `Unrune(r rune) error`               | `rune`          |
| Unmapper     | `Unmap(m Map) error`                 | map-like        |
| Unlister     | `Unlist(l List) error`               | slice/array     |
| Unniler      | `Unnil() error`                      | `nil`           |
| Untimer      | `Untime(t time.Time) error`          | `time.Time`     |
| Undurationer | `Unduration(d time.Duration) error`  | `time.Duration` |

## Tag grammar

```
`unface:"<name>[,<modifier>[=<value>]]..."`
```

Modifiers: `required`, `alias=X` (repeatable), `remainder`, `strict`, `inline`, `-` (skip), `match=exact|insensitive|fold`.

Struct-wide options via the zero-size marker field:

```go
type Config struct {
    _    struct{} `unface:",match=exact,unknown=error,tags=unface+json"`
    Port int      `unface:"port"`
}
```

## Built-in plugins

Scalars: `StringPlugin`, `BoolPlugin`, `BytesPlugin`, `RunePlugin`.
Numbers (atomic): `Int8Plugin` ... `Int64Plugin`, `IntPlugin`, `Uint8Plugin` ... `UintPlugin`, `Float32Plugin`, `Float64Plugin`, `Complex64Plugin`, `Complex128Plugin`, `BigIntPlugin`, `BigFloatPlugin`.
Structural: `StructPlugin`, `MapPlugin`, `ListPlugin`, `PointerPlugin`.
Special: `TimePlugin`, `JSONRawPlugin`.
Composites: `IntPluginBundle`, `UintPluginBundle`, `FloatPluginBundle`, `ComplexPluginBundle`, `NumberPlugin`, `PrimitivesPlugin`, `StandardPlugin`.

## When to reach for unface (vs. alternatives)

| Problem | Use |
|---|---|
| Bind a JSON body to a struct with exact field names | `encoding/json` |
| Bind a map to a struct, scalar-or-object fields, type coercion | **unface** |
| Bind with validation, default values, structural transforms | **unface** + your validator |

## Status

Pre-v1.0. The v0.x line may break the API if a serious design flaw surfaces; breaking changes are called out in `CHANGELOG.md`.

## License

MIT. See [LICENSE](LICENSE).
