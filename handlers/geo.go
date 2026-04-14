package handlers

import (
	"cmp"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"sftrails/models"
)

// UserLocation is the user's chosen location for distance sorting.
// Label is a short human-readable description (e.g. "my location" or "ZIP 33323").
type UserLocation struct {
	Lat   float64
	Lng   float64
	Label string
}

const (
	locCookieName    = "sft_loc"
	labelCookieName  = "sft_loc_label"
	maxLocCookieLen  = 48  // "-89.1234,-179.1234" is 18 bytes; 48 is a generous cap
	maxLabelCookie   = 64  // URL-encoded "ZIP 99999" fits easily; 64 rejects payload dumps
	defaultLocLabel  = "my location"
)

// labelPattern accepts only values the app intentionally generates.
// Keeps the HTML-rendered label out of reach of user-controlled junk.
var labelPattern = regexp.MustCompile(`^(my location|ZIP \d{5})$`)

// ReadUserLocation pulls the user's chosen sort location from cookies.
// Returns nil if no valid location cookie is present or the values are malformed.
func ReadUserLocation(r *http.Request) *UserLocation {
	c, err := r.Cookie(locCookieName)
	if err != nil || c.Value == "" || len(c.Value) > maxLocCookieLen {
		return nil
	}
	parts := strings.SplitN(c.Value, ",", 2)
	if len(parts) != 2 {
		return nil
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil || lat < -90 || lat > 90 {
		return nil
	}
	lng, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil || lng < -180 || lng > 180 {
		return nil
	}
	label := defaultLocLabel
	if lc, err := r.Cookie(labelCookieName); err == nil && lc.Value != "" && len(lc.Value) <= maxLabelCookie {
		if decoded, derr := url.QueryUnescape(lc.Value); derr == nil && labelPattern.MatchString(decoded) {
			label = decoded
		}
	}
	return &UserLocation{Lat: lat, Lng: lng, Label: label}
}

// HaversineMiles returns the great-circle distance in statute miles between
// two lat/lng points.
func HaversineMiles(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusMiles = 3958.8
	rad := math.Pi / 180
	dLat := (lat2 - lat1) * rad
	dLng := (lng2 - lng1) * rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*rad)*math.Cos(lat2*rad)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusMiles * c
}

// AttachDistanceAndSort populates DistanceMi/HasDistance on each trail and
// sorts the slice in place by ascending distance. No-op if loc is nil.
func AttachDistanceAndSort(trails []models.TrailWithStatus, loc *UserLocation) {
	if loc == nil {
		return
	}
	for i := range trails {
		trails[i].DistanceMi = HaversineMiles(loc.Lat, loc.Lng, trails[i].Latitude, trails[i].Longitude)
		trails[i].HasDistance = true
	}
	slices.SortFunc(trails, func(a, b models.TrailWithStatus) int {
		return cmp.Compare(a.DistanceMi, b.DistanceMi)
	})
}
