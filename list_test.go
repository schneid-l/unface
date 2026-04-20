package unface_test

import (
	"reflect"
	"testing"

	"github.com/schneid-l/unface"
)

func TestListOfSliceAny(t *testing.T) {
	l, ok := unface.ListOf([]any{1, 2, 3})
	if !ok {
		t.Fatal("ListOf must accept []any")
	}
	if l.Len() != 3 || l.At(1) != 2 {
		t.Fatalf("got len=%d at1=%v", l.Len(), l.At(1))
	}
}

func TestListOfTypedSlice(t *testing.T) {
	l, ok := unface.ListOf([]int{10, 20, 30})
	if !ok {
		t.Fatal("ListOf must accept []int")
	}
	if l.ElemType().Kind() != reflect.Int {
		t.Fatalf("elem kind=%v", l.ElemType().Kind())
	}
	var sum int
	l.Iter(func(_ int, v any) bool { sum += v.(int); return true })
	if sum != 60 {
		t.Fatalf("sum=%d", sum)
	}
}

func TestListOfArray(t *testing.T) {
	l, ok := unface.ListOf([3]string{"a", "b", "c"})
	if !ok {
		t.Fatal("ListOf must accept arrays")
	}
	if l.Len() != 3 || l.At(0) != "a" {
		t.Fatalf("got len=%d at0=%v", l.Len(), l.At(0))
	}
}

func TestListOfRejectsNonList(t *testing.T) {
	for _, in := range []any{"str", 42, map[string]int{}} {
		if _, ok := unface.ListOf(in); ok {
			t.Fatalf("ListOf(%T) should be !ok", in)
		}
	}
}

func TestListSlice(t *testing.T) {
	l, _ := unface.ListOf([]int{1, 2, 3, 4, 5})
	sl := l.Slice(1, 4)
	if sl.Len() != 3 || sl.At(0) != 2 {
		t.Fatalf("slice len=%d at0=%v", sl.Len(), sl.At(0))
	}
}

func TestListSliceOfArray(t *testing.T) {
	l, _ := unface.ListOf([5]int{10, 20, 30, 40, 50})
	sl := l.Slice(1, 4)
	if sl.Len() != 3 || sl.At(0) != 20 || sl.At(2) != 40 {
		t.Fatalf("slice values=%v,%v,%v", sl.At(0), sl.At(1), sl.At(2))
	}
}

func TestListToSlice(t *testing.T) {
	l, _ := unface.ListOf([]int{1, 2, 3})
	s := l.ToSlice()
	if len(s) != 3 || s[2] != 3 {
		t.Fatalf("got %v", s)
	}
}

func TestListMapFilter(t *testing.T) {
	l, _ := unface.ListOf([]int{1, 2, 3, 4})
	out := l.Filter(func(v any) bool { return v.(int)%2 == 0 }).
		Map(func(v any) any { return v.(int) * 10 })
	if out.Len() != 2 || out.At(0) != 20 || out.At(1) != 40 {
		t.Fatalf("got len=%d at=%v,%v", out.Len(), out.At(0), out.At(1))
	}
}

func TestListRaw(t *testing.T) {
	in := []int{1, 2, 3}
	l, _ := unface.ListOf(in)
	raw, ok := l.Raw().([]int)
	if !ok || len(raw) != 3 {
		t.Fatalf("raw=%v", l.Raw())
	}
}

func TestListIterEarlyExit(t *testing.T) {
	l, _ := unface.ListOf([]int{1, 2, 3, 4})
	count := 0
	l.Iter(func(i int, _ any) bool {
		count++
		return i < 1 // stop after index 1
	})
	if count != 2 {
		t.Fatalf("count=%d want 2", count)
	}
}
