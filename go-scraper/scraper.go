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

	// Convert main config format to scrapers.StateConfig format
	scraperConfigs := make(map[string]scrapers.StateConfig)
	for stateCode, stateConfig := range config.Scrapers {
		scraperConfigs[stateCode] = scrapers.StateConfig{
			BaseURL:   stateConfig.BaseURL,
			StateCode: stateConfig.StateCode,
			Homepage: scrapers.HomepageConfig{
				Strategy:  stateConfig.Pages.Homepage.Strategy,
				Selectors: stateConfig.Pages.Homepage.Selectors,
			},
			ParkPage: scrapers.ParkPageConfig{
				Strategy:  stateConfig.Pages.ParkPage.Strategy,
				Selectors: stateConfig.Pages.ParkPage.Selectors,
			},
		}
	}

	// Create multi-state scraper
	multiScraper := scrapers.NewMultiStateScraper(scraperConfigs)

	// Configure which states to scrape
	// Options:
	//   1. Specify states: statesToScrape := []string{"IL", "IN"}
	//   2. Scrape all: statesToScrape := []string{} or nil
	statesToScrape := []string{"IL", "IN"} // Change this as needed

	// Choose execution mode
	concurrent := false // Set to true for parallel scraping

	fmt.Println("\n=== Multi-State Park Scraper ===")
	fmt.Printf("States to scrape: %v\n", statesToScrape)
	fmt.Printf("Execution mode: ")
	if concurrent {
		fmt.Println("Concurrent (parallel)")
	} else {
		fmt.Println("Sequential")
	}
	fmt.Println()

	// Scrape parks
	parks, err := multiScraper.ScrapeStates(statesToScrape, concurrent)
	if err != nil {
		log.Fatalf("Failed to scrape parks: %v", err)
	}

	// Display results
	fmt.Printf("\n=== Results ===\n")
	fmt.Printf("Total parks scraped: %d\n\n", len(parks))

	// Group by state for display
	parksByState := make(map[string][]scrapers.Park)
	for _, park := range parks {
		parksByState[park.StateCode] = append(parksByState[park.StateCode], park)
	}

	for state, statePark := range parksByState {
		fmt.Printf("%s: %d parks\n", state, len(statePark))
	}

	// Save to JSON
	err = saveParksToJSON(parks, "parks.json")
	if err != nil {
		log.Fatalf("Failed to save parks to JSON: %v", err)
	}
	fmt.Printf("\n✓ Saved results to parks.json\n")
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