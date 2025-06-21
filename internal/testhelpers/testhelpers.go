package testhelpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"score-checker/internal/types"
)

// TestingInterface allows both *testing.T and *testing.B to be used
type TestingInterface interface {
	Helper()
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
}

// MockSonarrServer creates a mock Sonarr server for testing
func MockSonarrServer(t TestingInterface, series []types.Series, episodes map[int][]types.Episode, commandResponse *types.CommandResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v3/series":
			if r.Method != "GET" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			json.NewEncoder(w).Encode(series)

		case "/api/v3/episode":
			if r.Method != "GET" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			seriesIDStr := r.URL.Query().Get("seriesId")
			if seriesIDStr == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var seriesID int
			switch seriesIDStr {
			case "1":
				seriesID = 1
			case "2":
				seriesID = 2
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}

			if eps, ok := episodes[seriesID]; ok {
				json.NewEncoder(w).Encode(eps)
			} else {
				json.NewEncoder(w).Encode([]types.Episode{})
			}

		case "/api/v3/command":
			if r.Method != "POST" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if commandResponse != nil {
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(commandResponse)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// MockRadarrServer creates a mock Radarr server for testing
func MockRadarrServer(t TestingInterface, movies []types.MovieWithFile, commandResponse *types.CommandResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v3/movie":
			if r.Method != "GET" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			json.NewEncoder(w).Encode(movies)

		case "/api/v3/command":
			if r.Method != "POST" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if commandResponse != nil {
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(commandResponse)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// CreateTestSeries creates test series data
func CreateTestSeries() []types.Series {
	return []types.Series{
		{ID: 1, Title: "Breaking Bad"},
		{ID: 2, Title: "Better Call Saul"},
	}
}

// CreateTestEpisodes creates test episode data
func CreateTestEpisodes() map[int][]types.Episode {
	return map[int][]types.Episode{
		1: {
			{
				ID:            101,
				SeriesID:      1,
				Title:         "Pilot",
				SeasonNumber:  1,
				EpisodeNumber: 1,
				HasFile:       true,
				EpisodeFile: &types.EpisodeFile{
					ID:                201,
					CustomFormatScore: -10,
				},
			},
			{
				ID:            102,
				SeriesID:      1,
				Title:         "Cat's in the Bag...",
				SeasonNumber:  1,
				EpisodeNumber: 2,
				HasFile:       true,
				EpisodeFile: &types.EpisodeFile{
					ID:                202,
					CustomFormatScore: 5,
				},
			},
		},
		2: {
			{
				ID:            201,
				SeriesID:      2,
				Title:         "Uno",
				SeasonNumber:  1,
				EpisodeNumber: 1,
				HasFile:       true,
				EpisodeFile: &types.EpisodeFile{
					ID:                301,
					CustomFormatScore: -5,
				},
			},
		},
	}
}

// CreateTestMovies creates test movie data
func CreateTestMovies() []types.MovieWithFile {
	return []types.MovieWithFile{
		{
			ID:      1,
			Title:   "The Matrix",
			Year:    1999,
			HasFile: true,
			MovieFile: &types.MovieFile{
				ID:                101,
				CustomFormatScore: -15,
			},
		},
		{
			ID:      2,
			Title:   "Inception",
			Year:    2010,
			HasFile: true,
			MovieFile: &types.MovieFile{
				ID:                102,
				CustomFormatScore: 10,
			},
		},
		{
			ID:        3,
			Title:     "Interstellar",
			Year:      2014,
			HasFile:   false,
			MovieFile: nil,
		},
	}
}

// CreateTestConfig creates a test configuration
func CreateTestConfig() types.Config {
	return types.Config{
		SonarrInstances: []types.ServiceConfig{
			{
				Name:    "test-sonarr",
				BaseURL: "http://localhost:8989",
				APIKey:  "test-sonarr-key",
			},
		},
		RadarrInstances: []types.ServiceConfig{
			{
				Name:    "test-radarr",
				BaseURL: "http://localhost:7878",
				APIKey:  "test-radarr-key",
			},
		},
		TriggerSearch: false,
		BatchSize:     5,
	}
}

// CreateTestCommandResponse creates a test command response
func CreateTestCommandResponse() *types.CommandResponse {
	return &types.CommandResponse{
		ID:          123,
		Name:        "EpisodeSearch",
		CommandName: "EpisodeSearch",
		Status:      "queued",
	}
}
