package types

import (
	"testing"
	"time"
)

func TestServiceConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   ServiceConfig
		expected ServiceConfig
	}{
		{
			name: "valid service config",
			config: ServiceConfig{
				Name:    "test-instance",
				BaseURL: "http://localhost:8989",
				APIKey:  "test-api-key",
			},
			expected: ServiceConfig{
				Name:    "test-instance",
				BaseURL: "http://localhost:8989",
				APIKey:  "test-api-key",
			},
		},
		{
			name: "empty service config",
			config: ServiceConfig{
				Name:    "",
				BaseURL: "",
				APIKey:  "",
			},
			expected: ServiceConfig{
				Name:    "",
				BaseURL: "",
				APIKey:  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Name != tt.expected.Name {
				t.Errorf("expected Name %q, got %q", tt.expected.Name, tt.config.Name)
			}
			if tt.config.BaseURL != tt.expected.BaseURL {
				t.Errorf("expected BaseURL %q, got %q", tt.expected.BaseURL, tt.config.BaseURL)
			}
			if tt.config.APIKey != tt.expected.APIKey {
				t.Errorf("expected APIKey %q, got %q", tt.expected.APIKey, tt.config.APIKey)
			}
		})
	}
}

func TestConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected Config
	}{
		{
			name: "valid config with multiple instances",
			config: Config{
				SonarrInstances: []ServiceConfig{
					{Name: "main", BaseURL: "http://localhost:8989", APIKey: "sonarr-key"},
					{Name: "4k", BaseURL: "http://localhost:8990", APIKey: "sonarr-4k-key"},
				},
				RadarrInstances: []ServiceConfig{
					{Name: "main", BaseURL: "http://localhost:7878", APIKey: "radarr-key"},
				},
				TriggerSearch: true,
				BatchSize:     10,
				Interval:      time.Hour,
			},
			expected: Config{
				SonarrInstances: []ServiceConfig{
					{Name: "main", BaseURL: "http://localhost:8989", APIKey: "sonarr-key"},
					{Name: "4k", BaseURL: "http://localhost:8990", APIKey: "sonarr-4k-key"},
				},
				RadarrInstances: []ServiceConfig{
					{Name: "main", BaseURL: "http://localhost:7878", APIKey: "radarr-key"},
				},
				TriggerSearch: true,
				BatchSize:     10,
				Interval:      time.Hour,
			},
		},
		{
			name: "empty config",
			config: Config{
				SonarrInstances: nil,
				RadarrInstances: nil,
				TriggerSearch:   false,
				BatchSize:       0,
				Interval:        0,
			},
			expected: Config{
				SonarrInstances: nil,
				RadarrInstances: nil,
				TriggerSearch:   false,
				BatchSize:       0,
				Interval:        0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.config.SonarrInstances) != len(tt.expected.SonarrInstances) {
				t.Errorf("expected %d Sonarr instances, got %d", len(tt.expected.SonarrInstances), len(tt.config.SonarrInstances))
			}
			if len(tt.config.RadarrInstances) != len(tt.expected.RadarrInstances) {
				t.Errorf("expected %d Radarr instances, got %d", len(tt.expected.RadarrInstances), len(tt.config.RadarrInstances))
			}
			if tt.config.TriggerSearch != tt.expected.TriggerSearch {
				t.Errorf("expected TriggerSearch %v, got %v", tt.expected.TriggerSearch, tt.config.TriggerSearch)
			}
			if tt.config.BatchSize != tt.expected.BatchSize {
				t.Errorf("expected BatchSize %d, got %d", tt.expected.BatchSize, tt.config.BatchSize)
			}
			if tt.config.Interval != tt.expected.Interval {
				t.Errorf("expected Interval %v, got %v", tt.expected.Interval, tt.config.Interval)
			}
		})
	}
}

func TestSeries(t *testing.T) {
	series := Series{
		ID:    1,
		Title: "Breaking Bad",
	}

	if series.ID != 1 {
		t.Errorf("expected ID 1, got %d", series.ID)
	}
	if series.Title != "Breaking Bad" {
		t.Errorf("expected Title 'Breaking Bad', got %q", series.Title)
	}
}

func TestEpisode(t *testing.T) {
	episodeFile := &EpisodeFile{
		ID:                123,
		CustomFormatScore: -10,
	}

	episode := Episode{
		ID:            456,
		SeriesID:      1,
		Title:         "Pilot",
		SeasonNumber:  1,
		EpisodeNumber: 1,
		HasFile:       true,
		EpisodeFile:   episodeFile,
	}

	if episode.ID != 456 {
		t.Errorf("expected ID 456, got %d", episode.ID)
	}
	if episode.SeriesID != 1 {
		t.Errorf("expected SeriesID 1, got %d", episode.SeriesID)
	}
	if episode.Title != "Pilot" {
		t.Errorf("expected Title 'Pilot', got %q", episode.Title)
	}
	if episode.SeasonNumber != 1 {
		t.Errorf("expected SeasonNumber 1, got %d", episode.SeasonNumber)
	}
	if episode.EpisodeNumber != 1 {
		t.Errorf("expected EpisodeNumber 1, got %d", episode.EpisodeNumber)
	}
	if !episode.HasFile {
		t.Error("expected HasFile to be true")
	}
	if episode.EpisodeFile == nil {
		t.Fatal("expected EpisodeFile to not be nil")
	}
	if episode.EpisodeFile.ID != 123 {
		t.Errorf("expected EpisodeFile ID 123, got %d", episode.EpisodeFile.ID)
	}
	if episode.EpisodeFile.CustomFormatScore != -10 {
		t.Errorf("expected CustomFormatScore -10, got %d", episode.EpisodeFile.CustomFormatScore)
	}
}

