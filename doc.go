// Package unface transforms values of arbitrary type (any) into a
// destination value of arbitrary type. Its primary use cases are
// configuration loading, HTTP payload binding, and ingestion of generic
// map trees into typed Go structs.
//
// # Layered architecture
//
// The unface module is organised into four layers. Most users only need to
// import the root:
//
//   - unface           — the friendly public API (Unface + aliases).
//   - unface/engine    — the dispatcher ("under the hood"): Facer, options.
//   - unface/plugin    — the plugin contract: Plugin, Un*er interfaces,
//     Number/List/Map abstractions, errors.
//   - unface/unfacers  — the catalog of built-in plugins + StandardPlugin.
//
// The root package re-exports the most common symbols from engine, plugin,
// and unfacers so the 80 % use case is a single import.
//
// # Quick start
//
//	type Server struct {
//	    Host string `unface:"host"`
//	    Port int    `unface:"port,required"`
//	}
//
//	raw := map[string]any{"host": "localhost", "port": "8080"}
//	var s Server
//	if err := unface.Unface(raw, &s); err != nil {
//	    log.Fatal(err)
//	}
//	// s.Host == "localhost"; s.Port == 8080 (parsed from the string).
//
// # Interfaces
//
// A destination type can opt into custom coercion by implementing any of
// the Un*er interfaces. The library prefers Unfacer (a master switch) when
// present, then falls through to the source-specific interfaces
// (Unstringer, Unbooler, Unnumberer, Unmapper, Unlister, Untimer, etc.),
// and finally to the registered plugins.
//
// # Plugins
//
// Plugins provide default coercions for the built-in Go types. They are
// composed into a Facer via the With option. The package-level Default
// Facer is preloaded with StandardPlugin (every built-in plugin); Strict
// has no plugins loaded at all.
//
//	f := unface.New(unface.With(unface.StringPlugin, unface.IntPlugin))
//	f.Unface(42, &someStringVar)
//
// # Pointer resolution
//
// WithPointerResolve controls how multi-level pointers are handled on both
// src and dest. The default PointerResolveFlat flattens to the innermost
// element once. PointerResolveDeep tries each pointer level, and
// PointerResolveNone performs no dereferencing.
package unface
