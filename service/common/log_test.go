package common

import (
	"testing"
	"time"
)

func TestFormatLogMessageAt(t *testing.T) {
	ts := time.Date(2026, 7, 14, 10, 47, 49, 123000000, time.UTC)

	got := formatLogMessageAt(ts, "startup begin")

	want := "2026-07-14 10:47:49.123 | startup begin"
	if got != want {
		t.Fatalf("unexpected formatted log message: got %q want %q", got, want)
	}
}

func TestFormatLogMessageAtEmptyMessage(t *testing.T) {
	ts := time.Date(2026, 7, 14, 10, 47, 49, 0, time.UTC)

	got := formatLogMessageAt(ts, "")

	want := "2026-07-14 10:47:49.000"
	if got != want {
		t.Fatalf("unexpected formatted empty log message: got %q want %q", got, want)
	}
}
