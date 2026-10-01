//go:build !postgres

package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsAICrawlerIdentifiesAIFetchers(t *testing.T) {
	crawlers := []string{
		"Mozilla/5.0 (compatible; GPTBot/1.2; +https://openai.com/gptbot)",
		"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; ChatGPT-User/1.0",
		"Mozilla/5.0 (compatible; ClaudeBot/1.0; +claudebot@anthropic.com)",
		"Mozilla/5.0 (compatible; Claude-SearchBot/1.0)",
		"Claude-User/1.0",
		"Mozilla/5.0 (compatible; PerplexityBot/1.0)",
		"Perplexity-User/1.0",
		"Mozilla/5.0 (compatible; Bytespider; spider-feedback@bytedance.com)",
		"CCBot/2.0 (https://commoncrawl.org/faq/)",
		"meta-externalagent/1.1",
		"Mozilla/5.0 (Macintosh) Safari/605.1.15 Google-Extended",
		"Mozilla/5.0 (Macintosh) Applebot-Extended/0.1",
		"anthropic-ai",
		"cohere-ai",
		"OAI-SearchBot/1.0",
		"Amazonbot/0.1",
	}
	for _, ua := range crawlers {
		if !IsAICrawler(ua) {
			t.Errorf("IsAICrawler(%q) = false, want true", ua)
		}
	}
}

// TestIsAICrawlerAllowsSearchEnginesAndBrowsers is the load-bearing half of the
// design: search crawlers must be allowed through so they can fetch /metrics
// and read its X-Robots-Tag: noindex. Blocking them would hide that header and
// could leave the URL indexed from inbound links alone.
func TestIsAICrawlerAllowsSearchEnginesAndBrowsers(t *testing.T) {
	allowed := []string{
		"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		"Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
		"Mozilla/5.0 (compatible; Yahoo! Slurp; http://help.yahoo.com/help/us/ysearch/slurp)",
		"Mozilla/5.0 (compatible; Baiduspider/2.0)",
		"Mozilla/5.0 (compatible; YandexBot/3.0)",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Version/17.0 Safari/605.1.15 Applebot/0.1",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) Version/17.5 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:130.0) Gecko/20100101 Firefox/130.0",
		"curl/8.7.1",
		"",
	}
	for _, ua := range allowed {
		if IsAICrawler(ua) {
			t.Errorf("IsAICrawler(%q) = true, want false", ua)
		}
	}
}

func TestBlockAICrawlersMiddleware(t *testing.T) {
	cases := []struct {
		name       string
		userAgent  string
		wantStatus int
		wantCalled bool
	}{
		{"AI crawler", "Mozilla/5.0 (compatible; GPTBot/1.2)", http.StatusForbidden, false},
		{"search crawler", "Mozilla/5.0 (compatible; Googlebot/2.1)", http.StatusOK, true},
		{"browser", "Mozilla/5.0 (Macintosh) Chrome/131.0.0.0 Safari/537.36", http.StatusOK, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			called := false
			h := BlockAICrawlersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			}))
			req := httptest.NewRequest("GET", "/metrics", nil)
			req.Header.Set("User-Agent", c.userAgent)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			if w.Code != c.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, c.wantStatus)
			}
			if called != c.wantCalled {
				t.Errorf("handler called = %v, want %v", called, c.wantCalled)
			}
		})
	}
}

// TestAICrawlerListCoversRobotsTxt keeps the blocker and robots.txt from
// drifting apart: every named group that robots.txt disallows from /metrics
// must also be recognised by IsAICrawler, so a crawler that ignores robots.txt
// still gets turned away.
func TestAICrawlerListCoversRobotsTxt(t *testing.T) {
	h := setupTestHandler(t)
	oldDir, err := changeToProjectRoot()
	if err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer restoreDir(t, oldDir)

	req := httptest.NewRequest("GET", "/robots.txt", nil)
	w := httptest.NewRecorder()
	h.HandleRobotsTxt(w, req)

	checked := 0
	for _, g := range parseRobotsGroups(w.Body.String()) {
		if !g.disallowsMetrics {
			continue
		}
		for _, agent := range g.agents {
			if agent == "*" {
				continue
			}
			checked++
			if !IsAICrawler(agent) {
				t.Errorf("robots.txt disallows %q from /metrics, but IsAICrawler(%q) = false; "+
					"add it to aiCrawlerUserAgents", agent, agent)
			}
		}
	}
	if checked == 0 {
		t.Fatal("Expected robots.txt to name at least one crawler disallowed from /metrics")
	}
}
