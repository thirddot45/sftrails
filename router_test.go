//go:build !postgres

package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sftrails/db"
)

func testRouter(t *testing.T) http.Handler {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	if err := db.Initialize(context.Background(), d); err != nil {
		t.Fatal(err)
	}
	h, err := newHTTPHandler(d, nil, "direct")
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestPublicMetricsDiscovery(t *testing.T) {
	t.Setenv("METRICS_USER", "")
	t.Setenv("METRICS_PASSWORD", "")
	h := testRouter(t)
	for _, tc := range []struct {
		path, accept string
		code         int
	}{
		{"/metrics", "text/html", 200},
		{"/metrics?source=direct", "text/markdown", 200},
		{"/metrics.md", "text/markdown", 308},
	} {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			req.Header.Set("Accept", tc.accept)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			if w.Code != tc.code {
				t.Fatalf("status %d, want %d", w.Code, tc.code)
			}
			if !strings.Contains(w.Header().Get("X-Robots-Tag"), "noindex") {
				t.Error("missing noindex")
			}
			if !strings.Contains(w.Header().Get("Cache-Control"), "no-store") {
				t.Error("metrics may be cached by clients")
			}
			if w.Header().Get("WWW-Authenticate") != "" {
				t.Error("unexpected password challenge")
			}
			if tc.code == 308 {
				if w.Header().Get("Location") != "/metrics" {
					t.Error("wrong redirect")
				}
				return
			}
			if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
				t.Error("metrics must stay HTML")
			}
			body := w.Body.String()
			if !strings.Contains(body, "Site Metrics") || !strings.Contains(body, `name="robots" content="noindex`) {
				t.Error("missing public dashboard or meta noindex")
			}
			for _, forbidden := range []string{`rel="alternate"`, `rel="canonical"`, "application/ld+json", `property="og:`, `name="twitter:`} {
				if strings.Contains(body, forbidden) {
					t.Errorf("unexpected discovery metadata: %s", forbidden)
				}
			}
		})
	}
}

func TestDiscoveryDocumentsOmitMetrics(t *testing.T) {
	h := testRouter(t)
	for _, path := range []string{"/sitemap.xml", "/llms.txt", "/llms-full.txt", "/.well-known/agent-skills/index.json", "/.well-known/agent-skills/sftrails-status/SKILL.md"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 {
			t.Fatalf("%s: status %d", path, w.Code)
		}
		if strings.Contains(w.Body.String(), "/metrics") {
			t.Errorf("%s advertises metrics", path)
		}
	}
}

func TestMetricsHeadersOnErrorAndSecurityHeaders(t *testing.T) {
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	d.Close()
	h, err := newHTTPHandler(d, nil, "direct")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/metrics", "/missing"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if path == "/metrics" && (w.Code != 500 || !strings.Contains(w.Header().Get("X-Robots-Tag"), "noindex")) {
			t.Error("metrics error lacks indexing protection")
		}
		for _, key := range []string{"Strict-Transport-Security", "Content-Security-Policy", "X-Content-Type-Options", "X-Frame-Options"} {
			if w.Header().Get(key) == "" {
				t.Errorf("%s lacks %s", path, key)
			}
		}
	}
}
