package sonarr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"score-checker/internal/types"
)

func TestNewClient(t *testing.T) {
	config := types.ServiceConfig{
		Name:    "test",
		BaseURL: "http://localhost:8989",
		APIKey:  "test-api-key",
	}

	client := NewClient(config)

	if client == nil {
		t.Fatal("expected client to not be nil")
	}
	if client.config.Name != config.Name {
		t.Errorf("expected config name %q, got %q", config.Name, client.config.Name)
	}
	if client.config.BaseURL != config.BaseURL {
		t.Errorf("expected config baseurl %q, got %q", config.BaseURL, client.config.BaseURL)
	}
	if client.config.APIKey != config.APIKey {
		t.Errorf("expected config apikey %q, got %q", config.APIKey, client.config.APIKey)
	}
	if client.client == nil {
		t.Error("expected http client to not be nil")
	}
}

func TestGetSeries(t *testing.T) {
	tests := []struct {
		name           string
		responseCode   int
		responseBody   string
		expectedSeries []types.Series
		expectError    bool
	}{
		{
			name:         "successful response",
			responseCode: http.StatusOK,
			responseBody: `[
				{"id": 1, "title": "Breaking Bad"},
				{"id": 2, "title": "Better Call Saul"}
			]`,
			expectedSeries: []types.Series{
				{ID: 1, Title: "Breaking Bad"},
				{ID: 2, Title: "Better Call Saul"},
			},
			expectError: false,
		},
		{
			name:           "empty response",
			responseCode:   http.StatusOK,
			responseBody:   `[]`,
			expectedSeries: []types.Series{},
			expectError:    false,
		},
		{
			name:           "server error",
			responseCode:   http.StatusInternalServerError,
			responseBody:   `{"error": "internal server error"}`,
			expectedSeries: nil,
			expectError:    true,
		},
		{
			name:           "invalid json",
			responseCode:   http.StatusOK,
			responseBody:   `invalid json`,
			expectedSeries: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify API key header
				if r.Header.Get("X-Api-Key") != "test-api-key" {
					t.Errorf("expected X-Api-Key header 'test-api-key', got %q", r.Header.Get("X-Api-Key"))
				}

				// Verify URL path
				if r.URL.Path != "/api/v3/series" {
					t.Errorf("expected path '/api/v3/series', got %q", r.URL.Path)
				}

				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			config := types.ServiceConfig{
				Name:    "test",
				BaseURL: server.URL,
				APIKey:  "test-api-key",
			}
			client := NewClient(config)

			series, err := client.GetSeries()

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(series) != len(tt.expectedSeries) {
				t.Errorf("expected %d series, got %d", len(tt.expectedSeries), len(series))
				return
			}

			for i, expected := range tt.expectedSeries {
				if series[i].ID != expected.ID {
					t.Errorf("series[%d] expected ID %d, got %d", i, expected.ID, series[i].ID)
				}
				if series[i].Title != expected.Title {
					t.Errorf("series[%d] expected Title %q, got %q", i, expected.Title, series[i].Title)
				}
			}
		})
	}
}

