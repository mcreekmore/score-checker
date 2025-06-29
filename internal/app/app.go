package app

import (
	"fmt"
	"log/slog"
	"time"

	"score-checker/internal/config"
	"score-checker/internal/constants"
	"score-checker/internal/radarr"
	"score-checker/internal/sonarr"
	"score-checker/internal/types"
)

// findLowScoreEpisodes finds episodes with custom format scores below zero
// and optionally triggers searches for better versions
// batchSize limits how many episodes to process per run (0 = unlimited)
func findLowScoreEpisodes(client *sonarr.Client, cfg types.Config, instanceName string) ([]types.LowScoreEpisode, error) {
	// Get all series
	series, err := client.GetSeries()
	if err != nil {
		return nil, fmt.Errorf("getting series: %w", err)
	}

	var lowScoreEpisodes []types.LowScoreEpisode
	var episodesToSearch []int

	// Check each series
	processedCount := 0
	reachedLimit := false
	for _, s := range series {
		if reachedLimit {
			break
		}

		slog.Debug(fmt.Sprintf("[%s] Checking series: %s (ID: %d)", instanceName, s.Title, s.ID))

		episodes, err := client.GetEpisodes(s.ID)
		if err != nil {
			slog.Error(fmt.Sprintf("[%s] Warning: failed to get episodes for series %s: %v", instanceName, s.Title, err))
			continue
		}

		// Check each episode that has a file
		for _, episode := range episodes {
			if episode.HasFile && episode.EpisodeFile != nil {
				if episode.EpisodeFile.CustomFormatScore < 0 {
					lowScoreEpisodes = append(lowScoreEpisodes, types.LowScoreEpisode{
						Series:            s,
						Episode:           episode,
						CustomFormatScore: episode.EpisodeFile.CustomFormatScore,
					})

					// Collect episode IDs for search if enabled
					if cfg.TriggerSearch {
						episodesToSearch = append(episodesToSearch, episode.ID)
					}

					processedCount++
					// Stop if we've reached the batch limit
					if cfg.BatchSize > 0 && processedCount >= cfg.BatchSize {
						slog.Info(fmt.Sprintf("[%s] Reached batch limit of %d episodes", instanceName, cfg.BatchSize))
						reachedLimit = true
						break
					}
				}
			}
		}
	}

	// Trigger searches if enabled and we have episodes to search
	if cfg.TriggerSearch && len(episodesToSearch) > 0 {
		slog.Info(fmt.Sprintf("[%s] Triggering search for %d episode(s) with low scores...", instanceName, len(episodesToSearch)))

		// Search in batches to avoid overwhelming the system
		batchSize := constants.DefaultSearchBatchSize
		for i := 0; i < len(episodesToSearch); i += batchSize {
			end := min(i+batchSize, len(episodesToSearch))

			batch := episodesToSearch[i:end]
			resp, err := client.TriggerEpisodeSearch(batch)
			if err != nil {
				slog.Error(fmt.Sprintf("[%s] Warning: failed to trigger search for episodes %v: %v", instanceName, batch, err))
				continue
			}

			slog.Info(fmt.Sprintf("[%s] Search triggered for batch: %v (Command ID: %d, Status: %s)",
				instanceName, batch, resp.ID, resp.Status))
		}
	}

	return lowScoreEpisodes, nil
}

// findLowScoreMovies finds movies with custom format scores below zero
// and optionally triggers searches for better versions
func findLowScoreMovies(client *radarr.Client, cfg types.Config, instanceName string) ([]types.LowScoreMovie, error) {
	// Get all movies
	movies, err := client.GetMovies()
	if err != nil {
		return nil, fmt.Errorf("getting movies: %w", err)
	}

	var lowScoreMovies []types.LowScoreMovie
	var moviesToSearch []int

	// Check each movie that has a file
	processedCount := 0
	for _, movie := range movies {
		slog.Debug(fmt.Sprintf("[%s] Checking movie: %s (%d)", instanceName, movie.Title, movie.Year))

		if movie.HasFile && movie.MovieFile != nil {
			if movie.MovieFile.CustomFormatScore < 0 {
				lowScoreMovies = append(lowScoreMovies, types.LowScoreMovie{
					Movie:             movie,
					CustomFormatScore: movie.MovieFile.CustomFormatScore,
				})

				// Collect movie IDs for search if enabled
				if cfg.TriggerSearch {
					moviesToSearch = append(moviesToSearch, movie.ID)
				}

				processedCount++
				// Stop if we've reached the batch limit
				if cfg.BatchSize > 0 && processedCount >= cfg.BatchSize {
					slog.Info(fmt.Sprintf("[%s] Reached batch limit of %d movies", instanceName, cfg.BatchSize))
					break
				}
			}
		}
	}

	// Trigger searches if enabled and we have movies to search
	if cfg.TriggerSearch && len(moviesToSearch) > 0 {
		slog.Info(fmt.Sprintf("[%s] Triggering search for %d movie(s) with low scores...", instanceName, len(moviesToSearch)))

		// Search in batches to avoid overwhelming the system
		batchSize := constants.DefaultSearchBatchSize
		for i := 0; i < len(moviesToSearch); i += batchSize {
			end := min(i+batchSize, len(moviesToSearch))

			batch := moviesToSearch[i:end]
			resp, err := client.TriggerMovieSearch(batch)
			if err != nil {
				slog.Error(fmt.Sprintf("[%s] Warning: failed to trigger search for movies %v: %v", instanceName, batch, err))
				continue
			}

			slog.Info(fmt.Sprintf("[%s] Search triggered for batch: %v (Command ID: %d, Status: %s)",
				instanceName, batch, resp.ID, resp.Status))
		}
	}

	return lowScoreMovies, nil
}

