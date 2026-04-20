# Dispatch order

`unface.Unface(src, dest, opts...)` is a pipeline. Every call runs through the same sequence until something produces a result. This page describes that sequence exhaustively — it's the mental model you need to reason about why a particular coercion did or didn't fire.

## The steps

For every `Unface` call:

```
 1. Validate dest is a non-nil pointer            → else ErrInvalidDest
 2. Apply pointer-resolution mode                 (Flat / None / Deep)
 3. Check dest implements Unfacer                 → call it
 4. assignDirect: src is exactly the dest's type  → set and return
 5. Un<srcKind>er on dest                          (Unstringer, Unbooler, …)
 6. Plugin fallback (AdapterFactory matching)     (StringPlugin, …)
 7. ErrNoCoercion                                  (nothing handled it)
```

Steps 3–6 form one "attempt". Under `PointerResolveDeep` the dispatcher runs one attempt per (dest-layer, src-layer) combination and returns on the first success; under the other modes it runs a single attempt on the flattened (or as-passed) pair.

## Step 1 — dest validation

`dest` must be a **non-nil pointer**. Failure returns `ErrInvalidDest`.

```go
unface.Unface(42, nil)         // ErrInvalidDest
unface.Unface(42, (*int)(nil)) // ErrInvalidDest
var i int
unface.Unface(42, &i)          // ok
```

## Step 2 — pointer resolution

Controlled by `WithPointerResolve(mode)`:

| Mode | Behavior |
|---|---|
| `PointerResolveFlat` (default) | Before the attempt, both src and dest are flattened to their innermost non-pointer layer. Nil intermediates on dest are allocated along the way. A single attempt runs on the flattened pair. |
| `PointerResolveNone` | No dereferencing. Dispatch runs on the caller-supplied depth. A method declared on `*T` will not fire for a `**T` dest. Useful when you want strict exact-depth binding. |
| `PointerResolveDeep` | Tries every pointer level on both sides. Dest outer→inner; per dest level, src outer→inner. First non-soft outcome wins. Ideal for polymorphic dispatch where a type can handle sources at varied wrapping. |

### Illustration

Given `dest = ****T` where `*T` implements `Unfacer`:

- **Flat** → flattens to innermost `*T`, allocating the three intermediate pointers as needed. The `Unfacer` method fires.
- **None** → looks up `Unfacer` on `****T`, fails. No plugin matches. `ErrNoCoercion`.
- **Deep** → tries `****T` (miss), `***T` (miss), `**T` (miss), `*T` (hit — method fires).

## Step 3 — `Unfacer` on dest

If `dest` implements `Unface(src any) error`, it is called first. This is the **codegen fast path**: types that implement `Unfacer` skip reflection entirely.

```go
type URL struct { … }
func (u *URL) Unface(src any) error { /* hand-written OR unfacegen-emitted */ }
```

Return `ErrNotHandled` (or any error wrapping it) to let the pipeline continue. Any other error aborts.

## Step 4 — assignDirect

If `src` is exactly the pointed-to type of `dest`, the value is assigned directly by reflection. This is the fast path for identical types.

**Exception: slice / map / chan destinations** are deliberately *not* handled by `assignDirect` — direct assignment would alias the caller's underlying array/buckets/channel. Plugins (`BytesPlugin`, `ListPlugin`, `MapPlugin`) always deep-copy, guaranteeing the dest is independent after the call.

**Interface dest (*any, *MyIface)** is handled if the src value implements the interface. This lets `map[string]any` → struct fields typed `any` work transparently.

## Step 5 — `Un<srcKind>er` on dest

The dispatcher picks the interface based on `src`'s concrete kind and calls the corresponding method if `dest` implements it:

| src kind | interface |
|---|---|
| `bool` | `Unbooler.Unbool(bool) error` |
| `string` | `Unstringer.Unstring(string) error` |
| `[]byte` | `Unbyteser.Unbytes([]byte) error` (falls back to `Unstringer` if absent) |
| `time.Time` | `Untimer.Untime(time.Time) error` |
| `time.Duration` | `Undurationer.Unduration(time.Duration) error` |
| any numeric | `Unnumberer.Unnumber(Number) error` |
| slice / array | `Unlister.Unlist(List) error` |
| map | `Unmapper.Unmap(Map) error` |
| `nil` | `Unniler.Unnil() error` |

`Number`, `List`, `Map` are rich abstractions — see their godoc for typed accessors.

Returning `ErrNotHandled` from any of these sends the pipeline to step 6. Any other error aborts.

## Step 6 — plugin fallback

The dispatcher iterates the current plugin set in insertion order. For each `AdapterFactory` whose `Matches(destType)` is true, it builds an adapter and calls `adapter.Unface(src)`.

- A plugin may itself implement `Un*er` interfaces on its adapter; those are consulted the same way as in step 5.
- If an adapter returns `ErrNotHandled`, the dispatcher tries the next matching factory.
- The first hard error aborts the pipeline.

See [plugins.md](./plugins.md) for how to write one.

## Step 7 — terminal

If no step above returned nil or a hard error, the pipeline returns `ErrNoCoercion`.

## Error flow

Two kinds of errors:

- **Hard error**: any non-`ErrNotHandled` `error`. Aborts immediately. Propagated with the struct-walker field path wrapped in `*unface.Error` when applicable.
- **Soft error** (`ErrNotHandled`, any error wrapping it via `fmt.Errorf("%w", unface.ErrNotHandled)`, or any error implementing `Unhandled() bool`): tells the dispatcher "I declined; keep trying." Observable via `OnSoftError(handler)`.

```go
// An Un*er method that wants to fall through:
func (u *URL) Unstring(s string) error {
    if !strings.Contains(s, "://") {
        return unface.Skip()  // let defaults try (none, since dest is URL)
    }
    …
}
```

## The Un*er family

The interfaces are defined in the root package:

```go
type Unfacer      interface{ Unface(src any) error }
type Unstringer   interface{ Unstring(s string) error }
type Unbooler     interface{ Unbool(b bool) error }
type Unnumberer   interface{ Unnumber(n Number) error }
type Unbyteser    interface{ Unbytes(b []byte) error }
type Unruner      interface{ Unrune(r rune) error }
type Unmapper     interface{ Unmap(m Map) error }
type Unlister     interface{ Unlist(l List) error }
type Unniler      interface{ Unnil() error }
type Untimer      interface{ Untime(t time.Time) error }
type Undurationer interface{ Unduration(d time.Duration) error }
```

A type can implement any subset. `Unfacer` is the master: if you implement it, the more specific interfaces are bypassed. This is how the codegen path works — `unfacegen` emits an `Unface` method that dispatches to whichever specific methods you've hand-written.
