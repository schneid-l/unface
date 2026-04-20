package unface_test

import (
	"reflect"
	"testing"

	"github.com/schneid-l/unface"
)

func TestMapOfStringAny(t *testing.T) {
	m, ok := unface.MapOf(map[string]any{"a": 1, "b": "hi"})
	if !ok {
		t.Fatal("MapOf must accept map[string]any")
	}
	if m.Len() != 2 {
		t.Fatalf("len=%d", m.Len())
	}
	if v, ok := m.Get("a"); !ok || v != 1 {
		t.Fatalf("Get a=%v ok=%v", v, ok)
	}
}

func TestMapGetStringGetInt64GetBool(t *testing.T) {
	m, _ := unface.MapOf(map[string]any{
		"s": "hello",
		"n": 42,
		"b": true,
	})
	if s, ok := m.GetString("s"); !ok || s != "hello" {
		t.Fatalf("GetString got %q ok=%v", s, ok)
	}
	if n, ok := m.GetInt64("n"); !ok || n != 42 {
		t.Fatalf("GetInt64 got %d ok=%v", n, ok)
	}
	if b, ok := m.GetBool("b"); !ok || !b {
		t.Fatalf("GetBool got %v ok=%v", b, ok)
	}
}

func TestMapHasAndKeys(t *testing.T) {
	m, _ := unface.MapOf(map[string]int{"x": 1, "y": 2})
	if !m.Has("x") {
		t.Fatal("Has x")
	}
	if m.Has("z") {
		t.Fatal("Has z should be false")
	}
	keys := m.Keys()
	if len(keys) != 2 {
		t.Fatalf("keys=%d", len(keys))
	}
}

func TestMapOfRejectsNonMap(t *testing.T) {
	for _, in := range []any{"s", 1, []int{1}, nil} {
		if _, ok := unface.MapOf(in); ok {
			t.Fatalf("MapOf(%T) should be !ok", in)
		}
	}
}

func TestMapKeyValueTypes(t *testing.T) {
	m, _ := unface.MapOf(map[string]int{"a": 1})
	if m.KeyType().Kind() != reflect.String {
		t.Fatalf("KeyType=%v", m.KeyType().Kind())
	}
	if m.ValueType().Kind() != reflect.Int {
		t.Fatalf("ValueType=%v", m.ValueType().Kind())
	}
}

func TestMapGetNestedMapAndList(t *testing.T) {
	m, _ := unface.MapOf(map[string]any{
		"nested": map[string]any{"k": "v"},
		"items":  []any{1, 2, 3},
	})
	nm, ok := m.GetMap("nested")
	if !ok || nm.Len() != 1 {
		t.Fatalf("GetMap ok=%v", ok)
	}
	lst, ok := m.GetList("items")
	if !ok || lst.Len() != 3 {
		t.Fatalf("GetList ok=%v", ok)
	}
}

func TestMapIterEarlyExit(t *testing.T) {
	m, _ := unface.MapOf(map[string]int{"a": 1, "b": 2, "c": 3})
	count := 0
	m.Iter(func(_, _ any) bool {
		count++
		return count < 2
	})
	if count != 2 {
		t.Fatalf("count=%d", count)
	}
}

func TestMapGetMissing(t *testing.T) {
	m, _ := unface.MapOf(map[string]any{"a": 1})
	if _, ok := m.GetString("missing"); ok {
		t.Fatal("missing key should be !ok")
	}
	if _, ok := m.GetInt64("missing"); ok {
		t.Fatal("missing key should be !ok")
	}
	if _, ok := m.GetBool("missing"); ok {
		t.Fatal("missing key should be !ok")
	}
}

func TestMapRaw(t *testing.T) {
	in := map[string]int{"a": 1}
	m, _ := unface.MapOf(in)
	raw, ok := m.Raw().(map[string]int)
	if !ok || raw["a"] != 1 {
		t.Fatalf("raw=%v", m.Raw())
	}
}
