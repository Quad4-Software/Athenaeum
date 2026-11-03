package cli

import (
	"fmt"
	"net"
	"strings"
	"time"

	"athenaeum/internal/config"
)

// healthcheckTimeout bounds the probe; tests shrink it.
var healthcheckTimeout = 5 * time.Second

// WantsHealthcheck reports whether args request a loopback health probe.
func WantsHealthcheck(args []string) bool {
	if len(args) > 0 && args[0] == "healthcheck" {
		return true
	}
	for _, a := range args {
		if a == "--healthcheck" || a == "-healthcheck" {
			return true
		}
	}
	return false
}

// StripHealthcheck removes the healthcheck command or flag from args.
func StripHealthcheck(args []string) []string {
	out := make([]string, 0, len(args))
	if len(args) > 0 && args[0] == "healthcheck" {
		args = args[1:]
	}
	for _, a := range args {
		if a == "--healthcheck" || a == "-healthcheck" {
			continue
		}
		out = append(out, a)
	}
	return out
}

// healthcheckURL builds the loopback /api/health probe URL for a listen
// address in the same form --addr accepts.
func healthcheckURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		if i := strings.LastIndex(addr, ":"); i >= 0 {
			host, port = addr[:i], addr[i+1:]
		} else {
			host = ""
			port = addr
		}
	}
	if host == "" || host == "0.0.0.0" || host == "::" || host == "[::]" {
		host = "127.0.0.1"
	}
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		host = "[" + host + "]"
	}
	return fmt.Sprintf("http://%s:%s/api/health", host, port)
}

// RunHealthcheck probes a running server on loopback and reports whether
// /api/health answers 200. It honors --addr / ATHENAEUM_ADDR like the server.
func RunHealthcheck(args []string) error {
	cfg, err := config.Parse(args)
	if err != nil {
		return err
	}
	url := healthcheckURL(cfg.Addr)
	if err := waitHealthy(url, healthcheckTimeout); err != nil {
		return fmt.Errorf("%s: %w", url, err)
	}
	return nil
}
