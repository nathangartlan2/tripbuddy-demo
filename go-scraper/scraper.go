package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)


type park struct {
	Name string  `json:"name"`
	StateCode string `json:"stateCode"`
	Latitude float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

func main(){
	// Slice to store all parks
	var parks []park
	var mu sync.Mutex // Mutex to safely append to slice from concurrent requests

	cHomePage := colly.NewCollector()

	// Level 1 Collector: Get park list from JSON
	cJsonAPI := colly.NewCollector()

	// Level 2 Collector: Scrape individual park pages
	cParkPage := colly.NewCollector()

	// ===== Level 2: Finding Park API =======
	cHomePage.OnRequest(func(r *colly.Request){
		fmt.Println("[Level 0] Visiting:", r.URL)
	})

	cHomePage.OnHTML("[data-api-url]", func(e *colly.HTMLElement) {
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
		if strings.Contains(contentType, "application/json" ){
			fmt.Println("[Level 1] Processing JSON response")

			// Parse the JSON to extract park URLs
			var jsonData map[string]interface{}
			err := json.Unmarshal(r.Body, &jsonData)
			if err != nil {
				fmt.Printf("Error parsing JSON: %v\n", err)
				return
			}

			// Extract park URLs from JSON
			// The JSON structure typically has an array of items with park details
			if items, ok := jsonData["listItems"].([]interface{}); ok {
				fmt.Printf("[Level 1] Found %d parks\n", len(items))

				for _, item := range items {
					if parkItem, ok := item.(map[string]interface{}); ok {
						// Extract the park name
						if name, ok := parkItem["parkName"].(string); ok {
							fmt.Printf("[Level 1] Found park: %s\n", name)
						}

						// Extract the park page URL from nested "meta" object
						if meta, ok := parkItem["meta"].(map[string]interface{}); ok {
							if url, ok := meta["dynamicPageLink"].(string); ok {
								absoluteURL := r.Request.AbsoluteURL(url)
								fmt.Printf("[Level 1] Queueing park page: %s\n", absoluteURL)
								cParkPage.Visit(absoluteURL)
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

		// Extract park information
		parkName := e.ChildText("h1")
		latitudeStr := e.ChildText("div.cmp-contentfragment__element--parkLatitude p.cmp-contentfragment__element-value")
		longitudeStr := e.ChildText("div.cmp-contentfragment__element--parkLongitude p.cmp-contentfragment__element-value")

		fmt.Printf("[Level 2] Park Name: %s\n", parkName)
		fmt.Printf("[Level 2] Latitude: %s\n", latitudeStr)
		fmt.Printf("[Level 2] Longitude: %s\n", longitudeStr)

		// Convert lat/long strings to float32
		latitude, err1 := strconv.ParseFloat(latitudeStr, 32)
		longitude, err2 := strconv.ParseFloat(longitudeStr, 32)

		// Only add park if we have valid data
		if parkName != "" && err1 == nil && err2 == nil {
			p := park{
				Name:      parkName,
				StateCode: "IL", // Illinois - could be extracted from page if needed
				Latitude:  float32(latitude),
				Longitude: float32(longitude),
			}

			// Thread-safe append to parks slice
			mu.Lock()
			parks = append(parks, p)
			mu.Unlock()

			fmt.Printf("[Level 2] Added park: %s (%.3f, %.3f)\n", p.Name, p.Latitude, p.Longitude)
		}
	})

	// Start the scraping process
	cHomePage.Visit("https://dnr.illinois.gov/parks/allparks.html")

	// Wait for all requests to finish
	cHomePage.Wait()
	cJsonAPI.Wait()
	cParkPage.Wait()

	// Serialize parks slice to JSON file
	fmt.Printf("\n[Final] Scraped %d parks total\n", len(parks))

	jsonData, err := json.MarshalIndent(parks, "", "  ")
	if err != nil {
		fmt.Printf("Error serializing parks to JSON: %v\n", err)
		return
	}

	err = os.WriteFile("parks.json", jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing parks.json: %v\n", err)
		return
	}

	fmt.Println("[Final] Parks data saved to parks.json")
}