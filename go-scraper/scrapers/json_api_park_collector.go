package scrapers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

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

// JSONAPIParkCollector scrapes park URLs from a JSON API
type JSONAPIParkCollector struct {
	apiURLSelector string
	jsonSelectors  JSONAPISelectors
}

// NewJSONAPIParkCollector creates a new JSON API park collector
func NewJSONAPIParkCollector(apiURLSelector string, jsonSelectors JSONAPISelectors) *JSONAPIParkCollector {
	return &JSONAPIParkCollector{
		apiURLSelector: apiURLSelector,
		jsonSelectors:  jsonSelectors,
	}
}

// CollectParkURLs implements the ParkCollectorScraper interface
func (c *JSONAPIParkCollector) CollectParkURLs(homepageURL string) ([]string, error) {
	var parkURLs []string

	// Create collectors with browser-like settings
	cHomePage := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)
	cJsonAPI := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	// Add rate limiting
	cHomePage.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       1 * time.Second,
	})
	cJsonAPI.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       1 * time.Second,
	})

	// ===== Level 0: Finding Park API =======
	cHomePage.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Connection", "keep-alive")
		fmt.Println("[Collector] Visiting homepage:", r.URL)
	})

	cHomePage.OnHTML(c.apiURLSelector, func(e *colly.HTMLElement) {
		jsonURL := e.Request.AbsoluteURL(e.Attr("data-api-url"))
		fmt.Println("[Collector] Found JSON API:", jsonURL)
		cJsonAPI.Visit(jsonURL)
	})

	// ===== LEVEL 1: JSON Extraction =====
	cJsonAPI.OnRequest(func(r *colly.Request) {
		fmt.Println("[Collector] Visiting JSON API:", r.URL)
	})

	// Parse JSON response and extract park URLs
	cJsonAPI.OnResponse(func(r *colly.Response) {
		contentType := r.Headers.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			fmt.Println("[Collector] Processing JSON response")

			// Parse the JSON to extract park URLs
			var jsonData map[string]interface{}
			err := json.Unmarshal(r.Body, &jsonData)
			if err != nil {
				fmt.Printf("[Collector] Error parsing JSON: %v\n", err)
				return
			}

			// Extract park URLs from JSON using configured path
			if itemsData, ok := getNestedValue(jsonData, c.jsonSelectors.ParksListPath); ok {
				if items, ok := itemsData.([]interface{}); ok {
					fmt.Printf("[Collector] Found %d parks\n", len(items))

					for _, item := range items {
						if parkItem, ok := item.(map[string]interface{}); ok {
							// Extract the park name using configured path
							if nameData, ok := getNestedValue(parkItem, c.jsonSelectors.ParkNamePath); ok {
								if name, ok := nameData.(string); ok {
									fmt.Printf("[Collector] Found park: %s\n", name)
								}
							}

							// Extract the park page URL using configured path
							if urlData, ok := getNestedValue(parkItem, c.jsonSelectors.ParkURLPath); ok {
								if url, ok := urlData.(string); ok {
									absoluteURL := r.Request.AbsoluteURL(url)
									fmt.Printf("[Collector] Collected park URL: %s\n", absoluteURL)
									parkURLs = append(parkURLs, absoluteURL)
								}
							}
						}
					}
				}
			} else {
				// Fallback: print the JSON structure to understand it better
				fmt.Println("[Collector] JSON structure (first 500 chars):")
				jsonStr := string(r.Body)
				if len(jsonStr) > 500 {
					fmt.Println(jsonStr[:500])
				} else {
					fmt.Println(jsonStr)
				}
			}
		}
	})

	// Start the scraping process
	err := cHomePage.Visit(homepageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to visit homepage: %w", err)
	}

	// Wait for all collectors to finish
	cHomePage.Wait()
	cJsonAPI.Wait()

	fmt.Printf("[Collector] Collected %d park URLs total\n", len(parkURLs))

	return parkURLs, nil
}
