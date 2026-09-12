package tts

import (
	"testing"
	"time"

	"athenaeum/internal/models"
)

func at(h, m int) time.Time {
	return time.Date(2026, 3, 14, h, m, 0, 0, time.UTC)
}

func TestInWindowDisabled(t *testing.T) {
	if !InWindow(at(3, 0), models.TTSUserPrefs{}) {
		t.Fatal("disabled schedule should always be in window")
	}
}

func TestInWindowNormal(t *testing.T) {
	p := models.TTSUserPrefs{SchedEnabled: true, SchedStart: "02:00", SchedEnd: "06:00"}
	cases := []struct {
		t    time.Time
		want bool
	}{
		{at(1, 59), false},
		{at(2, 0), true},
		{at(4, 30), true},
		{at(5, 59), true},
		{at(6, 0), false},
	}
	for _, c := range cases {
		if got := InWindow(c.t, p); got != c.want {
			t.Fatalf("InWindow(%v) = %v, want %v", c.t, got, c.want)
		}
	}
}

func TestInWindowWrapsMidnight(t *testing.T) {
	p := models.TTSUserPrefs{SchedEnabled: true, SchedStart: "22:00", SchedEnd: "06:00"}
	if !InWindow(at(23, 30), p) || !InWindow(at(3, 0), p) {
		t.Fatal("expected inside wrap-around window")
	}
	if InWindow(at(12, 0), p) {
		t.Fatal("noon should be outside 22:00-06:00")
	}
}

func TestInWindowEqualTimesMeansAlways(t *testing.T) {
	p := models.TTSUserPrefs{SchedEnabled: true, SchedStart: "04:00", SchedEnd: "04:00"}
	if !InWindow(at(12, 0), p) {
		t.Fatal("equal start/end should mean always")
	}
}

func TestNextWindowStart(t *testing.T) {
	p := models.TTSUserPrefs{SchedEnabled: true, SchedStart: "02:00", SchedEnd: "06:00"}
	got := NextWindowStart(at(12, 0), p)
	if got.Hour() != 2 || got.Minute() != 0 || got.Day() != 15 {
		t.Fatalf("expected next-day 02:00, got %v", got)
	}
	if got := NextWindowStart(at(3, 0), p); !got.Equal(at(3, 0)) {
		t.Fatalf("in-window time should pass through, got %v", got)
	}
}

func TestNextWindowStartWrapsMidnight(t *testing.T) {
	p := models.TTSUserPrefs{SchedEnabled: true, SchedStart: "22:00", SchedEnd: "06:00"}
	got := NextWindowStart(at(12, 0), p)
	if got.Hour() != 22 || got.Day() != 14 {
		t.Fatalf("expected same-day 22:00, got %v", got)
	}
}

func TestValidSchedule(t *testing.T) {
	ok := models.TTSUserPrefs{SchedEnabled: true, SchedStart: "02:00", SchedEnd: "06:00"}
	bad := models.TTSUserPrefs{SchedEnabled: true, SchedStart: "2am", SchedEnd: "06:00"}
	if !ValidSchedule(ok) || ValidSchedule(bad) {
		t.Fatal("unexpected schedule validation result")
	}
}
