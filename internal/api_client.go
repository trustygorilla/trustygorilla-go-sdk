package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// flagsResponse mirrors the top-level JSON envelope returned by GET /v1/flags.
type flagsResponse struct {
	Flags map[string]RemoteFlag `json:"flags"`
}

// ApiClient fetches feature flags from the TrustyGorilla REST API.
// It uses ETag-based conditional requests to avoid redundant data transfers.
type ApiClient struct {
	apiURL string
	apiKey string
	http   *http.Client
	etag   string
}

// NewApiClient constructs an ApiClient with the given base URL and API key.
// Connect timeout is 5 s; overall request timeout is 10 s.
func NewApiClient(apiURL, apiKey string) *ApiClient {
	transport := &http.Transport{
		ResponseHeaderTimeout: 10 * time.Second,
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}
	return &ApiClient{
		apiURL: apiURL,
		apiKey: apiKey,
		http:   client,
	}
}

// FetchFlags calls GET /v1/flags and returns the parsed flags.
// It returns nil when the server responds with 304 (cache still valid).
// Non-200/304 responses are logged as warnings and an empty slice is returned.
func (c *ApiClient) FetchFlags() ([]RemoteFlag, error) {
	url := fmt.Sprintf("%s/v1/flags", c.apiURL)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("api_client: build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	if c.etag != "" {
		req.Header.Set("If-None-Match", c.etag)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api_client: request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotModified:
		// Cache is still valid; caller should keep using existing data.
		return nil, nil

	case http.StatusOK:
		var envelope flagsResponse
		if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
			return nil, fmt.Errorf("api_client: decode response: %w", err)
		}

		flags := make([]RemoteFlag, 0, len(envelope.Flags))
		for key, flag := range envelope.Flags {
			flag.Key = key
			flags = append(flags, flag)
		}

		if etag := resp.Header.Get("ETag"); etag != "" {
			c.etag = etag
		}

		return flags, nil

	default:
		log.Printf("[TrustyGorilla] WARNING: unexpected status %d from %s", resp.StatusCode, url)
		return []RemoteFlag{}, nil
	}
}
