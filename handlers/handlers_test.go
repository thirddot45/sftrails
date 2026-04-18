//go:build !postgres

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"sftrails/db"
	"sftrails/models"
)

func changeToProjectRoot() (string, error) {
	old, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if err := os.Chdir(".."); err != nil {
		return "", err
	}
	return old, nil
}

func restoreDir(t *testing.T, dir string) {
	t.Helper()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("restore chdir: %v", err)
	}
}

func setupTestHandler(t *testing.T) *Handler {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	if err := db.Initialize(context.Background(), d); err != nil {
		t.Fatalf("Failed to initialize test DB: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return NewHandler(d, nil)
}

func TestHandleIndex(t *testing.T) {
	h := setupTestHandler(t)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	h.HandleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Markham Park") {
		t.Error("Expected response to contain 'Markham Park'")
	}
	if !strings.Contains(body, "SF Trails") {
		t.Error("Expected response to contain 'SF Trails'")
	}
}

func TestHandleTrailsList(t *testing.T) {
	h := setupTestHandler(t)
	req := httptest.NewRequest("GET", "/trails-list", nil)
	w := httptest.NewRecorder()

	h.HandleTrailsList(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "trails-list") {
		t.Error("Expected response to contain 'trails-list' div")
	}
}

func TestHandleVote(t *testing.T) {
	h := setupTestHandler(t)

	form := url.Values{}
	form.Set("trail_id", "1")
	form.Set("vote", "open")
	form.Set("fingerprint", "testfp")

	req := httptest.NewRequest("POST", "/vote", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.HandleVote(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "trail-1") {
		t.Error("Expected response to contain 'trail-1'")
	}
}

func TestHandleVoteInvalidTrailID(t *testing.T) {
	h := setupTestHandler(t)

	form := url.Values{}
	form.Set("trail_id", "invalid")
	form.Set("vote", "open")
	form.Set("fingerprint", "testfp")

	req := httptest.NewRequest("POST", "/vote", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.HandleVote(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestHandleVoteInvalidVoteType(t *testing.T) {
	h := setupTestHandler(t)

	form := url.Values{}
	form.Set("trail_id", "1")
	form.Set("vote", "maybe")
	form.Set("fingerprint", "testfp")

	req := httptest.NewRequest("POST", "/vote", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.HandleVote(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	for i := range 3 {
		if !rl.Allow("127.0.0.1") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	if rl.Allow("127.0.0.1") {
		t.Error("Request 4 should be rate limited")
	}

	// Different IP should still be allowed
	if !rl.Allow("192.168.1.1") {
		t.Error("Different IP should be allowed")
	}
}

func TestHandleRobotsTxt(t *testing.T) {
	h := setupTestHandler(t)
	req := httptest.NewRequest("GET", "/robots.txt", nil)
	w := httptest.NewRecorder()

	h.HandleRobotsTxt(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "User-agent: *") {
		t.Error("Expected robots.txt to contain 'User-agent: *'")
	}
	if !strings.Contains(body, "Sitemap:") {
		t.Error("Expected robots.txt to contain 'Sitemap:'")
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Errorf("Expected Content-Type text/plain, got %s", ct)
	}
}

func TestHandleSitemap(t *testing.T) {
	h := setupTestHandler(t)
	req := httptest.NewRequest("GET", "/sitemap.xml", nil)
	w := httptest.NewRecorder()

	h.HandleSitemap(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "<urlset") {
		t.Error("Expected sitemap to contain '<urlset'")
	}
	if !strings.Contains(body, "sftrails.info") {
		t.Error("Expected sitemap to contain 'sftrails.info'")
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/xml") {
		t.Errorf("Expected Content-Type application/xml, got %s", ct)
	}
}

func TestHandleSignatureDirectory(t *testing.T) {
	h := setupTestHandler(t)
	req := httptest.NewRequest("GET", "/.well-known/http-message-signatures-directory", nil)
	w := httptest.NewRecorder()

	// The handler resolves the JWKS relative to the repo root; tests run
	// from handlers/, so temporarily chdir up one level.
	oldDir, err := changeToProjectRoot()
	if err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer restoreDir(t, oldDir)

	h.HandleSignatureDirectory(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "http-message-signatures-directory") {
		t.Errorf("Expected Content-Type http-message-signatures-directory+json, got %s", ct)
	}

	var set struct {
		Keys []struct {
			Kty, Crv, Alg, Kid, X string
		} `json:"keys"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &set); err != nil {
		t.Fatalf("Failed to decode JWKS: %v", err)
	}
	if len(set.Keys) == 0 {
		t.Fatal("Expected at least one key in JWKS")
	}
	k := set.Keys[0]
	if k.Kty != "OKP" || k.Crv != "Ed25519" || k.Alg != "EdDSA" {
		t.Errorf("Unexpected key params: %+v", k)
	}
	if k.Kid == "" || k.X == "" {
		t.Errorf("Expected non-empty kid and x, got %+v", k)
	}
}

func TestHandleIndexSEO(t *testing.T) {
	h := setupTestHandler(t)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	h.HandleIndex(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `<meta name="description"`) {
		t.Error("Expected meta description tag")
	}
	if !strings.Contains(body, `og:title`) {
		t.Error("Expected Open Graph title meta tag")
	}
	if !strings.Contains(body, `twitter:card`) {
		t.Error("Expected Twitter Card meta tag")
	}
	if !strings.Contains(body, `application/ld+json`) {
		t.Error("Expected JSON-LD structured data")
	}
	if !strings.Contains(body, `rel="canonical"`) {
		t.Error("Expected canonical link tag")
	}
}

func TestHandleAPITrails(t *testing.T) {
	h := setupTestHandler(t)
	req := httptest.NewRequest("GET", "/api/trails", nil)
	w := httptest.NewRecorder()

	h.HandleAPITrails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Expected JSON content type, got %s", ct)
	}

	var resp models.TrailsAPIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp.Trails) == 0 {
		t.Error("Expected trails in response")
	}

	if resp.GeneratedAt == "" {
		t.Error("Expected generated_at timestamp")
	}

	found := false
	for _, trail := range resp.Trails {
		if trail.Name == "Markham Park" {
			found = true
			if trail.City != "Sunrise" {
				t.Errorf("Expected city 'Sunrise', got '%s'", trail.City)
			}
			if trail.Latitude == 0 || trail.Longitude == 0 {
				t.Error("Expected non-zero coordinates")
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find 'Markham Park' in trails")
	}
}

func TestHandleAPITrailSingle(t *testing.T) {
	h := setupTestHandler(t)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/trails/{id}", h.HandleAPITrail)

	req := httptest.NewRequest("GET", "/api/trails/1", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var trail models.TrailAPIResponse
	if err := json.NewDecoder(w.Body).Decode(&trail); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if trail.Name != "Markham Park" {
		t.Errorf("Expected 'Markham Park', got '%s'", trail.Name)
	}
}

func TestHandleAPITrailNotFound(t *testing.T) {
	h := setupTestHandler(t)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/trails/{id}", h.HandleAPITrail)

	req := httptest.NewRequest("GET", "/api/trails/99999", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}

func TestHandleAPITrailInvalidID(t *testing.T) {
	h := setupTestHandler(t)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/trails/{id}", h.HandleAPITrail)

	req := httptest.NewRequest("GET", "/api/trails/abc", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestRobotsTxtAICrawlers(t *testing.T) {
	h := setupTestHandler(t)
	req := httptest.NewRequest("GET", "/robots.txt", nil)
	w := httptest.NewRecorder()

	h.HandleRobotsTxt(w, req)

	body := w.Body.String()
	crawlers := []string{"GPTBot", "ClaudeBot", "Claude-SearchBot", "PerplexityBot", "ChatGPT-User", "Google-Extended", "Applebot-Extended", "Amazonbot"}
	for _, crawler := range crawlers {
		if !strings.Contains(body, crawler) {
			t.Errorf("Expected robots.txt to contain AI crawler directive for %s", crawler)
		}
	}
	if !strings.Contains(body, "/api/") {
		t.Error("Expected robots.txt to allow /api/")
	}
	if !strings.Contains(body, "llms.txt") {
		t.Error("Expected robots.txt to reference llms.txt")
	}
}

func TestVotePersistsOnRefresh(t *testing.T) {
	h := setupTestHandler(t)

	// Cast a vote
	form := url.Values{}
	form.Set("trail_id", "1")
	form.Set("vote", "open")
	form.Set("fingerprint", "refresh-test-fp")

	req := httptest.NewRequest("POST", "/vote", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.HandleVote(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Vote failed with status %d", w.Code)
	}

	// "Refresh" by requesting the index page again
	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	h.HandleIndex(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "1/0") {
		t.Error("Expected vote count '1/0' to persist after refresh")
	}
}

func TestNoCacheHeadersOnHTMLPages(t *testing.T) {
	h := setupTestHandler(t)

	t.Run("index", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		h.HandleIndex(w, req)

		cc := w.Header().Get("Cache-Control")
		if !strings.Contains(cc, "no-cache") || !strings.Contains(cc, "no-store") {
			t.Errorf("Expected Cache-Control no-cache/no-store on index, got %q", cc)
		}
	})

	t.Run("trails-list", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/trails-list", nil)
		w := httptest.NewRecorder()
		h.HandleTrailsList(w, req)

		cc := w.Header().Get("Cache-Control")
		if !strings.Contains(cc, "no-cache") || !strings.Contains(cc, "no-store") {
			t.Errorf("Expected Cache-Control no-cache/no-store on trails-list, got %q", cc)
		}
	})
}

func TestMarkdownNegotiationIndex(t *testing.T) {
	h := setupTestHandler(t)
	handler := MarkdownNegotiationMiddleware(http.HandlerFunc(h.HandleIndex))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "text/markdown")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/markdown") {
		t.Errorf("Expected Content-Type text/markdown, got %q", ct)
	}
	if tok := w.Header().Get("X-Markdown-Tokens"); tok == "" {
		t.Error("Expected X-Markdown-Tokens header to be set")
	}
	if vary := w.Header().Get("Vary"); !strings.Contains(vary, "Accept") {
		t.Errorf("Expected Vary to include Accept, got %q", vary)
	}
	body := w.Body.String()
	if strings.Contains(body, "<html") || strings.Contains(body, "<div") {
		t.Errorf("Expected markdown body, but found HTML tags: %q", body[:min(200, len(body))])
	}
	if !strings.Contains(body, "Markham Park") {
		t.Error("Expected markdown body to contain 'Markham Park'")
	}
}

func TestMarkdownNegotiationPassthroughHTML(t *testing.T) {
	h := setupTestHandler(t)
	handler := MarkdownNegotiationMiddleware(http.HandlerFunc(h.HandleIndex))

	req := httptest.NewRequest("GET", "/", nil)
	// No Accept header - browsers should get HTML
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "<html") {
		t.Error("Expected HTML response for default Accept")
	}
	if ct := w.Header().Get("Content-Type"); strings.Contains(ct, "text/markdown") {
		t.Errorf("Did not expect text/markdown, got %q", ct)
	}
}

func TestMarkdownNegotiationMultipleAccept(t *testing.T) {
	h := setupTestHandler(t)
	handler := MarkdownNegotiationMiddleware(http.HandlerFunc(h.HandleIndex))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "text/html, text/markdown;q=0.9")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/markdown") {
		t.Errorf("Expected Content-Type text/markdown when client includes it, got %q", ct)
	}
}

func TestGetIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remote   string
		expected string
	}{
		{
			name:     "X-Forwarded-For",
			headers:  map[string]string{"X-Forwarded-For": "1.2.3.4, 5.6.7.8"},
			remote:   "127.0.0.1:1234",
			expected: "1.2.3.4",
		},
		{
			name:     "X-Real-IP",
			headers:  map[string]string{"X-Real-IP": "1.2.3.4"},
			remote:   "127.0.0.1:1234",
			expected: "1.2.3.4",
		},
		{
			name:     "RemoteAddr",
			headers:  map[string]string{},
			remote:   "127.0.0.1:1234",
			expected: "127.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remote
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			if got := GetIP(req); got != tt.expected {
				t.Errorf("GetIP() = %s, want %s", got, tt.expected)
			}
		})
	}
}
