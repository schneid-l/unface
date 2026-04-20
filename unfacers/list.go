package unfacers

import (
	"fmt"
	"reflect"

	"github.com/schneid-l/unface/engine"
	"github.com/schneid-l/unface/plugin"
)

type listAdapter struct {
	rv  reflect.Value // settable slice value
	cfg *plugin.Config
}

func (a listAdapter) WithConfig(cfg *plugin.Config) plugin.Unfacer {
	return listAdapter{rv: a.rv, cfg: cfg}
}

func (a listAdapter) Unface(src any) error {
	if src == nil {
		a.rv.Set(reflect.Zero(a.rv.Type()))
		return nil
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

	l, ok := plugin.ListOf(src)
	if !ok {
		// Scalar-to-singleton promotion.
		out := reflect.MakeSlice(a.rv.Type(), 1, 1)
		if err := f.Dispatch(cfg, src, out.Index(0).Addr().Interface()); err != nil {
			return err
		}
		a.rv.Set(out)
		return nil
	}
	n := l.Len()
	out := reflect.MakeSlice(a.rv.Type(), n, n)
	for i := range n {
		v := l.At(i)
		if err := f.Dispatch(cfg, v, out.Index(i).Addr().Interface()); err != nil {
			return fmt.Errorf("unface: list[%d]: %w", i, err)
		}
	}
	a.rv.Set(out)
	return nil
}

// ListPlugin coerces any list-like src into a []T dest, recursively
// unfacing each element. []byte is intentionally excluded (handled by
// BytesPlugin).
var ListPlugin = plugin.NewPlugin("list", plugin.FactoryFunc(
	func(t reflect.Type) bool {
		return t.Kind() == reflect.Slice && t != bytesType
	},
	func(ptr any) plugin.Adapter {
		return listAdapter{rv: reflect.ValueOf(ptr).Elem()}
	},
))
