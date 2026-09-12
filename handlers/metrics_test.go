//go:build !postgres

package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleMetricsRendersPublicly(t *testing.T) {
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
	if !strings.Contains(body, "Site Metrics") {
		t.Errorf("Expected metrics page body, got %q", body)
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

func TestMetricsSnapshotBoundsDatabaseReads(t *testing.T) {
	h := setupTestHandler(t)
	first := httptest.NewRecorder()
	h.HandleMetrics(first, httptest.NewRequest("GET", "/metrics", nil))
	if first.Code != 200 {
		t.Fatal("initial snapshot failed")
	}
	h.db.Close()
	cached := httptest.NewRecorder()
	h.HandleMetrics(cached, httptest.NewRequest("GET", "/metrics", nil))
	if cached.Code != 200 || cached.Body.String() != first.Body.String() {
		t.Error("repeated request did not reuse the cached snapshot")
	}
}
