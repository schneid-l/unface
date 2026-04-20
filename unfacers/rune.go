package unfacers

import (
	"fmt"
	"reflect"
	"unicode/utf8"

	"github.com/schneid-l/unface/plugin"
)

type runeAdapter struct{ ptr *rune }

func (a runeAdapter) Unface(src any) error {
	switch v := src.(type) {
	case rune:
		*a.ptr = v
		return nil
	case string:
		return a.parseString(v)
	}
	if n, ok := plugin.NumberOf(src); ok {
		if i, ok := n.Int64(); ok {
			*a.ptr = rune(i)
			return nil
		}
	}
	return plugin.ErrNotHandled
}

func (a runeAdapter) parseString(s string) error {
	if len(s) == 0 {
		return fmt.Errorf("unface/rune: empty string")
	}
	r, size := utf8.DecodeRuneInString(s)
	if size != len(s) {
		return fmt.Errorf("unface/rune: string %q contains more than one rune", s)
	}
	if r == utf8.RuneError {
		return fmt.Errorf("unface/rune: invalid utf-8 in %q", s)
	}
	*a.ptr = r
	return nil
}

var runeType = reflect.TypeOf(rune(0))

// RunePlugin coerces rune, single-rune strings, and integer numbers into
// *rune.
var RunePlugin = plugin.NewPlugin("rune", plugin.FactoryFunc(
	func(t reflect.Type) bool { return t == runeType },
	func(ptr any) plugin.Adapter { return runeAdapter{ptr: ptr.(*rune)} },
))
