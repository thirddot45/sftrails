package handlers

import (
	"context"
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

// defaultMetricsUser is the username for the metrics dashboard. It can be
// overridden via the METRICS_USER environment variable.
const defaultMetricsUser = "nimda"

// defaultMetricsPassword guards the metrics dashboard. It can be overridden via
// the METRICS_PASSWORD environment variable.
const defaultMetricsPassword = "!m00nsh0t!"

// metricsSalt is mixed into the visitor hash so the stored value cannot be
// reversed back to an IP address. It defaults to a fixed value (stable unique
// counts across restarts) but should be set via METRICS_SALT in production.
const defaultMetricsSalt = "sftrails-metrics-v1"

func metricsUser() string {
	if u := os.Getenv("METRICS_USER"); u != "" {
		return u
	}
	return defaultMetricsUser
}

func metricsPassword() string {
	if p := os.Getenv("METRICS_PASSWORD"); p != "" {
		return p
	}
	return defaultMetricsPassword
}

func metricsSalt() string {
	if s := os.Getenv("METRICS_SALT"); s != "" {
		return s
	}
	return defaultMetricsSalt
}

// visitorHash returns a salted, one-way hash identifying a visitor for
// unique-count purposes. The raw IP is never stored or displayed; combining it
// with the User-Agent and a server salt yields a stable but non-reversible id.
func visitorHash(r *http.Request) string {
	h := sha256.New()
	h.Write([]byte(metricsSalt()))
	h.Write([]byte{0})
	h.Write([]byte(GetIP(r)))
	h.Write([]byte{0})
	h.Write([]byte(r.UserAgent()))
	return hex.EncodeToString(h.Sum(nil))
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
// (METRICS_USER, defaulting to the built-in value) and password (METRICS_PASSWORD)
// are checked. Comparisons are constant-time and both run regardless of outcome
// to avoid leaking which field was wrong via timing.
func BasicAuthMiddleware(realm string, next http.Handler) http.Handler {
	wantUser := metricsUser()
	wantPass := metricsPassword()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
