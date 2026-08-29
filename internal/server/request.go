package server

import (
	"net"
	"net/http"
	"strings"
)

func (s *Server) clientIP(r *http.Request) string {
	if fwd := s.forwardedValue(r, "X-Forwarded-For"); fwd != "" {
		return fwd
	}
	if fwd := s.forwardedValue(r, "X-Real-IP"); fwd != "" {
		return fwd
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func (s *Server) requestBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(s.forwardedValue(r, "X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	host := r.Host
	if fwd := s.forwardedValue(r, "X-Forwarded-Host"); fwd != "" {
		host = fwd
	}
	return scheme + "://" + host
}
