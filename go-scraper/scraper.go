package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"scraper/scrapers"
)

func main() {
	// Load configuration
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create activity scraper
	activityScraperConfig := scrapers.ActivityScraperConfig{
		ActivitiesSelector: config.Scrapers.Illinois.Selectors.ParkPage.ActivitiesSelector,
	}
	activityScraper := scrapers.NewILParkActivityScraper(activityScraperConfig)

	// Create Illinois park scraper with full config
	ilConfig := scrapers.ScraperConfig{
		BaseURL:   config.Scrapers.Illinois.BaseURL,
		StateCode: config.Scrapers.Illinois.StateCode,
		Selectors: scrapers.SelectorConfig{
			Homepage: scrapers.HomepageSelectors{
				APIURLAttribute: config.Scrapers.Illinois.Selectors.Homepage.APIURLAttribute,
			},
			JSONAPI: scrapers.JSONAPISelectors{
				ParksListPath: config.Scrapers.Illinois.Selectors.JSONAPI.ParksListPath,
				ParkNamePath:  config.Scrapers.Illinois.Selectors.JSONAPI.ParkNamePath,
				ParkURLPath:   config.Scrapers.Illinois.Selectors.JSONAPI.ParkURLPath,
			},
			ParkPage: scrapers.ParkPageSelectors{
				NameSelector:      config.Scrapers.Illinois.Selectors.ParkPage.NameSelector,
				LatitudeSelector:  config.Scrapers.Illinois.Selectors.ParkPage.LatitudeSelector,
				LongitudeSelector: config.Scrapers.Illinois.Selectors.ParkPage.LongitudeSelector,
			},
		},
		ActivityScraper: activityScraper,
	}

	scraper := scrapers.NewILParkScraper(ilConfig)

	// Scrape all parks
	parks, err := scraper.ScrapeAll()
	if err != nil {
		log.Fatalf("Failed to scrape parks: %v", err)
	}

	// Save parks to JSON file
	err = saveParksToJSON(parks, "parks.json")
	if err != nil {
		log.Fatalf("Failed to save parks to JSON: %v", err)
	}

	fmt.Println("[Final] Parks data saved to parks.json")
}

// saveParksToJSON saves a slice of parks to a JSON file
func saveParksToJSON(parks []scrapers.Park, filename string) error {
	jsonData, err := json.MarshalIndent(parks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal parks to JSON: %w", err)
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}