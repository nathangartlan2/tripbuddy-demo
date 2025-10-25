package main

import (
	"encoding/json"
	"fmt"
	"os"

	"scraper/scrapers"
)

// Config holds the application configuration
type Config struct {
	Scrapers map[string]StateConfig `json:"-"` // Map of state code -> config
}

// configJSON is used for unmarshaling the JSON array
type configJSON struct {
	Scrapers []StateConfig `json:"scrapers"`
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
	Strategy  string                    `json:"strategy"`  // "json_api" or "static_html"
	Selectors scrapers.HomepageSelectors `json:"selectors"`
}

// ParkPage configuration for individual park pages
type ParkPage struct {
	Strategy  string                      `json:"strategy"`  // Usually "static_html"
	Selectors scrapers.ParkPageSelectors `json:"selectors"`
}

// loadConfig reads and parses the configuration file
func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// First unmarshal into the array structure
	var jsonConfig configJSON
	err = json.Unmarshal(data, &jsonConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Convert array to map keyed by state code
	config := &Config{
		Scrapers: make(map[string]StateConfig),
	}

	for _, scraperConfig := range jsonConfig.Scrapers {
		if scraperConfig.StateCode == "" {
			return nil, fmt.Errorf("scraper configuration missing stateCode")
		}
		config.Scrapers[scraperConfig.StateCode] = scraperConfig
	}

	return config, nil
}
