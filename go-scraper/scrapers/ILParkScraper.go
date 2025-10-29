package scrapers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type ILParkScraper struct{
	waitMS int;
	maxRetries int;
	attemptNumber int;
}


func NewILParkScraper(maxRetries int) *ILParkScraper {
	return &ILParkScraper{
		waitMS: 1,
		maxRetries: maxRetries,
		attemptNumber: 1,
	}
}


func (s *ILParkScraper) extractParkData(e *colly.HTMLElement, scrapedPark **Park) {
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
	activities := []ParkActivity{}

	e.ForEach("ul.cmp-contentfragment__element-linkList li a", func(_ int, el *colly.HTMLElement) {
		activityName := strings.TrimSpace(el.Text)
		href := el.Attr("href")
		ariaLabel := el.Attr("aria-label")

		activity := ParkActivity{
			Name:        activityName,
			Description: "",
		}

		activities = append(activities, activity)

		fmt.Printf("Activity: %s, URL: %s, Label: %s\n", activity, href, ariaLabel)
	})

	// Only add park if we have valid data
	if parkName != "" && err1 == nil && err2 == nil {
		p := Park{
			Name:       parkName,
			StateCode:  "IL", // Illinois - could be extracted from page if needed
			Latitude:   float32(latitude),
			Longitude:  float32(longitude),
			Activities: activities,
		}

		*scrapedPark = &p
	}
}



func (s *ILParkScraper) ScrapePark(url string) (*Park, time.Duration, error) {
	startTime := time.Now()

	for i := 0; i < s.maxRetries; i++ {
		Park , err := s.scrapeParkInternal(url)

		if err == nil {
			elapsed := time.Since(startTime)
			return Park, elapsed, nil
		} else {
			fmt.Printf("[Retry %d/%d] Error scraping URL: %s\n", i+1, s.maxRetries, url)
			fmt.Printf("  Error: %v\n", err)
			fmt.Printf("  Waiting %dms before retry...\n\n", s.waitMS)
			time.Sleep(time.Duration(s.waitMS) * time.Millisecond)
			s.waitMS *= 2
		}
	}

	elapsed := time.Since(startTime)
	return nil, elapsed, fmt.Errorf("failed to scrape park after %d retries", s.maxRetries)
}

func (s *ILParkScraper) scrapeParkInternal(url string) (*Park, error) {
	cParkPage := colly.NewCollector()

	var scrapedPark *Park

	cParkPage.OnRequest(func(r *colly.Request) {
		fmt.Println("[Level 2] Scraping park page:", r.URL)
	})

	// Extract park details from individual park pages
	cParkPage.OnHTML("body", func(e *colly.HTMLElement) {
		s.extractParkData(e, &scrapedPark)
	})

	err := cParkPage.Visit(url)
	if err != nil {
		return nil, err
	}

	return scrapedPark, nil
}

