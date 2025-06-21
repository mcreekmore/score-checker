package radarr

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"score-checker/internal/types"
)

func TestNewClient(t *testing.T) {
	config := types.ServiceConfig{
		Name:    "test",
		BaseURL: "http://localhost:7878",
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

func TestGetMovies(t *testing.T) {
	tests := []struct {
		name           string
		responseCode   int
		responseBody   string
		expectedMovies []types.MovieWithFile
		expectError    bool
	}{
		{
			name:         "successful response",
			responseCode: http.StatusOK,
			responseBody: `[
				{
					"id": 1,
					"title": "The Matrix",
					"year": 1999,
					"hasFile": true,
					"movieFile": {
						"id": 101,
						"customFormatScore": -15
					}
				},
				{
					"id": 2,
					"title": "Inception",
					"year": 2010,
					"hasFile": false,
					"movieFile": null
				}
			]`,
			expectedMovies: []types.MovieWithFile{
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
					ID:        2,
					Title:     "Inception",
					Year:      2010,
					HasFile:   false,
					MovieFile: nil,
				},
			},
			expectError: false,
		},
		{
			name:           "empty response",
			responseCode:   http.StatusOK,
			responseBody:   `[]`,
			expectedMovies: []types.MovieWithFile{},
			expectError:    false,
		},
		{
			name:           "server error",
			responseCode:   http.StatusInternalServerError,
			responseBody:   `{"error": "internal server error"}`,
			expectedMovies: nil,
			expectError:    true,
		},
		{
			name:           "invalid json",
			responseCode:   http.StatusOK,
			responseBody:   `invalid json`,
			expectedMovies: nil,
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
				if r.URL.Path != "/api/v3/movie" {
					t.Errorf("expected path '/api/v3/movie', got %q", r.URL.Path)
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

			movies, err := client.GetMovies()

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

			if len(movies) != len(tt.expectedMovies) {
				t.Errorf("expected %d movies, got %d", len(tt.expectedMovies), len(movies))
				return
			}

			for i, expected := range tt.expectedMovies {
				movie := movies[i]
				if movie.ID != expected.ID {
					t.Errorf("movie[%d] expected ID %d, got %d", i, expected.ID, movie.ID)
				}
				if movie.Title != expected.Title {
					t.Errorf("movie[%d] expected Title %q, got %q", i, expected.Title, movie.Title)
				}
				if movie.Year != expected.Year {
					t.Errorf("movie[%d] expected Year %d, got %d", i, expected.Year, movie.Year)
				}
				if movie.HasFile != expected.HasFile {
					t.Errorf("movie[%d] expected HasFile %v, got %v", i, expected.HasFile, movie.HasFile)
				}

				if expected.MovieFile != nil {
					if movie.MovieFile == nil {
						t.Errorf("movie[%d] expected MovieFile to not be nil", i)
						continue
					}
					if movie.MovieFile.ID != expected.MovieFile.ID {
						t.Errorf("movie[%d] expected MovieFile.ID %d, got %d", i, expected.MovieFile.ID, movie.MovieFile.ID)
					}
					if movie.MovieFile.CustomFormatScore != expected.MovieFile.CustomFormatScore {
						t.Errorf("movie[%d] expected MovieFile.CustomFormatScore %d, got %d", i, expected.MovieFile.CustomFormatScore, movie.MovieFile.CustomFormatScore)
					}
				} else if movie.MovieFile != nil {
					t.Errorf("movie[%d] expected MovieFile to be nil", i)
				}
			}
		})
	}
}

func TestTriggerMovieSearch(t *testing.T) {
	tests := []struct {
		name             string
		movieIDs         []int
		responseCode     int
		responseBody     string
		expectedResponse *types.CommandResponse
		expectError      bool
	}{
		{
			name:         "successful search trigger",
			movieIDs:     []int{1, 2, 3},
			responseCode: http.StatusCreated,
			responseBody: `{
				"id": 456,
				"name": "MoviesSearch",
				"commandName": "MoviesSearch",
				"status": "queued"
			}`,
			expectedResponse: &types.CommandResponse{
				ID:          456,
				Name:        "MoviesSearch",
				CommandName: "MoviesSearch",
				Status:      "queued",
			},
			expectError: false,
		},
		{
			name:         "successful search trigger with OK status",
			movieIDs:     []int{1},
			responseCode: http.StatusOK,
			responseBody: `{
				"id": 789,
				"name": "MoviesSearch",
				"commandName": "MoviesSearch",
				"status": "completed"
			}`,
			expectedResponse: &types.CommandResponse{
				ID:          789,
				Name:        "MoviesSearch",
				CommandName: "MoviesSearch",
				Status:      "completed",
			},
			expectError: false,
		},
		{
			name:             "empty movie IDs",
			movieIDs:         []int{},
			responseCode:     0, // won't be used
			responseBody:     "",
			expectedResponse: nil,
			expectError:      true,
		},
		{
			name:             "server error",
			movieIDs:         []int{1, 2},
			responseCode:     http.StatusBadRequest,
			responseBody:     `{"error": "bad request"}`,
			expectedResponse: nil,
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server

			if len(tt.movieIDs) > 0 {
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
					var cmdReq map[string]interface{}
					if err := json.NewDecoder(r.Body).Decode(&cmdReq); err != nil {
						t.Errorf("failed to decode request body: %v", err)
					}

					if cmdReq["name"] != "MoviesSearch" {
						t.Errorf("expected command name 'MoviesSearch', got %q", cmdReq["name"])
					}

					movieIDs, ok := cmdReq["movieIds"].([]interface{})
					if !ok {
						t.Error("expected movieIds to be an array")
					} else if len(movieIDs) != len(tt.movieIDs) {
						t.Errorf("expected %d movie IDs, got %d", len(tt.movieIDs), len(movieIDs))
					} else {
						for i, id := range tt.movieIDs {
							if int(movieIDs[i].(float64)) != id {
								t.Errorf("expected movie ID %d at index %d, got %v", id, i, movieIDs[i])
							}
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
					return "http://localhost:7878"
				}(),
				APIKey: "test-api-key",
			}
			client := NewClient(config)

			resp, err := client.TriggerMovieSearch(tt.movieIDs)

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
			endpoint:     "/api/v3/movie",
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
		{
			name:         "unauthorized",
			endpoint:     "/api/v3/movie",
			responseCode: http.StatusUnauthorized,
			responseBody: `{"error": "unauthorized"}`,
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
