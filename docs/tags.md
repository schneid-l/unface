# Tags

Struct tags control how `unface` walks your types. This page is the complete reference.

## Grammar

```
`unface:"<name>[,<modifier>[=<value>]]..."`
```

- `<name>` is the source key the walker will look for. Use `-` to skip the field.
- Each subsequent comma-separated token is a modifier; the order does not matter.
- Unknown modifiers cause the tag parser to return an error, which is swallowed by the struct walker (the field is then read through the fallback-tag path). Misspellings are thus soft failures — use strict CI checks to catch them if you rely on exotic modifiers.

## Field modifiers

| Modifier | Meaning |
|---|---|
| `<name>` *(positional, first)* | Source key to look up. Empty means "use the Go field name" unless overridden by fallback tags. |
| `-` *(as `<name>`)* | Skip this field entirely. |
| `required` | Error with `ErrRequired` if the source map is missing this key. |
| `alias=X` | Additional accepted source key. Repeat for multiple aliases: `alias=host_addr,alias=hostname`. |
| `remainder` | Catch-all field. Must be a `map[string]any` (or compatible). All source keys that didn't match any other field (and their values) are collected here. Suppresses the `UnknownError` policy. |
| `strict` | Disables plugin fallback for this field. Only the field type's own `Un*er` methods will be consulted. Useful for types that want to refuse "lenient" coercions (e.g. rejecting `"42"` into an `int` field when policy is strict). |
| `inline` | Walk this field's struct fields into the parent's key namespace, same as an anonymous embedded struct. Useful for composing config fragments. |
| `match=<mode>` | Override the match mode for this one field. Values: `fold` (default), `insensitive`, `exact` / `strict`. |
| `omitempty` | Reserved for future symmetry; currently a no-op. |

## Examples

```go
type Server struct {
    Host    string   `unface:"host,required,alias=hostname"`
    Port    int      `unface:"port"`
    Backup  *string  `unface:"backup"`               // pointer auto-allocated
    Timeout Duration `unface:"timeout_s"`            // renamed
    Secret  string   `unface:"-"`                    // skipped
    Extras  map[string]any `unface:",remainder"`     // catch-all
}
```

## Struct-wide options

Attach them to a zero-size marker field:

```go
type Config struct {
    _    struct{} `unface:",match=exact,unknown=error,tags=unface+json"`
    Port int      `unface:"port"`
}
```

The marker's `unface` tag is parsed for struct-level options only. The positional name is ignored (leave it empty).

| Modifier | Effect |
|---|---|
| `match=<mode>` | Sets the match mode for every field in this struct. Overridden by per-field `match=` if present. |
| `unknown=<policy>` | `ignore` \| `error` \| `warn`. Governs source keys that don't match any field (and aren't absorbed by `remainder`). |
| `tags=<a>+<b>+...` | Overrides the tag-fallback order (default: `unface+yaml+json`). The first tag whose name is non-empty wins. |

## Match modes

| Mode | Behavior |
|---|---|
| `fold` (default) | Case-insensitive + snake_case/kebab-case/CamelCase folded together. `"http_port"`, `"http-port"`, `"httpPort"`, and `"HTTPPort"` all match a field named `HTTPPort`. Full-Unicode folding. |
| `insensitive` | Case-insensitive only. Separators stay significant: `http_port` ≠ `HTTPPort`. |
| `exact` (alias: `strict`) | Byte-for-byte equality. |

## Unknown-key policy

Applies when the walker finishes and there are source keys that didn't match any field.

| Policy | Effect |
|---|---|
| `ignore` (default) | Silently drop. |
| `error` | Return an error wrapping `ErrUnknownField` with the offending key name. |
| `warn` | Call the handler set via `OnUnknown(UnknownWarn, handler)` once per unknown key, then continue. |

A `remainder`-tagged field absorbs all unknowns before this policy is checked, short-circuiting the logic for that struct.

## Tag fallback

When a field has no `unface` tag (or the `unface` tag has no name), the walker looks at the fallback list in order. For each tag it checks `Tag.Lookup(<name>)` and takes the first non-empty positional value as the source name.

Default order: `["unface", "yaml", "json"]`.

Only the **name** comes from fallback tags. Modifiers are `unface`-only.

If every fallback also yields nothing, the Go field name is used verbatim.

### Overriding the fallback

- Instance-wide: `unface.New(unface.WithTagFallback("unface", "toml"))`
- Disable fallback entirely: `unface.WithoutTagFallback()` (only the `unface` tag is consulted)
- Per-struct: use the marker field's `tags=` option

## Embedded structs

Anonymous struct fields are walked inline by default — their fields participate in the parent's key namespace:

```go
type Meta struct {
    Created string `unface:"created"`
}
type Config struct {
    Meta                      // inlined automatically
    Name string `unface:"name"`
}
```

A source `{created: "...", name: "..."}` populates both. Add an explicit `unface:"meta"` tag on the embedded field to nest it instead.

## Skip vs omitempty

`-` skips the field entirely: the walker never consults the source for it, and the value retains its current (zero) value. `omitempty` is reserved for a future serialization direction; it has no effect today.

## Required + remainder

You can combine them: required fields are checked after the remainder has absorbed unknowns. If `required` and `remainder` are both set on the same field, the remainder behavior wins (a remainder is by definition an optional catch-all).

## Pointer and zero handling

- `*T` fields whose source is present are auto-allocated and the element is coerced.
- `*T` fields whose source is absent stay `nil`.
- Value-type fields whose source is absent retain the zero value.
- Explicit `nil` in the source: if the field type has an `Unniler`, it is called; otherwise the field is zeroed.
