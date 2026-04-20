package reflectutil_test

import (
	"testing"

	"github.com/schneid-l/unface/internal/reflectutil"
)

func TestDerefDeeplyNested(t *testing.T) {
	var p ***int
	v, err := reflectutil.DerefToAddressable(&p)
	if err != nil {
		t.Fatal(err)
	}
	if !v.CanSet() {
		t.Fatal("must be settable")
	}
	// All nested pointers should now be non-nil.
	if p == nil || *p == nil || **p == nil {
		t.Fatalf("nested allocation failed: p=%v", p)
	}
}

func TestDerefNonPointerAny(t *testing.T) {
	for _, v := range []any{"str", 42, []int{1}, map[string]int{}} {
		if _, err := reflectutil.DerefToAddressable(v); err == nil {
			t.Errorf("DerefToAddressable(%T) expected error", v)
		}
	}
}

func TestIsNilLikeChannelFuncInterface(t *testing.T) {
	var c chan int
	if !reflectutil.IsNilLike(c) {
		t.Fatal("nil chan")
	}
	var f func()
	if !reflectutil.IsNilLike(f) {
		t.Fatal("nil func")
	}
	var i any
	if !reflectutil.IsNilLike(i) {
		t.Fatal("nil interface")
	}
}

func TestIsNilLikeNonNillable(t *testing.T) {
	if reflectutil.IsNilLike(0) {
		t.Fatal("0 is not nil-like")
	}
	if reflectutil.IsNilLike("") {
		t.Fatal("empty string is not nil-like")
	}
	if reflectutil.IsNilLike(struct{}{}) {
		t.Fatal("struct{} is not nil-like")
	}
}
