package testhelpers

import (
	"net/http"
	"strings"
	"testing"
)

func TestMockSonarrServer(t *testing.T) {
	series := CreateTestSeries()
	episodes := CreateTestEpisodes()
	commandResponse := CreateTestCommandResponse()

	server := MockSonarrServer(t, series, episodes, commandResponse)
	defer server.Close()

	// Test series endpoint
	resp, err := http.Get(server.URL + "/api/v3/series")
	if err != nil {
		t.Fatalf("failed to get series: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Test episodes endpoint
	resp, err = http.Get(server.URL + "/api/v3/episode?seriesId=1")
	if err != nil {
		t.Fatalf("failed to get episodes: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Test command endpoint
	resp, err = http.Post(server.URL+"/api/v3/command", "application/json", strings.NewReader(`{"name":"EpisodeSearch"}`))
	if err != nil {
		t.Fatalf("failed to post command: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
}

func TestMockRadarrServer(t *testing.T) {
	movies := CreateTestMovies()
	commandResponse := CreateTestCommandResponse()

	server := MockRadarrServer(t, movies, commandResponse)
	defer server.Close()

	// Test movies endpoint
	resp, err := http.Get(server.URL + "/api/v3/movie")
	if err != nil {
		t.Fatalf("failed to get movies: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Test command endpoint
	resp, err = http.Post(server.URL+"/api/v3/command", "application/json", strings.NewReader(`{"name":"MoviesSearch"}`))
	if err != nil {
		t.Fatalf("failed to post command: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
}

func TestCreateTestSeries(t *testing.T) {
	series := CreateTestSeries()

	if len(series) == 0 {
		t.Error("expected at least one test series")
	}

	for _, s := range series {
		if s.ID == 0 {
			t.Error("expected series to have non-zero ID")
		}
		if s.Title == "" {
			t.Error("expected series to have non-empty title")
		}
	}
}

func TestCreateTestEpisodes(t *testing.T) {
	episodes := CreateTestEpisodes()

	if len(episodes) == 0 {
		t.Error("expected at least one series with episodes")
	}

	for seriesID, episodeList := range episodes {
		if seriesID == 0 {
			t.Error("expected non-zero series ID")
		}

		if len(episodeList) == 0 {
			t.Error("expected at least one episode per series")
		}

		for _, ep := range episodeList {
			if ep.ID == 0 {
				t.Error("expected episode to have non-zero ID")
			}
			if ep.SeriesID != seriesID {
				t.Errorf("expected episode SeriesID %d to match map key %d", ep.SeriesID, seriesID)
			}
		}
	}
}

func TestCreateTestMovies(t *testing.T) {
	movies := CreateTestMovies()

	if len(movies) == 0 {
		t.Error("expected at least one test movie")
	}

	for _, m := range movies {
		if m.ID == 0 {
			t.Error("expected movie to have non-zero ID")
		}
		if m.Title == "" {
			t.Error("expected movie to have non-empty title")
		}
		if m.Year == 0 {
			t.Error("expected movie to have non-zero year")
		}
	}
}

func TestCreateTestConfig(t *testing.T) {
	config := CreateTestConfig()

	// Basic validation of config fields
	if config.BatchSize <= 0 {
		t.Error("expected positive batch size")
	}
	if config.Interval <= 0 {
		t.Error("expected positive interval")
	}
}

func TestCreateTestCommandResponse(t *testing.T) {
	response := CreateTestCommandResponse()

	if response.ID == 0 {
		t.Error("expected command response to have non-zero ID")
	}
	if response.Name == "" {
		t.Error("expected command response to have non-empty name")
	}
	if response.Status == "" {
		t.Error("expected command response to have non-empty status")
	}
}
