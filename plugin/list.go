package plugin

import "reflect"

// List is a type-neutral view over a slice/array source.
type List interface {
	Len() int
	At(i int) any
	ElemType() reflect.Type
	Iter(yield func(i int, v any) bool)
	Slice(lo, hi int) List
	Raw() any

	Map(fn func(any) any) List
	Filter(fn func(any) bool) List
	ToSlice() []any
}

// Unlister consumes a slice/array-like source via the List abstraction.
type Unlister interface {
	Unlist(l List) error
}

// ListOf wraps v in a List if v is a slice or array.
func ListOf(v any) (List, bool) {
	if v == nil {
		return nil, false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		return reflectList{rv: rv}, true
	default:
		return nil, false
	}
}

type reflectList struct{ rv reflect.Value }

func (l reflectList) Len() int               { return l.rv.Len() }
func (l reflectList) At(i int) any           { return l.rv.Index(i).Interface() }
func (l reflectList) ElemType() reflect.Type { return l.rv.Type().Elem() }

func (l reflectList) Iter(yield func(i int, v any) bool) {
	n := l.rv.Len()
	for i := 0; i < n; i++ {
		if !yield(i, l.rv.Index(i).Interface()) {
			return
		}
	}
}

func (l reflectList) Slice(lo, hi int) List {
	if l.rv.Kind() == reflect.Array {
		// Array slicing needs addressability; copy into a fresh slice.
		out := reflect.MakeSlice(reflect.SliceOf(l.rv.Type().Elem()), hi-lo, hi-lo)
		for i := 0; i < hi-lo; i++ {
			out.Index(i).Set(l.rv.Index(lo + i))
		}
		return reflectList{rv: out}
	}
	return reflectList{rv: l.rv.Slice(lo, hi)}
}

func (l reflectList) Raw() any { return l.rv.Interface() }

func (l reflectList) ToSlice() []any {
	out := make([]any, l.rv.Len())
	for i := range out {
		out[i] = l.rv.Index(i).Interface()
	}
	return out
}

func (l reflectList) Map(fn func(any) any) List {
	out := make([]any, l.rv.Len())
	for i := range out {
		out[i] = fn(l.rv.Index(i).Interface())
	}
	return reflectList{rv: reflect.ValueOf(out)}
}

func (l reflectList) Filter(fn func(any) bool) List {
	out := []any{}
	for i := 0; i < l.rv.Len(); i++ {
		v := l.rv.Index(i).Interface()
		if fn(v) {
			out = append(out, v)
		}
	}
	return reflectList{rv: reflect.ValueOf(out)}
}
