package unfacers

import (
	"fmt"
	"math"
	"reflect"
	"strconv"

	"github.com/schneid-l/unface/plugin"
)

type signedInt interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type signedAdapter[T signedInt] struct{ ptr *T }

func (a signedAdapter[T]) Unface(src any) error {
	if src == nil {
		var zero T
		*a.ptr = zero
		return nil
	}
	if n, ok := plugin.NumberOf(src); ok {
		return a.Unnumber(n)
	}
	if s, ok := src.(string); ok {
		return a.Unstring(s)
	}
	if b, ok := src.(bool); ok {
		return a.Unbool(b)
	}
	return plugin.ErrNotHandled
}

func (a signedAdapter[T]) Unnumber(n plugin.Number) error {
	if v, ok := n.Int64(); ok {
		return a.setFromInt64(v)
	}
	if u, ok := n.Uint64(); ok {
		if u > math.MaxInt64 {
			return fmt.Errorf("unface: %s overflows %T", n.String(), *a.ptr)
		}
		return a.setFromInt64(int64(u))
	}
	return plugin.ErrNotHandled
}

func (a signedAdapter[T]) setFromInt64(v int64) error {
	converted := T(v)
	if int64(converted) != v {
		return fmt.Errorf("unface: %d overflows %T", v, *a.ptr)
	}
	*a.ptr = converted
	return nil
}

func (a signedAdapter[T]) Unstring(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return fmt.Errorf("unface: parse int from %q: %w", s, err)
	}
	return a.setFromInt64(v)
}

func (a signedAdapter[T]) Unbool(b bool) error {
	if b {
		*a.ptr = 1
	} else {
		*a.ptr = 0
	}
	return nil
}

func newSignedPlugin[T signedInt](name string) plugin.Plugin {
	var zero T
	typ := reflect.TypeOf(zero)
	return plugin.NewPlugin(name, plugin.FactoryFunc(
		func(t reflect.Type) bool { return t == typ },
		func(ptr any) plugin.Adapter { return signedAdapter[T]{ptr: ptr.(*T)} },
	))
}

// Atomic signed-integer plugins.
var (
	Int8Plugin  = newSignedPlugin[int8]("int8")
	Int16Plugin = newSignedPlugin[int16]("int16")
	Int32Plugin = newSignedPlugin[int32]("int32")
	Int64Plugin = newSignedPlugin[int64]("int64")
	IntPlugin   = newSignedPlugin[int]("int")
)
