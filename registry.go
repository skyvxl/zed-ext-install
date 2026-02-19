package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	zedAPIBase    = "https://api.zed.dev"
	defaultSchema = 1
)

type ExtensionInfo struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Version        string   `json:"version"`
	Description    string   `json:"description"`
	Authors        []string `json:"authors"`
	Repository     string   `json:"repository"`
	SchemaVersion  int      `json:"schema_version"`
	WasmAPIVersion *string  `json:"wasm_api_version"`
	Provides       []string `json:"provides"`
	PublishedAt    string   `json:"published_at"`
	DownloadCount  int      `json:"download_count"`
}

type searchResponse struct {
	Data []ExtensionInfo `json:"data"`
}

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// SearchExtensions searches the Zed extension registry.
func SearchExtensions(query string) ([]ExtensionInfo, error) {
	url := fmt.Sprintf("%s/extensions?filter=%s&max_schema_version=%d", zedAPIBase, query, defaultSchema)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}
	return result.Data, nil
}

// FindExtension finds an exact extension by ID.
func FindExtension(id string) (*ExtensionInfo, error) {
	results, err := SearchExtensions(id)
	if err != nil {
		return nil, err
	}
	for _, ext := range results {
		if ext.ID == id {
			return &ext, nil
		}
	}
	return nil, fmt.Errorf("extension %q not found", id)
}

// GetDownloadURL returns the API download URL for an extension archive.
// The API handles the pre-signed redirect internally.
func GetDownloadURL(id, version string) string {
	return fmt.Sprintf("%s/extensions/%s/%s/download", zedAPIBase, id, version)
}
