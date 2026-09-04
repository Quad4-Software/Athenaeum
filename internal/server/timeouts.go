package server

import "time"

// httpClientTimeout is the shared outbound HTTP request timeout (OIDC discovery and similar).
const httpClientTimeout = 10 * time.Second
