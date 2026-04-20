# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-04-19

### Added
- Initial release.
- `Facer`, `New`, `Default`, `Strict`, package-level `Unface`.
- `Un*er` interface family: `Unfacer`, `Unstringer`, `Unbooler`, `Unnumberer`, `Unbyteser`, `Unruner`, `Unmapper`, `Unlister`, `Unniler`, `Untimer`, `Undurationer`.
- Rich source abstractions: `Number`, `List`, `Map`.
- Plugin system: `Plugin`, `AdapterFactory`, `NewPlugin`, `FactoryFunc`, `Compose`.
- Options: `With`, `Without`, `WithoutNamed`, `Only`, `WithFieldMatch`, `OnUnknown`, `WithTagFallback`, `WithoutTagFallback`, `OnSoftError`, `WithPointerResolve`.
- Pointer resolution modes: `PointerResolveFlat` (default), `PointerResolveNone`, `PointerResolveDeep`.
- Error sentinels and `Error` path-tracing type.
- Built-in plugins: String, Bool, Bytes, Rune, Int8..Int64, Int, Uint8..Uint64, Uint, Float32/64, Complex64/128, BigInt, BigFloat, Time, JSONRaw, Pointer, List, Map, Struct.
- Composite plugins: IntPluginBundle, UintPluginBundle, FloatPluginBundle, ComplexPluginBundle, NumberPlugin, PrimitivesPlugin, StandardPlugin.
- Struct walker with tag grammar (required, alias, remainder, strict, inline, skip, match), YAML/JSON tag fallback, configurable match modes and unknown-key policy.
