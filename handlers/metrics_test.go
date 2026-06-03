//go:build !postgres

package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBasicAuthMiddlewareRejectsMissingCredentials(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
	h := BasicAuthMiddleware("test", next)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
	if called {
		t.Error("Expected next handler not to be called without credentials")
	}
	if !strings.Contains(w.Header().Get("WWW-Authenticate"), "Basic") {
		t.Errorf("Expected WWW-Authenticate challenge, got %q", w.Header().Get("WWW-Authenticate"))
	}
}

func TestBasicAuthMiddlewareRejectsWrongPassword(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h := BasicAuthMiddleware("test", next)

	req := httptest.NewRequest("GET", "/metrics", nil)
	req.SetBasicAuth("admin", "wrong")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for wrong password, got %d", w.Code)
	}
}

func TestBasicAuthMiddlewareAcceptsCorrectPassword(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
	h := BasicAuthMiddleware("test", next)

	req := httptest.NewRequest("GET", "/metrics", nil)
	req.SetBasicAuth("anyuser", defaultMetricsPassword)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if !called {
		t.Error("Expected next handler to be called with correct password")
	}
}

func TestHandleMetricsRendersBehindAuth(t *testing.T) {
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
