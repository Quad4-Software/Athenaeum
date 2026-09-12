package storage

import (
	"context"
	"testing"
	"time"

	"athenaeum/internal/models"
)

func seedSessionBook(t *testing.T, ctx context.Context, s *Store, title, relPath string) int64 {
	t.Helper()
	id, err := s.UpsertBook(ctx, &models.Book{Title: title, Format: models.FormatEPUB, RelPath: relPath}, 1)
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	return id
}

func sessionCount(t *testing.T, ctx context.Context, s *Store) int {
	t.Helper()
	var n int
	if err := s.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM reading_sessions`).Scan(&n); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	return n
}

func TestAddReadSecondsGroupsHeartbeats(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	bookID := seedSessionBook(t, ctx, s, "Dune", "dune.epub")
	uid := models.AnonymousUserID

	now := time.Now().Unix()
	if err := s.addReadSecondsAt(ctx, uid, bookID, 300, now); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := s.addReadSecondsAt(ctx, uid, bookID, 120, now+10*60); err != nil {
		t.Fatalf("add 10min later: %v", err)
	}

	sessions, err := s.ListReadingSessions(ctx, uid, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("sessions = %d, want 1 merged session", len(sessions))
	}
	got := sessions[0]
	if got.Seconds != 420 {
		t.Errorf("seconds = %d, want 300+120=420", got.Seconds)
	}
	if got.StartedAt.Unix() != now {
		t.Errorf("started_at = %d, want first heartbeat %d", got.StartedAt.Unix(), now)
	}
	if got.EndedAt.Unix() != now+10*60 {
		t.Errorf("ended_at = %d, want second heartbeat %d", got.EndedAt.Unix(), now+10*60)
	}

	// A heartbeat more than the merge window later opens a new session.
	if err := s.addReadSecondsAt(ctx, uid, bookID, 60, now+45*60); err != nil {
		t.Fatalf("add 45min later: %v", err)
	}
	sessions, err = s.ListReadingSessions(ctx, uid, 0)
	if err != nil {
		t.Fatalf("list 2: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("sessions = %d, want 2", len(sessions))
	}
	if sessions[0].Seconds != 60 {
		t.Errorf("newest session seconds = %d, want 60", sessions[0].Seconds)
	}
}

func TestRecordSessionMergeWindowBoundary(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	bookID := seedSessionBook(t, ctx, s, "Edge", "edge.epub")
	uid := models.AnonymousUserID

	base := time.Now().Unix()
	cases := []struct {
		name    string
		gapSecs int64
		want    int
	}{
		{"inside window", 29 * 60, 1},
		{"at window edge", 30 * 60, 1},
		{"past window", 30*60 + 1, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := s.DB().ExecContext(ctx, `DELETE FROM reading_sessions`); err != nil {
				t.Fatalf("reset: %v", err)
			}
			if err := s.recordSession(ctx, uid, bookID, 100, base); err != nil {
				t.Fatalf("seed: %v", err)
			}
			if err := s.recordSession(ctx, uid, bookID, 50, base+tc.gapSecs); err != nil {
				t.Fatalf("extend: %v", err)
			}
			n := sessionCount(t, ctx, s)
			if n != tc.want {
				t.Fatalf("sessions = %d, want %d", n, tc.want)
			}
		})
	}
}

func TestListReadingSessionsJoinsTitleNewestFirst(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	uid := models.AnonymousUserID
	bookA := seedSessionBook(t, ctx, s, "Alpha", "a.epub")
	bookB := seedSessionBook(t, ctx, s, "Beta", "b.epub")

	now := time.Now().Unix()
	if err := s.addReadSecondsAt(ctx, uid, bookA, 30, now-3600); err != nil {
		t.Fatalf("seed a: %v", err)
	}
	if err := s.addReadSecondsAt(ctx, uid, bookB, 45, now); err != nil {
		t.Fatalf("seed b: %v", err)
	}

	sessions, err := s.ListReadingSessions(ctx, uid, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("sessions = %d, want 2", len(sessions))
	}
	if sessions[0].Title != "Beta" || sessions[1].Title != "Alpha" {
		t.Errorf("order = %q, %q; want Beta, Alpha (newest first)", sessions[0].Title, sessions[1].Title)
	}
	if sessions[0].BookID != bookB || sessions[1].BookID != bookA {
		t.Errorf("book ids = %d, %d; want %d, %d", sessions[0].BookID, sessions[1].BookID, bookB, bookA)
	}

	limited, err := s.ListReadingSessions(ctx, uid, 1)
	if err != nil {
		t.Fatalf("list limit: %v", err)
	}
	if len(limited) != 1 || limited[0].Title != "Beta" {
		t.Errorf("limit=1 got %+v", limited)
	}
}

func TestReadingStatsSessionWindows(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	uid := models.AnonymousUserID
	bookID := seedSessionBook(t, ctx, s, "Windows", "win.epub")

	now := time.Now().Unix()
	day := int64(24 * 3600)
	// Seed oldest first so each stays its own session.
	for _, seed := range []struct {
		ageDays int64
		secs    int64
	}{
		{40, 700},
		{10, 500},
		{3, 300},
	} {
		if err := s.recordSession(ctx, uid, bookID, seed.secs, now-seed.ageDays*day); err != nil {
			t.Fatalf("seed %dd: %v", seed.ageDays, err)
		}
	}

	st, err := s.ReadingStats(ctx, uid)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if st.Sessions7d != 1 {
		t.Errorf("sessions7d = %d, want 1", st.Sessions7d)
	}
	if st.Seconds7d != 300 {
		t.Errorf("seconds7d = %d, want 300", st.Seconds7d)
	}
	if st.Seconds30d != 800 {
		t.Errorf("seconds30d = %d, want 300+500=800", st.Seconds30d)
	}
}

func TestNonPositiveDeltasRecordNoSession(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	uid := models.AnonymousUserID
	bookID := seedSessionBook(t, ctx, s, "Zero", "zero.epub")

	for _, secs := range []int64{0, -120} {
		if err := s.AddReadSeconds(ctx, uid, bookID, secs); err != nil {
			t.Fatalf("add %d: %v", secs, err)
		}
	}
	if err := s.SaveProgress(ctx, uid, models.Progress{BookID: bookID, Location: "1", Percent: 0.1}); err != nil {
		t.Fatalf("save progress: %v", err)
	}
	if n := sessionCount(t, ctx, s); n != 0 {
		t.Fatalf("sessions = %d, want 0 for zero/negative deltas", n)
	}
}

func TestSaveProgressRecordsSession(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	uid := models.AnonymousUserID
	bookID := seedSessionBook(t, ctx, s, "ViaProgress", "vp.epub")

	if err := s.SaveProgress(ctx, uid, models.Progress{BookID: bookID, Location: "10", Percent: 0.2, ReadSeconds: 90}); err != nil {
		t.Fatalf("save: %v", err)
	}
	sessions, err := s.ListReadingSessions(ctx, uid, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(sessions) != 1 || sessions[0].Seconds != 90 {
		t.Fatalf("sessions = %+v, want one session of 90s", sessions)
	}
}
