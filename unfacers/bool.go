package unfacers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/schneid-l/unface/plugin"
)

type boolAdapter struct{ ptr *bool }

func (a boolAdapter) Unface(src any) error {
	switch v := src.(type) {
	case bool:
		*a.ptr = v
		return nil
	case string:
		return a.parseString(v)
	}
	if n, ok := plugin.NumberOf(src); ok {
		return a.parseNumber(n)
	}
	return plugin.ErrNotHandled
}

func (a boolAdapter) parseString(s string) error {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "yes", "y", "on", "1", "enabled":
		*a.ptr = true
		return nil
	case "false", "no", "n", "off", "0", "disabled":
		*a.ptr = false
		return nil
	}
	return fmt.Errorf("unface/bool: cannot parse %q as bool", s)
}

func (a boolAdapter) parseNumber(n plugin.Number) error {
	if v, ok := n.Int64(); ok {
		*a.ptr = v != 0
		return nil
	}
	if v, ok := n.Float64(); ok {
		*a.ptr = v != 0
		return nil
	}
	return plugin.ErrNotHandled
}

var boolType = reflect.TypeOf(false)

// BoolPlugin coerces bool-like strings ("true","yes","on","1","enabled",
// and their negations) and numbers into *bool. Matches the exact bool type
// only; named types with underlying bool are not auto-handled (register a
// custom plugin or implement Unbooler on the named type).
var BoolPlugin = plugin.NewPlugin("bool", plugin.FactoryFunc(
	func(t reflect.Type) bool { return t == boolType },
	func(ptr any) plugin.Adapter { return boolAdapter{ptr: ptr.(*bool)} },
))
