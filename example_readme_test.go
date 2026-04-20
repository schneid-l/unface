package unface_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/schneid-l/unface"
)

// URL is the README example. It implements Unstringer so the library can
// parse the shorthand "https://host:port/path"; the map form falls through
// to the default struct walker.
type readmeURL struct {
	Scheme string `unface:"scheme"`
	Host   string `unface:"host,required"`
	Port   int    `unface:"port"`
	Path   string `unface:"path"`
}

func (u *readmeURL) Unstring(s string) error {
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

type readmeConfig struct {
	Name string    `unface:"name"`
	API  readmeURL `unface:"api"`
	Web  readmeURL `unface:"web"`
}

func TestReadmeExample(t *testing.T) {
	raw := map[string]any{
		"name": "demo",
		"api":  "https://api.example.com:8080/v1",
		"web": map[string]any{
			"scheme": "https",
			"host":   "example.com",
			"path":   "/",
		},
	}
	var cfg readmeConfig
	if err := unface.Unface(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "demo" {
		t.Fatalf("Name=%q", cfg.Name)
	}
	if cfg.API.Scheme != "https" || cfg.API.Host != "api.example.com" ||
		cfg.API.Port != 8080 || cfg.API.Path != "/v1" {
		t.Fatalf("API=%+v", cfg.API)
	}
	if cfg.Web.Scheme != "https" || cfg.Web.Host != "example.com" ||
		cfg.Web.Path != "/" || cfg.Web.Port != 0 {
		t.Fatalf("Web=%+v", cfg.Web)
	}
}
