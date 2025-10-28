package scrapers

import (
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

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
	configs      map[string]StateConfig
	logger       *slog.Logger
	requestDelay time.Duration
}

// NewMultiStateScraper creates a new multi-state scraper
func NewMultiStateScraper(configs map[string]StateConfig, logger *slog.Logger, requestDelaySeconds int) *MultiStateScraper {
	return &MultiStateScraper{
		configs:      configs,
		logger:       logger,
		requestDelay: time.Duration(requestDelaySeconds) * time.Second,
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
		m.logger.Info("Starting sequential scrape", "state", stateCode)

		parks, err := m.scrapeState(stateCode)
		if err != nil {
			return nil, fmt.Errorf("failed to scrape %s: %w", stateCode, err)
		}

		m.logger.Info("Completed scraping state", "state", stateCode, "parks_count", len(parks))
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

			m.logger.Info("Starting concurrent scrape", "state", code)

			parks, err := m.scrapeState(code)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				m.logger.Error("Failed to scrape state", "state", code, "error", err)
				errors = append(errors, fmt.Errorf("failed to scrape %s: %w", code, err))
			} else {
				m.logger.Info("Completed scraping state", "state", code, "parks_count", len(parks))
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
	m.logger.Debug("Collecting park URLs from homepage", "state", stateCode, "base_url", config.BaseURL)
	parkURLs, err := parkCollector.CollectParkURLs(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to collect park URLs: %w", err)
	}
	m.logger.Debug("Found park URLs", "state", stateCode, "url_count", len(parkURLs))

	// Step 3: Scrape each park page
	return m.scrapeParkPages(stateCode, config, parkURLs)
}

// scrapeParkPages scrapes details from a list of park page URLs
func (m *MultiStateScraper) scrapeParkPages(stateCode string, config StateConfig, parkURLs []string) ([]Park, error) {
	var parks []Park
	var mu sync.Mutex

	// Create collector for park pages with browser-like settings
	cParkPage := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	// Add rate limiting to avoid being blocked
	cParkPage.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       m.requestDelay,
	})

	// Set headers to look more like a real browser and log requests
	cParkPage.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-Site", "none")

		m.logger.Debug("Scraping park page", "state", stateCode, "url", r.URL.String())
	})

	// Extract park details from individual park pages
	cParkPage.OnHTML("body", func(e *colly.HTMLElement) {
		// Extract park information using configured selectors
		parkName := e.ChildText(config.ParkPage.Selectors.NameSelector)
		latitudeStr := e.ChildText(config.ParkPage.Selectors.LatitudeSelector)
		longitudeStr := e.ChildText(config.ParkPage.Selectors.LongitudeSelector)

		m.logger.Debug("Extracted park data",
			"state", stateCode,
			"park_name", parkName,
			"latitude", latitudeStr,
			"longitude", longitudeStr)

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

			m.logger.Debug("Added park",
				"state", stateCode,
				"park_name", p.Name,
				"latitude", p.Latitude,
				"longitude", p.Longitude)
		} else {
			m.logger.Warn("Skipped park due to missing/invalid data",
				"state", stateCode,
				"park_name", parkName,
				"has_lat", err1 == nil,
				"has_lon", err2 == nil)
		}
	})

	cParkPage.OnError(func(r *colly.Response, err error) {
		m.logger.Error("Error scraping page",
			"state", stateCode,
			"url", r.Request.URL.String(),
			"error", err)
	})

	// Visit all park URLs
	for _, url := range parkURLs {
		err := cParkPage.Visit(url)
		if err != nil {
			m.logger.Error("Error visiting URL",
				"state", stateCode,
				"url", url,
				"error", err)
		}
	}

	// Wait for collector to finish
	cParkPage.Wait()

	m.logger.Debug("Completed scraping park pages",
		"state", stateCode,
		"parks_scraped", len(parks))

	return parks, nil
}
