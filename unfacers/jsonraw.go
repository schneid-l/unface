package unfacers

import (
	"encoding/json"
	"reflect"

	"github.com/schneid-l/unface/plugin"
)

type jsonRawAdapter struct{ ptr *json.RawMessage }

func (a jsonRawAdapter) Unface(src any) error {
	if src == nil {
		*a.ptr = json.RawMessage("null")
		return nil
	}
	if b, ok := src.(json.RawMessage); ok {
		*a.ptr = append(json.RawMessage(nil), b...)
		return nil
	}
	if b, ok := src.([]byte); ok {
		*a.ptr = append(json.RawMessage(nil), b...)
		return nil
	}
	buf, err := json.Marshal(src)
	if err != nil {
		return err
	}
	*a.ptr = buf
	return nil
}

var jsonRawType = reflect.TypeOf(json.RawMessage{})

// JSONRawPlugin re-encodes any src as JSON and stores it in json.RawMessage.
var JSONRawPlugin = plugin.NewPlugin("jsonraw", plugin.FactoryFunc(
	func(t reflect.Type) bool { return t == jsonRawType },
	func(ptr any) plugin.Adapter { return jsonRawAdapter{ptr: ptr.(*json.RawMessage)} },
))
