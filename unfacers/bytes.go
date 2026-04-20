package unfacers

import (
	"reflect"

	"github.com/schneid-l/unface/plugin"
)

type bytesAdapter struct{ ptr *[]byte }

func (a bytesAdapter) Unface(src any) error {
	switch v := src.(type) {
	case []byte:
		out := make([]byte, len(v))
		copy(out, v)
		*a.ptr = out
		return nil
	case string:
		*a.ptr = []byte(v)
		return nil
	case nil:
		*a.ptr = nil
		return nil
	}
	return plugin.ErrNotHandled
}

var bytesType = reflect.TypeOf([]byte(nil))

// BytesPlugin coerces string and []byte into *[]byte. []byte sources are
// copied defensively so the destination never aliases the caller's buffer.
var BytesPlugin = plugin.NewPlugin("bytes", plugin.FactoryFunc(
	func(t reflect.Type) bool { return t == bytesType },
	func(ptr any) plugin.Adapter { return bytesAdapter{ptr: ptr.(*[]byte)} },
))