func TestGetEpisodes(t *testing.T) {
	tests := []struct {
		name             string
		seriesID         int
		responseCode     int
		responseBody     string
		expectedEpisodes []types.Episode
		expectError      bool
	}{
		{
			name:         "successful response",
			seriesID:     1,
			responseCode: http.StatusOK,
			responseBody: `[
				{
					"id": 101,
					"seriesId": 1,
					"title": "Pilot",
					"seasonNumber": 1,
					"episodeNumber": 1,
					"hasFile": true,
					"episodeFile": {
						"id": 201,
						"customFormatScore": -10
					}
				}
			]`,
			expectedEpisodes: []types.Episode{
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
			},
			expectError: false,
		},
		{
			name:             "server error",
			seriesID:         1,
			responseCode:     http.StatusInternalServerError,
			responseBody:     `{"error": "internal server error"}`,
			expectedEpisodes: nil,
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify API key header
				if r.Header.Get("X-Api-Key") != "test-api-key" {
					t.Errorf("expected X-Api-Key header 'test-api-key', got %q", r.Header.Get("X-Api-Key"))
				}

				// Verify URL path and query parameters
				if r.URL.Path != "/api/v3/episode" {
					t.Errorf("expected path '/api/v3/episode', got %q", r.URL.Path)
				}

				seriesID := r.URL.Query().Get("seriesId")
				expectedSeriesID := fmt.Sprintf("%d", tt.seriesID)
				if seriesID != expectedSeriesID {
					t.Errorf("expected seriesId query param %q, got %q", expectedSeriesID, seriesID)
				}

				includeEpisodeFile := r.URL.Query().Get("includeEpisodeFile")
				if includeEpisodeFile != "true" {
					t.Errorf("expected includeEpisodeFile query param 'true', got %q", includeEpisodeFile)
				}

				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			config := types.ServiceConfig{
				Name:    "test",
				BaseURL: server.URL,
				APIKey:  "test-api-key",
			}
			client := NewClient(config)

			episodes, err := client.GetEpisodes(tt.seriesID)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(episodes) != len(tt.expectedEpisodes) {
				t.Errorf("expected %d episodes, got %d", len(tt.expectedEpisodes), len(episodes))
				return
			}

			for i, expected := range tt.expectedEpisodes {
				episode := episodes[i]
				if episode.ID != expected.ID {
					t.Errorf("episode[%d] expected ID %d, got %d", i, expected.ID, episode.ID)
				}
				if episode.SeriesID != expected.SeriesID {
					t.Errorf("episode[%d] expected SeriesID %d, got %d", i, expected.SeriesID, episode.SeriesID)
				}
				if episode.Title != expected.Title {
					t.Errorf("episode[%d] expected Title %q, got %q", i, expected.Title, episode.Title)
				}
				if episode.SeasonNumber != expected.SeasonNumber {
					t.Errorf("episode[%d] expected SeasonNumber %d, got %d", i, expected.SeasonNumber, episode.SeasonNumber)
				}
				if episode.EpisodeNumber != expected.EpisodeNumber {
					t.Errorf("episode[%d] expected EpisodeNumber %d, got %d", i, expected.EpisodeNumber, episode.EpisodeNumber)
				}
				if episode.HasFile != expected.HasFile {
					t.Errorf("episode[%d] expected HasFile %v, got %v", i, expected.HasFile, episode.HasFile)
				}

				if expected.EpisodeFile != nil {
					if episode.EpisodeFile == nil {
						t.Errorf("episode[%d] expected EpisodeFile to not be nil", i)
						continue
					}
					if episode.EpisodeFile.ID != expected.EpisodeFile.ID {
						t.Errorf("episode[%d] expected EpisodeFile.ID %d, got %d", i, expected.EpisodeFile.ID, episode.EpisodeFile.ID)
					}
					if episode.EpisodeFile.CustomFormatScore != expected.EpisodeFile.CustomFormatScore {
						t.Errorf("episode[%d] expected EpisodeFile.CustomFormatScore %d, got %d", i, expected.EpisodeFile.CustomFormatScore, episode.EpisodeFile.CustomFormatScore)
					}
				} else if episode.EpisodeFile != nil {
					t.Errorf("episode[%d] expected EpisodeFile to be nil", i)
				}
			}
		})
	}
}

