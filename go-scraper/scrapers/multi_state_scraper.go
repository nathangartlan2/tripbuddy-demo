package scrapers

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/gocolly/colly"
)

// StateConfig holds the configuration for a single state scraper
type StateConfig struct {
	BaseURL   string
	StateCode string
	Homepage  HomepageConfig
	ParkPage  ParkPageConfig
}

// HomepageConfig holds homepage scraping configuration
type HomepageConfig struct {
	Strategy  string
	Selectors HomepageSelectors
}

// ParkPageConfig holds park page scraping configuration
type ParkPageConfig struct {
	Strategy  string
	Selectors ParkPageSelectors
}

// MultiStateScraper orchestrates scraping across multiple states
type MultiStateScraper struct {
	configs map[string]StateConfig
}

// NewMultiStateScraper creates a new multi-state scraper
func NewMultiStateScraper(configs map[string]StateConfig) *MultiStateScraper {
	return &MultiStateScraper{
		configs: configs,
	}
}

// ScrapeStates scrapes parks from the specified states
// If stateCodes is empty, scrapes all configured states
// If concurrent is true, scrapes states in parallel
func (m *MultiStateScraper) ScrapeStates(stateCodes []string, concurrent bool) ([]Park, error) {
	// If no state codes specified, scrape all
	if len(stateCodes) == 0 {
		stateCodes = make([]string, 0, len(m.configs))
		for code := range m.configs {
			stateCodes = append(stateCodes, code)
		}
	}

	// Validate that all requested states exist in config
	for _, code := range stateCodes {
		if _, exists := m.configs[code]; !exists {
			return nil, fmt.Errorf("no configuration found for state: %s", code)
		}
	}

	if concurrent {
		return m.scrapeConcurrent(stateCodes)
	}
	return m.scrapeSequential(stateCodes)
}

// scrapeSequential scrapes states one at a time
func (m *MultiStateScraper) scrapeSequential(stateCodes []string) ([]Park, error) {
	var allParks []Park

	for _, stateCode := range stateCodes {
		fmt.Printf("\n=== Scraping %s ===\n", stateCode)

		parks, err := m.scrapeState(stateCode)
		if err != nil {
			return nil, fmt.Errorf("failed to scrape %s: %w", stateCode, err)
		}

		fmt.Printf("✓ Scraped %d parks from %s\n", len(parks), stateCode)
		allParks = append(allParks, parks...)
	}

	return allParks, nil
}

// scrapeConcurrent scrapes states in parallel using goroutines
func (m *MultiStateScraper) scrapeConcurrent(stateCodes []string) ([]Park, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	allParks := []Park{}
	errors := []error{}

	for _, stateCode := range stateCodes {
		wg.Add(1)

		go func(code string) {
			defer wg.Done()

			fmt.Printf("\n=== Scraping %s (concurrent) ===\n", code)

			parks, err := m.scrapeState(code)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors = append(errors, fmt.Errorf("failed to scrape %s: %w", code, err))
			} else {
				fmt.Printf("✓ Scraped %d parks from %s\n", len(parks), code)
				allParks = append(allParks, parks...)
			}
		}(stateCode)
	}

	wg.Wait()

	// Return first error if any occurred
	if len(errors) > 0 {
		return allParks, errors[0]
	}

	return allParks, nil
}

// scrapeState scrapes all parks for a single state
func (m *MultiStateScraper) scrapeState(stateCode string) ([]Park, error) {
	config := m.configs[stateCode]

	// Step 1: Create park collector using factory
	parkCollector, err := NewParkCollector(config.Homepage.Strategy, config.Homepage.Selectors)
	if err != nil {
		return nil, fmt.Errorf("failed to create park collector: %w", err)
	}

	// Step 2: Collect park URLs from homepage
	fmt.Printf("[%s] Collecting park URLs from homepage...\n", stateCode)
	parkURLs, err := parkCollector.CollectParkURLs(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to collect park URLs: %w", err)
	}
	fmt.Printf("[%s] Found %d park URLs\n", stateCode, len(parkURLs))

	// Step 3: Scrape each park page
	return m.scrapeParkPages(stateCode, config, parkURLs)
}

// scrapeParkPages scrapes details from a list of park page URLs
func (m *MultiStateScraper) scrapeParkPages(stateCode string, config StateConfig, parkURLs []string) ([]Park, error) {
	var parks []Park
	var mu sync.Mutex

	// Create collector for park pages
	cParkPage := colly.NewCollector()

	cParkPage.OnRequest(func(r *colly.Request) {
		fmt.Printf("[%s] Scraping park page: %s\n", stateCode, r.URL)
	})

	// Extract park details from individual park pages
	cParkPage.OnHTML("body", func(e *colly.HTMLElement) {
		// Extract park information using configured selectors
		parkName := e.ChildText(config.ParkPage.Selectors.NameSelector)
		latitudeStr := e.ChildText(config.ParkPage.Selectors.LatitudeSelector)
		longitudeStr := e.ChildText(config.ParkPage.Selectors.LongitudeSelector)

		fmt.Printf("[%s] Park: %s (Lat: %s, Lon: %s)\n", stateCode, parkName, latitudeStr, longitudeStr)

		// Convert lat/long strings to float32
		latitude, err1 := strconv.ParseFloat(latitudeStr, 32)
		longitude, err2 := strconv.ParseFloat(longitudeStr, 32)

		// TODO: Add activity scraping when needed
		// For now, we'll leave activities empty
		var activities []ParkActivity

		// Only add park if we have valid data
		if parkName != "" && err1 == nil && err2 == nil {
			p := Park{
				Name:       parkName,
				StateCode:  stateCode,
				Latitude:   float32(latitude),
				Longitude:  float32(longitude),
				Activities: activities,
			}

			// Thread-safe append to parks slice
			mu.Lock()
			parks = append(parks, p)
			mu.Unlock()

			fmt.Printf("[%s] ✓ Added park: %s (%.3f, %.3f)\n",
				stateCode, p.Name, p.Latitude, p.Longitude)
		} else {
			fmt.Printf("[%s] ⚠ Skipped park due to missing/invalid data\n", stateCode)
		}
	})

	cParkPage.OnError(func(r *colly.Response, err error) {
		fmt.Printf("[%s] ✗ Error scraping %s: %v\n", stateCode, r.Request.URL, err)
	})

	// Visit all park URLs
	for _, url := range parkURLs {
		err := cParkPage.Visit(url)
		if err != nil {
			fmt.Printf("[%s] Error visiting %s: %v\n", stateCode, url, err)
		}
	}

	// Wait for collector to finish
	cParkPage.Wait()

	fmt.Printf("[%s] Completed: %d parks scraped\n", stateCode, len(parks))

	return parks, nil
}
