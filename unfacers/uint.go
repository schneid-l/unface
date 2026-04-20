package unfacers

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/schneid-l/unface/plugin"
)

type unsignedInt interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type unsignedAdapter[T unsignedInt] struct{ ptr *T }

func (a unsignedAdapter[T]) Unface(src any) error {
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

func (a unsignedAdapter[T]) Unnumber(n plugin.Number) error {
	v, ok := n.Uint64()
	if !ok {
		return fmt.Errorf("unface: %s cannot convert to %T", n.String(), *a.ptr)
	}
	converted := T(v)
	if uint64(converted) != v {
		return fmt.Errorf("unface: %d overflows %T", v, *a.ptr)
	}
	*a.ptr = converted
	return nil
}

func (a unsignedAdapter[T]) Unstring(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		return fmt.Errorf("unface: parse uint from %q: %w", s, err)
	}
	converted := T(v)
	if uint64(converted) != v {
		return fmt.Errorf("unface: %d overflows %T", v, *a.ptr)
	}
	*a.ptr = converted
	return nil
}

func (a unsignedAdapter[T]) Unbool(b bool) error {
	if b {
		*a.ptr = 1
	} else {
		*a.ptr = 0
	}
	return nil
}

func newUnsignedPlugin[T unsignedInt](name string) plugin.Plugin {
	var zero T
	typ := reflect.TypeOf(zero)
	return plugin.NewPlugin(name, plugin.FactoryFunc(
		func(t reflect.Type) bool { return t == typ },
		func(ptr any) plugin.Adapter { return unsignedAdapter[T]{ptr: ptr.(*T)} },
	))
}

// Atomic unsigned-integer plugins.
var (
	Uint8Plugin  = newUnsignedPlugin[uint8]("uint8")
	Uint16Plugin = newUnsignedPlugin[uint16]("uint16")
	Uint32Plugin = newUnsignedPlugin[uint32]("uint32")
	Uint64Plugin = newUnsignedPlugin[uint64]("uint64")
	UintPlugin   = newUnsignedPlugin[uint]("uint")
)
