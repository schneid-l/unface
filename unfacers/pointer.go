package unfacers

import (
	"reflect"

	"github.com/schneid-l/unface/engine"
	"github.com/schneid-l/unface/plugin"
)

type pointerAdapter struct {
	rv  reflect.Value
	cfg *plugin.Config
}

func (a pointerAdapter) WithConfig(cfg *plugin.Config) plugin.Unfacer {
	return pointerAdapter{rv: a.rv, cfg: cfg}
}

func (a pointerAdapter) Unface(src any) error {
	if src == nil {
		a.rv.Set(reflect.Zero(a.rv.Type()))
		return nil
	}
	elem := reflect.New(a.rv.Type().Elem())
	cfg := a.cfg
	if cfg == nil {
		// Fallback path when called outside the dispatcher (e.g. raw adapter
		// tests). Use Default's cfg if available, otherwise a fresh default.
		if Default != nil {
			cfg = Default.Config()
		} else {
			cfg = plugin.NewDefaultConfig()
		}
	}
	f := engine.New()
	if err := f.Dispatch(cfg, src, elem.Interface()); err != nil {
		return err
	}
	a.rv.Set(elem)
	return nil
}

// PointerPlugin auto-allocates nil pointer destinations and forwards the
// coercion to the element type, using the Facer's active plugin set.
var PointerPlugin = plugin.NewPlugin("pointer", plugin.FactoryFunc(
	func(t reflect.Type) bool { return t.Kind() == reflect.Pointer },
	func(ptr any) plugin.Adapter {
		return pointerAdapter{rv: reflect.ValueOf(ptr).Elem()}
	},
))
