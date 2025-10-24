package scrapers

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

// ILParkActivityScraper scrapes activities from Illinois park pages using Colly
type ILParkActivityScraper struct {
	config ActivityScraperConfig
}

// NewILParkActivityScraper creates a new Illinois park activity scraper
func NewILParkActivityScraper(config ActivityScraperConfig) *ILParkActivityScraper {
	return &ILParkActivityScraper{
		config: config,
	}
}

// ScrapeActivities implements the ParkActivityScraper interface
func (s *ILParkActivityScraper) ScrapeActivities(url string) ([]ParkActivity, error) {
	activities := []ParkActivity{}

	// Create collector for scraping the page
	c := colly.NewCollector()

	// Extract activities using configured selector
	c.OnHTML("body", func(e *colly.HTMLElement) {
		e.ForEach(s.config.ActivitiesSelector, func(_ int, el *colly.HTMLElement) {
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
	})

	// Handle errors
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Error scraping activities from %s: %v\n", url, err)
	})

	// Visit the URL
	err := c.Visit(url)
	if err != nil {
		return nil, fmt.Errorf("failed to visit park page: %w", err)
	}

	c.Wait()

	return activities, nil
}
