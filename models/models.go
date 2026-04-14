package models

import "time"

type VoteType string

const (
	VoteOpen   VoteType = "open"
	VoteClosed VoteType = "closed"
)

type TrailStatus string

const (
	StatusOpen    TrailStatus = "open"
	StatusClosed  TrailStatus = "closed"
	StatusUnknown TrailStatus = "unknown"
)

// VoteOutcome indicates whether a user's vote was recorded or rejected as a
// duplicate. Used to show immediate feedback in the UI after a vote.
type VoteOutcome int

const (
	VoteOutcomeNone VoteOutcome = iota
	VoteOutcomeRecorded
	VoteOutcomeDuplicate
)

type Trail struct {
	ID          int64
	Name        string
	Location    string
	City        string
	Description string
	Latitude    float64
	Longitude   float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Vote struct {
	ID          int64
	TrailID     int64
	Vote        VoteType
	IPAddress   string
	Fingerprint string
	CreatedAt   time.Time
}

type TrailWithStatus struct {
	Trail
	Status      TrailStatus
	OpenVotes   int
	ClosedVotes int
	TotalVotes  int
	// Weather fields (populated once daily from Open-Meteo)
	HasWeather       bool
	WeatherIcon      string
	WeatherDesc      string
	WeatherTempHighF float64
	WeatherRainPct   int
	// Transient vote outcome set by HandleVote so the TrailCard can show
	// an ephemeral toast. Not populated by DB queries.
	VoteOutcome VoteOutcome
}

// API response types for JSON endpoints

type TrailAPIResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Location    string  `json:"location"`
	City        string  `json:"city"`
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Status      string  `json:"status"`
	OpenVotes   int     `json:"open_votes"`
	ClosedVotes int     `json:"closed_votes"`
	TotalVotes  int     `json:"total_votes"`
}

type TrailsAPIResponse struct {
	Trails      []TrailAPIResponse `json:"trails"`
	GeneratedAt string             `json:"generated_at"`
}

func TrailToAPI(t TrailWithStatus) TrailAPIResponse {
	return TrailAPIResponse{
		ID:          t.ID,
		Name:        t.Name,
		Location:    t.Location,
		City:        t.City,
		Description: t.Description,
		Latitude:    t.Latitude,
		Longitude:   t.Longitude,
		Status:      string(t.Status),
		OpenVotes:   t.OpenVotes,
		ClosedVotes: t.ClosedVotes,
		TotalVotes:  t.TotalVotes,
	}
}
