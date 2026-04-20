package engine

import "github.com/schneid-l/unface/plugin"

// Option configures a Facer at construction time or a single Unface call.
type Option func(*plugin.Config)

// With appends plugins to the current set. Composites are flattened to
// their atomic leaves on insertion so Without(composite) removes children
// correctly even across nested composition.
func With(plugins ...plugin.Plugin) Option {
	return func(c *plugin.Config) {
		for _, p := range plugins {
			c.Plugins = append(c.Plugins, plugin.Flatten(p)...)
		}
	}
}

// Without removes plugins from the current set by identity. Composites are
// expanded: removing a composite also removes all of its named children.
func Without(plugins ...plugin.Plugin) Option {
	names := make(map[string]struct{}, len(plugins))
	for _, p := range plugins {
		names[p.Name()] = struct{}{}
		for _, child := range p.ChildNames() {
			names[child] = struct{}{}
		}
	}
	return func(c *plugin.Config) {
		c.Plugins = filterPlugins(c.Plugins, names)
	}
}

// WithoutNamed removes plugins by name. Matches both top-level plugins and
// named children exposed by composites.
func WithoutNamed(names ...string) Option {
	set := make(map[string]struct{}, len(names))
	for _, n := range names {
		set[n] = struct{}{}
	}
	return func(c *plugin.Config) {
		c.Plugins = filterPlugins(c.Plugins, set)
	}
}

// Only replaces the current set with the given plugins.
func Only(plugins ...plugin.Plugin) Option {
	return func(c *plugin.Config) {
		c.Plugins = append(c.Plugins[:0:0], plugins...)
	}
}

// WithFieldMatch sets the struct-field matching mode.
func WithFieldMatch(m plugin.MatchMode) Option {
	return func(c *plugin.Config) { c.Match = m }
}

// OnUnknown sets the unknown-key policy and, for UnknownWarn, its handler.
func OnUnknown(p plugin.UnknownPolicy, handler ...plugin.UnknownHandler) Option {
	return func(c *plugin.Config) {
		c.OnUnknown = p
		if len(handler) > 0 {
			c.UnknownHandler = handler[0]
		}
	}
}

// WithTagFallback sets the struct-tag fallback order (default:
// ["unface","yaml","json"]).
func WithTagFallback(tags ...string) Option {
	return func(c *plugin.Config) {
		c.TagFallback = append(c.TagFallback[:0:0], tags...)
	}
}

// WithoutTagFallback disables the yaml/json tag fallback; only the "unface"
// tag is read.
func WithoutTagFallback() Option {
	return func(c *plugin.Config) { c.TagFallback = []string{"unface"} }
}

// OnSoftError installs an observer invoked when a dispatcher step returns a
// soft error.
func OnSoftError(h plugin.SoftErrorHandler) Option {
	return func(c *plugin.Config) { c.SoftHandler = h }
}

// WithPointerResolve sets the pointer-resolution mode (default Flat).
func WithPointerResolve(mode plugin.PointerResolution) Option {
	return func(c *plugin.Config) { c.PointerMode = mode }
}

func filterPlugins(in []plugin.Plugin, drop map[string]struct{}) []plugin.Plugin {
	out := in[:0:0]
	for _, p := range in {
		if _, skip := drop[p.Name()]; skip {
			continue
		}
		out = append(out, p)
	}
	return out
}