// printLowScoreEpisodes prints episodes with low custom format scores to console
func printLowScoreEpisodes(episodes []types.LowScoreEpisode, triggerSearch bool, instanceName string) {
	if len(episodes) == 0 {
		slog.Info(fmt.Sprintf("[%s] No episodes found with custom format scores below zero.", instanceName))
		return
	}

	slog.Info(fmt.Sprintf("[%s] Found %d episode(s) with custom format scores below zero:", instanceName, len(episodes)))
	if triggerSearch {
		slog.Info(fmt.Sprintf("[%s] (Searches have been triggered for these episodes)", instanceName))
	} else {
		slog.Info(fmt.Sprintf("[%s] (Set SCORECHECK_TRIGGER_SEARCH=true to automatically trigger searches)", instanceName))
	}

	for _, ep := range episodes {
		slog.Debug(fmt.Sprintf("[%s] Series: %s", instanceName, ep.Series.Title))
		slog.Debug(fmt.Sprintf("[%s]   Episode: S%02dE%02d - %s", instanceName,
			ep.Episode.SeasonNumber,
			ep.Episode.EpisodeNumber,
			ep.Episode.Title))
		slog.Debug(fmt.Sprintf("[%s]   Custom Format Score: %d", instanceName, ep.CustomFormatScore))
		slog.Debug(fmt.Sprintf("[%s]   Episode ID: %d", instanceName, ep.Episode.ID))
	}
}

// printLowScoreMovies prints movies with low custom format scores to console
func printLowScoreMovies(movies []types.LowScoreMovie, triggerSearch bool, instanceName string) {
	if len(movies) == 0 {
		slog.Info(fmt.Sprintf("[%s] No movies found with custom format scores below zero.", instanceName))
		return
	}

	slog.Info(fmt.Sprintf("[%s] Found %d movie(s) with custom format scores below zero:", instanceName, len(movies)))
	if triggerSearch {
		slog.Info(fmt.Sprintf("[%s] (Searches have been triggered for these movies)", instanceName))
	} else {
		slog.Info(fmt.Sprintf("[%s] (Set SCORECHECK_TRIGGER_SEARCH=true to automatically trigger searches)", instanceName))
	}

	for _, movie := range movies {
		slog.Debug(fmt.Sprintf("[%s] Movie: %s (%d)", instanceName, movie.Movie.Title, movie.Movie.Year))
		slog.Debug(fmt.Sprintf("[%s]   Custom Format Score: %d", instanceName, movie.CustomFormatScore))
		slog.Debug(fmt.Sprintf("[%s]   Movie ID: %d", instanceName, movie.Movie.ID))
	}
}

// RunOnce runs the score checker once
func RunOnce() {
	cfg := config.Load()

	if cfg.TriggerSearch {
		slog.Info("Search triggering is ENABLED - will automatically search for better versions")
	} else {
		slog.Info("Search triggering is DISABLED - will only report findings")
	}
	slog.Info(fmt.Sprintf("Batch size: %d items per run", cfg.BatchSize))
	slog.Debug(fmt.Sprintf("Log level: %s", cfg.LogLevel))

	// Process each Sonarr instance
	if len(cfg.SonarrInstances) > 0 {
		slog.Info(fmt.Sprintf("Found %d Sonarr instance(s)", len(cfg.SonarrInstances)))
		for _, instance := range cfg.SonarrInstances {
			slog.Info(fmt.Sprintf("=== Checking Sonarr Instance: %s ===", instance.Name))

			client := sonarr.NewClient(instance)
			slog.Info(fmt.Sprintf("[%s] Fetching series and checking custom format scores...", instance.Name))

			lowScoreEpisodes, err := findLowScoreEpisodes(client, cfg, instance.Name)
			if err != nil {
				slog.Error(fmt.Sprintf("[%s] Error finding low score episodes: %v", instance.Name, err))
				continue
			}

			printLowScoreEpisodes(lowScoreEpisodes, cfg.TriggerSearch, instance.Name)
		}
	}

	// Process each Radarr instance
	if len(cfg.RadarrInstances) > 0 {
		slog.Info(fmt.Sprintf("Found %d Radarr instance(s)", len(cfg.RadarrInstances)))
		for _, instance := range cfg.RadarrInstances {
			slog.Info(fmt.Sprintf("=== Checking Radarr Instance: %s ===", instance.Name))

			client := radarr.NewClient(instance)
			slog.Info(fmt.Sprintf("[%s] Fetching movies and checking custom format scores...", instance.Name))

			lowScoreMovies, err := findLowScoreMovies(client, cfg, instance.Name)
			if err != nil {
				slog.Error(fmt.Sprintf("[%s] Error finding low score movies: %v", instance.Name, err))
				continue
			}

			printLowScoreMovies(lowScoreMovies, cfg.TriggerSearch, instance.Name)
		}
	}

	if len(cfg.SonarrInstances) == 0 && len(cfg.RadarrInstances) == 0 {
		slog.Info("No Sonarr or Radarr instances configured. Please check your configuration.")
	}
}

// RunDaemon runs the score checker as a daemon
func RunDaemon() {
	cfg := config.Load()
	slog.Info(fmt.Sprintf("Starting daemon mode with interval: %v", cfg.Interval))

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	// Run once immediately
	RunOnce()

	// Then run on schedule
	for range ticker.C {
		slog.Info(fmt.Sprintf("=== Scheduled run at %s ===", time.Now().Format("2006-01-02 15:04:05")))
		RunOnce()
	}
}