func TestLowScoreEpisode(t *testing.T) {
	series := Series{ID: 1, Title: "Breaking Bad"}
	episode := Episode{ID: 456, Title: "Pilot", SeasonNumber: 1, EpisodeNumber: 1}

	lowScoreEpisode := LowScoreEpisode{
		Series:            series,
		Episode:           episode,
		CustomFormatScore: -15,
	}

	if lowScoreEpisode.Series.ID != 1 {
		t.Errorf("expected Series ID 1, got %d", lowScoreEpisode.Series.ID)
	}
	if lowScoreEpisode.Episode.ID != 456 {
		t.Errorf("expected Episode ID 456, got %d", lowScoreEpisode.Episode.ID)
	}
	if lowScoreEpisode.CustomFormatScore != -15 {
		t.Errorf("expected CustomFormatScore -15, got %d", lowScoreEpisode.CustomFormatScore)
	}
}

func TestMovie(t *testing.T) {
	movie := Movie{
		ID:    1,
		Title: "The Matrix",
		Year:  1999,
	}

	if movie.ID != 1 {
		t.Errorf("expected ID 1, got %d", movie.ID)
	}
	if movie.Title != "The Matrix" {
		t.Errorf("expected Title 'The Matrix', got %q", movie.Title)
	}
	if movie.Year != 1999 {
		t.Errorf("expected Year 1999, got %d", movie.Year)
	}
}

func TestMovieWithFile(t *testing.T) {
	movieFile := &MovieFile{
		ID:                789,
		CustomFormatScore: -20,
	}

	movie := MovieWithFile{
		ID:        1,
		Title:     "The Matrix",
		Year:      1999,
		HasFile:   true,
		MovieFile: movieFile,
	}

	if movie.ID != 1 {
		t.Errorf("expected ID 1, got %d", movie.ID)
	}
	if movie.Title != "The Matrix" {
		t.Errorf("expected Title 'The Matrix', got %q", movie.Title)
	}
	if movie.Year != 1999 {
		t.Errorf("expected Year 1999, got %d", movie.Year)
	}
	if !movie.HasFile {
		t.Error("expected HasFile to be true")
	}
	if movie.MovieFile == nil {
		t.Fatal("expected MovieFile to not be nil")
	}
	if movie.MovieFile.ID != 789 {
		t.Errorf("expected MovieFile ID 789, got %d", movie.MovieFile.ID)
	}
	if movie.MovieFile.CustomFormatScore != -20 {
		t.Errorf("expected CustomFormatScore -20, got %d", movie.MovieFile.CustomFormatScore)
	}
}

func TestLowScoreMovie(t *testing.T) {
	movie := MovieWithFile{ID: 1, Title: "The Matrix", Year: 1999}

	lowScoreMovie := LowScoreMovie{
		Movie:             movie,
		CustomFormatScore: -25,
	}

	if lowScoreMovie.Movie.ID != 1 {
		t.Errorf("expected Movie ID 1, got %d", lowScoreMovie.Movie.ID)
	}
	if lowScoreMovie.CustomFormatScore != -25 {
		t.Errorf("expected CustomFormatScore -25, got %d", lowScoreMovie.CustomFormatScore)
	}
}

func TestCommandRequest(t *testing.T) {
	cmdReq := CommandRequest{
		Name:       "EpisodeSearch",
		EpisodeIDs: []int{1, 2, 3},
	}

	if cmdReq.Name != "EpisodeSearch" {
		t.Errorf("expected Name 'EpisodeSearch', got %q", cmdReq.Name)
	}
	if len(cmdReq.EpisodeIDs) != 3 {
		t.Errorf("expected 3 EpisodeIDs, got %d", len(cmdReq.EpisodeIDs))
	}
	expectedIDs := []int{1, 2, 3}
	for i, id := range cmdReq.EpisodeIDs {
		if id != expectedIDs[i] {
			t.Errorf("expected EpisodeID %d at index %d, got %d", expectedIDs[i], i, id)
		}
	}
}

func TestCommandResponse(t *testing.T) {
	cmdResp := CommandResponse{
		ID:          123,
		Name:        "EpisodeSearch",
		CommandName: "EpisodeSearch",
		Status:      "queued",
	}

	if cmdResp.ID != 123 {
		t.Errorf("expected ID 123, got %d", cmdResp.ID)
	}
	if cmdResp.Name != "EpisodeSearch" {
		t.Errorf("expected Name 'EpisodeSearch', got %q", cmdResp.Name)
	}
	if cmdResp.CommandName != "EpisodeSearch" {
		t.Errorf("expected CommandName 'EpisodeSearch', got %q", cmdResp.CommandName)
	}
	if cmdResp.Status != "queued" {
		t.Errorf("expected Status 'queued', got %q", cmdResp.Status)
	}
}
