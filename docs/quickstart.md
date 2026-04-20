# Quickstart

**What this page covers.** A five-minute tour: install the library, coerce a map into a struct, add one `Un*er` method to accept shorthand, and run the example.

## 1. Install

```bash
go get github.com/schneid-l/unface
```

Requires Go 1.25+.

## 2. Coerce a map into a struct

The top-level `unface.Unface(src, dest)` uses a default `Facer` preloaded with every built-in plugin. The struct walker reads `unface`, `yaml`, and `json` tags (in that order) and coerces scalar types for free.

```go
package main

import (
	"fmt"

	"github.com/schneid-l/unface"
)

type Server struct {
	Host string `unface:"host,required"`
	Port int    `unface:"port"`
}

func main() {
	raw := map[string]any{
		"host": "api.example.com",
		"port": "8080", // string → int coerced by IntPlugin
	}
	var s Server
	if err := unface.Unface(raw, &s); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", s) // {Host:api.example.com Port:8080}
}
```

No tag? The walker falls back to `yaml:"..."` then `json:"..."`, then the Go field name with fold-matching.

## 3. Accept both a string and a map for the same field

Implement `Unstring(s string) error` on your destination type. The dispatcher calls it whenever the source is a string; map sources still go through the default struct walker.

```go
type URL struct {
	Scheme string `unface:"scheme"`
	Host   string `unface:"host,required"`
}

// Unstring teaches the library how to parse the shorthand form.
func (u *URL) Unstring(s string) error {
	u.Scheme = "https"
	u.Host = s
	return nil
}

type Config struct {
	API URL `unface:"api"`
	Web URL `unface:"web"`
}

func main() {
	raw := map[string]any{
		"api": "api.example.com",                              // string → Unstring
		"web": map[string]any{"scheme": "https", "host": "x"}, // map → walker
	}
	var c Config
	_ = unface.Unface(raw, &c)
}
```

## 4. Know what just happened

1. `Unface(src, &dst)` enters the dispatcher ([dispatch-order.md](./dispatch-order.md)).
2. For each struct field, the walker extracts the matching key from the source map.
3. For each field value it re-enters the dispatcher:
   - If the field type implements `Unfacer` / `Un<srcKind>er`, that wins (step 5 in the dispatch order).
   - Otherwise, a built-in plugin (`IntPlugin`, `StringPlugin`, `StructPlugin`, ...) handles it.
4. Missing-but-`required` keys return `ErrRequired`. Type-mismatch returns a `*unface.Error` with a dotted path.

## 5. Where to next

- [dispatch-order.md](./dispatch-order.md) — the exact pipeline.
- [tags.md](./tags.md) — every struct-tag modifier.
- [guides/http-binding.md](./guides/http-binding.md) — bind a JSON body.
- [guides/authoring-plugins.md](./guides/authoring-plugins.md) — write your own plugin.
- [reference/options.md](./reference/options.md) — every option, with examples.
- [faq.md](./faq.md) — common questions.
