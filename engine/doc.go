// Package engine hosts the Facer dispatcher and the option builders that
// configure it. The Facer is what actually walks the src/dest pair through
// the Un*er fast path and then into the registered plugin set.
//
// Most callers will use the package-level unface.Unface entry point or the
// unface.New constructor (which are thin re-exports of this package). The
// engine package is factored out so that plugin authors and codegen can
// depend on dispatch behavior without pulling in the entire root package's
// built-in plugin set.
package engine
