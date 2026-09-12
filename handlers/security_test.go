package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientIPTrustBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, mode, remote string
		do                 []string
		want               string
	}{
		{"direct ignores spoofed headers", "direct", "192.0.2.1:80", []string{"203.0.113.2"}, "192.0.2.1"},
		{"platform client", "digitalocean", "10.0.0.1:80", []string{"203.0.113.2"}, "203.0.113.2"},
		{"platform missing falls back", "digitalocean", "10.0.0.1:80", nil, "10.0.0.1"},
		{"invalid list rejected", "digitalocean", "10.0.0.1:80", []string{"1.2.3.4, 5.6.7.8"}, "10.0.0.1"},
		{"duplicate header rejected", "digitalocean", "10.0.0.1:80", []string{"1.2.3.4", "5.6.7.8"}, "10.0.0.1"},
		{"mapped IPv4 normalized", "digitalocean", "10.0.0.1:80", []string{"::ffff:203.0.113.2"}, "203.0.113.2"},
		{"IPv6", "digitalocean", "[::1]:80", []string{"2001:db8::1"}, "2001:db8::1"},
		{"IPv6 zone rejected", "digitalocean", "10.0.0.1:80", []string{"fe80::1%eth0"}, "10.0.0.1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			middleware, err := ClientIPMiddleware(tc.mode)
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tc.remote
			req.Header["Do-Connecting-Ip"] = tc.do
			req.Header.Set("X-Forwarded-For", "198.51.100.9")
			req.Header.Set("X-Real-IP", "198.51.100.8")
			middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := GetIP(r); got != tc.want {
					t.Errorf("IP = %s, want %s", got, tc.want)
				}
			})).ServeHTTP(httptest.NewRecorder(), req)
		})
	}
	if _, err := ClientIPMiddleware("typo"); err == nil {
		t.Error("invalid mode accepted")
	}
}

func TestRateLimiterBoundAndExpiry(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	for i := range maxRateLimitClients {
		if !rl.Allow(fmt.Sprint(i)) {
			t.Fatal("unexpected capacity rejection")
		}
	}
	if rl.Allow("extra") {
		t.Fatal("unbounded map growth")
	}
	for key := range rl.requests {
		rl.requests[key] = []time.Time{time.Now().Add(-2 * time.Minute)}
	}
	rl.lastCleanup = time.Now().Add(-2 * time.Minute)
	if !rl.Allow("extra") {
		t.Error("expired capacity not reclaimed")
	}
}

func TestPlatformLimiterUsesClientAddress(t *testing.T) {
	middleware, err := ClientIPMiddleware("digitalocean")
	if err != nil {
		t.Fatal(err)
	}
	limiter := NewRateLimiter(1, time.Minute)
	h := middleware(limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })))
	for i, ip := range []string{"203.0.113.1", "203.0.113.1", "203.0.113.2"} {
		req := httptest.NewRequest("POST", "/vote", nil)
		req.RemoteAddr = "10.0.0.1:80"
		req.Header.Set("DO-Connecting-IP", ip)
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("198.51.100.%d", i))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		want := 204
		if i == 1 {
			want = 429
		}
		if w.Code != want {
			t.Errorf("request %d: %d, want %d", i, w.Code, want)
		}
	}
}
