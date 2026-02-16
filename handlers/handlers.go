package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

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
