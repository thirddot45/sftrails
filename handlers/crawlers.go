package handlers

import (
	"net/http"
	"strings"
)

// aiCrawlerUserAgents lists lowercase User-Agent substrings for the AI crawlers
// and assistant fetchers that robots.txt disallows from /metrics. It must stay
// in step with the named groups in static/robots.txt; a test enforces that.
//
// Ordinary search crawlers are deliberately absent. Search engines are kept out
// of the index by X-Robots-Tag: noindex, which they can only honor if they are
// allowed to fetch the page and read the header — so blocking them here would
// defeat the very thing it looks like it is enforcing. Matching is also done
// against explicit names rather than a generic "bot" token for the same reason:
// "bot" would catch Googlebot.
var aiCrawlerUserAgents = []string{
	"amazonbot",
	"anthropic-ai",
	"applebot-extended",
	"bytespider",
	"ccbot",
	"chatgpt-user",
	"claude-searchbot",
	"claude-user",
	"claudebot",
	"cohere-ai",
	"google-extended",
	"gptbot",
	"meta-externalagent",
	"meta-externalfetcher",
	"mistralai-user",
	"oai-searchbot",
	"omgili",
	"perplexity-user",
	"perplexitybot",
}

// IsAICrawler reports whether a User-Agent belongs to a known AI crawler. An
// empty User-Agent is not treated as one: unidentified clients are usually
// scripts or privacy tools driven by a person.
func IsAICrawler(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	if ua == "" {
		return false
	}
	for _, name := range aiCrawlerUserAgents {
		if strings.Contains(ua, name) {
			return true
		}
	}
	return false
}

// BlockAICrawlersMiddleware refuses requests from known AI crawlers with 403.
//
// robots.txt already disallows these crawlers from /metrics, but that is a
// cooperative control; this enforces it for the ones that ask anyway. Search
// crawlers and ordinary visitors pass through untouched, so the noindex
// signalling in MetricsDiscoveryMiddleware still reaches the clients that need
// to read it.
func BlockAICrawlersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsAICrawler(r.UserAgent()) {
			http.Error(w, "Forbidden: this page is not available to AI crawlers", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
