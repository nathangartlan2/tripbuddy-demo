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

	// Create park collector based on homepage strategy
	var parkCollector scrapers.ParkCollectorScraper
	switch config.Scrapers.Indiana.Pages.Homepage.Strategy {
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
		// Convert config types to scraper types
		staticSelectors := scrapers.StaticHTMLSelectors{
			Section: scrapers.HTMLSection{
				ID:       config.Scrapers.Illinois.Pages.Homepage.Selectors.StaticHTML.Section.ID,
				Class:    config.Scrapers.Illinois.Pages.Homepage.Selectors.StaticHTML.Section.Class,
				Selector: config.Scrapers.Illinois.Pages.Homepage.Selectors.StaticHTML.Section.Selector,
			},
			URLElement: scrapers.URLElement{
				HrefPattern:       config.Scrapers.Illinois.Pages.Homepage.Selectors.StaticHTML.URLElement.HrefPattern,
				ParkNameAttribute: config.Scrapers.Illinois.Pages.Homepage.Selectors.StaticHTML.URLElement.ParkNameAttribute,
			},
		}
		parkCollector = scrapers.NewStaticHTMLParkCollector(staticSelectors)
	default:
		log.Fatalf("Unknown homepage strategy: %s", config.Scrapers.Illinois.Pages.Homepage.Strategy)
	}

	// TEST: Just collect park URLs from homepage
	fmt.Println("\n=== Testing Homepage Collector ===")
	fmt.Printf("Strategy: %s\n", config.Scrapers.Indiana.Pages.Homepage.Strategy)
	fmt.Printf("Base URL: %s\n\n", config.Scrapers.Indiana.BaseURL)

	parkURLs, err := parkCollector.CollectParkURLs(config.Scrapers.Indiana.BaseURL)
	if err != nil {
		log.Fatalf("Failed to collect park URLs: %v", err)
	}

	fmt.Printf("\n=== Results ===\n")
	fmt.Printf("Collected %d park URLs:\n\n", len(parkURLs))
	for i, url := range parkURLs {
		fmt.Printf("%3d. %s\n", i+1, url)
	}
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