package server

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"athenaeum/internal/auth"
	"athenaeum/internal/brand"
	"athenaeum/internal/version"
)

var httpRequestsTotal atomic.Uint64

func (s *Server) registerMetricsRoute(mux *http.ServeMux) {
	mux.HandleFunc("GET /metrics", s.handleMetrics)
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.store.GetServerConfig(r.Context(), true)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !cfg.MetricsEnabled {
		http.NotFound(w, r)
		return
	}
	if cfg.MetricsAuth {
		if !s.metricsAuthorized(r, cfg.MetricsUsername, cfg.MetricsPassword) {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s metrics"`, brand.Name))
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	} else if !metricsClientIsLoopback(r) {
		// Unauthenticated metrics stay on loopback only (scrape via SSH tunnel or sidecar).
		http.Error(w, "metrics without auth are loopback-only", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	bookCount, userCount := s.metricCounts(r)
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	p := brand.MetricsPrefix
	lines := []string{
		fmt.Sprintf("# HELP %sinfo %s build metadata.", p, brand.Name),
		fmt.Sprintf("# TYPE %sinfo gauge", p),
		fmt.Sprintf(`%sinfo{version=%q,web_version=%q} 1`, p, version.Version, version.WebVersion),
		fmt.Sprintf("# HELP %shttp_requests_total Total HTTP requests served.", p),
		fmt.Sprintf("# TYPE %shttp_requests_total counter", p),
		fmt.Sprintf("%shttp_requests_total %d", p, httpRequestsTotal.Load()),
		fmt.Sprintf("# HELP %sbooks_total Indexed books in the library.", p),
		fmt.Sprintf("# TYPE %sbooks_total gauge", p),
		fmt.Sprintf("%sbooks_total %d", p, bookCount),
		fmt.Sprintf("# HELP %susers_total Registered user accounts.", p),
		fmt.Sprintf("# TYPE %susers_total gauge", p),
		fmt.Sprintf("%susers_total %d", p, userCount),
		fmt.Sprintf("# HELP %sgo_goroutines Current goroutines.", p),
		fmt.Sprintf("# TYPE %sgo_goroutines gauge", p),
		fmt.Sprintf("%sgo_goroutines %d", p, runtime.NumGoroutine()),
		fmt.Sprintf("# HELP %sgo_mem_alloc_bytes Bytes allocated and in use.", p),
		fmt.Sprintf("# TYPE %sgo_mem_alloc_bytes gauge", p),
		fmt.Sprintf("%sgo_mem_alloc_bytes %d", p, ms.Alloc),
		fmt.Sprintf("# HELP %sup Whether the %s process is serving requests.", p, brand.Name),
		fmt.Sprintf("# TYPE %sup gauge", p),
		fmt.Sprintf("%sup 1", p),
		fmt.Sprintf("# HELP %sprocess_start_time_seconds Start time of the process since unix epoch.", p),
		fmt.Sprintf("# TYPE %sprocess_start_time_seconds gauge", p),
		fmt.Sprintf("%sprocess_start_time_seconds %d", p, processStart.Unix()),
	}
	_, _ = w.Write([]byte(strings.Join(lines, "\n") + "\n")) // #nosec G705 -- Prometheus text exposition; numeric metrics only
}

func (s *Server) metricsAuthorized(r *http.Request, username, passwordHash string) bool {
	u, p, ok := r.BasicAuth()
	if !ok || u != username {
		return false
	}
	return auth.CheckPassword(passwordHash, p)
}

func metricsClientIsLoopback(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (s *Server) metricCounts(r *http.Request) (books, users int64) {
	stats, err := s.store.Stats(r.Context(), 0, 0)
	if err == nil {
		books = stats.TotalBooks
	}
	required, err := s.store.AuthRequired(r.Context())
	if err == nil && required {
		users, _ = s.store.UserCount(r.Context())
	}
	return books, users
}

var processStart = time.Now()
