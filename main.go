package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sftrails/db"
	"sftrails/handlers"
)

func startVoteResetScheduler(database *sql.DB) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		time.Sleep(time.Until(next))
		deleted, err := db.ResetVotes(database)
		if err != nil {
			log.Printf("Midnight vote reset failed: %v", err)
		} else {
			log.Printf("Midnight vote reset: deleted %d votes", deleted)
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
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	if err := db.Initialize(database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	go startVoteResetScheduler(database)

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
		Addr:    ":8080",
		Handler: handlers.LoggingMiddleware(mux),
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	log.Println("SF Trails server starting on http://localhost:8080")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
