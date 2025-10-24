package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	Scrapers ScraperConfig `json:"scrapers"`
}

// ScraperConfig holds scraper-specific configuration
type ScraperConfig struct {
	Illinois StateConfig `json:"illinois"`
	Indiana StateConfig `json:"indiana"`
}

// StateConfig holds Illinois scraper configuration
type StateConfig struct {
	BaseURL   string        `json:"baseURL"`
	StateCode string        `json:"stateCode"`
	Pages     PagesConfig   `json:"pages"`
}

// PagesConfig holds configuration for different page types
type PagesConfig struct {
	Homepage HomePage `json:"homepage"`
	ParkPage ParkPage `json:"parkPage"`
}

// HomePage configuration for the homepage/list page
type HomePage struct {
	Strategy  string            `json:"strategy"`  // "json_api" or "static_html"
	Selectors HomepageSelectors `json:"selectors"`
}

// ParkPage configuration for individual park pages
type ParkPage struct {
	Strategy  string            `json:"strategy"`  // Usually "static_html"
	Selectors ParkPageSelectors `json:"selectors"`
}

// HomepageSelectors holds selectors for discovering parks on homepage
type HomepageSelectors struct {
	// For JSON API strategy
	APIURLAttribute string           `json:"apiURLAttribute,omitempty"` // CSS selector for JSON API URL
	JSONAPI         JSONAPISelectors `json:"jsonAPI,omitempty"`         // JSON parsing config

	// For Static HTML strategy
	StaticHTML StaticHTMLConfig `json:"staticHTML,omitempty"` // Static HTML parsing config
}

// JSONAPISelectors for parsing the JSON response
type JSONAPISelectors struct {
	ParksListPath string `json:"parksListPath"` // e.g., "listItems"
	ParkNamePath  string `json:"parkNamePath"`  // e.g., "parkName"
	ParkURLPath   string `json:"parkURLPath"`   // e.g., "meta.dynamicPageLink"
}

// StaticHTMLConfig for parsing static HTML park lists
type StaticHTMLConfig struct {
	Section    HTMLSection `json:"section"`    // Parent section containing park links
	URLElement URLElement  `json:"urlElement"` // URL element configuration
}

// HTMLSection identifies the parent container for park links
type HTMLSection struct {
	ID       string `json:"id,omitempty"`       // HTML id attribute
	Class    string `json:"class,omitempty"`    // HTML class attribute
	Selector string `json:"selector,omitempty"` // CSS selector (most flexible option)
}

// URLElement configures how to identify and extract park URLs
type URLElement struct {
	HrefPattern       string `json:"hrefPattern"`                 // Pattern to match href (e.g., "/dnr/state-parks/parks-lakes/*")
	ParkNameAttribute string `json:"parkNameAttribute,omitempty"` // Attribute to get park name (e.g., "text", "title", "aria-label")
}

// ParkPageSelectors for extracting park details from HTML
type ParkPageSelectors struct {
	NameSelector       string `json:"nameSelector"`       // CSS selector for park name
	LatitudeSelector   string `json:"latitudeSelector"`   // CSS selector for latitude
	LongitudeSelector  string `json:"longitudeSelector"`  // CSS selector for longitude
	ActivitiesSelector string `json:"activitiesSelector"` // CSS selector for activities list
}

// loadConfig reads and parses the configuration file
func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
}
