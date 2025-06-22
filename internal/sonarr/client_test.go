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

func TestMakeRequest(t *testing.T) {
	tests := []struct {
		name         string
		endpoint     string
		responseCode int
		responseBody string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "successful request",
			endpoint:     "/api/v3/series",
			responseCode: http.StatusOK,
			responseBody: `[{"id": 1, "title": "Test Series"}]`,
			expectError:  false,
		},
		{
			name:         "server error",
			endpoint:     "/api/v3/series",
			responseCode: http.StatusInternalServerError,
			responseBody: `{"error": "internal server error"}`,
			expectError:  true,
			errorMessage: "API request failed with status 500",
		},
		{
			name:         "not found error",
			endpoint:     "/api/v3/series",
			responseCode: http.StatusNotFound,
			responseBody: `{"error": "not found"}`,
			expectError:  true,
			errorMessage: "API request failed with status 404",
		},
		{
			name:         "unauthorized error",
			endpoint:     "/api/v3/series",
			responseCode: http.StatusUnauthorized,
			responseBody: `{"error": "unauthorized"}`,
			expectError:  true,
			errorMessage: "API request failed with status 401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				_, _ = w.Write([]byte(tt.responseBody))
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
				} else if !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("expected error message to contain %q, got %q", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if string(body) != tt.responseBody {
					t.Errorf("expected body %q, got %q", tt.responseBody, string(body))
				}
			}
		})
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
				_, _ = w.Write([]byte(tt.responseBody))
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

func verifyEpisodesRequest(t *testing.T, r *http.Request, expectedSeriesID int) {
	if r.Header.Get("X-Api-Key") != "test-api-key" {
		t.Errorf("expected X-Api-Key header 'test-api-key', got %q", r.Header.Get("X-Api-Key"))
	}

	if r.URL.Path != "/api/v3/episode" {
		t.Errorf("expected path '/api/v3/episode', got %q", r.URL.Path)
	}

	seriesID := r.URL.Query().Get("seriesId")
	expectedSeriesIDStr := fmt.Sprintf("%d", expectedSeriesID)
	if seriesID != expectedSeriesIDStr {
		t.Errorf("expected seriesId query param %q, got %q", expectedSeriesIDStr, seriesID)
	}

	includeEpisodeFile := r.URL.Query().Get("includeEpisodeFile")
	if includeEpisodeFile != "true" {
		t.Errorf("expected includeEpisodeFile query param 'true', got %q", includeEpisodeFile)
	}
}

func validateEpisode(t *testing.T, episode, expected types.Episode, index int) {
	if episode.ID != expected.ID {
		t.Errorf("episode[%d] expected ID %d, got %d", index, expected.ID, episode.ID)
	}
	if episode.SeriesID != expected.SeriesID {
		t.Errorf("episode[%d] expected SeriesID %d, got %d", index, expected.SeriesID, episode.SeriesID)
	}
	if episode.Title != expected.Title {
		t.Errorf("episode[%d] expected Title %q, got %q", index, expected.Title, episode.Title)
	}
	if episode.SeasonNumber != expected.SeasonNumber {
		t.Errorf("episode[%d] expected SeasonNumber %d, got %d", index, expected.SeasonNumber, episode.SeasonNumber)
	}
	if episode.EpisodeNumber != expected.EpisodeNumber {
		t.Errorf("episode[%d] expected EpisodeNumber %d, got %d", index, expected.EpisodeNumber, episode.EpisodeNumber)
	}
	if episode.HasFile != expected.HasFile {
		t.Errorf("episode[%d] expected HasFile %v, got %v", index, expected.HasFile, episode.HasFile)
	}
	validateEpisodeFile(t, episode.EpisodeFile, expected.EpisodeFile, index)
}

func validateEpisodeFile(t *testing.T, actual, expected *types.EpisodeFile, index int) {
	if expected != nil {
		if actual == nil {
			t.Errorf("episode[%d] expected EpisodeFile to not be nil", index)
			return
		}
		if actual.ID != expected.ID {
			t.Errorf("episode[%d] expected EpisodeFile.ID %d, got %d", index, expected.ID, actual.ID)
		}
		if actual.CustomFormatScore != expected.CustomFormatScore {
			t.Errorf("episode[%d] expected EpisodeFile.CustomFormatScore %d, got %d", index, expected.CustomFormatScore, actual.CustomFormatScore)
		}
	} else if actual != nil {
		t.Errorf("episode[%d] expected EpisodeFile to be nil", index)
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
				verifyEpisodesRequest(t, r, tt.seriesID)
				w.WriteHeader(tt.responseCode)
				_, _ = w.Write([]byte(tt.responseBody))
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
				validateEpisode(t, episodes[i], expected, i)
			}
		})
	}
}

func verifyEpisodeSearchRequest(t *testing.T, r *http.Request, expectedEpisodeIDs []int) {
	if r.Header.Get("X-Api-Key") != "test-api-key" {
		t.Errorf("expected X-Api-Key header 'test-api-key', got %q", r.Header.Get("X-Api-Key"))
	}

	if r.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type header 'application/json', got %q", r.Header.Get("Content-Type"))
	}

	if r.URL.Path != "/api/v3/command" {
		t.Errorf("expected path '/api/v3/command', got %q", r.URL.Path)
	}

	if r.Method != "POST" {
		t.Errorf("expected POST method, got %q", r.Method)
	}

	var cmdReq types.CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&cmdReq); err != nil {
		t.Errorf("failed to decode request body: %v", err)
		return
	}

	verifyCommandRequest(t, cmdReq, expectedEpisodeIDs)
}

func verifyCommandRequest(t *testing.T, cmdReq types.CommandRequest, expectedEpisodeIDs []int) {
	if cmdReq.Name != "EpisodeSearch" {
		t.Errorf("expected command name 'EpisodeSearch', got %q", cmdReq.Name)
	}

	if len(cmdReq.EpisodeIDs) != len(expectedEpisodeIDs) {
		t.Errorf("expected %d episode IDs, got %d", len(expectedEpisodeIDs), len(cmdReq.EpisodeIDs))
		return
	}

	for i, id := range expectedEpisodeIDs {
		if cmdReq.EpisodeIDs[i] != id {
			t.Errorf("expected episode ID %d at index %d, got %d", id, i, cmdReq.EpisodeIDs[i])
		}
	}
}

func validateCommandResponse(t *testing.T, actual, expected *types.CommandResponse) {
	if actual.ID != expected.ID {
		t.Errorf("expected response ID %d, got %d", expected.ID, actual.ID)
	}
	if actual.Name != expected.Name {
		t.Errorf("expected response Name %q, got %q", expected.Name, actual.Name)
	}
	if actual.CommandName != expected.CommandName {
		t.Errorf("expected response CommandName %q, got %q", expected.CommandName, actual.CommandName)
	}
	if actual.Status != expected.Status {
		t.Errorf("expected response Status %q, got %q", expected.Status, actual.Status)
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
			responseCode:     0,
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
					verifyEpisodeSearchRequest(t, r, tt.episodeIDs)
					w.WriteHeader(tt.responseCode)
					_, _ = w.Write([]byte(tt.responseBody))
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

			validateCommandResponse(t, resp, tt.expectedResponse)
		})
	}
}
