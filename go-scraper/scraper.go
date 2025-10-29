package main

import (
	"fmt"
	"log"
	"scraper/config"
	"scraper/events"
	"scraper/extractors"
	"scraper/models"
	"scraper/scrapers"
	"scraper/writers"
	"time"
)

func main() {
	// Load URL configuration from urls.json
	urlConfig, err := config.LoadURLConfig("urls.json")
	if err != nil {
		log.Fatalf("Failed to load URL config: %v", err)
	}

	// Get state-to-URLs dictionary
	stateURLs := urlConfig.GetAllURLs()

	// Create extractor factory
	extractorFactory := extractors.NewExtractorFactory()

	// Create event publisher
	publisher := events.NewParkEventPublisher()
	defer publisher.Close()

	// Create and subscribe JSON writer
	jsonWriter := writers.NewParkJSONWriter("output")
	publisher.Subscribe(jsonWriter)

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