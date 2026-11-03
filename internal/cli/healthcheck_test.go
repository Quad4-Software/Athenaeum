package cli

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWantsHealthcheck(t *testing.T) {
	cases := []struct {
		args []string
		want bool
	}{
		{nil, false},
		{[]string{"--addr", ":8080"}, false},
		{[]string{"--healthcheck"}, true},
		{[]string{"-healthcheck"}, true},
		{[]string{"healthcheck"}, true},
		{[]string{"healthcheck", "--addr", ":9999"}, true},
		{[]string{"--addr", ":9999", "--healthcheck"}, true},
		{[]string{"users", "list"}, false},
		{[]string{"self-check"}, false},
	}
	for _, tc := range cases {
		if got := WantsHealthcheck(tc.args); got != tc.want {
			t.Fatalf("WantsHealthcheck(%v)=%v want %v", tc.args, got, tc.want)
		}
	}
}

func TestStripHealthcheck(t *testing.T) {
	got := StripHealthcheck([]string{"healthcheck", "--addr", ":9"})
	if len(got) != 2 || got[0] != "--addr" || got[1] != ":9" {
		t.Fatalf("got %#v", got)
	}
	got = StripHealthcheck([]string{"--addr", ":9", "--healthcheck", "--debug"})
	if len(got) != 3 || got[0] != "--addr" || got[2] != "--debug" {
		t.Fatalf("got %#v", got)
	}
}

func TestHealthcheckURL(t *testing.T) {
	cases := map[string]string{
		":8080":          "http://127.0.0.1:8080/api/health",
		"8080":           "http://127.0.0.1:8080/api/health",
		"0.0.0.0:8080":   "http://127.0.0.1:8080/api/health",
		"127.0.0.1:9000": "http://127.0.0.1:9000/api/health",
		"::1:9000":       "http://[::1]:9000/api/health",
	}
	for addr, want := range cases {
		if got := healthcheckURL(addr); got != want {
			t.Fatalf("healthcheckURL(%q)=%q want %q", addr, got, want)
		}
	}
}

func TestRunHealthcheckHealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/health" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_, port, err := net.SplitHostPort(srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	if err := RunHealthcheck([]string{"--addr", "127.0.0.1:" + port}); err != nil {
		t.Fatal(err)
	}
}

func TestRunHealthcheckUnhealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	_, port, err := net.SplitHostPort(srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	old := healthcheckTimeout
	healthcheckTimeout = 200 * time.Millisecond
	defer func() { healthcheckTimeout = old }()
	start := time.Now()
	err = RunHealthcheck([]string{"--addr", "127.0.0.1:" + port})
	if err == nil {
		t.Fatal("expected error for 503 health endpoint")
	}
	if time.Since(start) > healthcheckTimeout+2*time.Second {
		t.Fatalf("healthcheck outlived its timeout: %s", time.Since(start))
	}
}

func TestRunHealthcheckNoServer(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	old := healthcheckTimeout
	healthcheckTimeout = 200 * time.Millisecond
	defer func() { healthcheckTimeout = old }()
	if err := RunHealthcheck([]string{"--addr", addr}); err == nil {
		t.Fatal("expected error when nothing listens")
	}
}