func TestTriggerEpisodeSearch(t *testing.T) {
	tests := []struct {
		name             string
		episodeIDs       []int
		responseCode     int
		responseBody     string
		expectedResponse *types.CommandResponse
		expectError      bool
	}{
		{
			name:         "successful search trigger",
			episodeIDs:   []int{1, 2, 3},
			responseCode: http.StatusCreated,
			responseBody: `{
				"id": 123,
				"name": "EpisodeSearch",
				"commandName": "EpisodeSearch",
				"status": "queued"
			}`,
			expectedResponse: &types.CommandResponse{
				ID:          123,
				Name:        "EpisodeSearch",
				CommandName: "EpisodeSearch",
				Status:      "queued",
			},
			expectError: false,
		},
		{
			name:             "empty episode IDs",
			episodeIDs:       []int{},
			responseCode:     0, // won't be used
			responseBody:     "",
			expectedResponse: nil,
			expectError:      true,
		},
		{
			name:             "server error",
			episodeIDs:       []int{1, 2},
			responseCode:     http.StatusBadRequest,
			responseBody:     `{"error": "bad request"}`,
			expectedResponse: nil,
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server

			if len(tt.episodeIDs) > 0 {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Verify API key header
					if r.Header.Get("X-Api-Key") != "test-api-key" {
						t.Errorf("expected X-Api-Key header 'test-api-key', got %q", r.Header.Get("X-Api-Key"))
					}

					// Verify Content-Type header
					if r.Header.Get("Content-Type") != "application/json" {
						t.Errorf("expected Content-Type header 'application/json', got %q", r.Header.Get("Content-Type"))
					}

					// Verify URL path
					if r.URL.Path != "/api/v3/command" {
						t.Errorf("expected path '/api/v3/command', got %q", r.URL.Path)
					}

					// Verify HTTP method
					if r.Method != "POST" {
						t.Errorf("expected POST method, got %q", r.Method)
					}

					// Verify request body
					var cmdReq types.CommandRequest
					if err := json.NewDecoder(r.Body).Decode(&cmdReq); err != nil {
						t.Errorf("failed to decode request body: %v", err)
					}

					if cmdReq.Name != "EpisodeSearch" {
						t.Errorf("expected command name 'EpisodeSearch', got %q", cmdReq.Name)
					}

					if len(cmdReq.EpisodeIDs) != len(tt.episodeIDs) {
						t.Errorf("expected %d episode IDs, got %d", len(tt.episodeIDs), len(cmdReq.EpisodeIDs))
					}

					for i, id := range tt.episodeIDs {
						if cmdReq.EpisodeIDs[i] != id {
							t.Errorf("expected episode ID %d at index %d, got %d", id, i, cmdReq.EpisodeIDs[i])
						}
					}

					w.WriteHeader(tt.responseCode)
					w.Write([]byte(tt.responseBody))
				}))
				defer server.Close()
			}

			config := types.ServiceConfig{
				Name: "test",
				BaseURL: func() string {
					if server != nil {
						return server.URL
					}
					return "http://localhost:8989"
				}(),
				APIKey: "test-api-key",
			}
			client := NewClient(config)

			resp, err := client.TriggerEpisodeSearch(tt.episodeIDs)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Fatal("expected response to not be nil")
			}

			if resp.ID != tt.expectedResponse.ID {
				t.Errorf("expected response ID %d, got %d", tt.expectedResponse.ID, resp.ID)
			}
			if resp.Name != tt.expectedResponse.Name {
				t.Errorf("expected response Name %q, got %q", tt.expectedResponse.Name, resp.Name)
			}
			if resp.CommandName != tt.expectedResponse.CommandName {
				t.Errorf("expected response CommandName %q, got %q", tt.expectedResponse.CommandName, resp.CommandName)
			}
			if resp.Status != tt.expectedResponse.Status {
				t.Errorf("expected response Status %q, got %q", tt.expectedResponse.Status, resp.Status)
			}
		})
	}
}

func TestMakeRequest(t *testing.T) {
	tests := []struct {
		name         string
		endpoint     string
		responseCode int
		responseBody string
		expectError  bool
	}{
		{
			name:         "successful request",
			endpoint:     "/api/v3/series",
			responseCode: http.StatusOK,
			responseBody: `{"result": "success"}`,
			expectError:  false,
		},
		{
			name:         "not found",
			endpoint:     "/api/v3/notfound",
			responseCode: http.StatusNotFound,
			responseBody: `{"error": "not found"}`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasSuffix(r.URL.Path, tt.endpoint) {
					t.Errorf("expected endpoint to end with %q, got %q", tt.endpoint, r.URL.Path)
				}

				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			config := types.ServiceConfig{
				Name:    "test",
				BaseURL: server.URL,
				APIKey:  "test-api-key",
			}
			client := NewClient(config)

			body, err := client.makeRequest(tt.endpoint, nil)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if string(body) != tt.responseBody {
				t.Errorf("expected response body %q, got %q", tt.responseBody, string(body))
			}
		})
	}
}
