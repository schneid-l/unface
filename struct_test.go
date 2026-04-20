package unface_test

import (
	"errors"
	"testing"

	"github.com/schneid-l/unface"
)

// For these tests we build a "kitchen-sink" facer with the plugins needed
// by nested field coercions.
func structFacer(extra ...unface.Option) *unface.Facer {
	opts := make([]unface.Option, 0, 1+len(extra))
	opts = append(opts, unface.With(
		unface.StructPlugin,
		unface.MapPlugin,
		unface.ListPlugin,
		unface.StringPlugin,
		unface.BoolPlugin,
		unface.IntPlugin,
		unface.Int64Plugin,
		unface.Float64Plugin,
	))
	opts = append(opts, extra...)
	return unface.New(opts...)
}

type serverCfg struct {
	Host string `unface:"host"`
	Port int    `unface:"port,required"`
}

type appCfg struct {
	Name   string            `unface:"name"`
	Server serverCfg         `unface:"server"`
	Tags   []string          `unface:"tags"`
	Labels map[string]string `unface:"labels"`
	Secret string            `unface:"-"`
	Remain map[string]any    `unface:",remainder"`
}

func TestStructBasic(t *testing.T) {
	var c appCfg
	src := map[string]any{
		"name":   "demo",
		"server": map[string]any{"host": "localhost", "port": 8080},
		"tags":   []any{"a", "b"},
		"labels": map[string]any{"env": "prod"},
		"extra":  123,
	}
	if err := structFacer().Unface(src, &c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "demo" || c.Server.Host != "localhost" || c.Server.Port != 8080 {
		t.Fatalf("c=%+v", c)
	}
	if len(c.Tags) != 2 || c.Tags[0] != "a" {
		t.Fatalf("tags=%v", c.Tags)
	}
	if c.Labels["env"] != "prod" {
		t.Fatalf("labels=%v", c.Labels)
	}
	if v, ok := c.Remain["extra"]; !ok || v != 123 {
		t.Fatalf("remain=%v", c.Remain)
	}
}

func TestStructRequiredMissing(t *testing.T) {
	var s serverCfg
	err := structFacer().Unface(map[string]any{"host": "h"}, &s)
	if !errors.Is(err, unface.ErrRequired) {
		t.Fatalf("err=%v", err)
	}
}

func TestStructAliases(t *testing.T) {
	type T struct {
		Addr string `unface:"address,alias=addr,alias=host_addr"`
	}
	var v T
	if err := structFacer().Unface(map[string]any{"addr": "1.2.3.4"}, &v); err != nil {
		t.Fatal(err)
	}
	if v.Addr != "1.2.3.4" {
		t.Fatalf("Addr=%q", v.Addr)
	}
}

func TestStructInline(t *testing.T) {
	type Inner struct {
		A int `unface:"a"`
		B int `unface:"b"`
	}
	type Outer struct {
		I Inner `unface:",inline"`
		C int   `unface:"c"`
	}
	var o Outer
	src := map[string]any{"a": 1, "b": 2, "c": 3}
	if err := structFacer().Unface(src, &o); err != nil {
		t.Fatal(err)
	}
	if o.I.A != 1 || o.I.B != 2 || o.C != 3 {
		t.Fatalf("o=%+v", o)
	}
}

func TestStructEmbedded(t *testing.T) {
	type Base struct {
		Version int `unface:"version"`
	}
	type Full struct {
		Base
		Name string `unface:"name"`
	}
	var f Full
	if err := structFacer().Unface(map[string]any{"version": 1, "name": "x"}, &f); err != nil {
		t.Fatal(err)
	}
	if f.Version != 1 || f.Name != "x" {
		t.Fatalf("f=%+v", f)
	}
}

func TestStructMatchFold(t *testing.T) {
	type T struct {
		HTTPPort int `unface:"http_port"`
	}
	var v T
	if err := structFacer().Unface(map[string]any{"HTTP_PORT": 80}, &v); err != nil {
		t.Fatal(err)
	}
	if v.HTTPPort != 80 {
		t.Fatalf("got %d", v.HTTPPort)
	}
}

func TestStructUnknownErrorPolicy(t *testing.T) {
	type T struct {
		A int `unface:"a"`
	}
	var v T
	err := structFacer(unface.OnUnknown(unface.UnknownError)).Unface(
		map[string]any{"a": 1, "b": 2}, &v,
	)
	if !errors.Is(err, unface.ErrUnknownField) {
		t.Fatalf("err=%v", err)
	}
}

func TestStructMarkerOverride(t *testing.T) {
	type T struct {
		_    struct{} `unface:",match=exact,unknown=error"`
		Port int      `unface:"Port"`
	}
	var v T
	err := structFacer().Unface(map[string]any{"port": 80}, &v)
	if !errors.Is(err, unface.ErrUnknownField) {
		t.Fatalf("err=%v", err)
	}
}

func TestStructYAMLTagFallback(t *testing.T) {
	type T struct {
		Port int `yaml:"port"`
	}
	var v T
	if err := structFacer().Unface(map[string]any{"port": 8080}, &v); err != nil {
		t.Fatal(err)
	}
	if v.Port != 8080 {
		t.Fatalf("port=%d", v.Port)
	}
}

func TestStructDefaultFieldName(t *testing.T) {
	type T struct {
		Port int
	}
	var v T
	if err := structFacer().Unface(map[string]any{"port": 8080}, &v); err != nil {
		t.Fatal(err)
	}
	if v.Port != 8080 {
		t.Fatalf("port=%d", v.Port)
	}
}
