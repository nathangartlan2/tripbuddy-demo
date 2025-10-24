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
	Illinois IllinoisConfig `json:"illinois"`
}

// IllinoisConfig holds Illinois scraper configuration
type IllinoisConfig struct {
	BaseURL   string         `json:"baseURL"`
	StateCode string         `json:"stateCode"`
	Selectors ScraperSelectors `json:"selectors"`
}

// ScraperSelectors holds CSS selectors and JSON paths for scraping
type ScraperSelectors struct {
	// Homepage selectors
	Homepage HomepageSelectors `json:"homepage"`

	// JSON API paths
	JSONAPI JSONAPISelectors `json:"jsonAPI"`

	// Park page selectors
	ParkPage ParkPageSelectors `json:"parkPage"`
}

// HomepageSelectors for finding the JSON API
type HomepageSelectors struct {
	APIURLAttribute string `json:"apiURLAttribute"`
}

// JSONAPISelectors for parsing the JSON response
type JSONAPISelectors struct {
	ParksListPath     string `json:"parksListPath"`      // e.g., "listItems"
	ParkNamePath      string `json:"parkNamePath"`       // e.g., "parkName"
	ParkURLPath       string `json:"parkURLPath"`        // e.g., "meta.dynamicPageLink"
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
