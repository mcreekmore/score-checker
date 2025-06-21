package app

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"score-checker/internal/radarr"
	"score-checker/internal/sonarr"
	"score-checker/internal/testhelpers"
	"score-checker/internal/types"
)

func TestFindLowScoreEpisodes(t *testing.T) {
	tests := []struct {
		name                   string
		config                 types.Config
		instanceName           string
		series                 []types.Series
		episodes               map[int][]types.Episode
		expectedLowScoreCount  int
		expectCommandTriggered bool
	}{
		{
			name: "finds low score episodes without triggering search",
			config: types.Config{
				TriggerSearch: false,
				BatchSize:     5,
			},
			instanceName:           "test",
			series:                 testhelpers.CreateTestSeries(),
			episodes:               testhelpers.CreateTestEpisodes(),
			expectedLowScoreCount:  2, // episodes with scores -10 and -5
			expectCommandTriggered: false,
		},
		{
			name: "finds low score episodes and triggers search",
			config: types.Config{
				TriggerSearch: true,
				BatchSize:     5,
			},
			instanceName:           "test",
			series:                 testhelpers.CreateTestSeries(),
			episodes:               testhelpers.CreateTestEpisodes(),
			expectedLowScoreCount:  2,
			expectCommandTriggered: true,
		},
		{
			name: "respects batch size limit",
			config: types.Config{
				TriggerSearch: false,
				BatchSize:     1,
			},
			instanceName:           "test",
			series:                 testhelpers.CreateTestSeries(),
			episodes:               testhelpers.CreateTestEpisodes(),
			expectedLowScoreCount:  1, // limited by batch size
			expectCommandTriggered: false,
		},
		{
			name: "handles empty series list",
			config: types.Config{
				TriggerSearch: false,
				BatchSize:     5,
			},
			instanceName:           "test",
			series:                 []types.Series{},
			episodes:               map[int][]types.Episode{},
			expectedLowScoreCount:  0,
			expectCommandTriggered: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var commandResponse *types.CommandResponse
			if tt.expectCommandTriggered {
				commandResponse = testhelpers.CreateTestCommandResponse()
			}

			server := testhelpers.MockSonarrServer(t, tt.series, tt.episodes, commandResponse)
			defer server.Close()

			config := types.ServiceConfig{
				Name:    tt.instanceName,
				BaseURL: server.URL,
				APIKey:  "test-api-key",
			}
			client := sonarr.NewClient(config)

			lowScoreEpisodes, err := findLowScoreEpisodes(client, tt.config, tt.instanceName)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(lowScoreEpisodes) != tt.expectedLowScoreCount {
				t.Errorf("expected %d low score episodes, got %d", tt.expectedLowScoreCount, len(lowScoreEpisodes))
			}

			// Verify that all returned episodes have negative scores
			for i, episode := range lowScoreEpisodes {
				if episode.CustomFormatScore >= 0 {
					t.Errorf("episode[%d] expected negative score, got %d", i, episode.CustomFormatScore)
				}
			}
		})
	}
}

func TestFindLowScoreMovies(t *testing.T) {
	tests := []struct {
		name                   string
		config                 types.Config
		instanceName           string
		movies                 []types.MovieWithFile
		expectedLowScoreCount  int
		expectCommandTriggered bool
	}{
		{
			name: "finds low score movies without triggering search",
			config: types.Config{
				TriggerSearch: false,
				BatchSize:     5,
			},
			instanceName:           "test",
			movies:                 testhelpers.CreateTestMovies(),
			expectedLowScoreCount:  1, // only The Matrix has negative score
			expectCommandTriggered: false,
		},
		{
			name: "finds low score movies and triggers search",
			config: types.Config{
				TriggerSearch: true,
				BatchSize:     5,
			},
			instanceName:           "test",
			movies:                 testhelpers.CreateTestMovies(),
			expectedLowScoreCount:  1,
			expectCommandTriggered: true,
		},
		{
			name: "respects batch size limit",
			config: types.Config{
				TriggerSearch: false,
				BatchSize:     0, // should not limit when set to 0
			},
			instanceName:           "test",
			movies:                 testhelpers.CreateTestMovies(),
			expectedLowScoreCount:  1,
			expectCommandTriggered: false,
		},
		{
			name: "handles empty movies list",
			config: types.Config{
				TriggerSearch: false,
				BatchSize:     5,
			},
			instanceName:           "test",
			movies:                 []types.MovieWithFile{},
			expectedLowScoreCount:  0,
			expectCommandTriggered: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var commandResponse *types.CommandResponse
			if tt.expectCommandTriggered {
				commandResponse = &types.CommandResponse{
					ID:          456,
					Name:        "MoviesSearch",
					CommandName: "MoviesSearch",
					Status:      "queued",
				}
			}

			server := testhelpers.MockRadarrServer(t, tt.movies, commandResponse)
			defer server.Close()

			config := types.ServiceConfig{
				Name:    tt.instanceName,
				BaseURL: server.URL,
				APIKey:  "test-api-key",
			}
			client := radarr.NewClient(config)

			lowScoreMovies, err := findLowScoreMovies(client, tt.config, tt.instanceName)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(lowScoreMovies) != tt.expectedLowScoreCount {
				t.Errorf("expected %d low score movies, got %d", tt.expectedLowScoreCount, len(lowScoreMovies))
			}

			// Verify that all returned movies have negative scores
			for i, movie := range lowScoreMovies {
				if movie.CustomFormatScore >= 0 {
					t.Errorf("movie[%d] expected negative score, got %d", i, movie.CustomFormatScore)
				}
			}
		})
	}
}

