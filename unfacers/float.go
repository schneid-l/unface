package unfacers

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/schneid-l/unface/plugin"
)

type floatKind interface {
	~float32 | ~float64
}

type floatAdapter[T floatKind] struct{ ptr *T }

func (a floatAdapter[T]) Unface(src any) error {
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
	return plugin.ErrNotHandled
}

func (a floatAdapter[T]) Unnumber(n plugin.Number) error {
	v, ok := n.Float64()
	if !ok {
		return fmt.Errorf("unface: %s cannot convert to %T", n.String(), *a.ptr)
	}
	*a.ptr = T(v)
	return nil
}

func (a floatAdapter[T]) Unstring(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("unface: parse float from %q: %w", s, err)
	}
	*a.ptr = T(v)
	return nil
}

func newFloatPlugin[T floatKind](name string) plugin.Plugin {
	var zero T
	typ := reflect.TypeOf(zero)
	return plugin.NewPlugin(name, plugin.FactoryFunc(
		func(t reflect.Type) bool { return t == typ },
		func(ptr any) plugin.Adapter { return floatAdapter[T]{ptr: ptr.(*T)} },
	))
}

// Atomic float plugins.
var (
	Float32Plugin = newFloatPlugin[float32]("float32")
	Float64Plugin = newFloatPlugin[float64]("float64")
)
