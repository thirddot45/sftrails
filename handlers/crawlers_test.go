//go:build !postgres

package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsCrawlerIdentifiesSearchAndAIBots(t *testing.T) {
	crawlers := []string{
		"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		"Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
		"Mozilla/5.0 (compatible; GPTBot/1.2; +https://openai.com/gptbot)",
		"Mozilla/5.0 (compatible; ClaudeBot/1.0; +claudebot@anthropic.com)",
		"Mozilla/5.0 (compatible; PerplexityBot/1.0)",
		"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; ChatGPT-User/1.0",
		"Mozilla/5.0 (compatible; Bytespider; spider-feedback@bytedance.com)",
		"Mozilla/5.0 (compatible; Yahoo! Slurp; http://help.yahoo.com/help/us/ysearch/slurp)",
		"Mozilla/5.0 (compatible; Baiduspider/2.0)",
		"Mozilla/5.0 (compatible; YandexBot/3.0)",
		"anthropic-ai",
		"cohere-ai",
		"Mozilla/5.0 (Macintosh) AppleWebKit/605.1.15 Version/17.0 Safari/605.1.15 Google-Extended",
		"meta-externalagent/1.1",
		"CCBot/2.0 (https://commoncrawl.org/faq/)",
		"SomeFutureCrawler/1.0",
	}
	for _, ua := range crawlers {
		if !IsCrawler(ua) {
			t.Errorf("IsCrawler(%q) = false, want true", ua)
		}
	}
}

func TestIsCrawlerAllowsRealBrowsers(t *testing.T) {
	browsers := []string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:130.0) Gecko/20100101 Firefox/130.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0",
		"curl/8.7.1",
		// Unidentified clients are treated as human; noindex headers still apply.
		"",
	}
	for _, ua := range browsers {
		if IsCrawler(ua) {
			t.Errorf("IsCrawler(%q) = true, want false", ua)
		}
	}
}

func TestNoCrawlMiddlewareBlocksCrawlers(t *testing.T) {
	called := false
	h := NoCrawlMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))

	req := httptest.NewRequest("GET", "/metrics", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GPTBot/1.2; +https://openai.com/gptbot)")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403 for a crawler, got %d", w.Code)
	}
	if called {
		t.Error("Expected the wrapped handler not to run for a crawler")
	}
	if got := w.Header().Get("X-Robots-Tag"); !strings.Contains(got, "noindex") {
		t.Errorf("Expected X-Robots-Tag on the blocked response, got %q", got)
	}
}

func TestNoCrawlMiddlewareServesHumansWithNoIndexHeader(t *testing.T) {
	called := false
	h := NoCrawlMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))

	req := httptest.NewRequest("GET", "/metrics", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 for a browser, got %d", w.Code)
	}
	if !called {
		t.Error("Expected the wrapped handler to run for a browser")
	}
	got := w.Header().Get("X-Robots-Tag")
	for _, directive := range []string{"noindex", "nofollow", "noai"} {
		if !strings.Contains(got, directive) {
			t.Errorf("Expected X-Robots-Tag to contain %q, got %q", directive, got)
		}
	}
}

func TestMetricsHasNoMarkdownVariant(t *testing.T) {
	// /metrics.md must not fall through to the HTML dashboard; the router
	// should never see a rewritten path for it.
	var seen string
	h := MarkdownSuffixMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Path
	}))

	req := httptest.NewRequest("GET", "/metrics.md", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if seen != "/metrics.md" {
		t.Errorf("Expected /metrics.md to pass through unrewritten (so it 404s), got %q", seen)
	}

	// Pages that do publish markdown are still rewritten.
	req = httptest.NewRequest("GET", "/trail/markham-park.md", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if seen != "/trail/markham-park" {
		t.Errorf("Expected /trail/markham-park.md to be rewritten, got %q", seen)
	}
}
