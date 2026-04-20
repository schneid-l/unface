package unfacers

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/schneid-l/unface/plugin"
)

type complexKind interface {
	~complex64 | ~complex128
}

type complexAdapter[T complexKind] struct{ ptr *T }

func (a complexAdapter[T]) Unface(src any) error {
	if src == nil {
		var zero T
		*a.ptr = zero
		return nil
	}
	if n, ok := plugin.NumberOf(src); ok {
		c, ok := n.Complex128()
		if !ok {
			return plugin.ErrNotHandled
		}
		*a.ptr = T(c)
		return nil
	}
	if s, ok := src.(string); ok {
		c, err := strconv.ParseComplex(s, 128)
		if err != nil {
			return fmt.Errorf("unface: parse complex from %q: %w", s, err)
		}
		*a.ptr = T(c)
		return nil
	}
	return plugin.ErrNotHandled
}

func newComplexPlugin[T complexKind](name string) plugin.Plugin {
	var zero T
	typ := reflect.TypeOf(zero)
	return plugin.NewPlugin(name, plugin.FactoryFunc(
		func(t reflect.Type) bool { return t == typ },
		func(ptr any) plugin.Adapter { return complexAdapter[T]{ptr: ptr.(*T)} },
	))
}

// Atomic complex plugins.
var (
	Complex64Plugin  = newComplexPlugin[complex64]("complex64")
	Complex128Plugin = newComplexPlugin[complex128]("complex128")
)
