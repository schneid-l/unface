package unfacers

import (
	"fmt"
	"reflect"
	"time"

	"github.com/schneid-l/unface/plugin"
)

type timeAdapter struct{ ptr *time.Time }

// timeFormats is the layout list tried in order when parsing a string into
// a time.Time. Covers the most common wire formats; users needing exotic
// layouts should implement Untimer on their own type.
var timeFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
	time.RFC1123,
}

func (a timeAdapter) Unface(src any) error {
	switch v := src.(type) {
	case time.Time:
		*a.ptr = v
		return nil
	case string:
		return a.parseString(v)
	case nil:
		*a.ptr = time.Time{}
		return nil
	}
	if n, ok := plugin.NumberOf(src); ok {
		if s, ok := n.Int64(); ok {
			*a.ptr = time.Unix(s, 0).UTC()
			return nil
		}
	}
	return plugin.ErrNotHandled
}

func (a timeAdapter) parseString(s string) error {
	for _, layout := range timeFormats {
		if t, err := time.Parse(layout, s); err == nil {
			*a.ptr = t
			return nil
		}
	}
	return fmt.Errorf("unface/time: cannot parse %q as time.Time", s)
}

type durationAdapter struct{ ptr *time.Duration }

func (a durationAdapter) Unface(src any) error {
	switch v := src.(type) {
	case time.Duration:
		*a.ptr = v
		return nil
	case string:
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("unface/time: parse duration from %q: %w", v, err)
		}
		*a.ptr = d
		return nil
	case nil:
		*a.ptr = 0
		return nil
	}
	if n, ok := plugin.NumberOf(src); ok {
		if s, ok := n.Int64(); ok {
			*a.ptr = time.Duration(s) * time.Second
			return nil
		}
		if f, ok := n.Float64(); ok {
			*a.ptr = time.Duration(f * float64(time.Second))
			return nil
		}
	}
	return plugin.ErrNotHandled
}

var (
	timeType     = reflect.TypeOf(time.Time{})
	durationType = reflect.TypeOf(time.Duration(0))
)

// TimePlugin coerces strings, numbers, and time.Time sources into time.Time
// or time.Duration destinations. Strings are parsed against a list of
// common layouts (RFC3339, date-only, RFC1123); integers into time.Time
// are interpreted as Unix seconds; numbers into time.Duration as seconds.
var TimePlugin = plugin.NewPlugin("time",
	plugin.FactoryFunc(
		func(t reflect.Type) bool { return t == timeType },
		func(ptr any) plugin.Adapter { return timeAdapter{ptr: ptr.(*time.Time)} },
	),
	plugin.FactoryFunc(
		func(t reflect.Type) bool { return t == durationType },
		func(ptr any) plugin.Adapter { return durationAdapter{ptr: ptr.(*time.Duration)} },
	),
)
