package plugin

import "reflect"

// Map is a type-neutral view over a map source.
type Map interface {
	Len() int
	Keys() []any
	KeyType() reflect.Type
	ValueType() reflect.Type
	Has(key any) bool
	Get(key any) (any, bool)
	Iter(yield func(k, v any) bool)
	Raw() any

	GetString(key any) (string, bool)
	GetInt64(key any) (int64, bool)
	GetBool(key any) (bool, bool)
	GetMap(key any) (Map, bool)
	GetList(key any) (List, bool)
}

// Unmapper consumes a map-like source via the Map abstraction.
type Unmapper interface {
	Unmap(m Map) error
}

// MapOf wraps v in a Map if v is a map.
func MapOf(v any) (Map, bool) {
	if v == nil {
		return nil, false
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map {
		return nil, false
	}
	return reflectMap{rv: rv}, true
}

type reflectMap struct{ rv reflect.Value }

func (m reflectMap) Len() int                { return m.rv.Len() }
func (m reflectMap) KeyType() reflect.Type   { return m.rv.Type().Key() }
func (m reflectMap) ValueType() reflect.Type { return m.rv.Type().Elem() }

func (m reflectMap) Keys() []any {
	iter := m.rv.MapRange()
	out := make([]any, 0, m.rv.Len())
	for iter.Next() {
		out = append(out, iter.Key().Interface())
	}
	return out
}

func (m reflectMap) Has(key any) bool {
	_, ok := m.Get(key)
	return ok
}

func (m reflectMap) Get(key any) (any, bool) {
	kv, ok := m.convertKey(key)
	if !ok {
		return nil, false
	}
	v := m.rv.MapIndex(kv)
	if !v.IsValid() {
		return nil, false
	}
	return v.Interface(), true
}

func (m reflectMap) convertKey(key any) (reflect.Value, bool) {
	kt := m.rv.Type().Key()
	kv := reflect.ValueOf(key)
	if !kv.IsValid() {
		return reflect.Value{}, false
	}
	if kv.Type() == kt {
		return kv, true
	}
	if kv.Type().ConvertibleTo(kt) {
		return kv.Convert(kt), true
	}
	return reflect.Value{}, false
}

func (m reflectMap) Iter(yield func(k, v any) bool) {
	iter := m.rv.MapRange()
	for iter.Next() {
		if !yield(iter.Key().Interface(), iter.Value().Interface()) {
			return
		}
	}
}

func (m reflectMap) Raw() any { return m.rv.Interface() }

func (m reflectMap) GetString(key any) (string, bool) {
	v, ok := m.Get(key)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func (m reflectMap) GetInt64(key any) (int64, bool) {
	v, ok := m.Get(key)
	if !ok {
		return 0, false
	}
	n, ok := NumberOf(v)
	if !ok {
		return 0, false
	}
	return n.Int64()
}

func (m reflectMap) GetBool(key any) (bool, bool) {
	v, ok := m.Get(key)
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

func (m reflectMap) GetMap(key any) (Map, bool) {
	v, ok := m.Get(key)
	if !ok {
		return nil, false
	}
	return MapOf(v)
}

func (m reflectMap) GetList(key any) (List, bool) {
	v, ok := m.Get(key)
	if !ok {
		return nil, false
	}
	return ListOf(v)
}
