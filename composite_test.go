package unface_test

import (
	"testing"

	"github.com/schneid-l/unface"
)

func TestStandardPluginStringToInt(t *testing.T) {
	var i int
	if err := unface.Unface("42", &i); err != nil {
		t.Fatal(err)
	}
	if i != 42 {
		t.Fatalf("i=%d", i)
	}
}

func TestStandardPluginMapToStruct(t *testing.T) {
	type T struct {
		A int    `unface:"a"`
		B string `unface:"b"`
	}
	var v T
	if err := unface.Unface(map[string]any{"a": 1, "b": "hi"}, &v); err != nil {
		t.Fatal(err)
	}
	if v.A != 1 || v.B != "hi" {
		t.Fatalf("v=%+v", v)
	}
}

func TestStandardPluginNestedYAMLLike(t *testing.T) {
	type Server struct {
		Host string `unface:"host"`
		Port int    `unface:"port"`
	}
	type Config struct {
		Name   string   `unface:"name"`
		Server Server   `unface:"server"`
		Tags   []string `unface:"tags"`
	}
	src := map[string]any{
		"name":   "demo",
		"server": map[string]any{"host": "localhost", "port": "8080"},
		"tags":   []any{"prod", "eu"},
	}
	var cfg Config
	if err := unface.Unface(src, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "demo" || cfg.Server.Host != "localhost" || cfg.Server.Port != 8080 {
		t.Fatalf("cfg=%+v", cfg)
	}
	if len(cfg.Tags) != 2 || cfg.Tags[0] != "prod" {
		t.Fatalf("tags=%v", cfg.Tags)
	}
}

func TestComposeRemoval(t *testing.T) {
	// Removing NumberPlugin (composite) should drop int coercion.
	f := unface.New(unface.With(unface.StandardPlugin), unface.Without(unface.NumberPlugin))
	var i int
	err := f.Unface("42", &i)
	if err == nil {
		t.Fatal("expected failure after removing NumberPlugin")
	}
}

func TestPackageLevelUnfaceUsesDefault(t *testing.T) {
	// Exercises that Default was wired by init().
	var i int
	if err := unface.Unface(42, &i); err != nil {
		t.Fatal(err)
	}
}

func TestPerCallOnlyOverride(t *testing.T) {
	var i int
	err := unface.Default.Unface("42", &i, unface.Only())
	if err == nil {
		t.Fatal("expected failure with Only() killing all plugins")
	}
}
