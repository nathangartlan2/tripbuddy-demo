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
		ActivitiesSelector: config.Scrapers.Illinois.Pages.ParkPage.Selectors.ActivitiesSelector,
	}
	activityScraper := scrapers.NewILParkActivityScraper(activityScraperConfig)

	// Create park collector based on homepage strategy
	var parkCollector scrapers.ParkCollectorScraper
	switch config.Scrapers.Illinois.Pages.Homepage.Strategy {
	case "json_api":
		jsonSelectors := scrapers.JSONAPISelectors{
			ParksListPath: config.Scrapers.Illinois.Pages.Homepage.Selectors.JSONAPI.ParksListPath,
			ParkNamePath:  config.Scrapers.Illinois.Pages.Homepage.Selectors.JSONAPI.ParkNamePath,
			ParkURLPath:   config.Scrapers.Illinois.Pages.Homepage.Selectors.JSONAPI.ParkURLPath,
		}
		parkCollector = scrapers.NewJSONAPIParkCollector(
			config.Scrapers.Illinois.Pages.Homepage.Selectors.APIURLAttribute,
			jsonSelectors,
		)
	case "static_html":
		staticSelectors := scrapers.StaticHTMLSelectors{
			ParkLinksSelector: config.Scrapers.Illinois.Pages.Homepage.Selectors.StaticHTML.ParkLinksSelector,
			ParkNameAttribute: config.Scrapers.Illinois.Pages.Homepage.Selectors.StaticHTML.ParkNameAttribute,
		}
		parkCollector = scrapers.NewStaticHTMLParkCollector(staticSelectors)
	default:
		log.Fatalf("Unknown homepage strategy: %s", config.Scrapers.Illinois.Pages.Homepage.Strategy)
	}

	// Create Illinois park scraper with full config
	ilConfig := scrapers.ScraperConfig{
		BaseURL:       config.Scrapers.Illinois.BaseURL,
		StateCode:     config.Scrapers.Illinois.StateCode,
		ParkCollector: parkCollector,
		ParkPageSelectors: scrapers.ParkPageSelectors{
			NameSelector:      config.Scrapers.Illinois.Pages.ParkPage.Selectors.NameSelector,
			LatitudeSelector:  config.Scrapers.Illinois.Pages.ParkPage.Selectors.LatitudeSelector,
			LongitudeSelector: config.Scrapers.Illinois.Pages.ParkPage.Selectors.LongitudeSelector,
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