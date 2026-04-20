package unface_test

import (
	"reflect"
	"testing"

	"github.com/schneid-l/unface"
)

// --- List edges ---

func TestListOfNilRejected(t *testing.T) {
	if _, ok := unface.ListOf(nil); ok {
		t.Fatal("ListOf(nil) must be !ok")
	}
}

func TestListEmpty(t *testing.T) {
	l, _ := unface.ListOf([]int{})
	if l.Len() != 0 {
		t.Fatalf("empty slice Len=%d", l.Len())
	}
	if s := l.ToSlice(); len(s) != 0 {
		t.Fatalf("ToSlice=%v", s)
	}
	count := 0
	l.Iter(func(_ int, _ any) bool { count++; return true })
	if count != 0 {
		t.Fatalf("Iter on empty: count=%d", count)
	}
}

func TestListFilterEmpty(t *testing.T) {
	l, _ := unface.ListOf([]int{1, 2, 3})
	out := l.Filter(func(_ any) bool { return false })
	if out.Len() != 0 {
		t.Fatalf("empty filter len=%d", out.Len())
	}
}

func TestListFilterAll(t *testing.T) {
	l, _ := unface.ListOf([]int{1, 2, 3})
	out := l.Filter(func(_ any) bool { return true })
	if out.Len() != 3 {
		t.Fatalf("full filter len=%d", out.Len())
	}
}

func TestListElemTypeAnyForSliceOfAny(t *testing.T) {
	l, _ := unface.ListOf([]any{1, "s"})
	k := l.ElemType().Kind()
	if k != reflect.Interface {
		t.Fatalf("ElemType=%v", k)
	}
}

// --- Map edges ---

func TestMapOfNilRejected(t *testing.T) {
	if _, ok := unface.MapOf(nil); ok {
		t.Fatal("MapOf(nil) must be !ok")
	}
}

func TestMapEmpty(t *testing.T) {
	m, _ := unface.MapOf(map[string]int{})
	if m.Len() != 0 || len(m.Keys()) != 0 {
		t.Fatal("empty map non-zero")
	}
}

func TestMapConvertKeyDifferentType(t *testing.T) {
	// map[int]int with string key lookup — not convertible, must miss.
	m, _ := unface.MapOf(map[int]int{1: 10})
	// int → int succeeds
	if v, ok := m.Get(1); !ok || v != 10 {
		t.Fatalf("Get(1)=%v ok=%v", v, ok)
	}
	// string → int: not convertible
	if _, ok := m.Get("1"); ok {
		t.Fatal("string key should not convert to int key")
	}
}

func TestMapConvertKeyInvalidInterface(t *testing.T) {
	m, _ := unface.MapOf(map[string]int{"a": 1})
	// reflect.Invalid key (nil any can't convert to string) → miss
	if _, ok := m.Get(nil); ok {
		t.Fatal("nil key should miss")
	}
}

func TestMapGetInt64FromString(t *testing.T) {
	m, _ := unface.MapOf(map[string]any{"n": "not-a-number"})
	if _, ok := m.GetInt64("n"); ok {
		t.Fatal("GetInt64 on non-numeric string should be !ok")
	}
}

func TestMapGetMapMissing(t *testing.T) {
	m, _ := unface.MapOf(map[string]any{})
	if _, ok := m.GetMap("x"); ok {
		t.Fatal("GetMap missing key should be !ok")
	}
}

func TestMapGetListMissing(t *testing.T) {
	m, _ := unface.MapOf(map[string]any{})
	if _, ok := m.GetList("x"); ok {
		t.Fatal("GetList missing key should be !ok")
	}
}

func TestMapGetMapWrongValueType(t *testing.T) {
	m, _ := unface.MapOf(map[string]any{"x": 42})
	if _, ok := m.GetMap("x"); ok {
		t.Fatal("GetMap on int value should be !ok")
	}
}

// --- Error type edges ---

func TestErrorWithPathNilReceiver(t *testing.T) {
	var e *unface.Error
	if got := e.WithPath("x"); got != nil {
		t.Fatalf("nil receiver must return nil, got %v", got)
	}
}

func TestErrorPathPreservesWithMultipleWraps(t *testing.T) {
	inner := &unface.Error{Path: []string{"port"}, Err: errorsNew("bad")}
	mid := inner.WithPath("server")
	outer := mid.WithPath("config")
	if outer.Error() != "unface: config.server.port: bad" {
		t.Fatalf("got %q", outer.Error())
	}
	// Intermediate must be unchanged.
	if mid.Error() != "unface: server.port: bad" {
		t.Fatalf("mid=%q", mid.Error())
	}
	if inner.Error() != "unface: port: bad" {
		t.Fatalf("inner=%q", inner.Error())
	}
}

// local helper to avoid importing errors from every test file
func errorsNew(msg string) error { return &stringErr{msg: msg} }

type stringErr struct{ msg string }

func (e *stringErr) Error() string { return e.msg }
