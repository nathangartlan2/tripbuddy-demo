package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// ScraperConfig holds the configuration for the Illinois scraper
type ScraperConfig struct {
	LogLevel     string `json:"logLevel"`
	RequestDelay int    `json:"requestDelay"` // in seconds
	Illinois     IllinoisConfig
}

// IllinoisConfig holds Illinois-specific scraper configuration
type IllinoisConfig struct {
	BaseURL  string
	Homepage HomepageConfig
	ParkPage ParkPageConfig
}

// HomepageConfig holds configuration for scraping the homepage
type HomepageConfig struct {
	Strategy        string
	APIURLAttribute string
	JSONAPIConfig   JSONAPIConfig
}

// JSONAPIConfig holds JSON API scraping configuration
type JSONAPIConfig struct {
	ParksListPath string // e.g., "listItems"
	ParkNamePath  string // e.g., "parkName"
	ParkURLPath   string // e.g., "meta.dynamicPageLink"
}

// ParkPageConfig holds configuration for scraping individual park pages
type ParkPageConfig struct {
	Strategy           string
	NameSelector       string
	LatitudeSelector   string
	LongitudeSelector  string
	ActivitiesSelector string
}

// LoadScraperConfig loads and parses the config.json file
func LoadScraperConfig(filename string) (*ScraperConfig, error) {
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON into intermediate structure
	var rawConfig struct {
		LogLevel     string `json:"logLevel"`
		RequestDelay int    `json:"requestDelay"`
		Scrapers     []struct {
			BaseURL   string `json:"baseURL"`
			StateCode string `json:"stateCode"`
			Pages     struct {
				Homepage struct {
					Strategy  string `json:"strategy"`
					Selectors struct {
						APIURLAttribute string `json:"apiURLAttribute"`
						JSONAPI         struct {
							ParksListPath string `json:"parksListPath"`
							ParkNamePath  string `json:"parkNamePath"`
							ParkURLPath   string `json:"parkURLPath"`
						} `json:"jsonAPI"`
					} `json:"selectors"`
				} `json:"homepage"`
				ParkPage struct {
					Strategy  string `json:"strategy"`
					Selectors struct {
						NameSelector       string `json:"nameSelector"`
						LatitudeSelector   string `json:"latitudeSelector"`
						LongitudeSelector  string `json:"longitudeSelector"`
						ActivitiesSelector string `json:"activitiesSelector"`
					} `json:"selectors"`
				} `json:"parkPage"`
			} `json:"pages"`
		} `json:"scrapers"`
	}

	err = json.Unmarshal(data, &rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Find Illinois config
	var illinoisRaw *struct {
		BaseURL   string `json:"baseURL"`
		StateCode string `json:"stateCode"`
		Pages     struct {
			Homepage struct {
				Strategy  string `json:"strategy"`
				Selectors struct {
					APIURLAttribute string `json:"apiURLAttribute"`
					JSONAPI         struct {
						ParksListPath string `json:"parksListPath"`
						ParkNamePath  string `json:"parkNamePath"`
						ParkURLPath   string `json:"parkURLPath"`
					} `json:"jsonAPI"`
				} `json:"selectors"`
			} `json:"homepage"`
			ParkPage struct {
				Strategy  string `json:"strategy"`
				Selectors struct {
					NameSelector       string `json:"nameSelector"`
					LatitudeSelector   string `json:"latitudeSelector"`
					LongitudeSelector  string `json:"longitudeSelector"`
					ActivitiesSelector string `json:"activitiesSelector"`
				} `json:"selectors"`
			} `json:"parkPage"`
		} `json:"pages"`
	}

	for i := range rawConfig.Scrapers {
		if rawConfig.Scrapers[i].StateCode == "IL" {
			illinoisRaw = &rawConfig.Scrapers[i]
			break
		}
	}

	if illinoisRaw == nil {
		return nil, fmt.Errorf("Illinois (IL) configuration not found in config.json")
	}

	// Build clean config structure
	config := &ScraperConfig{
		LogLevel:     rawConfig.LogLevel,
		RequestDelay: rawConfig.RequestDelay,
		Illinois: IllinoisConfig{
			BaseURL: illinoisRaw.BaseURL,
			Homepage: HomepageConfig{
				Strategy:        illinoisRaw.Pages.Homepage.Strategy,
				APIURLAttribute: illinoisRaw.Pages.Homepage.Selectors.APIURLAttribute,
				JSONAPIConfig: JSONAPIConfig{
					ParksListPath: illinoisRaw.Pages.Homepage.Selectors.JSONAPI.ParksListPath,
					ParkNamePath:  illinoisRaw.Pages.Homepage.Selectors.JSONAPI.ParkNamePath,
					ParkURLPath:   illinoisRaw.Pages.Homepage.Selectors.JSONAPI.ParkURLPath,
				},
			},
			ParkPage: ParkPageConfig{
				Strategy:           illinoisRaw.Pages.ParkPage.Strategy,
				NameSelector:       illinoisRaw.Pages.ParkPage.Selectors.NameSelector,
				LatitudeSelector:   illinoisRaw.Pages.ParkPage.Selectors.LatitudeSelector,
				LongitudeSelector:  illinoisRaw.Pages.ParkPage.Selectors.LongitudeSelector,
				ActivitiesSelector: illinoisRaw.Pages.ParkPage.Selectors.ActivitiesSelector,
			},
		},
	}

	// Set defaults
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	if config.RequestDelay == 0 {
		config.RequestDelay = 1
	}

	return config, nil
}
