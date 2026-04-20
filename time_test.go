package unface_test

import (
	"testing"
	"time"

	"github.com/schneid-l/unface"
)

func TestTimeFromRFC3339(t *testing.T) {
	var got time.Time
	f := unface.New(unface.With(unface.TimePlugin))
	if err := f.Unface("2026-04-19T12:34:56Z", &got); err != nil {
		t.Fatal(err)
	}
	want, _ := time.Parse(time.RFC3339, "2026-04-19T12:34:56Z")
	if !got.Equal(want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
}

func TestTimeFromUnixSeconds(t *testing.T) {
	var got time.Time
	f := unface.New(unface.With(unface.TimePlugin))
	if err := f.Unface(int64(1700000000), &got); err != nil {
		t.Fatal(err)
	}
	if got.Unix() != 1700000000 {
		t.Fatalf("got=%v", got)
	}
}

func TestTimeFromDate(t *testing.T) {
	var got time.Time
	f := unface.New(unface.With(unface.TimePlugin))
	if err := f.Unface("2026-04-19", &got); err != nil {
		t.Fatal(err)
	}
	if got.Year() != 2026 {
		t.Fatalf("got=%v", got)
	}
}

func TestTimeFromTime(t *testing.T) {
	var got time.Time
	f := unface.New(unface.With(unface.TimePlugin))
	want := time.Unix(42, 0).UTC()
	if err := f.Unface(want, &got); err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatalf("got=%v", got)
	}
}

func TestTimeParseFails(t *testing.T) {
	var got time.Time
	f := unface.New(unface.With(unface.TimePlugin))
	if err := f.Unface("not-a-time", &got); err == nil {
		t.Fatal("expected error")
	}
}

func TestDurationFromString(t *testing.T) {
	var got time.Duration
	f := unface.New(unface.With(unface.TimePlugin))
	if err := f.Unface("1h30m", &got); err != nil {
		t.Fatal(err)
	}
	if got != 90*time.Minute {
		t.Fatalf("got=%v", got)
	}
}

func TestDurationFromInt(t *testing.T) {
	var got time.Duration
	f := unface.New(unface.With(unface.TimePlugin))
	if err := f.Unface(60, &got); err != nil {
		t.Fatal(err)
	}
	if got != 60*time.Second {
		t.Fatalf("got=%v", got)
	}
}

func TestDurationFromFloat(t *testing.T) {
	var got time.Duration
	f := unface.New(unface.With(unface.TimePlugin))
	if err := f.Unface(1.5, &got); err != nil {
		t.Fatal(err)
	}
	if got != 1500*time.Millisecond {
		t.Fatalf("got=%v", got)
	}
}

func TestDurationInvalidString(t *testing.T) {
	var got time.Duration
	f := unface.New(unface.With(unface.TimePlugin))
	if err := f.Unface("not-a-duration", &got); err == nil {
		t.Fatal("expected error")
	}
}
