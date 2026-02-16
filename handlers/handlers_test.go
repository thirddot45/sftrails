package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"sftrails/db"
)

func setupTestHandler(t *testing.T) *Handler {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	if err := db.Initialize(d); err != nil {
		t.Fatalf("Failed to initialize test DB: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return NewHandler(d)
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
