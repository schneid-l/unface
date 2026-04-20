package plugin

import "reflect"

// Adapter is the bridge between a destination and the Un*er family. An
// adapter wraps a destination pointer and itself implements Unfacer (at
// minimum); optionally it implements any subset of the Un*er interfaces for
// direct dispatch.
type Adapter = Unfacer

// AdapterFactory builds an Adapter for a compatible destination type.
type AdapterFactory interface {
	// Matches reports whether this factory can produce an adapter for
	// destType. destType is the type of *dest (the pointed-to element), not
	// *destType.
	Matches(destType reflect.Type) bool
	// For constructs an Adapter wrapping destPtr. destPtr is the raw pointer
	// provided to Unface; factories downcast to their expected concrete
	// pointer type.
	For(destPtr any) Adapter
}

// Plugin is a named bag of AdapterFactorys. Plugins are composed into a
// Facer via the With option.
type Plugin interface {
	Name() string
	Factories() []AdapterFactory
	// ChildNames returns the names of sub-plugins when this is a composite.
	// Atomic plugins return nil.
	ChildNames() []string
	// Children returns direct sub-plugins when this is a composite, nil
	// otherwise. Used by With to expand composites into their atomic leaves.
	Children() []Plugin
}

// NewPlugin returns an atomic Plugin with the given name and factories.
func NewPlugin(name string, factories ...AdapterFactory) Plugin {
	return atomicPlugin{name: name, factories: factories}
}

// FactoryFunc builds an AdapterFactory from two closures.
func FactoryFunc(
	matches func(reflect.Type) bool,
	build func(destPtr any) Adapter,
) AdapterFactory {
	return funcFactory{match: matches, build: build}
}

// Compose groups plugins under a new name. The composite exposes its
// children so With can expand to atomic leaves (and Without can drop them
// by name).
func Compose(name string, plugins ...Plugin) Plugin {
	children := make([]Plugin, len(plugins))
	copy(children, plugins)
	return compositePlugin{name: name, children: children}
}

type atomicPlugin struct {
	name      string
	factories []AdapterFactory
}

func (p atomicPlugin) Name() string                { return p.name }
func (p atomicPlugin) Factories() []AdapterFactory { return p.factories }
func (p atomicPlugin) ChildNames() []string        { return nil }
func (p atomicPlugin) Children() []Plugin          { return nil }

type compositePlugin struct {
	name     string
	children []Plugin
}

func (p compositePlugin) Name() string { return p.name }

func (p compositePlugin) Factories() []AdapterFactory {
	var out []AdapterFactory
	for _, c := range p.children {
		out = append(out, c.Factories()...)
	}
	return out
}

func (p compositePlugin) ChildNames() []string {
	names := make([]string, 0, len(p.children))
	for _, c := range p.children {
		names = append(names, c.Name())
		names = append(names, c.ChildNames()...)
	}
	return names
}

func (p compositePlugin) Children() []Plugin { return p.children }

// Flatten recursively expands composites into their atomic leaves.
// Atomic plugins return themselves as a single-element slice.
func Flatten(p Plugin) []Plugin {
	children := p.Children()
	if len(children) == 0 {
		return []Plugin{p}
	}
	out := make([]Plugin, 0, len(children))
	for _, c := range children {
		out = append(out, Flatten(c)...)
	}
	return out
}

type funcFactory struct {
	match func(reflect.Type) bool
	build func(any) Adapter
}

func (f funcFactory) Matches(t reflect.Type) bool { return f.match(t) }
func (f funcFactory) For(ptr any) Adapter         { return f.build(ptr) }
