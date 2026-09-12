package main

import (
	"database/sql"
	"net/http"
	"time"

	"sftrails/handlers"
	"sftrails/weather"
)

func newHTTPHandler(database *sql.DB, ws *weather.Store, clientIPMode string) (http.Handler, error) {
	clientIP, err := handlers.ClientIPMiddleware(clientIPMode)
	if err != nil {
		return nil, err
	}
	h := handlers.NewHandler(database, ws)
	rl := handlers.NewRateLimiter(30, time.Minute)

	md := handlers.MarkdownNegotiationMiddleware

	mux := http.NewServeMux()
	mux.Handle("GET /{$}", md(http.HandlerFunc(h.HandleIndex)))
	mux.Handle("GET /trail/{slug}", md(http.HandlerFunc(h.HandleTrailDetail)))
	mux.Handle("GET /trails-list", md(http.HandlerFunc(h.HandleTrailsList)))
	mux.Handle("POST /vote", rl.Middleware(md(http.HandlerFunc(h.HandleVote))))
	mux.Handle("GET /status", md(http.HandlerFunc(h.HandleStatus)))
	mux.Handle("GET /metrics", handlers.MetricsDiscoveryMiddleware(http.HandlerFunc(h.HandleMetrics)))
	mux.Handle("GET /metrics.md", handlers.MetricsDiscoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusPermanentRedirect)
	})))
	mux.HandleFunc("GET /.well-known/http-message-signatures-directory", h.HandleSignatureDirectory)
	mux.HandleFunc("GET /.well-known/agent-skills/index.json", h.HandleAgentSkillsIndex)
	mux.HandleFunc("GET /.well-known/agent-skills/{path...}", h.HandleAgentSkillFile)
	mux.HandleFunc("GET /robots.txt", h.HandleRobotsTxt)
	mux.HandleFunc("GET /sitemap.xml", h.HandleSitemap)
	mux.HandleFunc("GET /api/trails", h.HandleAPITrails)
	mux.HandleFunc("GET /api/trails/{id}", h.HandleAPITrail)
	mux.HandleFunc("GET /llms.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/llms.txt")
	})
	mux.HandleFunc("GET /llms-full.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/llms-full.txt")
	})
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	return handlers.SecurityHeadersMiddleware(clientIP(handlers.LoggingMiddleware(handlers.MarkdownSuffixMiddleware(handlers.MetricsMiddleware(database)(mux))))), nil
}
