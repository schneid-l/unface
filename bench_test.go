package unface_test

import (
	"testing"

	"github.com/schneid-l/unface"
)

// BenchmarkScalarCoercion measures string→int coercion hot-path cost.
func BenchmarkScalarCoercion(b *testing.B) {
	f := unface.New(unface.With(unface.Int64Plugin))
	b.ReportAllocs()
	for range b.N {
		var i int64
		if err := f.Unface("1234567890", &i); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDirectAssignment measures the fast-path (identical types, no
// plugin needed).
func BenchmarkDirectAssignment(b *testing.B) {
	f := unface.New()
	b.ReportAllocs()
	for range b.N {
		var i int
		if err := f.Unface(42, &i); err != nil {
			b.Fatal(err)
		}
	}
}

type benchServer struct {
	Host string `unface:"host"`
	Port int    `unface:"port"`
}

type benchConfig struct {
	Name    string        `unface:"name"`
	Server  benchServer   `unface:"server"`
	Servers []benchServer `unface:"servers"`
	Tags    []string      `unface:"tags"`
}

// BenchmarkStructWalk measures the end-to-end cost of a typical YAML-ish
// config: nested struct + slice of structs.
func BenchmarkStructWalk(b *testing.B) {
	f := unface.New(unface.With(unface.StandardPlugin))
	src := map[string]any{
		"name":   "demo",
		"server": map[string]any{"host": "localhost", "port": "8080"},
		"servers": []any{
			map[string]any{"host": "a", "port": 1},
			map[string]any{"host": "b", "port": 2},
			map[string]any{"host": "c", "port": 3},
		},
		"tags": []any{"prod", "eu", "web"},
	}
	b.ReportAllocs()
	for range b.N {
		var cfg benchConfig
		if err := f.Unface(src, &cfg); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPluginLookup measures the cost of iterating the plugin set to
// find a match; exercises the factory dispatch loop.
func BenchmarkPluginLookup(b *testing.B) {
	f := unface.New(unface.With(unface.StandardPlugin))
	b.ReportAllocs()
	for range b.N {
		var s string
		if err := f.Unface(42, &s); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPointerResolveDeep measures the worst-case overhead of Deep
// pointer resolution on a 3-level pointer dest.
func BenchmarkPointerResolveDeep(b *testing.B) {
	f := unface.New(
		unface.With(unface.IntPlugin),
		unface.WithPointerResolve(unface.PointerResolveDeep),
	)
	b.ReportAllocs()
	for range b.N {
		var p ***int
		if err := f.Unface(42, &p); err != nil {
			b.Fatal(err)
		}
	}
}
