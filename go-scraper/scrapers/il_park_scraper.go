package scrapers

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/gocolly/colly"
)

// ScraperConfig holds the configuration for a state park scraper
type ScraperConfig struct {
	BaseURL          string
	StateCode        string
	ParkCollector    ParkCollectorScraper
	ParkPageSelectors ParkPageSelectors
	ActivityScraper   ParkActivityScraper
}

// ParkPageSelectors for extracting park details
type ParkPageSelectors struct {
	NameSelector      string
	LatitudeSelector  string
	LongitudeSelector string
}

// HomepageSelectors holds selectors for discovering parks on homepage
type HomepageSelectors struct {
	// For JSON API strategy
	APIURLAttribute string
	JSONAPI         JSONAPISelectors

	// For Static HTML strategy
	StaticHTML StaticHTMLSelectors
}

// JSONAPISelectors for parsing JSON responses
type JSONAPISelectors struct {
	ParksListPath string
	ParkNamePath  string
	ParkURLPath   string
}

// StaticHTMLSelectors for parsing static HTML park lists
type StaticHTMLSelectors struct {
	ParkLinksSelector string
	ParkNameAttribute string
}

// ILParkScraper scrapes parks from the Illinois DNR website
type ILParkScraper struct {
	config ScraperConfig
}

// NewILParkScraper creates a new Illinois park scraper
func NewILParkScraper(config ScraperConfig) *ILParkScraper {
	return &ILParkScraper{
		config: config,
	}
}

// ScrapeAll implements the ParkScraper interface
func (s *ILParkScraper) ScrapeAll() ([]Park, error) {
	// Step 1: Collect park URLs from homepage using the configured collector
	parkURLs, err := s.config.ParkCollector.CollectParkURLs(s.config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to collect park URLs: %w", err)
	}

	// Step 2: Scrape each park page
	return s.scrapeParkPages(parkURLs)
}

// scrapeParkPages scrapes details from a list of park page URLs
func (s *ILParkScraper) scrapeParkPages(parkURLs []string) ([]Park, error) {
	// Slice to store all parks
	var parks []Park
	var mu sync.Mutex // Mutex to safely append to slice from concurrent requests

	// Create collector for park pages
	cParkPage := colly.NewCollector()

	// ===== Park Page HTML Scraping =====
	cParkPage.OnRequest(func(r *colly.Request) {
		fmt.Println("[ParkPage] Scraping park page:", r.URL)
	})

	// Extract park details from individual park pages
	cParkPage.OnHTML("body", func(e *colly.HTMLElement) {
		fmt.Println("[Level 2] Extracting park details from:", e.Request.URL)

		// Extract park information using configured selectors
		parkName := e.ChildText(s.config.ParkPageSelectors.NameSelector)
		latitudeStr := e.ChildText(s.config.ParkPageSelectors.LatitudeSelector)
		longitudeStr := e.ChildText(s.config.ParkPageSelectors.LongitudeSelector)

		fmt.Printf("[Level 2] Park Name: %s\n", parkName)
		fmt.Printf("[Level 2] Latitude: %s\n", latitudeStr)
		fmt.Printf("[Level 2] Longitude: %s\n", longitudeStr)

		// Convert lat/long strings to float32
		latitude, err1 := strconv.ParseFloat(latitudeStr, 32)
		longitude, err2 := strconv.ParseFloat(longitudeStr, 32)

		// Use the ActivityScraper interface to scrape activities
		activities, err := s.config.ActivityScraper.ScrapeActivities(e.Request.URL.String())
		if err != nil {
			fmt.Printf("[Level 2] Error scraping activities: %v\n", err)
			activities = []ParkActivity{} // Use empty slice on error
		}

		// Only add park if we have valid data
		if parkName != "" && err1 == nil && err2 == nil {
			p := Park{
				Name:       parkName,
				StateCode:  s.config.StateCode, // Use configured state code
				Latitude:   float32(latitude),
				Longitude:  float32(longitude),
				Activities: activities,
			}

			// Thread-safe append to parks slice
			mu.Lock()
			parks = append(parks, p)
			mu.Unlock()

			fmt.Printf("[Level 2] Added park: %s (%.3f, %.3f) with %d activities\n",
				p.Name, p.Latitude, p.Longitude, len(p.Activities))
		}
	})

	// Visit all park URLs
	for _, url := range parkURLs {
		err := cParkPage.Visit(url)
		if err != nil {
			fmt.Printf("[ParkPage] Error visiting %s: %v\n", url, err)
		}
	}

	// Wait for collector to finish
	cParkPage.Wait()

	fmt.Printf("\n[Final] Scraped %d parks total\n", len(parks))

	return parks, nil
}
