//go:build !postgres

package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVoteInputGuards(t *testing.T) {
	h := setupTestHandler(t)
	for _, tc := range []struct {
		name, body, origin, site string
		want                     int
	}{
		{"oversized", "trail_id=1&vote=open&fingerprint=" + strings.Repeat("x", 5000), "", "", 413},
		{"long fingerprint", "trail_id=1&vote=open&fingerprint=" + strings.Repeat("x", 65), "", "", 400},
		{"cross origin", "trail_id=1&vote=open", "https://other.example", "", 403},
		{"opaque origin", "trail_id=1&vote=open", "null", "", 403},
		{"cross site", "trail_id=1&vote=open", "", "cross-site", 403},
		{"missing trail", "trail_id=999999&vote=open", "", "", 404},
		{"same origin", "trail_id=1&vote=open", "https://sftrails.info", "same-origin", 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "https://sftrails.info/vote", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Sec-Fetch-Site", tc.site)
			w := httptest.NewRecorder()
			h.HandleVote(w, req)
			if w.Code != tc.want {
				t.Errorf("status %d, want %d", w.Code, tc.want)
			}
		})
	}
	var votes int
	if err := h.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM votes").Scan(&votes); err != nil {
		t.Fatal(err)
	}
	if votes != 1 {
		t.Errorf("invalid submissions wrote votes: %d", votes)
	}
}

func TestLocationRejectsNonFiniteValues(t *testing.T) {
	for _, value := range []string{"NaN,0", "0,NaN", "Inf,0", "0,-Inf"} {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "sft_loc", Value: value})
		if ReadUserLocation(req) != nil {
			t.Errorf("accepted %s", value)
		}
	}
}
