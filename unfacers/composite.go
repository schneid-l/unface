package unfacers

import (
	"github.com/schneid-l/unface/engine"
	"github.com/schneid-l/unface/plugin"
)

// Composite plugins group atomic plugins into named bundles. Removing a
// composite (via Without or WithoutNamed) recursively drops its children.
var (
	// IntPluginBundle bundles every signed-integer atomic plugin.
	IntPluginBundle = plugin.Compose("int.bundle",
		Int8Plugin, Int16Plugin, Int32Plugin, Int64Plugin, IntPlugin)
	// UintPluginBundle bundles every unsigned-integer atomic plugin.
	UintPluginBundle = plugin.Compose("uint.bundle",
		Uint8Plugin, Uint16Plugin, Uint32Plugin, Uint64Plugin, UintPlugin)
	// FloatPluginBundle bundles the float plugins.
	FloatPluginBundle = plugin.Compose("float.bundle", Float32Plugin, Float64Plugin)
	// ComplexPluginBundle bundles the complex plugins.
	ComplexPluginBundle = plugin.Compose("complex.bundle", Complex64Plugin, Complex128Plugin)

	// NumberPlugin bundles every numeric scalar plus big.Int/big.Float.
	NumberPlugin = plugin.Compose("number",
		IntPluginBundle, UintPluginBundle, FloatPluginBundle, ComplexPluginBundle,
		BigIntPlugin, BigFloatPlugin,
	)

	// PrimitivesPlugin bundles every scalar coercion.
	PrimitivesPlugin = plugin.Compose("primitives",
		NumberPlugin, BoolPlugin, StringPlugin, BytesPlugin, RunePlugin,
	)

	// StandardPlugin is the out-of-the-box bundle: everything the library
	// provides. Used by the package-level Default Facer.
	StandardPlugin = plugin.Compose("standard",
		PrimitivesPlugin,
		TimePlugin,
		ListPlugin,
		MapPlugin,
		StructPlugin,
		PointerPlugin,
		JSONRawPlugin,
	)
)

// Default is a Facer preloaded with StandardPlugin. It is initialised in
// init() to avoid a package-initialisation cycle between the plugin vars
// above and the Facer construction.
var Default *engine.Facer

// init wires the package-level Default Facer with StandardPlugin. Ordering
// is guaranteed by Go's package-level dependency analysis: every plugin var
// referenced above is declared at package scope and resolved before init().
func init() {
	Default = engine.New(engine.With(StandardPlugin))
}
