package types

import "time"

// ServiceConfig holds connection details for a single service
type ServiceConfig struct {
	Name    string
	BaseURL string
	APIKey  string
}

// Config holds application configuration
type Config struct {
	SonarrInstances []ServiceConfig
	RadarrInstances []ServiceConfig
	TriggerSearch   bool          // Whether to actually trigger searches or just report
	BatchSize       int           // Number of items to check per run
	Interval        time.Duration // How often to run the check
}

// Series represents a Sonarr series (minimal fields needed)
type Series struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// Episode represents a Sonarr episode
type Episode struct {
	ID            int          `json:"id"`
	SeriesID      int          `json:"seriesId"`
	Title         string       `json:"title"`
	SeasonNumber  int          `json:"seasonNumber"`
	EpisodeNumber int          `json:"episodeNumber"`
	HasFile       bool         `json:"hasFile"`
	EpisodeFile   *EpisodeFile `json:"episodeFile"`
}

// EpisodeFile represents episode file info
type EpisodeFile struct {
	ID                int `json:"id"`
	CustomFormatScore int `json:"customFormatScore"`
}

// CommandRequest represents a command to be sent to Sonarr
type CommandRequest struct {
	Name       string `json:"name"`
	EpisodeIDs []int  `json:"episodeIds,omitempty"`
}

// CommandResponse represents the response from a command request
type CommandResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	CommandName string `json:"commandName"`
	Status      string `json:"status"`
}

// LowScoreEpisode represents an episode with a low custom format score
type LowScoreEpisode struct {
	Series            Series
	Episode           Episode
	CustomFormatScore int
}

// Movie represents a Radarr movie (minimal fields needed)
type Movie struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Year  int    `json:"year"`
}

// MovieFile represents movie file info
type MovieFile struct {
	ID                int `json:"id"`
	CustomFormatScore int `json:"customFormatScore"`
}

// MovieWithFile represents a movie with its file information
type MovieWithFile struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	Year      int        `json:"year"`
	HasFile   bool       `json:"hasFile"`
	MovieFile *MovieFile `json:"movieFile"`
}

// LowScoreMovie represents a movie with a low custom format score
type LowScoreMovie struct {
	Movie             MovieWithFile
	CustomFormatScore int
}