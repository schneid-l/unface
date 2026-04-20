package reflectutil_test

import (
	"reflect"
	"testing"

	"github.com/schneid-l/unface/internal/reflectutil"
)

func TestDerefToAddressable(t *testing.T) {
	var i int
	v, err := reflectutil.DerefToAddressable(&i)
	if err != nil {
		t.Fatal(err)
	}
	if !v.CanSet() {
		t.Fatal("value must be settable")
	}
	if v.Kind() != reflect.Int {
		t.Fatalf("kind=%v want int", v.Kind())
	}
}

func TestDerefRejectsNonPointer(t *testing.T) {
	if _, err := reflectutil.DerefToAddressable(42); err == nil {
		t.Fatal("expected error for non-pointer")
	}
}

func TestDerefRejectsNilPointer(t *testing.T) {
	var p *int
	if _, err := reflectutil.DerefToAddressable(p); err == nil {
		t.Fatal("expected error for nil pointer")
	}
}

func TestDerefRejectsNilAny(t *testing.T) {
	if _, err := reflectutil.DerefToAddressable(nil); err == nil {
		t.Fatal("expected error for nil any")
	}
}

func TestDerefAllocatesNestedPointer(t *testing.T) {
	var p *int
	v, err := reflectutil.DerefToAddressable(&p)
	if err != nil {
		t.Fatal(err)
	}
	if v.Kind() != reflect.Int {
		t.Fatalf("kind=%v want int", v.Kind())
	}
	if p == nil {
		t.Fatal("nested pointer should have been allocated")
	}
}

func TestIsNilLike(t *testing.T) {
	if !reflectutil.IsNilLike(nil) {
		t.Fatal("nil any")
	}
	var p *int
	if !reflectutil.IsNilLike(p) {
		t.Fatal("nil pointer")
	}
	var m map[string]int
	if !reflectutil.IsNilLike(m) {
		t.Fatal("nil map")
	}
	var s []int
	if !reflectutil.IsNilLike(s) {
		t.Fatal("nil slice")
	}
	if reflectutil.IsNilLike(42) {
		t.Fatal("int 42 is not nil-like")
	}
	x := 1
	if reflectutil.IsNilLike(&x) {
		t.Fatal("non-nil pointer")
	}
}
