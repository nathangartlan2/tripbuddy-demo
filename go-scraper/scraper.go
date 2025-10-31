package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"scraper/configHelper"
	"scraper/events"
	"scraper/extractors"
	"scraper/models"
	"scraper/scrapers"
	"scraper/services"
	"scraper/writers"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Parse command line arguments
	statesFlag := flag.String("states", "", "Comma-separated list of state codes to scrape (e.g., 'IL,IN'). If empty, scrapes all states.")
	flag.Parse()

	// Load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load("config/.env")

	// Load URL configuration from urls.json
	urlConfig, err := configHelper.LoadURLConfig("config/urls.json")
	if err != nil {
		log.Fatalf("Failed to load URL config: %v", err)
	}

	// Parse state filter
	var statesToScrape []string
	if *statesFlag != "" {
		statesToScrape = strings.Split(*statesFlag, ",")
		// Trim whitespace from each state code
		for i := range statesToScrape {
			statesToScrape[i] = strings.TrimSpace(statesToScrape[i])
		}
		fmt.Printf("Filtering to scrape only states: %v\n", statesToScrape)
	} else {
		fmt.Println("No state filter provided, scraping all states")
	}

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
	jsonWriter := writers.NewParkJSONWriter("data")

	// Get API URL from environment variable, default to localhost
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
		log.Println("API_URL not set, using default: http://localhost:8080")
	} else {
		log.Printf("Using API_URL: %s", apiURL)
	}

	apiWriter := writers.NewAPIParkWriter(apiURL)
	publisher.Subscribe(jsonWriter)
	publisher.Subscribe((apiWriter))

	// Scrape parks for each state
	results := scrapeAllStates(urlConfig, extractorFactory, publisher, statesToScrape)

	// Wait for all events to be processed
	publisher.WaitForQueue()

	// Print summary
	fmt.Printf("\n=== Scraping Summary ===\n")
	for state, parks := range results {
		fmt.Printf("%s: %d parks scraped\n", state, len(parks))
	}
}

// scrapeAllStates takes the URL config and scrapes all parks for all states (or filtered states)
func scrapeAllStates(urlConfig *configHelper.URLConfig, factory *extractors.ExtractorFactory, publisher *events.ParkEventPublisher, stateFilter []string) map[string][]*models.Park {
	results := make(map[string][]*models.Park)

	// Create a map for quick lookup if filtering
	filterMap := make(map[string]bool)
	if len(stateFilter) > 0 {
		for _, state := range stateFilter {
			filterMap[state] = true
		}
	}

	for _, stateCode := range urlConfig.GetAllStates() {
		// Skip if not in filter (when filter is provided)
		if len(stateFilter) > 0 && !filterMap[stateCode] {
			fmt.Printf("Skipping %s (not in filter)\n", stateCode)
			continue
		}

		baseURL, ok := urlConfig.GetBaseURLByState(stateCode)
		if !ok || baseURL == "" {
			log.Printf("No base URL found for state: %s, skipping", stateCode)
			continue
		}

		homePageUrl, ok := urlConfig.GetHomePageURLByState(stateCode)

		if !ok || homePageUrl == "" {
			log.Printf("No homePage URL found for state: %s, skipping", stateCode)
			continue
		}
		fmt.Printf("\n=== Scraping %s ===\n", stateCode)
		parks := scrapeParksByState(stateCode, baseURL, homePageUrl, factory, publisher)
		results[stateCode] = parks
	}

	return results
}

// scrapeParksByState scrapes all parks for a given state
func scrapeParksByState(stateCode string, baseUrl string, homePageUrl string, factory *extractors.ExtractorFactory, publisher *events.ParkEventPublisher) []*models.Park {
	parks := make([]*models.Park, 0)

	// Get appropriate extractor for state using factory
	extractor := factory.CreateExtractor(stateCode)
	if extractor == nil {
		log.Printf("No extractor found for state: %s", stateCode)
		return parks
	}

	// Create appropriate URL gatherer based on state
	var gatherer scrapers.ParkUrlGatherer
	switch stateCode {
	case "IL":
		gatherer = scrapers.NewJSONParkUrlGatherer(baseUrl)
	case "IN":
		gatherer = scrapers.NewINHTMLParkUrlGatherer(baseUrl)
	default:
		log.Printf("No URL gatherer configured for state: %s", stateCode)
		return parks
	}

	// Create callback function for when a park is scraped
	onParkScraped := func(park *models.Park, duration time.Duration, timestamp time.Time) {
		// Validate park data
		if park == nil {
			log.Printf("Error: received nil park in callback")
			return
		}

		// Print park info with error handling for potentially invalid data
		fmt.Printf("  ✓ %s (%.3f, %.3f) - %d activities - %v\n",
			park.Name, park.Latitude, park.Longitude, len(park.Activities), duration)

		// Publish event for scraped park
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Error publishing park event for %s: %v", park.Name, r)
			}
		}()

		publisher.Publish(events.ParkScrapedEvent{
			Park:      park,
			StateCode: stateCode,
			URL:       "", // URL not available in callback context
			Duration:  duration,
			Timestamp: timestamp,
		})
	}

	// Create scraper
	scraper := scrapers.NewBaseParkScraper(5, extractor, gatherer, onParkScraped)

	scraper.ScrapeAllParks(homePageUrl)

	return parks
}