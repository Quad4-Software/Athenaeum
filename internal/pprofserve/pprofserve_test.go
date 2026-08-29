package pprofserve

import "testing"

func TestIsLoopbackAddr(t *testing.T) {
	cases := []struct {
		addr string
		ok   bool
	}{
		{"127.0.0.1:6060", true},
		{"localhost:6060", true},
		{"[::1]:6060", true},
		{":6060", false},
		{"0.0.0.0:6060", false},
		{"192.168.1.1:6060", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isLoopbackAddr(tc.addr); got != tc.ok {
			t.Fatalf("isLoopbackAddr(%q)=%v want %v", tc.addr, got, tc.ok)
		}
	}
}
