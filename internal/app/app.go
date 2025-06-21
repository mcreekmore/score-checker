package app

import (
	"fmt"
	"log"
	"time"

	"score-checker/internal/config"
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
	for _, s := range series {
		fmt.Printf("[%s] Checking series: %s (ID: %d)\n", instanceName, s.Title, s.ID)

		episodes, err := client.GetEpisodes(s.ID)
		if err != nil {
			log.Printf("[%s] Warning: failed to get episodes for series %s: %v", instanceName, s.Title, err)
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
						fmt.Printf("[%s] Reached batch limit of %d episodes\n", instanceName, cfg.BatchSize)
						return lowScoreEpisodes, nil
					}
				}
			}
		}
	}

	// Trigger searches if enabled and we have episodes to search
	if cfg.TriggerSearch && len(episodesToSearch) > 0 {
		fmt.Printf("\n[%s] Triggering search for %d episode(s) with low scores...\n", instanceName, len(episodesToSearch))

		// Search in batches to avoid overwhelming the system
		batchSize := 10
		for i := 0; i < len(episodesToSearch); i += batchSize {
			end := min(i+batchSize, len(episodesToSearch))

			batch := episodesToSearch[i:end]
			resp, err := client.TriggerEpisodeSearch(batch)
			if err != nil {
				log.Printf("[%s] Warning: failed to trigger search for episodes %v: %v", instanceName, batch, err)
				continue
			}

			fmt.Printf("[%s] Search triggered for batch: %v (Command ID: %d, Status: %s)\n",
				instanceName, batch, resp.ID, resp.Status)
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
		fmt.Printf("[%s] Checking movie: %s (%d)\n", instanceName, movie.Title, movie.Year)

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
					fmt.Printf("[%s] Reached batch limit of %d movies\n", instanceName, cfg.BatchSize)
					break
				}
			}
		}
	}

	// Trigger searches if enabled and we have movies to search
	if cfg.TriggerSearch && len(moviesToSearch) > 0 {
		fmt.Printf("\n[%s] Triggering search for %d movie(s) with low scores...\n", instanceName, len(moviesToSearch))

		// Search in batches to avoid overwhelming the system
		batchSize := 10
		for i := 0; i < len(moviesToSearch); i += batchSize {
			end := min(i+batchSize, len(moviesToSearch))

			batch := moviesToSearch[i:end]
			resp, err := client.TriggerMovieSearch(batch)
			if err != nil {
				log.Printf("[%s] Warning: failed to trigger search for movies %v: %v", instanceName, batch, err)
				continue
			}

			fmt.Printf("[%s] Search triggered for batch: %v (Command ID: %d, Status: %s)\n",
				instanceName, batch, resp.ID, resp.Status)
		}
	}

	return lowScoreMovies, nil
}

// printLowScoreEpisodes prints episodes with low custom format scores to console
func printLowScoreEpisodes(episodes []types.LowScoreEpisode, triggerSearch bool, instanceName string) {
	if len(episodes) == 0 {
		fmt.Printf("[%s] No episodes found with custom format scores below zero.\n", instanceName)
		return
	}

	fmt.Printf("\n[%s] Found %d episode(s) with custom format scores below zero:\n", instanceName, len(episodes))
	if triggerSearch {
		fmt.Printf("[%s] (Searches have been triggered for these episodes)\n", instanceName)
	} else {
		fmt.Printf("[%s] (Set SCORECHECK_TRIGGER_SEARCH=true to automatically trigger searches)\n", instanceName)
	}
	fmt.Println()

	for _, ep := range episodes {
		fmt.Printf("[%s] Series: %s\n", instanceName, ep.Series.Title)
		fmt.Printf("[%s]   Episode: S%02dE%02d - %s\n", instanceName,
			ep.Episode.SeasonNumber,
			ep.Episode.EpisodeNumber,
			ep.Episode.Title)
		fmt.Printf("[%s]   Custom Format Score: %d\n", instanceName, ep.CustomFormatScore)
		fmt.Printf("[%s]   Episode ID: %d\n\n", instanceName, ep.Episode.ID)
	}
}

