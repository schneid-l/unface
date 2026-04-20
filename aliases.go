package unface

import (
	"github.com/schneid-l/unface/engine"
	"github.com/schneid-l/unface/plugin"
	"github.com/schneid-l/unface/unfacers"
)

// This file re-exports the most common symbols from the three sub-packages
// (engine, plugin, unfacers) so that users of the unface package don't need
// to import them explicitly for the common cases. Users who need advanced
// plugin-authoring or engine control can still import the sub-packages
// directly.
//
// See the package doc for an overview of when to use each layer.

// ----- Plugin-layer types -----

//nolint:revive // aliases preserve documentation on the originating plugin package.
type (
	Adapter        = plugin.Adapter
	AdapterFactory = plugin.AdapterFactory
	Plugin         = plugin.Plugin
	Unfacer        = plugin.Unfacer
	Unstringer     = plugin.Unstringer
	Unbooler       = plugin.Unbooler
	Unbyteser      = plugin.Unbyteser
	Unruner        = plugin.Unruner
	Unniler        = plugin.Unniler
	Untimer        = plugin.Untimer
	Undurationer   = plugin.Undurationer
	Unnumberer     = plugin.Unnumberer
	Unmapper       = plugin.Unmapper
	Unlister       = plugin.Unlister
	Number         = plugin.Number
	List           = plugin.List
	Map            = plugin.Map
	Unhandled      = plugin.Unhandled
	Error          = plugin.Error

	MatchMode         = plugin.MatchMode
	UnknownPolicy     = plugin.UnknownPolicy
	UnknownHandler    = plugin.UnknownHandler
	SoftErrorHandler  = plugin.SoftErrorHandler
	PointerResolution = plugin.PointerResolution
)

// ----- Engine-layer types -----

//nolint:revive // aliases preserve documentation on the originating engine package.
type (
	Facer  = engine.Facer
	Option = engine.Option
)

// ----- Plugin-layer constants -----

// Aliases for the enum values declared in the plugin package. See
// MatchMode, UnknownPolicy, and PointerResolution for the surrounding
// context.
const (
	MatchFold          = plugin.MatchFold
	MatchInsensitive   = plugin.MatchInsensitive
	MatchExact         = plugin.MatchExact
	UnknownIgnore      = plugin.UnknownIgnore
	UnknownError       = plugin.UnknownError
	UnknownWarn        = plugin.UnknownWarn
	PointerResolveFlat = plugin.PointerResolveFlat
	PointerResolveNone = plugin.PointerResolveNone
	PointerResolveDeep = plugin.PointerResolveDeep
)

// ----- Plugin-layer functions & sentinels -----

//nolint:revive // aliases preserve documentation on the originating plugin package.
var (
	NewPlugin   = plugin.NewPlugin
	FactoryFunc = plugin.FactoryFunc
	Compose     = plugin.Compose
	NumberOf    = plugin.NumberOf
	ListOf      = plugin.ListOf
	MapOf       = plugin.MapOf
	Skip        = plugin.Skip
	IsUnhandled = plugin.IsUnhandled

	ErrNotHandled   = plugin.ErrNotHandled
	ErrInvalidDest  = plugin.ErrInvalidDest
	ErrNoCoercion   = plugin.ErrNoCoercion
	ErrSrcNil       = plugin.ErrSrcNil
	ErrUnknownField = plugin.ErrUnknownField
	ErrRequired     = plugin.ErrRequired
)

// ----- Engine-layer functions & option builders -----

//nolint:revive // aliases preserve documentation on the originating engine package.
var (
	New                = engine.New
	With               = engine.With
	Without            = engine.Without
	WithoutNamed       = engine.WithoutNamed
	Only               = engine.Only
	WithFieldMatch     = engine.WithFieldMatch
	OnUnknown          = engine.OnUnknown
	WithTagFallback    = engine.WithTagFallback
	WithoutTagFallback = engine.WithoutTagFallback
	OnSoftError        = engine.OnSoftError
	WithPointerResolve = engine.WithPointerResolve
	Strict             = engine.Strict
)

// ----- Built-in plugin catalog (from the unfacers sub-package) -----

//nolint:revive // aliases preserve documentation on the originating unfacers package.
var (
	// Default is the Facer used by the top-level Unface function, preloaded
	// with StandardPlugin (every built-in plugin).
	Default = unfacers.Default

	// Atomic scalar plugins.
	StringPlugin = unfacers.StringPlugin
	BoolPlugin   = unfacers.BoolPlugin
	BytesPlugin  = unfacers.BytesPlugin
	RunePlugin   = unfacers.RunePlugin

	// Atomic integer plugins.
	Int8Plugin   = unfacers.Int8Plugin
	Int16Plugin  = unfacers.Int16Plugin
	Int32Plugin  = unfacers.Int32Plugin
	Int64Plugin  = unfacers.Int64Plugin
	IntPlugin    = unfacers.IntPlugin
	Uint8Plugin  = unfacers.Uint8Plugin
	Uint16Plugin = unfacers.Uint16Plugin
	Uint32Plugin = unfacers.Uint32Plugin
	Uint64Plugin = unfacers.Uint64Plugin
	UintPlugin   = unfacers.UintPlugin

	// Atomic float and complex plugins.
	Float32Plugin    = unfacers.Float32Plugin
	Float64Plugin    = unfacers.Float64Plugin
	Complex64Plugin  = unfacers.Complex64Plugin
	Complex128Plugin = unfacers.Complex128Plugin

	// Arbitrary-precision plugins.
	BigIntPlugin   = unfacers.BigIntPlugin
	BigFloatPlugin = unfacers.BigFloatPlugin

	// Special-purpose plugins.
	TimePlugin    = unfacers.TimePlugin
	PointerPlugin = unfacers.PointerPlugin
	JSONRawPlugin = unfacers.JSONRawPlugin

	// Structural plugins.
	ListPlugin   = unfacers.ListPlugin
	MapPlugin    = unfacers.MapPlugin
	StructPlugin = unfacers.StructPlugin

	// Composite bundles.
	IntPluginBundle     = unfacers.IntPluginBundle
	UintPluginBundle    = unfacers.UintPluginBundle
	FloatPluginBundle   = unfacers.FloatPluginBundle
	ComplexPluginBundle = unfacers.ComplexPluginBundle
	NumberPlugin        = unfacers.NumberPlugin
	PrimitivesPlugin    = unfacers.PrimitivesPlugin
	StandardPlugin      = unfacers.StandardPlugin
)
