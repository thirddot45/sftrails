package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sftrails/db"
	"sftrails/handlers"
)

func startVoteResetScheduler(ctx context.Context, database *sql.DB) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		deleted, err := db.ResetVotes(ctx, database)
		if err != nil {
			slog.Error("midnight vote reset failed", "error", err)
		} else {
			slog.Info("midnight vote reset", "deleted", deleted)
		}
	}
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./sftrails.db"
	}
	database, err := db.Open(dbPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := db.Initialize(ctx, database); err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	go startVoteResetScheduler(ctx, database)

	h := handlers.NewHandler(database)
	rl := handlers.NewRateLimiter(30, time.Minute)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", h.HandleIndex)
	mux.HandleFunc("GET /trails-list", h.HandleTrailsList)
	mux.Handle("POST /vote", rl.Middleware(http.HandlerFunc(h.HandleVote)))
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

	server := &http.Server{
		Addr:              ":8080",
		Handler:           handlers.LoggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("shutting down")
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		server.Shutdown(shutdownCtx)
	}()

	slog.Info("server starting", "addr", "http://localhost:8080")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
