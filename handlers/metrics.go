package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
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

// ephemeralSalt is generated once at startup and used for the visitor hash when
// METRICS_SALT is not set. It keeps unique-visitor counting working without a
// baked-in salt; visitor identities change across restarts unless METRICS_SALT is provided.
var ephemeralSalt = randomSalt()

func randomSalt() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("metrics: unable to generate random salt: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// metricsSalt is the secret key used for pseudonymous visitor identifiers. Set METRICS_SALT in production for stable
// unique counts across restarts; otherwise a random per-process salt is used.
func metricsSalt() string {
	if s := os.Getenv("METRICS_SALT"); s != "" {
		return s
	}
	return ephemeralSalt
}

// visitorHash is a pseudonymous analytics identifier keyed by METRICS_SALT.
// Keep the key secret; a holder can recompute hashes for candidate IP/UA pairs.
func visitorHash(r *http.Request) string {
	mac := hmac.New(sha256.New, []byte(metricsSalt()))
	mac.Write([]byte(GetIP(r)))
	mac.Write([]byte{0})
	mac.Write([]byte(r.UserAgent()))
	return hex.EncodeToString(mac.Sum(nil))
}

// shouldTrack reports whether a request should be counted as a page view. We
// track GET requests to human-facing pages and skip assets, machine endpoints,
// and the public metrics dashboard itself.
func shouldTrack(r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}
	p := r.URL.Path
	return p == "/" || p == "/trails-list" || p == "/status" ||
		(strings.HasPrefix(p, "/trail/") && len(p) <= 160)
}

// MetricsMiddleware records a page view for tracked requests after the response
// is served. Recording happens asynchronously and never blocks or fails the
// request. No IP address is stored.
func MetricsMiddleware(database *sql.DB) func(http.Handler) http.Handler {
	// Bound background writes when traffic exceeds database throughput.
	pending := make(chan struct{}, 128)
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
			select {
			case pending <- struct{}{}:
			default:
				return
			}
			hash := visitorHash(r)
			go func() {
				defer func() { <-pending }()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := db.RecordPageView(ctx, database, path, hash); err != nil {
					slog.Error("failed to record page view", "path", path, "error", err)
				}
			}()
		})
	}
}

// MetricsDiscoveryMiddleware keeps the public dashboard out of indexing.
// Crawlers need to fetch this header to honor noindex; it is not access control.
func MetricsDiscoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex, nofollow, nosnippet, noarchive")
		w.Header().Set("Content-Signal", "search=no, ai-train=no, ai-input=no")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// HandleMetrics returns public aggregate counts. Cache the database snapshot
// for one minute and serialize refreshes so anonymous traffic cannot force a
// full-table scan on every request. No visitor identifiers leave this handler.
func (h *Handler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	h.metricsMu.Lock()
	m := h.metricsCache
	var err error
	if time.Now().After(h.metricsExpires) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		m, err = db.GetSiteMetrics(ctx, h.db)
		cancel()
		if err == nil {
			h.metricsCache = m
			h.metricsExpires = time.Now().Add(time.Minute)
		}
	}
	h.metricsMu.Unlock()
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
