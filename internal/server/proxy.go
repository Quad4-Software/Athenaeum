package server

import (
	"net"
	"net/http"
	"strings"
	"sync"
)

type proxyTrust struct {
	mu      sync.RWMutex
	raw     string
	network []*net.IPNet
	ips     []net.IP
}

func (p *proxyTrust) set(raw string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if raw == p.raw {
		return
	}
	p.raw = raw
	p.network = nil
	p.ips = nil
	for part := range strings.SplitSeq(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "/") {
			_, n, err := net.ParseCIDR(part)
			if err == nil {
				p.network = append(p.network, n)
			}
			continue
		}
		if ip := net.ParseIP(part); ip != nil {
			p.ips = append(p.ips, ip)
		}
	}
}

func (p *proxyTrust) trusted(addr string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.network) == 0 && len(p.ips) == 0 {
		return false
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, n := range p.network {
		if n.Contains(ip) {
			return true
		}
	}
	for _, trusted := range p.ips {
		if trusted.Equal(ip) {
			return true
		}
	}
	return false
}

func (s *Server) proxyTrusted(r *http.Request) bool {
	return s.proxies.trusted(r.RemoteAddr)
}

func (s *Server) forwardedValue(r *http.Request, header string) string {
	if !s.proxyTrusted(r) {
		return ""
	}
	v := r.Header.Get(header)
	if v == "" {
		return ""
	}
	return strings.TrimSpace(strings.Split(v, ",")[0])
}
