package scrapers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ILJSONParkUrlGatherer implements ParkUrlGatherer by parsing JSON responses
type ILJSONParkUrlGatherer struct {
	userAgent string
	baseURL   string // Optional base URL to prepend to relative URLs
}

// NewJSONParkUrlGatherer creates a new JSONParkUrlGatherer
func NewJSONParkUrlGatherer(baseURL string) *ILJSONParkUrlGatherer {
	return &ILJSONParkUrlGatherer{
		userAgent: "TripBuddyBot/1.0 (Educational Park Data Scraper; +https://github.com/nathangartlan2/tripbuddy-demo)",
		baseURL:   baseURL,
	}
}

// JSONResponse represents the structure of the JSON response
type JSONResponse struct {
	ListItems []ListItem `json:"listItems"`
}

// ListItem represents an individual item in the listItems array
type ListItem struct {
	Meta Meta `json:"meta"`
}

// Meta contains metadata including the dynamic page link
type Meta struct {
	DynamicPageLink string `json:"dynamicPageLink"`
}

// GatherUrls fetches and parses the JSON from the main page URL to extract park URLs
func (g *ILJSONParkUrlGatherer) GatherUrls(mainPageUrl string) ([]string, error) {
	// Create HTTP request
	req, err := http.NewRequest("GET", mainPageUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", g.userAgent)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON
	var jsonResponse JSONResponse
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract URLs from listItems
	urls := make([]string, 0, len(jsonResponse.ListItems))
	for _, item := range jsonResponse.ListItems {
		if item.Meta.DynamicPageLink != "" {
			// Prepend base URL if provided and link is relative
			url := item.Meta.DynamicPageLink
			if g.baseURL != "" && url[0] == '/' {
				url = g.baseURL + url
			}
			urls = append(urls, url)
		}
	}

	return urls, nil
}
