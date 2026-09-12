package tts

import (
	"fmt"
	"time"

	"athenaeum/internal/models"
)

// parseHHMM parses a "HH:MM" daily time. Returns -1 for empty or invalid.
func parseHHMM(s string) int {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return -1
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return -1
	}
	return h*60 + m
}

// ValidSchedule reports whether the preference's schedule fields parse.
func ValidSchedule(p models.TTSUserPrefs) bool {
	if !p.SchedEnabled {
		return true
	}
	return parseHHMM(p.SchedStart) >= 0 && parseHHMM(p.SchedEnd) >= 0
}

// InWindow reports whether t falls inside the user's daily generation
// window. Disabled scheduling, or an empty/equal window, means always.
// Windows may wrap midnight (22:00-06:00).
func InWindow(t time.Time, p models.TTSUserPrefs) bool {
	if !p.SchedEnabled {
		return true
	}
	start := parseHHMM(p.SchedStart)
	end := parseHHMM(p.SchedEnd)
	if start < 0 || end < 0 || start == end {
		return true
	}
	cur := t.Hour()*60 + t.Minute()
	if start < end {
		return cur >= start && cur < end
	}
	return cur >= start || cur < end
}

// NextWindowStart returns the next time at or after t that lands in the
// window. Always-in-window prefs return t.
func NextWindowStart(t time.Time, p models.TTSUserPrefs) time.Time {
	if InWindow(t, p) {
		return t
	}
	start := parseHHMM(p.SchedStart)
	if start < 0 {
		return t
	}
	cand := time.Date(t.Year(), t.Month(), t.Day(), start/60, start%60, 0, 0, t.Location())
	if !cand.After(t) {
		cand = cand.Add(24 * time.Hour)
	}
	return cand
}
