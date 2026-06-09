package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"sftrails/db"
	"sftrails/templates"
)

// metricsUser and metricsPassword are read from the environment. No defaults are
// baked in: if either is unset the dashboard is unreachable (fail closed).
func metricsUser() string     { return os.Getenv("METRICS_USER") }
func metricsPassword() string { return os.Getenv("METRICS_PASSWORD") }

// ephemeralSalt is generated once at startup and used for the visitor hash when
// METRICS_SALT is not set. It keeps unique-visitor counting working without a
// baked-in salt; counts reset across restarts unless METRICS_SALT is provided.
var ephemeralSalt = randomSalt()

func randomSalt() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("metrics: unable to generate random salt: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// metricsSalt is the secret key for the visitor HMAC. Because the IPv4 space is
// small enough to brute-force, the salt must be kept secret and unpredictable
// for the hash to resist reversal — treat it like any other credential. Set
// METRICS_SALT in production for stable unique counts across restarts;
// otherwise a random per-process salt is used and counts reset on restart.
func metricsSalt() string {
	if s := os.Getenv("METRICS_SALT"); s != "" {
		return s
	}
	return ephemeralSalt
}

// visitorHash returns a keyed (HMAC-SHA256) hash identifying a visitor for
// unique-count purposes. The raw IP is never stored or displayed. Keying with
// the secret salt — rather than just hashing salt+IP — means the hash cannot be
// reversed by brute-forcing the (small) IP space without also knowing the key.
func visitorHash(r *http.Request) string {
	mac := hmac.New(sha256.New, []byte(metricsSalt()))
	mac.Write([]byte(GetIP(r)))
	mac.Write([]byte{0})
	mac.Write([]byte(r.UserAgent()))
	return hex.EncodeToString(mac.Sum(nil))
}

// shouldTrack reports whether a request should be counted as a page view. We
// track GET requests to human-facing pages and skip assets, machine endpoints,
// and the password-protected metrics dashboard itself.
func shouldTrack(r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}
	p := r.URL.Path
	switch {
	case p == "/metrics",
		p == "/robots.txt",
		p == "/sitemap.xml",
		p == "/favicon.ico",
		strings.HasPrefix(p, "/static/"),
		strings.HasPrefix(p, "/api/"),
		strings.HasPrefix(p, "/.well-known/"),
		strings.HasPrefix(p, "/llms"):
		return false
	}
	return true
}

// MetricsMiddleware records a page view for tracked requests after the response
// is served. Recording happens asynchronously and never blocks or fails the
// request. No IP address is stored.
func MetricsMiddleware(database *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !shouldTrack(r) {
				next.ServeHTTP(w, r)
				return
			}
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			if sw.status >= 400 {
				return
			}
			path := r.URL.Path
			hash := visitorHash(r)
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := db.RecordPageView(ctx, database, path, hash); err != nil {
					slog.Error("failed to record page view", "path", path, "error", err)
				}
			}()
		})
	}
}

// BasicAuthMiddleware protects a handler with HTTP Basic Auth. Both the username
// (METRICS_USER) and password (METRICS_PASSWORD) are checked. Comparisons are
// constant-time and both run regardless of outcome to avoid leaking which field
// was wrong via timing. If either credential is unset the dashboard is unreachable
// (fail closed) so no usable default ever ships.
func BasicAuthMiddleware(realm string, next http.Handler) http.Handler {
	wantUser := metricsUser()
	wantPass := metricsPassword()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wantUser == "" || wantPass == "" {
			http.Error(w, "Metrics dashboard not configured", http.StatusServiceUnavailable)
			return
		}
		user, pass, ok := r.BasicAuth()
		userOK := subtle.ConstantTimeCompare([]byte(user), []byte(wantUser)) == 1
		passOK := subtle.ConstantTimeCompare([]byte(pass), []byte(wantPass)) == 1
		if !ok || !userOK || !passOK {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// HandleMetrics renders the site metrics dashboard. It must be mounted behind
// BasicAuthMiddleware.
func (h *Handler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	m, err := db.GetSiteMetrics(r.Context(), h.db)
	if err != nil {
		slog.Error("failed to get site metrics", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if err := templates.MetricsPage(m).Render(r.Context(), w); err != nil {
		slog.Error("failed to render metrics", "error", err)
	}
}
