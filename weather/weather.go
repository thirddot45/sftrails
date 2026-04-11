package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
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
	mu        sync.RWMutex
	forecasts map[int64]Forecast
	locations []Location
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
	s.mu.Unlock()
	slog.Info("weather refreshed", "trails", len(updated))
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
