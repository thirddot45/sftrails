//go:build !postgres

package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleMetricsIsPublicAndNoIndex(t *testing.T) {
	h := setupTestHandler(t)
	dir, err := changeToProjectRoot()
	if err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer restoreDir(t, dir)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	h.HandleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "SF Trails Metrics") {
		t.Errorf("Expected metrics page body, got %q", body)
	}
	if got := w.Header().Get("X-Robots-Tag"); !strings.Contains(got, "noindex") || !strings.Contains(got, "noai") {
		t.Errorf("Expected X-Robots-Tag to opt out of search and AI indexing, got %q", got)
	}
	if !strings.Contains(body, `<meta name="robots"`) || !strings.Contains(body, "noindex") {
		t.Error("Expected a robots noindex meta tag in the metrics page head")
	}
	// No discovery markup: nothing here should invite a crawler to index,
	// preview, or follow the page.
	for _, marker := range []string{`rel="canonical"`, `type="text/markdown"`, "og:image", "og:title", "ld+json", "twitter:card"} {
		if strings.Contains(body, marker) {
			t.Errorf("Metrics page must not emit discovery markup %q", marker)
		}
	}
}

func TestHandleMetricsRendersOnlyDataPoints(t *testing.T) {
	h := setupTestHandler(t)
	dir, err := changeToProjectRoot()
	if err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer restoreDir(t, dir)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	h.HandleMetrics(w, req)
	body := w.Body.String()

	// Every aggregate the dashboard exists to report is present.
	for _, label := range []string{"Total visits", "Unique visitors", "Visits today", "Unique today", "Last 7 days", "Top pages"} {
		if !strings.Contains(body, label) {
			t.Errorf("Expected data point %q in metrics page", label)
		}
	}

	// ...and nothing else: no styling, no scripts, no site chrome.
	for _, marker := range []string{"tailwind", "htmx", "<script", "<style", "fingerprint.js", "sort.js", "class=", "<header", "<footer", "South Florida Mountain Bike Trail Status"} {
		if strings.Contains(strings.ToLower(body), strings.ToLower(marker)) {
			t.Errorf("Metrics page should be bare data points, but contains %q", marker)
		}
	}
}

func TestShouldTrackFiltersNonPageRequests(t *testing.T) {
	cases := []struct {
		method, path string
		want         bool
	}{
		{"GET", "/", true},
		{"GET", "/trail/markham-park", true},
		{"GET", "/status", true},
		{"GET", "/metrics", false},
		{"GET", "/static/sort.js", false},
		{"GET", "/api/trails", false},
		{"GET", "/.well-known/agent-skills/index.json", false},
		{"GET", "/llms.txt", false},
		{"GET", "/robots.txt", false},
		{"GET", "/sitemap.xml", false},
		{"POST", "/vote", false},
	}
	for _, c := range cases {
		req := httptest.NewRequest(c.method, c.path, nil)
		if got := shouldTrack(req); got != c.want {
			t.Errorf("shouldTrack(%s %s) = %v, want %v", c.method, c.path, got, c.want)
		}
	}
}

func TestVisitorHashOmitsRawIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.7:12345"
	req.Header.Set("User-Agent", "test-agent")

	hash := visitorHash(req)
	if strings.Contains(hash, "203.0.113.7") {
		t.Error("visitor hash must not contain the raw IP address")
	}
	// Deterministic for the same visitor.
	if hash != visitorHash(req) {
		t.Error("visitor hash should be stable for the same visitor")
	}
}