// printLowScoreMovies prints movies with low custom format scores to console
func printLowScoreMovies(movies []types.LowScoreMovie, triggerSearch bool, instanceName string) {
	if len(movies) == 0 {
		fmt.Printf("[%s] No movies found with custom format scores below zero.\n", instanceName)
		return
	}

	fmt.Printf("\n[%s] Found %d movie(s) with custom format scores below zero:\n", instanceName, len(movies))
	if triggerSearch {
		fmt.Printf("[%s] (Searches have been triggered for these movies)\n", instanceName)
	} else {
		fmt.Printf("[%s] (Set SCORECHECK_TRIGGER_SEARCH=true to automatically trigger searches)\n", instanceName)
	}
	fmt.Println()

	for _, movie := range movies {
		fmt.Printf("[%s] Movie: %s (%d)\n", instanceName, movie.Movie.Title, movie.Movie.Year)
		fmt.Printf("[%s]   Custom Format Score: %d\n", instanceName, movie.CustomFormatScore)
		fmt.Printf("[%s]   Movie ID: %d\n\n", instanceName, movie.Movie.ID)
	}
}

// RunOnce runs the score checker once
func RunOnce() {
	cfg := config.Load()

	if cfg.TriggerSearch {
		fmt.Println("Search triggering is ENABLED - will automatically search for better versions")
	} else {
		fmt.Println("Search triggering is DISABLED - will only report findings")
	}
	fmt.Printf("Batch size: %d items per run\n", cfg.BatchSize)

	// Process each Sonarr instance
	if len(cfg.SonarrInstances) > 0 {
		fmt.Printf("Found %d Sonarr instance(s)\n", len(cfg.SonarrInstances))
		for _, instance := range cfg.SonarrInstances {
			fmt.Printf("\n=== Checking Sonarr Instance: %s ===\n", instance.Name)
			
			client := sonarr.NewClient(instance)
			fmt.Printf("[%s] Fetching series and checking custom format scores...\n", instance.Name)
			
			lowScoreEpisodes, err := findLowScoreEpisodes(client, cfg, instance.Name)
			if err != nil {
				log.Printf("[%s] Error finding low score episodes: %v", instance.Name, err)
				continue
			}
			
			printLowScoreEpisodes(lowScoreEpisodes, cfg.TriggerSearch, instance.Name)
		}
	}

	// Process each Radarr instance
	if len(cfg.RadarrInstances) > 0 {
		fmt.Printf("Found %d Radarr instance(s)\n", len(cfg.RadarrInstances))
		for _, instance := range cfg.RadarrInstances {
			fmt.Printf("\n=== Checking Radarr Instance: %s ===\n", instance.Name)
			
			client := radarr.NewClient(instance)
			fmt.Printf("[%s] Fetching movies and checking custom format scores...\n", instance.Name)
			
			lowScoreMovies, err := findLowScoreMovies(client, cfg, instance.Name)
			if err != nil {
				log.Printf("[%s] Error finding low score movies: %v", instance.Name, err)
				continue
			}
			
			printLowScoreMovies(lowScoreMovies, cfg.TriggerSearch, instance.Name)
		}
	}

	if len(cfg.SonarrInstances) == 0 && len(cfg.RadarrInstances) == 0 {
		fmt.Println("No Sonarr or Radarr instances configured. Please check your configuration.")
	}
}

// RunDaemon runs the score checker as a daemon
func RunDaemon() {
	cfg := config.Load()
	fmt.Printf("Starting daemon mode with interval: %v\n", cfg.Interval)
	fmt.Printf("Batch size: %d items per run\n", cfg.BatchSize)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	// Run once immediately
	RunOnce()

	// Then run on schedule
	for range ticker.C {
		fmt.Printf("\n=== Scheduled run at %s ===\n", time.Now().Format("2006-01-02 15:04:05"))
		RunOnce()
	}
}