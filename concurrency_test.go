package unface_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/schneid-l/unface"
)

// TestFacerConcurrentUnface exercises Facer.Unface from many goroutines
// simultaneously. Shared Facer + independent dests must remain correct
// under -race.
func TestFacerConcurrentUnface(t *testing.T) {
	type Config struct {
		Name string `unface:"name"`
		Port int    `unface:"port"`
	}

	f := unface.New(unface.With(unface.StandardPlugin))
	const goroutines = 64
	const iterations = 50

	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	for g := range goroutines {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := range iterations {
				src := map[string]any{
					"name": fmt.Sprintf("svc-%d-%d", g, i),
					"port": fmt.Sprintf("%d", 8000+i),
				}
				var cfg Config
				if err := f.Unface(src, &cfg); err != nil {
					errs <- err
					return
				}
				if cfg.Port != 8000+i {
					errs <- fmt.Errorf("g=%d i=%d port=%d", g, i, cfg.Port)
					return
				}
			}
		}(g)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

// TestPackageLevelUnfaceConcurrent verifies the package-level Unface is
// safe when called from many goroutines.
func TestPackageLevelUnfaceConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var i int
			if err := unface.Unface("42", &i); err != nil || i != 42 {
				t.Errorf("i=%d err=%v", i, err)
			}
		}()
	}
	wg.Wait()
}

// TestPerCallOptionsDoNotAffectInstance confirms that per-call options
// clone the config and don't leak into the instance's state.
func TestPerCallOptionsDoNotAffectInstance(t *testing.T) {
	f := unface.New(unface.With(unface.StandardPlugin))

	// Per-call: remove NumberPlugin — should fail this one call.
	var i int
	if err := f.Unface("42", &i, unface.Without(unface.NumberPlugin)); err == nil {
		t.Fatal("expected failure when NumberPlugin is removed for this call")
	}

	// Following call on the same instance should still succeed.
	var j int
	if err := f.Unface("42", &j); err != nil || j != 42 {
		t.Fatalf("j=%d err=%v — instance mutated by per-call Without", j, err)
	}
}

// TestStrictInstanceHasNoPlugins confirms the exported Strict facer
// refuses coercion that needs defaults.
func TestStrictInstanceHasNoPlugins(t *testing.T) {
	var i int
	err := unface.Strict.Unface("42", &i)
	if err == nil {
		t.Fatal("Strict must not coerce string→int")
	}
}
