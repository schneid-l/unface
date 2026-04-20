package plugin

// Config carries the effective per-call configuration for a Facer dispatch.
// It is defined in the plugin package (rather than engine) so that adapters
// can implement CfgAware without creating an import cycle between the
// engine and the adapter packages.
//
// Fields are exported to make them visible to adapters (in particular the
// recursive plugins: pointer, list, map, struct), but user code typically
// manipulates a Config via the option builders exposed from the engine
// package rather than touching fields directly.
type Config struct {
	Plugins        []Plugin
	Match          MatchMode
	OnUnknown      UnknownPolicy
	UnknownHandler UnknownHandler
	TagFallback    []string
	SoftHandler    SoftErrorHandler
	PointerMode    PointerResolution
}

// NewDefaultConfig returns a Config preloaded with sensible defaults:
// MatchFold, UnknownIgnore, tag fallback ["unface","yaml","json"], and
// PointerResolveFlat.
func NewDefaultConfig() *Config {
	return &Config{
		Match:       MatchFold,
		OnUnknown:   UnknownIgnore,
		TagFallback: []string{"unface", "yaml", "json"},
		PointerMode: PointerResolveFlat,
	}
}

// Clone returns a deep-ish copy of the Config, safe to mutate in an option
// callback without disturbing the original.
func (c *Config) Clone() *Config {
	out := *c
	out.Plugins = append([]Plugin(nil), c.Plugins...)
	out.TagFallback = append([]string(nil), c.TagFallback...)
	return &out
}

// CfgAware is the dispatch-layer contract between the Facer and its
// built-in adapters. An adapter opts into receiving the effective per-call
// Config by implementing CfgAware. This is primarily needed by the struct
// walker and the recursive plugins (PointerPlugin, ListPlugin, MapPlugin)
// so they can dispatch children using the same plugin set + options the
// caller passed.
//
// User-authored plugins that need access to the config can either accept
// it via closure at construction time or implement this interface.
type CfgAware interface {
	WithConfig(*Config) Unfacer
}
