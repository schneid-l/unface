package unfacers

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/schneid-l/unface/plugin"
)

type bigIntAdapter struct{ ptr *big.Int }

func (a bigIntAdapter) Unface(src any) error {
	if src == nil {
		a.ptr.SetInt64(0)
		return nil
	}
	if n, ok := plugin.NumberOf(src); ok {
		b, ok := n.BigInt()
		if !ok {
			return plugin.ErrNotHandled
		}
		a.ptr.Set(b)
		return nil
	}
	if s, ok := src.(string); ok {
		if _, ok := a.ptr.SetString(s, 0); !ok {
			return fmt.Errorf("unface: cannot parse big.Int from %q", s)
		}
		return nil
	}
	return plugin.ErrNotHandled
}

type bigFloatAdapter struct{ ptr *big.Float }

func (a bigFloatAdapter) Unface(src any) error {
	if src == nil {
		a.ptr.SetFloat64(0)
		return nil
	}
	if n, ok := plugin.NumberOf(src); ok {
		b, ok := n.BigFloat()
		if !ok {
			return plugin.ErrNotHandled
		}
		a.ptr.Set(b)
		return nil
	}
	if s, ok := src.(string); ok {
		if _, ok := a.ptr.SetString(s); !ok {
			return fmt.Errorf("unface: cannot parse big.Float from %q", s)
		}
		return nil
	}
	return plugin.ErrNotHandled
}

var (
	bigIntType   = reflect.TypeOf(big.Int{})
	bigFloatType = reflect.TypeOf(big.Float{})
)

// BigIntPlugin coerces int/uint/float/complex/string into *big.Int.
var BigIntPlugin = plugin.NewPlugin("big.Int", plugin.FactoryFunc(
	func(t reflect.Type) bool { return t == bigIntType },
	func(ptr any) plugin.Adapter { return bigIntAdapter{ptr: ptr.(*big.Int)} },
))

// BigFloatPlugin coerces numbers/strings into *big.Float.
var BigFloatPlugin = plugin.NewPlugin("big.Float", plugin.FactoryFunc(
	func(t reflect.Type) bool { return t == bigFloatType },
	func(ptr any) plugin.Adapter { return bigFloatAdapter{ptr: ptr.(*big.Float)} },
))
