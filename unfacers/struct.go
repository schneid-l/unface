package unfacers

import (
	"reflect"

	"github.com/schneid-l/unface/engine"
	"github.com/schneid-l/unface/internal/walker"
	"github.com/schneid-l/unface/plugin"
)

type structAdapter struct {
	rv  reflect.Value
	cfg *plugin.Config
}

func (a structAdapter) WithConfig(cfg *plugin.Config) plugin.Unfacer {
	return structAdapter{rv: a.rv, cfg: cfg}
}

func (a structAdapter) Unface(src any) error {
	cfg := a.cfg
	if cfg == nil {
		if Default != nil {
			cfg = Default.Config()
		} else {
			cfg = plugin.NewDefaultConfig()
		}
	}
	// Reusable engines: the dispatcher itself is stateless across calls
	// once Dispatch(cfg, ...) is invoked with the per-call config. Strict
	// has no plugins loaded.
	f := engine.New()
	strict := plugin.NewDefaultConfig()
	return walker.StructWalk(src, a.rv, cfg,
		func(s, d any) error { return f.Dispatch(cfg, s, d) },
		func(s, d any) error { return f.Dispatch(strict, s, d) },
	)
}

// StructPlugin walks a map-like src into a struct dest, honoring the tag
// grammar and instance/struct-level match/unknown policy.
var StructPlugin = plugin.NewPlugin("struct", plugin.FactoryFunc(
	func(t reflect.Type) bool { return t.Kind() == reflect.Struct },
	func(ptr any) plugin.Adapter {
		return structAdapter{rv: reflect.ValueOf(ptr).Elem()}
	},
))
