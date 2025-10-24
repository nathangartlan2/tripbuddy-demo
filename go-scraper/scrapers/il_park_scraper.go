package scrapers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

// ScraperConfig holds the configuration for a state park scraper
type ScraperConfig struct {
	BaseURL         string
	StateCode       string
	Selectors       SelectorConfig
	ActivityScraper ParkActivityScraper
}

// SelectorConfig holds all the selectors needed for scraping
type SelectorConfig struct {
	Homepage  HomepageSelectors
	JSONAPI   JSONAPISelectors
	ParkPage  ParkPageSelectors
}

// HomepageSelectors for finding the JSON API
type HomepageSelectors struct {
	APIURLAttribute string
}

// JSONAPISelectors for parsing JSON responses
type JSONAPISelectors struct {
	ParksListPath string
	ParkNamePath  string
	ParkURLPath   string
}

// ParkPageSelectors for extracting park details
type ParkPageSelectors struct {
	NameSelector      string
	LatitudeSelector  string
	LongitudeSelector string
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

// getNestedValue retrieves a value from a nested map using a dot-separated path
func getNestedValue(data map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current, ok = m[part]
			if !ok {
				return nil, false
			}
		} else {
			return nil, false
		}
	}
	return current, true
}

// ScrapeAll implements the ParkScraper interface
func (s *ILParkScraper) ScrapeAll() ([]Park, error) {
	// Slice to store all parks
	var parks []Park
	var mu sync.Mutex // Mutex to safely append to slice from concurrent requests

	// Create collectors
	cHomePage := colly.NewCollector()
	cJsonAPI := colly.NewCollector()
	cParkPage := colly.NewCollector()

	// ===== Level 0: Finding Park API =======
	cHomePage.OnRequest(func(r *colly.Request) {
		fmt.Println("[Level 0] Visiting:", r.URL)
	})

	cHomePage.OnHTML(s.config.Selectors.Homepage.APIURLAttribute, func(e *colly.HTMLElement) {
		jsonURL := e.Request.AbsoluteURL(e.Attr("data-api-url"))
		fmt.Println("[Level 0] Found JSON API:", jsonURL)
		cJsonAPI.Visit(jsonURL)
	})

	// ===== LEVEL 1: JSON Extraction =====
	cJsonAPI.OnRequest(func(r *colly.Request) {
		fmt.Println("[Level 1] Visiting:", r.URL)
	})

	// Parse JSON response and extract park URLs
	cJsonAPI.OnResponse(func(r *colly.Response) {
		contentType := r.Headers.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			fmt.Println("[Level 1] Processing JSON response")

			// Parse the JSON to extract park URLs
			var jsonData map[string]interface{}
			err := json.Unmarshal(r.Body, &jsonData)
			if err != nil {
				fmt.Printf("Error parsing JSON: %v\n", err)
				return
			}

			// Extract park URLs from JSON using configured path
			parksListPath := s.config.Selectors.JSONAPI.ParksListPath
			if itemsData, ok := getNestedValue(jsonData, parksListPath); ok {
				if items, ok := itemsData.([]interface{}); ok {
					fmt.Printf("[Level 1] Found %d parks\n", len(items))

					for _, item := range items {
						if parkItem, ok := item.(map[string]interface{}); ok {
							// Extract the park name using configured path
							parkNamePath := s.config.Selectors.JSONAPI.ParkNamePath
							if nameData, ok := getNestedValue(parkItem, parkNamePath); ok {
								if name, ok := nameData.(string); ok {
									fmt.Printf("[Level 1] Found park: %s\n", name)
								}
							}

							// Extract the park page URL using configured path
							parkURLPath := s.config.Selectors.JSONAPI.ParkURLPath
							if urlData, ok := getNestedValue(parkItem, parkURLPath); ok {
								if url, ok := urlData.(string); ok {
									absoluteURL := r.Request.AbsoluteURL(url)
									fmt.Printf("[Level 1] Queueing park page: %s\n", absoluteURL)
									cParkPage.Visit(absoluteURL)
								}
							}
						}
					}
				}
			} else {
				// Fallback: print the JSON structure to understand it better
				fmt.Println("[Level 1] JSON structure (first 500 chars):")
				jsonStr := string(r.Body)
				if len(jsonStr) > 500 {
					fmt.Println(jsonStr[:500])
				} else {
					fmt.Println(jsonStr)
				}
			}
		}
	})

	// ===== LEVEL 2: HTML Scraping =====
	cParkPage.OnRequest(func(r *colly.Request) {
		fmt.Println("[Level 2] Scraping park page:", r.URL)
	})

	// Extract park details from individual park pages
	cParkPage.OnHTML("body", func(e *colly.HTMLElement) {
		fmt.Println("[Level 2] Extracting park details from:", e.Request.URL)

		// Extract park information using configured selectors
		parkName := e.ChildText(s.config.Selectors.ParkPage.NameSelector)
		latitudeStr := e.ChildText(s.config.Selectors.ParkPage.LatitudeSelector)
		longitudeStr := e.ChildText(s.config.Selectors.ParkPage.LongitudeSelector)

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

	// Start the scraping process
	err := cHomePage.Visit(s.config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to visit homepage: %w", err)
	}

	// Wait for all collectors to finish
	cHomePage.Wait()
	cJsonAPI.Wait()
	cParkPage.Wait()

	fmt.Printf("\n[Final] Scraped %d parks total\n", len(parks))

	return parks, nil
}
