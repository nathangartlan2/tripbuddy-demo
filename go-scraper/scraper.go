package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

type Direction int 
 const (
      North Direction = iota  
      East                  
      South             
	  West
  )

type latLong struct {
	Magnitude float32 `json:"magnitude"`
	Direction Direction `json:"direction"`
}

type park struct {
	Name string  `json:"name"`
	StateCode string `json:"stateCode"`
	Latitude latLong `json:"latitude"`
	Longitude latLong `json:"longitude"`
}

func main(){
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
		parkName := e.ChildText("h1") // Adjust selector based on actual HTML structure
		fmt.Printf("[Level 2] Park Name: %s\n", parkName)

		// TODO: Add more specific selectors once we see the park page structure
		// Examples of what we might extract:
		// - Latitude/Longitude
		// - Address
		// - Description
		// - Activities/Amenities
		// - Contact information
	})

	// Start the scraping process
	cHomePage.Visit("https://dnr.illinois.gov/parks/allparks.html")
}