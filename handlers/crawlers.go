package handlers

import (
	"net/http"
	"strings"

	"sftrails/templates"
)

// crawlerUserAgents lists lowercase User-Agent substrings for search engine and
// AI crawlers that do not carry one of the generic crawlerTokens below. Names
// like "ChatGPT-User", "Google-Extended" or "Slurp" are invisible to a
// bot/crawler/spider match, so they are enumerated explicitly.
var crawlerUserAgents = []string{
	// AI crawlers and assistant fetchers.
	"anthropic-ai",
	"applebot-extended",
	"chatgpt-user",
	"claude-user",
	"claude-searchbot",
	"cohere-ai",
	"google-extended",
	"gptbot",
	"meta-externalagent",
	"meta-externalfetcher",
	"mistralai-user",
	"oai-searchbot",
	"omgili",
	"perplexity-user",
	// Search engines.
	"applebot",
	"baiduspider",
	"slurp",
	"sogou",
	"yandex",
}

// crawlerTokens are generic markers that appear in nearly every automated
// crawler's User-Agent. Matching them catches bots that are not in the
// enumerated list above (including ones that do not exist yet) without
// affecting real browsers, none of which carry these words.
var crawlerTokens = []string{
	"bot",
	"crawler",
	"spider",
	"scraper",
}

// IsCrawler reports whether a User-Agent belongs to a search engine or AI
// crawler. An empty User-Agent is not treated as a crawler: unidentified
// clients are usually scripts or privacy tools operated by a human, and the
// noindex headers already cover anything that does index the response.
func IsCrawler(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	if ua == "" {
		return false
	}
	for _, token := range crawlerTokens {
		if strings.Contains(ua, token) {
			return true
		}
	}
	for _, name := range crawlerUserAgents {
		if strings.Contains(ua, name) {
			return true
		}
	}
	return false
}

// setNoIndexHeaders marks a response as off-limits to search indexes and AI
// crawlers. robots.txt keeps well-behaved crawlers from requesting the page at
// all; this covers the ones that fetch it anyway.
func setNoIndexHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Robots-Tag", templates.NoIndexDirectives)
}

// NoCrawlMiddleware serves a page to humans while refusing search engine and AI
// crawlers. Crawlers get 403; everyone else gets the page with headers telling
// any indexer that reaches it not to keep the content.
func NoCrawlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setNoIndexHeaders(w)
		if IsCrawler(r.UserAgent()) {
			http.Error(w, "Forbidden: this page is not available to crawlers", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
