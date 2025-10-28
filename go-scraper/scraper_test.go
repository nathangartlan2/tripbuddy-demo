package main

import (
	"log/slog"
	"os"
	"testing"

	"scraper/scrapers"
)

// loadTestConfig loads the config for benchmarks
func loadTestConfig(b *testing.B) *Config {
	config, err := loadConfig("config.json")
	if err != nil {
		b.Fatalf("Failed to load config: %v", err)
	}
	return config
}

// createTestScraper creates a MultiStateScraper for testing
func createTestScraper(b *testing.B, config *Config) *scrapers.MultiStateScraper {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Quiet logging during benchmarks
	}))

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

	return scrapers.NewMultiStateScraper(scraperConfigs, logger, config.RequestDelay)
}

// BenchmarkSequentialScraping benchmarks sequential scraping
// Note: This will make real network requests - use with caution
// Run with: go test -bench=BenchmarkSequential -benchtime=1x
func BenchmarkSequentialScraping(b *testing.B) {
	config := loadTestConfig(b)
	multiScraper := createTestScraper(b, config)

	// Get list of states to scrape (use a subset for testing)
	var statesToScrape = []string{"IL"}


	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := multiScraper.ScrapeStates(statesToScrape, false)
		if err != nil {
			b.Fatalf("Sequential scraping failed: %v", err)
		}
	}
}

// BenchmarkParallelScraping benchmarks parallel scraping
// Note: This will make real network requests - use with caution
// Run with: go test -bench=BenchmarkParallel -benchtime=1x
func BenchmarkParallelScraping(b *testing.B) {
	config := loadTestConfig(b)
	multiScraper := createTestScraper(b, config)

	// Get list of states to scrape (use a subset for testing)
	var statesToScrape = []string{"IL"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := multiScraper.ScrapeStates(statesToScrape, true)
		if err != nil {
			b.Fatalf("Parallel scraping failed: %v", err)
		}
	}
}

// BenchmarkSequentialAllStates benchmarks sequential scraping of all states
// Run with: go test -bench=BenchmarkSequentialAllStates -benchtime=1x
func BenchmarkSequentialAllStates(b *testing.B) {
	config := loadTestConfig(b)
	multiScraper := createTestScraper(b, config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := multiScraper.ScrapeStates([]string{}, false)
		if err != nil {
			b.Fatalf("Sequential scraping failed: %v", err)
		}
	}
}

// BenchmarkParallelAllStates benchmarks parallel scraping of all states
// Run with: go test -bench=BenchmarkParallelAllStates -benchtime=1x
func BenchmarkParallelAllStates(b *testing.B) {
	config := loadTestConfig(b)
	multiScraper := createTestScraper(b, config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := multiScraper.ScrapeStates([]string{}, true)
		if err != nil {
			b.Fatalf("Parallel scraping failed: %v", err)
		}
	}
}
