package server

import (
	"fmt"
	"strings"
	"testing"
)

// PROVED_DEFAULT_CSP_NO_OBJECT
// Guarantee: default CSP blocks plugins/object embeds (native PDF XSS surface).

func TestDefaultCSPObjectNoneOracle(t *testing.T) {
	if !strings.Contains(defaultCSP, "object-src 'none'") {
		t.Fatalf("defaultCSP missing object-src none: %s", defaultCSP)
	}
	if strings.Contains(defaultCSP, "object-src 'self'") {
		t.Fatal("defaultCSP still allows object-src self")
	}
	fmt.Println("PROVED_DEFAULT_CSP_NO_OBJECT: object-src none in default CSP")
}
