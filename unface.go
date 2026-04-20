package unface

import (
	"github.com/schneid-l/unface/engine"
	"github.com/schneid-l/unface/unfacers"
)

// Unface coerces src into dest using the default Facer (preloaded with
// unfacers.StandardPlugin). For custom plugin sets or dedicated instances,
// use unface.New with option builders.
func Unface(src, dest any, opts ...engine.Option) error {
	return unfacers.Default.Unface(src, dest, opts...)
}
