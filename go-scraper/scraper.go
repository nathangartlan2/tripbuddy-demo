package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"scraper/scrapers"
)

func main() {
	// Parse command line flags
	concurrent := flag.Bool("concurrent", false, "Enable concurrent/parallel scraping")
	states := flag.String("states", "", "Comma-separated list of state codes to scrape (e.g., IL,IN). Leave empty to scrape all.")
	flag.Parse()

	// Load configuration
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup logger with configured level
	logger := setupLogger(config.LogLevel)
	slog.SetDefault(logger)

	logger.Info("Starting scraper",
		"log_level", config.LogLevel,
		"request_delay_seconds", config.RequestDelay,
		"states_configured", len(config.Scrapers))

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

	// Create multi-state scraper with logger and request delay
	multiScraper := scrapers.NewMultiStateScraper(scraperConfigs, logger, config.RequestDelay)

	// Parse states to scrape from command line or default to all
	var statesToScrape []string
	if *states != "" {
		// Split comma-separated states
		statesToScrape = parseStates(*states)
	} else {
		// Default: scrape all configured states
		statesToScrape = []string{}
	}

	fmt.Println("\n=== Multi-State Park Scraper ===")
	fmt.Printf("States to scrape: %v\n", statesToScrape)
	fmt.Printf("Execution mode: ")
	if *concurrent {
		fmt.Println("Concurrent (parallel)")
	} else {
		fmt.Println("Sequential")
	}
	fmt.Println()

	logger.Info("Starting park scraping",
		"states", statesToScrape,
		"concurrent", *concurrent)

	// Scrape parks and collect performance metrics
	parks, metrics, err := multiScraper.ScrapeStates(statesToScrape, *concurrent)
	if err != nil {
		logger.Error("Failed to scrape parks", "error", err)
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

	// Display performance metrics
	fmt.Printf("\n=== Performance Metrics ===\n")
	fmt.Printf("Execution mode: %s\n", getExecutionMode(metrics.Concurrent))
	fmt.Printf("Total duration: %v\n", metrics.TotalDuration)
	fmt.Printf("States scraped: %d\n", metrics.StatesScraped)
	fmt.Printf("Parks collected: %d\n", metrics.ParksCollected)
	if metrics.StatesScraped > 0 {
		fmt.Printf("Average time per state: %v\n", metrics.AvgTimePerState)
	}
	if metrics.ParksCollected > 0 {
		avgTimePerPark := metrics.TotalDuration / time.Duration(metrics.ParksCollected)
		fmt.Printf("Average time per park: %v\n", avgTimePerPark)
	}

	// Show per-state timings
	if len(metrics.StateTimings) > 0 {
		fmt.Printf("\nPer-state timings:\n")
		for state, duration := range metrics.StateTimings {
			parkCount := len(parksByState[state])
			fmt.Printf("  %s: %v (%d parks)\n", state, duration, parkCount)
		}
	}

	logger.Info("Scraping completed",
		"total_parks", len(parks),
		"states_count", len(parksByState),
		"total_duration", metrics.TotalDuration,
		"concurrent", metrics.Concurrent)

	// Save to JSON
	err = saveParksToJSON(parks, "parks.json")
	if err != nil {
		logger.Error("Failed to save parks to JSON", "error", err)
		log.Fatalf("Failed to save parks to JSON: %v", err)
	}

	logger.Info("Saved results to file", "filename", "parks.json")
	fmt.Printf("\n✓ Saved results to parks.json\n")
}

// getExecutionMode returns a human-readable string for the execution mode
func getExecutionMode(concurrent bool) string {
	if concurrent {
		return "Concurrent (parallel)"
	}
	return "Sequential"
}

// parseStates splits a comma-separated string into a slice of state codes
func parseStates(statesStr string) []string {
	if statesStr == "" {
		return []string{}
	}

	// Split by comma and trim whitespace
	var states []string
	for _, state := range strings.Split(statesStr, ",") {
		trimmed := strings.TrimSpace(state)
		if trimmed != "" {
			states = append(states, trimmed)
		}
	}
	return states
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