// Command url demonstrates the URL-as-string-or-object case: a single
// field type that accepts either a shorthand string or a structured map.
//
// Run with:
//
//	go run ./examples/url
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/schneid-l/unface"
)

// URL is a config-friendly URL: accepts "https://host:port/path" OR a
// structured {scheme, host, port, path} map.
type URL struct {
	Scheme string `unface:"scheme"`
	Host   string `unface:"host,required"`
	Port   int    `unface:"port"`
	Path   string `unface:"path"`
}

// Unstring implements Unstringer so unface can parse the shorthand.
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
		if _, err := fmt.Sscanf(rest[i+1:], "%d", &u.Port); err != nil {
			return err
		}
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

func Load(raw map[string]any) (Config, error) {
	var cfg Config
	err := unface.Unface(raw, &cfg)
	return cfg, err
}

func main() {
	raw := map[string]any{
		"name": "demo",
		"api":  "https://api.example.com:8080/v1",
		"web": map[string]any{
			"scheme": "https",
			"host":   "example.com",
			"path":   "/",
		},
	}
	cfg, err := Load(raw)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("name=%s\n", cfg.Name)
	fmt.Printf("api =%+v\n", cfg.API)
	fmt.Printf("web =%+v\n", cfg.Web)
}
