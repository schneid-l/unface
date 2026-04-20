package unfacers

import (
	"fmt"
	"reflect"

	"github.com/schneid-l/unface/plugin"
)

type strAdapter struct{ ptr *string }

func (a strAdapter) Unface(src any) error {
	switch v := src.(type) {
	case string:
		*a.ptr = v
		return nil
	case []byte:
		*a.ptr = string(v)
		return nil
	case bool:
		*a.ptr = fmt.Sprintf("%t", v)
		return nil
	case nil:
		*a.ptr = ""
		return nil
	}
	if n, ok := plugin.NumberOf(src); ok {
		*a.ptr = n.String()
		return nil
	}
	return plugin.ErrNotHandled
}

var stringType = reflect.TypeOf("")

// StringPlugin coerces string, []byte, bool, and any numeric source into
// *string.
var StringPlugin = plugin.NewPlugin("string", plugin.FactoryFunc(
	func(t reflect.Type) bool { return t == stringType },
	func(ptr any) plugin.Adapter { return strAdapter{ptr: ptr.(*string)} },
))
