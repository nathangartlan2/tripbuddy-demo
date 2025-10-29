package main

import (
	"fmt"
	"log"
	"os"
	"scraper/config"
	"scraper/events"
	"scraper/extractors"
	"scraper/models"
	"scraper/scrapers"
	"scraper/services"
	"scraper/writers"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	// Load URL configuration from urls.json
	urlConfig, err := config.LoadURLConfig("urls.json")
	if err != nil {
		log.Fatalf("Failed to load URL config: %v", err)
	}

	// Get state-to-URLs dictionary
	stateURLs := urlConfig.GetAllURLs()

	// Initialize geocoding service
	mapboxAPIKey := os.Getenv("MAPBOX_API_KEY")
	if mapboxAPIKey == "" {
		log.Println("Warning: MAPBOX_API_KEY environment variable not set. Geocoding will not work.")
	}
	geocodingService := services.NewGeocodingService(mapboxAPIKey)

	// Create extractor factory
	extractorFactory := extractors.NewExtractorFactory(geocodingService)

	// Create event publisher
	publisher := events.NewParkEventPublisher()
	defer publisher.Close()

	// Create and subscribe JSON writer
	jsonWriter := writers.NewParkJSONWriter("output")
	apiWriter := writers.NewAPIParkWriter("http://localhost:8080")
	publisher.Subscribe(jsonWriter)
	publisher.Subscribe((apiWriter))

	// Scrape parks for each state
	results := scrapeAllStates(stateURLs, extractorFactory, publisher)

	// Wait for all events to be processed
	publisher.WaitForQueue()

	// Print summary
	fmt.Printf("\n=== Scraping Summary ===\n")
	for state, parks := range results {
		fmt.Printf("%s: %d parks scraped\n", state, len(parks))
	}
}

// scrapeAllStates takes a map of stateCode -> []urls and scrapes all parks
func scrapeAllStates(stateURLs map[string][]string, factory *extractors.ExtractorFactory, publisher *events.ParkEventPublisher) map[string][]*models.Park {
	results := make(map[string][]*models.Park)

	for stateCode, urls := range stateURLs {
		fmt.Printf("\n=== Scraping %s (%d parks) ===\n", stateCode, len(urls))
		parks := scrapeParksByState(stateCode, urls, factory, publisher)
		results[stateCode] = parks
	}

	return results
}

// scrapeParksByState scrapes all parks for a given state
func scrapeParksByState(stateCode string, urls []string, factory *extractors.ExtractorFactory, publisher *events.ParkEventPublisher) []*models.Park {
	parks := make([]*models.Park, 0, len(urls))

	// Get appropriate extractor for state using factory
	extractor := factory.CreateExtractor(stateCode)
	if extractor == nil {
		log.Printf("No extractor found for state: %s", stateCode)
		return parks
	}

	// Create scraper
	scraper := scrapers.NewBaseParkScraper(5, extractor)

	// Scrape each URL
	for i, url := range urls {
		fmt.Printf("[%d/%d] Scraping: %s\n", i+1, len(urls), url)

		park, duration, err := scraper.ScrapePark(url)
		if err != nil {
			log.Printf("Error scraping %s: %v", url, err)
			continue
		}

		fmt.Printf("  ✓ %s (%.3f, %.3f) - %d activities - %v\n",
			park.Name, park.Latitude, park.Longitude, len(park.Activities), duration)

		// Publish event for scraped park
		publisher.Publish(events.ParkScrapedEvent{
			Park:      park,
			StateCode: stateCode,
			URL:       url,
			Duration:  duration,
			Timestamp: time.Now(),
		})

		parks = append(parks, park)
	}

	return parks
}