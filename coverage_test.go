package unface_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/schneid-l/unface"
	"github.com/schneid-l/unface/engine"
	"github.com/schneid-l/unface/plugin"
	"github.com/schneid-l/unface/unfacers"
)

// --- Facer.Config() accessor -----------------------------------------------

func TestFacerConfigAccessor(t *testing.T) {
	f := engine.New(engine.With(unfacers.StringPlugin))
	cfg := f.Config()
	if cfg == nil {
		t.Fatal("Config() must return non-nil")
	}
	if len(cfg.Plugins) == 0 {
		t.Fatal("Config should surface configured plugins")
	}
}

// --- PointerPlugin via unface.Unface (exercises Unface + WithConfig) -------

func TestPointerPluginViaDefault(t *testing.T) {
	var p *int
	if err := unface.Unface(42, &p); err != nil {
		t.Fatal(err)
	}
	if p == nil || *p != 42 {
		t.Fatalf("p=%v", p)
	}
}

// Under PointerResolveNone the dispatcher doesn't flatten, so the
// PointerPlugin's Unface + WithConfig paths run directly.
func TestPointerPluginNoneModeExercisesAdapter(t *testing.T) {
	var p *int
	f := engine.New(
		engine.With(unfacers.PointerPlugin, unfacers.IntPlugin),
		engine.WithPointerResolve(unface.PointerResolveNone),
	)
	if err := f.Unface(42, &p); err != nil {
		t.Fatal(err)
	}
	if p == nil || *p != 42 {
		t.Fatalf("p=%v", p)
	}
}

func TestPointerPluginNoneModeNilSrc(t *testing.T) {
	x := 10
	p := &x
	f := engine.New(
		engine.With(unfacers.PointerPlugin, unfacers.IntPlugin),
		engine.WithPointerResolve(unface.PointerResolveNone),
	)
	if err := f.Unface(nil, &p); err != nil {
		t.Fatal(err)
	}
	if p != nil {
		t.Fatalf("p=%v, want nil after nil src", p)
	}
}

func TestPointerPluginRecursiveStruct(t *testing.T) {
	type Inner struct {
		A int `unface:"a"`
	}
	type Outer struct {
		I *Inner `unface:"i"`
	}
	var o Outer
	src := map[string]any{"i": map[string]any{"a": 5}}
	if err := unface.Unface(src, &o); err != nil {
		t.Fatal(err)
	}
	if o.I == nil || o.I.A != 5 {
		t.Fatalf("o=%+v", o.I)
	}
}

// --- Int plugin Unnumber uint64 branch -------------------------------------

func TestInt64PluginFromUint64InRange(t *testing.T) {
	var x int64
	if err := unface.Unface(uint64(100), &x); err != nil {
		t.
			Fatal(err)
	}
	if x != 100 {
		t.Fatalf("x=%d", x)
	}
}

// Forces the Int64-fails/Uint64-succeeds branch: a uint64 larger than
// MaxInt64 must overflow *int64.
func TestInt64PluginFromUint64OverflowsSignedMax(t *testing.T) {
	var x int64
	const oversized = uint64(1) << 63 // MaxInt64 + 1
	err := unface.Unface(oversized, &x)
	if err == nil {
		t.Fatal("expected overflow error")
	}
}

// --- Unsigned adapter: successful conversion path -------------------------

func TestUint32PluginFromPositiveInt(t *testing.T) {
	var x uint32
	if err := unface.Unface(int(42), &x); err != nil {
		t.Fatal(err)
	}
	if x != 42 {
		t.Fatalf("x=%d", x)
	}
}

// --- Uint narrowing overflow: uint64 → uint8 --------------------------------

func TestUint8PluginFromUint64Overflow(t *testing.T) {
	var x uint8
	err := unface.Unface(uint64(500), &x)
	if err == nil {
		t.Fatal("expected overflow")
	}
}

// --- Big{Int,Float} uncovered complex-source branches ---------------------

func TestBigIntFromComplexWithImagFails(t *testing.T) {
	var x big.Int
	err := unface.Unface(complex(1, 2), &x)
	if err == nil {
		t.Fatal("complex with imag should not convert to big.Int")
	}
}

// --- Map convertKey: string-to-int key conversion --------------------------

func TestMapGetWithConvertibleKey(t *testing.T) {
	raw := map[int]string{1: "a"}
	m, ok := plugin.MapOf(raw)
	if !ok {
		t.Fatal("MapOf")
	}
	// int64(1) should convert to int key.
	v, ok := m.Get(int64(1))
	if !ok || v != "a" {
		t.Fatalf("Get=%v ok=%v", v, ok)
	}
}

// --- applyUnknownPolicy: UnknownWarn without handler is a no-op -----------

func TestStructUnknownWarnNoHandlerSilentSuccess(t *testing.T) {
	type T struct {
		A int `unface:"a"`
	}
	var v T
	// No handler passed; must not error.
	err := unface.Unface(
		map[string]any{"a": 1, "b": 2},
		&v,
		unface.OnUnknown(unface.UnknownWarn),
	)
	if err != nil {
		t.Fatal(err)
	}
}

// --- Unface aliases resolve to the right implementation ------------------

func TestRootAliasesPointToSubpackage(t *testing.T) {
	// Smoke test: the alias references the real symbol.
	if !errors.Is(unface.ErrNotHandled, plugin.ErrNotHandled) {
		t.Fatal("ErrNotHandled alias drift")
	}
	if unface.StandardPlugin == nil {
		t.Fatal("StandardPlugin alias is nil")
	}
	if unface.StandardPlugin.Name() != "standard" {
		t.Fatalf("StandardPlugin name=%q", unface.StandardPlugin.Name())
	}
}

// --- concurrent per-call options ------------------------------------------

func TestPerCallOptionClonePreservesInstance(t *testing.T) {
	f := unface.New(unface.With(unfacers.StandardPlugin))
	// One call with a per-call option that drops a plugin.
	var i int
	errBad := f.Unface("42", &i, unface.Without(unfacers.NumberPlugin))
	if errBad == nil {
		t.Fatal("expected failure with NumberPlugin removed")
	}
	// Subsequent call on the same facer should still succeed — the per-call
	// Option must not have mutated the instance's plugin set.
	var j int
	if err := f.Unface("42", &j); err != nil {
		t.Fatalf("instance mutated: %v", err)
	}
	if j != 42 {
		t.Fatalf("j=%d", j)
	}
}
