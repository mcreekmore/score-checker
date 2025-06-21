package sonarr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"score-checker/internal/types"
)

// Client handles API interactions with Sonarr
type Client struct {
	config types.ServiceConfig
	client *http.Client
}

// NewClient creates a new Sonarr API client
func NewClient(config types.ServiceConfig) *Client {
	return &Client{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest handles common HTTP request logic with authentication
func (c *Client) makeRequest(endpoint string, params url.Values) ([]byte, error) {
	// Build URL
	u, err := url.Parse(c.config.BaseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Add query parameters
	if params != nil {
		u.RawQuery = params.Encode()
	}

	// Create request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Add API key authentication
	req.Header.Set("X-Api-Key", c.config.APIKey)

	// Make request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return body, nil
}

// GetSeries fetches all series from Sonarr
func (c *Client) GetSeries() ([]types.Series, error) {
	body, err := c.makeRequest("/api/v3/series", nil)
	if err != nil {
		return nil, fmt.Errorf("fetching series: %w", err)
	}

	var series []types.Series
	if err := json.Unmarshal(body, &series); err != nil {
		return nil, fmt.Errorf("unmarshaling series: %w", err)
	}

	return series, nil
}

// GetEpisodes fetches episodes for a specific series with episode file information
func (c *Client) GetEpisodes(seriesID int) ([]types.Episode, error) {
	params := url.Values{}
	params.Set("seriesId", strconv.Itoa(seriesID))
	params.Set("includeEpisodeFile", "true")

	body, err := c.makeRequest("/api/v3/episode", params)
	if err != nil {
		return nil, fmt.Errorf("fetching episodes for series %d: %w", seriesID, err)
	}

	var episodes []types.Episode
	if err := json.Unmarshal(body, &episodes); err != nil {
		return nil, fmt.Errorf("unmarshaling episodes: %w", err)
	}

	return episodes, nil
}

// TriggerEpisodeSearch triggers a search for better versions of specific episodes
func (c *Client) TriggerEpisodeSearch(episodeIDs []int) (*types.CommandResponse, error) {
	if len(episodeIDs) == 0 {
		return nil, fmt.Errorf("no episode IDs provided")
	}

	// Create command request
	commandReq := types.CommandRequest{
		Name:       "EpisodeSearch",
		EpisodeIDs: episodeIDs,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(commandReq)
	if err != nil {
		return nil, fmt.Errorf("marshaling command request: %w", err)
	}

	// Build URL
	u, err := url.Parse(c.config.BaseURL + "/api/v3/command")
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Create POST request
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	req.Header.Set("X-Api-Key", c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// Parse response
	var commandResp types.CommandResponse
	if err := json.Unmarshal(body, &commandResp); err != nil {
		return nil, fmt.Errorf("unmarshaling command response: %w", err)
	}

	return &commandResp, nil
}
