package unfacers

import (
	"fmt"
	"reflect"

	"github.com/schneid-l/unface/engine"
	"github.com/schneid-l/unface/plugin"
)

type mapAdapter struct {
	rv  reflect.Value // settable map value
	cfg *plugin.Config
}

func (a mapAdapter) WithConfig(cfg *plugin.Config) plugin.Unfacer {
	return mapAdapter{rv: a.rv, cfg: cfg}
}

func (a mapAdapter) Unface(src any) error {
	if src == nil {
		a.rv.Set(reflect.Zero(a.rv.Type()))
		return nil
	}
	m, ok := plugin.MapOf(src)
	if !ok {
		return fmt.Errorf("unface: cannot coerce %T to %s", src, a.rv.Type())
	}
	cfg := a.cfg
	if cfg == nil {
		if Default != nil {
			cfg = Default.Config()
		} else {
			cfg = plugin.NewDefaultConfig()
		}
	}
	f := engine.New()

	out := reflect.MakeMapWithSize(a.rv.Type(), m.Len())
	keyT := a.rv.Type().Key()
	valT := a.rv.Type().Elem()
	var walkErr error
	m.Iter(func(k, v any) bool {
		kp := reflect.New(keyT)
		if err := f.Dispatch(cfg, k, kp.Interface()); err != nil {
			walkErr = fmt.Errorf("unface: map key %v: %w", k, err)
			return false
		}
		vp := reflect.New(valT)
		if err := f.Dispatch(cfg, v, vp.Interface()); err != nil {
			walkErr = fmt.Errorf("unface: map value for %v: %w", k, err)
			return false
		}
		out.SetMapIndex(kp.Elem(), vp.Elem())
		return true
	})
	if walkErr != nil {
		return walkErr
	}
	a.rv.Set(out)
	return nil
}

// MapPlugin coerces a map src into a typed map dest, recursively unfacing
// each key and value. It does not handle struct destinations; use
// StructPlugin for those.
var MapPlugin = plugin.NewPlugin("map", plugin.FactoryFunc(
	func(t reflect.Type) bool { return t.Kind() == reflect.Map },
	func(ptr any) plugin.Adapter {
		return mapAdapter{rv: reflect.ValueOf(ptr).Elem()}
	},
))