func TestPrintLowScoreEpisodes(t *testing.T) {
	tests := []struct {
		name          string
		episodes      []types.LowScoreEpisode
		triggerSearch bool
		instanceName  string
		expectedOutput []string
	}{
		{
			name: "prints episodes with search disabled",
			episodes: []types.LowScoreEpisode{
				{
					Series: types.Series{ID: 1, Title: "Breaking Bad"},
					Episode: types.Episode{
						ID:            101,
						Title:         "Pilot",
						SeasonNumber:  1,
						EpisodeNumber: 1,
					},
					CustomFormatScore: -10,
				},
			},
			triggerSearch:  false,
			instanceName:   "test",
			expectedOutput: []string{"[test]", "Breaking Bad", "S01E01", "Pilot", "Custom Format Score: -10", "Episode ID: 101", "Set SCORECHECK_TRIGGER_SEARCH=true"},
		},
		{
			name: "prints episodes with search enabled",
			episodes: []types.LowScoreEpisode{
				{
					Series: types.Series{ID: 1, Title: "Better Call Saul"},
					Episode: types.Episode{
						ID:            201,
						Title:         "Uno",
						SeasonNumber:  1,
						EpisodeNumber: 1,
					},
					CustomFormatScore: -5,
				},
			},
			triggerSearch:  true,
			instanceName:   "test",
			expectedOutput: []string{"[test]", "Better Call Saul", "S01E01", "Uno", "Custom Format Score: -5", "Episode ID: 201", "Searches have been triggered"},
		},
		{
			name:          "handles empty episodes list",
			episodes:      []types.LowScoreEpisode{},
			triggerSearch: false,
			instanceName:  "test",
			expectedOutput: []string{"[test]", "No episodes found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printLowScoreEpisodes(tt.episodes, tt.triggerSearch, tt.instanceName)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check that all expected strings are present
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got: %s", expected, output)
				}
			}
		})
	}
}

func TestPrintLowScoreMovies(t *testing.T) {
	tests := []struct {
		name          string
		movies        []types.LowScoreMovie
		triggerSearch bool
		instanceName  string
		expectedOutput []string
	}{
		{
			name: "prints movies with search disabled",
			movies: []types.LowScoreMovie{
				{
					Movie: types.MovieWithFile{
						ID:    1,
						Title: "The Matrix",
						Year:  1999,
					},
					CustomFormatScore: -15,
				},
			},
			triggerSearch:  false,
			instanceName:   "test",
			expectedOutput: []string{"[test]", "The Matrix (1999)", "Custom Format Score: -15", "Movie ID: 1", "Set SCORECHECK_TRIGGER_SEARCH=true"},
		},
		{
			name: "prints movies with search enabled",
			movies: []types.LowScoreMovie{
				{
					Movie: types.MovieWithFile{
						ID:    2,
						Title: "Blade Runner",
						Year:  1982,
					},
					CustomFormatScore: -20,
				},
			},
			triggerSearch:  true,
			instanceName:   "test",
			expectedOutput: []string{"[test]", "Blade Runner (1982)", "Custom Format Score: -20", "Movie ID: 2", "Searches have been triggered"},
		},
		{
			name:          "handles empty movies list",
			movies:        []types.LowScoreMovie{},
			triggerSearch: false,
			instanceName:  "test",
			expectedOutput: []string{"[test]", "No movies found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printLowScoreMovies(tt.movies, tt.triggerSearch, tt.instanceName)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check that all expected strings are present
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got: %s", expected, output)
				}
			}
		})
	}
}

// Integration test for the RunOnce function
func TestRunOnceIntegration(t *testing.T) {
	// This test would require mocking the config.Load() function
	// which is challenging without dependency injection
	// In a real-world scenario, we'd refactor RunOnce to accept a config parameter
	t.Skip("Integration test requires config mocking - would need refactoring for better testability")
}


// Benchmark for findLowScoreEpisodes
func BenchmarkFindLowScoreEpisodes(b *testing.B) {
	series := testhelpers.CreateTestSeries()
	episodes := testhelpers.CreateTestEpisodes()
	config := types.Config{
		TriggerSearch: false,
		BatchSize:     100,
	}

	server := testhelpers.MockSonarrServer(b, series, episodes, nil)
	defer server.Close()

	serviceConfig := types.ServiceConfig{
		Name:    "benchmark",
		BaseURL: server.URL,
		APIKey:  "test-api-key",
	}
	client := sonarr.NewClient(serviceConfig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := findLowScoreEpisodes(client, config, "benchmark")
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// Benchmark for findLowScoreMovies
func BenchmarkFindLowScoreMovies(b *testing.B) {
	movies := testhelpers.CreateTestMovies()
	config := types.Config{
		TriggerSearch: false,
		BatchSize:     100,
	}

	server := testhelpers.MockRadarrServer(b, movies, nil)
	defer server.Close()

	serviceConfig := types.ServiceConfig{
		Name:    "benchmark",
		BaseURL: server.URL,
		APIKey:  "test-api-key",
	}
	client := radarr.NewClient(serviceConfig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := findLowScoreMovies(client, config, "benchmark")
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}