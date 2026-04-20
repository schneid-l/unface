// Package reflectutil contains reflection helpers shared across unface.
// It is internal: APIs may change without notice.
package reflectutil

import (
	"errors"
	"reflect"
)

// ErrInvalidDest is returned when DerefToAddressable receives a value that
// is not a non-nil pointer. Exposed so callers can re-wrap with the public
// unface.ErrInvalidDest sentinel.
var ErrInvalidDest = errors.New("reflectutil: dest must be a non-nil pointer")

// DerefToAddressable takes an any that must be a non-nil pointer, walks
// through any number of pointer-to-pointer indirections (allocating nested
// nil pointers along the way), and returns a settable reflect.Value for the
// innermost non-pointer element.
func DerefToAddressable(dest any) (reflect.Value, error) {
	if dest == nil {
		return reflect.Value{}, ErrInvalidDest
	}
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return reflect.Value{}, ErrInvalidDest
	}
	v = v.Elem()
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			if !v.CanSet() {
				return reflect.Value{}, ErrInvalidDest
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v, nil
}

// IsNilLike reports whether v is nil or a nillable kind holding nil.
func IsNilLike(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
