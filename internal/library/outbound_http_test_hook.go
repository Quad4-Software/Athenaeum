package library

import "net/http"

// SwapCoverHTTPClient replaces the shared cover download client.
// The returned function restores the previous client.
func SwapCoverHTTPClient(c *http.Client) func() {
	prev := coverHTTPClient
	coverHTTPClient = c
	return func() { coverHTTPClient = prev }
}
