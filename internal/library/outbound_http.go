package library

import (
	"context"
	"net"
	"net/http"
	"time"
)

var (
	metadataAPITransport = newPooledTransport(nil)
	coverHTTPTransport   = newPooledTransport(dialAllowed)

	sharedMetadataLookupClient = &http.Client{
		Timeout:   8 * time.Second,
		Transport: metadataAPITransport,
	}
	sharedMetadataSearchClient = &http.Client{
		Timeout:   12 * time.Second,
		Transport: metadataAPITransport,
	}
	coverHTTPClient = &http.Client{
		Timeout:   15 * time.Second,
		Transport: coverHTTPTransport,
	}
)

func newPooledTransport(dialContext func(context.Context, string, string) (net.Conn, error)) *http.Transport {
	t := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          64,
		MaxIdleConnsPerHost:   8,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}
	if dialContext != nil {
		t.DialContext = dialContext
	}
	return t
}

// CloseOutboundHTTP closes idle connections on shared metadata and cover clients.
func CloseOutboundHTTP() {
	metadataAPITransport.CloseIdleConnections()
	coverHTTPTransport.CloseIdleConnections()
}
