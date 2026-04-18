package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"sftrails/db"
	"sftrails/models"
	"sftrails/templates"
	"sftrails/weather"
)

type Handler struct {
	db      *sql.DB
	weather *weather.Store
}

func NewHandler(database *sql.DB, ws *weather.Store) *Handler {
	return &Handler{db: database, weather: ws}
}

// attachWeatherOne populates weather fields on a single trail from the store.
func (h *Handler) attachWeatherOne(trail *models.TrailWithStatus) {
	if h.weather == nil || trail == nil {
		return
	}
	f, ok := h.weather.Get(trail.ID)
	if !ok {
		return
	}
	trail.HasWeather = true
	trail.WeatherIcon = f.Icon
	trail.WeatherDesc = f.Desc
	trail.WeatherTempHighF = f.TempHighF
	trail.WeatherRainPct = f.RainChance
}

// attachWeather populates weather fields on each trail from the weather store.
func (h *Handler) attachWeather(trails []models.TrailWithStatus) {
	if h.weather == nil {
		return
	}
	for i := range trails {
		h.attachWeatherOne(&trails[i])
	}
}

func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	trails, err := db.GetTrailsWithStatus(r.Context(), h.db)
	if err != nil {
		slog.Error("failed to get trails", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	loc := ReadUserLocation(r)
	AttachDistanceAndSort(trails, loc)
	h.attachWeather(trails)
	sortLabel := ""
	if loc != nil {
		sortLabel = loc.Label
	}
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Add("Link", `</api/trails>; rel="service-desc"; type="application/json"`)
	w.Header().Add("Link", `</llms.txt>; rel="service-doc"; type="text/plain"`)
	w.Header().Add("Link", `</llms-full.txt>; rel="service-doc"; type="text/plain"`)
	w.Header().Add("Link", `</sitemap.xml>; rel="sitemap"; type="application/xml"`)
	if err := templates.IndexPage(trails, sortLabel).Render(r.Context(), w); err != nil {
		slog.Error("failed to render index", "error", err)
	}
}

func (h *Handler) HandleTrailsList(w http.ResponseWriter, r *http.Request) {
	trails, err := db.GetTrailsWithStatus(r.Context(), h.db)
	if err != nil {
		slog.Error("failed to get trails", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	AttachDistanceAndSort(trails, ReadUserLocation(r))
	h.attachWeather(trails)
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if err := templates.TrailsList(trails).Render(r.Context(), w); err != nil {
		slog.Error("failed to render trails list", "error", err)
	}
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

	ctx := r.Context()
	ip := GetIP(r)
	if err := db.CastVote(ctx, h.db, trailID, voteType, ip, fingerprint); err != nil {
		slog.Error("failed to cast vote", "trail_id", trailID, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	trail, err := db.GetTrailWithStatus(ctx, h.db, trailID)
	if err != nil || trail == nil {
		slog.Error("failed to get trail after vote", "trail_id", trailID, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.attachWeatherOne(trail)
	trails := []models.TrailWithStatus{*trail}
	AttachDistanceAndSort(trails, ReadUserLocation(r))
	*trail = trails[0]

	if err := templates.TrailCard(*trail).Render(ctx, w); err != nil {
		slog.Error("failed to render trail card", "trail_id", trailID, "error", err)
	}
}

func (h *Handler) HandleAPITrails(w http.ResponseWriter, r *http.Request) {
	trails, err := db.GetTrailsWithStatus(r.Context(), h.db)
	if err != nil {
		slog.Error("failed to get trails for API", "error", err)
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
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode trails response", "error", err)
	}
}

func (h *Handler) HandleAPITrail(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"invalid trail ID"}`, http.StatusBadRequest)
		return
	}

	trail, err := db.GetTrailWithStatus(r.Context(), h.db, id)
	if err != nil {
		slog.Error("failed to get trail for API", "trail_id", id, "error", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	if trail == nil {
		http.Error(w, `{"error":"trail not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=60")
	if err := json.NewEncoder(w).Encode(models.TrailToAPI(*trail)); err != nil {
		slog.Error("failed to encode trail response", "trail_id", id, "error", err)
	}
}

func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	data := templates.StatusData{
		Endpoints: weather.CheckEndpoints(),
	}
	if h.weather != nil {
		cached, total, refresh := h.weather.Stats()
		data.WeatherCached = cached
		data.WeatherTotal = total
		data.WeatherRefresh = refresh
	}
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if err := templates.StatusPage(data).Render(r.Context(), w); err != nil {
		slog.Error("failed to render status", "error", err)
	}
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
    <loc>https://sftrails.info/</loc>
    <lastmod>%s</lastmod>
    <changefreq>hourly</changefreq>
    <priority>1.0</priority>
  </url>
</urlset>`, lastmod)
}
