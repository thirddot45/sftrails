package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"sftrails/db"
	"sftrails/models"
	"sftrails/templates"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(database *sql.DB) *Handler {
	return &Handler{db: database}
}

func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	trails, err := db.GetTrailsWithStatus(h.db)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	templates.IndexPage(trails).Render(r.Context(), w)
}

func (h *Handler) HandleTrailsList(w http.ResponseWriter, r *http.Request) {
	trails, err := db.GetTrailsWithStatus(h.db)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	templates.TrailsList(trails).Render(r.Context(), w)
}

func (h *Handler) HandleVote(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	trailIDStr := r.FormValue("trail_id")
	voteStr := r.FormValue("vote")
	fingerprint := r.FormValue("fingerprint")

	trailID, err := strconv.ParseInt(trailIDStr, 10, 64)
	if err != nil || trailID <= 0 {
		http.Error(w, "Invalid trail ID", http.StatusBadRequest)
		return
	}

	var voteType models.VoteType
	switch voteStr {
	case "open":
		voteType = models.VoteOpen
	case "closed":
		voteType = models.VoteClosed
	default:
		http.Error(w, "Invalid vote type", http.StatusBadRequest)
		return
	}

	ip := GetIP(r)
	if err := db.CastVote(h.db, trailID, voteType, ip, fingerprint); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	trail, err := db.GetTrailWithStatus(h.db, trailID)
	if err != nil || trail == nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	templates.TrailCard(*trail).Render(r.Context(), w)
}

func (h *Handler) HandleAPITrails(w http.ResponseWriter, r *http.Request) {
	trails, err := db.GetTrailsWithStatus(h.db)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	apiTrails := make([]models.TrailAPIResponse, len(trails))
	for i, t := range trails {
		apiTrails[i] = models.TrailToAPI(t)
	}

	resp := models.TrailsAPIResponse{
		Trails:      apiTrails,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=60")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleAPITrail(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"invalid trail ID"}`, http.StatusBadRequest)
		return
	}

	trail, err := db.GetTrailWithStatus(h.db, id)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	if trail == nil {
		http.Error(w, `{"error":"trail not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=60")
	json.NewEncoder(w).Encode(models.TrailToAPI(*trail))
}

func (h *Handler) HandleRobotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, `User-agent: *
Allow: /
Allow: /api/
Allow: /llms.txt
Allow: /llms-full.txt
Disallow: /vote
Disallow: /trails-list

Sitemap: /sitemap.xml

User-agent: GPTBot
Allow: /

User-agent: ChatGPT-User
Allow: /

User-agent: ClaudeBot
Allow: /

User-agent: Claude-SearchBot
Allow: /

User-agent: PerplexityBot
Allow: /

User-agent: Google-Extended
Allow: /

User-agent: Applebot-Extended
Allow: /

User-agent: Amazonbot
Allow: /

User-agent: Bytespider
Allow: /

User-agent: cohere-ai
Allow: /
`)
}

func (h *Handler) HandleSitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	lastmod := time.Now().UTC().Format("2006-01-02")
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://sftrails.com/</loc>
    <lastmod>%s</lastmod>
    <changefreq>hourly</changefreq>
    <priority>1.0</priority>
  </url>
</urlset>`, lastmod)
}
