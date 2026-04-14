package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Forecast holds the daily weather forecast for a single trail location.
type Forecast struct {
	Icon       string
	Desc       string
	TempHighF  float64
	RainChance int
}

// Location identifies a trail's coordinates for weather lookup.
type Location struct {
	TrailID int64
	Lat     float64
	Lng     float64
}

// Store holds cached daily forecasts keyed by trail ID.
type Store struct {
	mu          sync.RWMutex
	forecasts   map[int64]Forecast
	locations   []Location
	lastRefresh time.Time
}

// NewStore creates a weather store for the given trail locations.
func NewStore(locations []Location) *Store {
	return &Store{
		forecasts: make(map[int64]Forecast),
		locations: locations,
	}
}

// Get returns the cached forecast for a trail, if available.
func (s *Store) Get(trailID int64) (Forecast, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.forecasts[trailID]
	return f, ok
}

// Refresh fetches fresh forecasts for all trail locations from Open-Meteo.
func (s *Store) Refresh() {
	updated := make(map[int64]Forecast, len(s.locations))
	client := &http.Client{Timeout: 10 * time.Second}

	for _, loc := range s.locations {
		f, err := fetchForecast(client, loc.Lat, loc.Lng)
		if err != nil {
			slog.Warn("weather fetch failed", "trail_id", loc.TrailID, "error", err)
			// Keep stale data for this trail if we have it
			s.mu.RLock()
			if old, ok := s.forecasts[loc.TrailID]; ok {
				updated[loc.TrailID] = old
			}
			s.mu.RUnlock()
			continue
		}
		updated[loc.TrailID] = *f
	}

	s.mu.Lock()
	s.forecasts = updated
	s.lastRefresh = time.Now()
	s.mu.Unlock()
	slog.Info("weather refreshed", "trails", len(updated))
}

// Stats returns a snapshot of weather store state for diagnostics.
func (s *Store) Stats() (cached, total int, lastRefresh time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.forecasts), len(s.locations), s.lastRefresh
}

// EndpointCheck is the result of a single outbound endpoint health check.
// Error is a short error category ("timeout", "dns", "tls", "http"); raw
// upstream errors are kept out of public responses to avoid info disclosure.
type EndpointCheck struct {
	Name    string
	URL     string
	OK      bool
	Status  int
	Latency time.Duration
	Error   string
}

// checkCache rate-limits public /status calls to one outbound probe per TTL.
var (
	checkCacheMu   sync.Mutex
	checkCacheData []EndpointCheck
	checkCacheAt   time.Time
	checkCacheTTL  = 60 * time.Second
)

// CheckEndpoints runs lightweight health checks against outbound endpoints
// the app depends on. Results are cached for 60 seconds so repeated /status
// requests can't amplify outbound traffic.
func CheckEndpoints() []EndpointCheck {
	checkCacheMu.Lock()
	defer checkCacheMu.Unlock()

	if checkCacheData != nil && time.Since(checkCacheAt) < checkCacheTTL {
		out := make([]EndpointCheck, len(checkCacheData))
		copy(out, checkCacheData)
		return out
	}

	client := &http.Client{Timeout: 5 * time.Second}
	checks := []struct{ name, url string }{
		{
			name: "Open-Meteo",
			url:  "https://api.open-meteo.com/v1/forecast?latitude=0&longitude=0&daily=weather_code&forecast_days=1&timezone=UTC",
		},
	}

	results := make([]EndpointCheck, 0, len(checks))
	for _, c := range checks {
		start := time.Now()
		resp, err := client.Get(c.url)
		elapsed := time.Since(start)

		r := EndpointCheck{Name: c.name, URL: c.url, Latency: elapsed}
		if err != nil {
			r.Error = classifyErr(err)
			results = append(results, r)
			continue
		}
		resp.Body.Close()
		r.Status = resp.StatusCode
		r.OK = resp.StatusCode >= 200 && resp.StatusCode < 300
		if !r.OK {
			r.Error = "http"
		}
		results = append(results, r)
	}

	checkCacheData = results
	checkCacheAt = time.Now()
	out := make([]EndpointCheck, len(results))
	copy(out, results)
	return out
}

// classifyErr maps a net/http error into a short, public-safe category.
// We intentionally avoid returning raw error strings (which can leak
// internal hostnames, proxy URLs, or cert chain details).
func classifyErr(err error) string {
	s := err.Error()
	switch {
	case strings.Contains(s, "context deadline exceeded"), strings.Contains(s, "Client.Timeout"), strings.Contains(s, "i/o timeout"):
		return "timeout"
	case strings.Contains(s, "no such host"), strings.Contains(s, "dns"):
		return "dns"
	case strings.Contains(s, "x509"), strings.Contains(s, "tls:"), strings.Contains(s, "certificate"):
		return "tls"
	default:
		return "http"
	}
}

// StartScheduler fetches weather immediately, then refreshes once every 24 hours.
// It blocks until ctx is cancelled.
func (s *Store) StartScheduler(ctx context.Context) {
	s.Refresh()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.Refresh()
		}
	}
}

// Open-Meteo API response structure (daily forecast).
type openMeteoResponse struct {
	Daily struct {
		WeatherCode              []int     `json:"weather_code"`
		Temperature2mMax         []float64 `json:"temperature_2m_max"`
		PrecipitationProbability []int     `json:"precipitation_probability_max"`
	} `json:"daily"`
}

func fetchForecast(client *http.Client, lat, lng float64) (*Forecast, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f"+
			"&daily=weather_code,temperature_2m_max,precipitation_probability_max"+
			"&temperature_unit=fahrenheit&timezone=America%%2FNew_York&forecast_days=1",
		lat, lng,
	)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var data openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	if len(data.Daily.WeatherCode) == 0 {
		return nil, fmt.Errorf("no daily data returned")
	}

	code := data.Daily.WeatherCode[0]
	tempHigh := 0.0
	if len(data.Daily.Temperature2mMax) > 0 {
		tempHigh = data.Daily.Temperature2mMax[0]
	}
	rainChance := 0
	if len(data.Daily.PrecipitationProbability) > 0 {
		rainChance = data.Daily.PrecipitationProbability[0]
	}

	icon, desc := weatherInfo(code)

	return &Forecast{
		Icon:       icon,
		Desc:       desc,
		TempHighF:  tempHigh,
		RainChance: rainChance,
	}, nil
}

// weatherInfo maps a WMO weather code to an icon and short description.
func weatherInfo(code int) (string, string) {
	switch {
	case code == 0:
		return "☀️", "Clear"
	case code <= 2:
		return "🌤️", "Partly Cloudy"
	case code == 3:
		return "☁️", "Overcast"
	case code <= 48:
		return "🌫️", "Foggy"
	case code <= 57:
		return "🌦️", "Drizzle"
	case code <= 67:
		return "🌧️", "Rain"
	case code <= 77:
		return "🌨️", "Snow"
	case code <= 82:
		return "🌧️", "Showers"
	case code <= 86:
		return "🌨️", "Snow Showers"
	default:
		return "⛈️", "Thunderstorm"
	}
}
