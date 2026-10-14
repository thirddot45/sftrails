//go:build !postgres

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

func TestReportSummaryAcrossPageAndHTMXResponses(t *testing.T) {
	h := setupTestHandler(t)
	form := url.Values{"trail_id": {"1"}, "vote": {"open"}, "fingerprint": {"report-summary"}}
	req := httptest.NewRequest("POST", "/vote", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	voteResponse := httptest.NewRecorder()
	h.HandleVote(voteResponse, req)
	if voteResponse.Code != http.StatusOK {
		t.Fatalf("vote status: %d", voteResponse.Code)
	}
	trail, err := db.GetTrailWithStatus(req.Context(), h.db, 1)
	if err != nil || trail == nil || trail.LastReportAt.IsZero() {
		t.Fatalf("missing report time: %v, %v", trail, err)
	}
	wantTime := `datetime="` + trail.LastReportAt.UTC().Format(time.RFC3339) + `"`
	check := func(body string) {
		t.Helper()
		for _, want := range []string{"1 rideable · 0 closed", "(past 12h)", "Last report:", wantTime} {
			if !strings.Contains(body, want) {
				t.Errorf("missing %q", want)
			}
		}
	}
	check(voteResponse.Body.String())
	for _, handler := range []http.HandlerFunc{h.HandleIndex, h.HandleTrailsList} {
		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest("GET", "/", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("read status: %d", w.Code)
		}
		check(w.Body.String())
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /trail/{slug}", h.HandleTrailDetail)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/trail/markham-park", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("detail status: %d", w.Code)
	}
	check(w.Body.String())
}
