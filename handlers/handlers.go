package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
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
	w.Header().Add("Link", `</index.md>; rel="alternate"; type="text/markdown"`)
	if err := templates.IndexPage(trails, sortLabel).Render(r.Context(), w); err != nil {
		slog.Error("failed to render index", "error", err)
	}
}

// Slugify converts a trail name to a URL-safe kebab-case slug.
// "Markham Park" -> "markham-park", "Oleta River State Park" -> "oleta-river-state-park".
func Slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	b.Grow(len(s))
	prevDash := true
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.TrimRight(b.String(), "-")
}

func (h *Handler) HandleTrailDetail(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	trails, err := db.GetTrailsWithStatus(r.Context(), h.db)
	if err != nil {
		slog.Error("failed to get trails", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	for i := range trails {
		if Slugify(trails[i].Name) != slug {
			continue
		}
		trail := trails[i]
		AttachDistanceAndSort([]models.TrailWithStatus{trail}, ReadUserLocation(r))
		h.attachWeatherOne(&trail)
		others := otherTrails(trails, i, 6)
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Link", fmt.Sprintf(`</trail/%s.md>; rel="alternate"; type="text/markdown"`, slug))
		w.Header().Add("Link", `</>; rel="up"`)
		if err := templates.TrailDetailPage(trail, others).Render(r.Context(), w); err != nil {
			slog.Error("failed to render trail detail", "slug", slug, "error", err)
		}
		return
	}
	http.NotFound(w, r)
}

// otherTrails returns up to n trails other than the one at idx, preserving the
// caller's slice order so distance-sorted callers surface the nearest peers
// first. The returned slice does not alias the input.
func otherTrails(trails []models.TrailWithStatus, idx, n int) []models.TrailWithStatus {
	if n <= 0 || len(trails) <= 1 {
		return nil
	}
	out := make([]models.TrailWithStatus, 0, n)
	for i, t := range trails {
		if i == idx {
			continue
		}
		out = append(out, t)
		if len(out) >= n {
			break
		}
	}
	return out
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

func (h *Handler) HandleSignatureDirectory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/http-message-signatures-directory+json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, "./static/.well-known/http-message-signatures-directory")
}

func (h *Handler) HandleAgentSkillsIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	http.ServeFile(w, r, "./static/.well-known/agent-skills/index.json")
}

func (h *Handler) HandleAgentSkillFile(w http.ResponseWriter, r *http.Request) {
	rel := r.PathValue("path")
	if rel == "" || strings.Contains(rel, "..") {
		http.NotFound(w, r)
		return
	}
	if !strings.HasSuffix(rel, "/SKILL.md") {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	http.ServeFile(w, r, "./static/.well-known/agent-skills/"+rel)
}

func (h *Handler) HandleRobotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, "./static/robots.txt")
}

func (h *Handler) HandleSitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	lastmod := time.Now().UTC().Format("2006-01-02")

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	fmt.Fprintf(&b, "  <url>\n    <loc>https://sftrails.info/</loc>\n    <lastmod>%s</lastmod>\n    <changefreq>hourly</changefreq>\n    <priority>1.0</priority>\n  </url>\n", lastmod)

	if trails, err := db.GetTrailsWithStatus(r.Context(), h.db); err == nil {
		for _, t := range trails {
			fmt.Fprintf(&b, "  <url>\n    <loc>https://sftrails.info/trail/%s</loc>\n    <lastmod>%s</lastmod>\n    <changefreq>hourly</changefreq>\n    <priority>0.8</priority>\n  </url>\n", Slugify(t.Name), lastmod)
		}
	} else {
		slog.Warn("sitemap: trail lookup failed; serving index-only sitemap", "error", err)
	}

	// Reference docs and machine-readable feeds. These are stable artifacts
	// that change rarely but should be discoverable for both search and AI
	// crawlers walking the sitemap.
	fmt.Fprintf(&b, "  <url>\n    <loc>https://sftrails.info/llms.txt</loc>\n    <lastmod>%s</lastmod>\n    <changefreq>weekly</changefreq>\n    <priority>0.5</priority>\n  </url>\n", lastmod)
	fmt.Fprintf(&b, "  <url>\n    <loc>https://sftrails.info/llms-full.txt</loc>\n    <lastmod>%s</lastmod>\n    <changefreq>weekly</changefreq>\n    <priority>0.5</priority>\n  </url>\n", lastmod)
	fmt.Fprintf(&b, "  <url>\n    <loc>https://sftrails.info/api/trails</loc>\n    <lastmod>%s</lastmod>\n    <changefreq>hourly</changefreq>\n    <priority>0.4</priority>\n  </url>\n", lastmod)

	b.WriteString(`</urlset>`)
	if _, err := w.Write([]byte(b.String())); err != nil {
		slog.Error("failed to write sitemap", "error", err)
	}
}
