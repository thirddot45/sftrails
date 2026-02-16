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
}
