package radarr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"score-checker/internal/types"
)

// Client handles API interactions with Radarr
type Client struct {
	config types.ServiceConfig
	client *http.Client
}

// NewClient creates a new Radarr API client
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

// GetMovies fetches all movies from Radarr with file information
func (c *Client) GetMovies() ([]types.MovieWithFile, error) {
	body, err := c.makeRequest("/api/v3/movie", nil)
	if err != nil {
		return nil, fmt.Errorf("fetching movies: %w", err)
	}

	var movies []types.MovieWithFile
	if err := json.Unmarshal(body, &movies); err != nil {
		return nil, fmt.Errorf("unmarshaling movies: %w", err)
	}

	return movies, nil
}

// TriggerMovieSearch triggers a search for better versions of specific movies
func (c *Client) TriggerMovieSearch(movieIDs []int) (*types.CommandResponse, error) {
	if len(movieIDs) == 0 {
		return nil, fmt.Errorf("no movie IDs provided")
	}

	// Create command request - note: Radarr uses "movieIds" instead of "episodeIds"
	commandReq := map[string]interface{}{
		"name":     "MoviesSearch",
		"movieIds": movieIDs,
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
