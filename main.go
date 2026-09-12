package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"sftrails/db"
	"sftrails/weather"
)

func startVoteResetScheduler(ctx context.Context, database *sql.DB) {
	for {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
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
	dbDSN := os.Getenv("DATABASE_URL")
	if dbDSN == "" {
		dbDSN = os.Getenv("DB_PATH")
		if dbDSN == "" {
			dbDSN = "./sftrails.db"
		}
	}
	database, err := db.Open(dbDSN)
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

	// Build weather store from trail locations and start daily refresh.
	// Weather is non-critical: if trail load fails we start without it
	// rather than blocking app startup. The initial refresh runs inside
	// StartScheduler's goroutine so slow outbound calls don't delay the
	// HTTP listener.
	var ws *weather.Store
	if trails, err := db.GetTrailsWithStatus(ctx, database); err != nil {
		slog.Warn("skipping weather: failed to load trails", "error", err)
	} else {
		locs := make([]weather.Location, len(trails))
		for i, t := range trails {
			locs[i] = weather.Location{TrailID: t.ID, Lat: t.Latitude, Lng: t.Longitude}
		}
		ws = weather.NewStore(locs)
		go ws.StartScheduler(ctx)
	}

	appHandler, err := newHTTPHandler(database, ws, os.Getenv("CLIENT_IP_MODE"))
	if err != nil {
		slog.Error("invalid server configuration", "error", err)
		os.Exit(1)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 {
		slog.Error("PORT must be between 1 and 65535")
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           appHandler,
		MaxHeaderBytes:    32 << 10,
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

	slog.Info("server starting", "addr", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
