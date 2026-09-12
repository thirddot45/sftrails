package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"
)

type RateLimiter struct {
	mu          sync.Mutex
	requests    map[string][]time.Time
	maxRequests int
	window      time.Duration
	lastCleanup time.Time
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests:    make(map[string][]time.Time),
		maxRequests: maxRequests,
		window:      window,
	}
	return rl
}

const maxRateLimitClients = 10000

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if now.Sub(rl.lastCleanup) >= rl.window {
		for key, times := range rl.requests {
			if len(times) == 0 || now.Sub(times[len(times)-1]) >= rl.window {
				delete(rl.requests, key)
			}
		}
		rl.lastCleanup = now
	}
	if _, exists := rl.requests[ip]; !exists && len(rl.requests) >= maxRateLimitClients {
		return false
	}
	var valid []time.Time
	for _, t := range rl.requests[ip] {
		if now.Sub(t) < rl.window {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.maxRequests {
		rl.requests[ip] = valid
		return false
	}

	rl.requests[ip] = append(valid, now)
	return true
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := GetIP(r)
		if !rl.Allow(ip) {
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", rl.window.Seconds()))
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(sw, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "status", sw.status, "duration", time.Since(start))
	})
}

type clientIPKey struct{}

// ClientIPMiddleware opts into App Platform's ingress-provided address.
// Enable digitalocean only behind App Platform ingress; header presence alone
// never enables trust. Local/direct deployments ignore all forwarding headers.
func ClientIPMiddleware(mode string) (func(http.Handler) http.Handler, error) {
	if mode != "" && mode != "direct" && mode != "digitalocean" {
		return nil, fmt.Errorf("CLIENT_IP_MODE must be direct or digitalocean")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := connectionIP(r)
			if mode == "digitalocean" {
				values := r.Header.Values("DO-Connecting-IP")
				if len(values) == 1 {
					if addr, err := netip.ParseAddr(strings.TrimSpace(values[0])); err == nil && addr.Zone() == "" {
						ip = addr.Unmap().String()
					}
				}
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), clientIPKey{}, ip)))
		})
	}, nil
}

func GetIP(r *http.Request) string {
	if ip, ok := r.Context().Value(clientIPKey{}).(string); ok {
		return ip
	}
	return connectionIP(r)
}

func connectionIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if addr, err := netip.ParseAddr(host); err == nil {
		return addr.Unmap().String()
	}
	return host
}

// SecurityHeadersMiddleware also applies to errors and static responses.
// Inline styles/events remain in the existing UI; external assets are local.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	const csp = "default-src 'self'; script-src 'self' 'unsafe-inline'; " +
		"style-src 'self' 'unsafe-inline'; img-src 'self' data:; " +
		"connect-src 'self' https://api.zippopotam.us; " +
		"object-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", csp)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		next.ServeHTTP(w, r)
	})
}
