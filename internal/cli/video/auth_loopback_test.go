package video

import "testing"

// TestRequireLoopbackHostPort guards the CDP dial-SSRF defense (HARDENING
// 2026-06-04 second pass): the WebSocket debugger host is derived from Chrome's
// /json/list response, so getCookiesViaWebSocket must refuse to dial anything
// that is not omni's own loopback Chrome.
func TestRequireLoopbackHostPort(t *testing.T) {
	loopback := []string{"127.0.0.1:9222", "localhost:9222", "[::1]:9222", "127.0.0.1"}
	for _, h := range loopback {
		if err := requireLoopbackHostPort(h); err != nil {
			t.Errorf("requireLoopbackHostPort(%q) = %v, want nil (loopback is allowed)", h, err)
		}
	}

	nonLoopback := []string{"169.254.169.254:80", "evil.example.com:443", "10.0.0.5:9222", "8.8.8.8:80", "0.0.0.0:9222"}
	for _, h := range nonLoopback {
		if err := requireLoopbackHostPort(h); err == nil {
			t.Errorf("requireLoopbackHostPort(%q) = nil, want refusal (non-loopback target)", h)
		}
	}
}
