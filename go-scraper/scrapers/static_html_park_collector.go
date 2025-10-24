package scrapers

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

// StaticHTMLParkCollector scrapes park URLs from static HTML
type StaticHTMLParkCollector struct {
	selectors StaticHTMLSelectors
}

// NewStaticHTMLParkCollector creates a new static HTML park collector
func NewStaticHTMLParkCollector(selectors StaticHTMLSelectors) *StaticHTMLParkCollector {
	return &StaticHTMLParkCollector{
		selectors: selectors,
	}
}

// CollectParkURLs implements the ParkCollectorScraper interface
func (c *StaticHTMLParkCollector) CollectParkURLs(homepageURL string) ([]string, error) {
	var parkURLs []string

	// Create collector
	cHomePage := colly.NewCollector()

	// ===== Extract park links from static HTML =======
	cHomePage.OnRequest(func(r *colly.Request) {
		fmt.Println("[Collector] Visiting homepage:", r.URL)
	})

	cHomePage.OnHTML(c.selectors.ParkLinksSelector, func(e *colly.HTMLElement) {
		parkURL := e.Request.AbsoluteURL(e.Attr("href"))
		parkName := ""

		// Extract park name based on configured attribute
		switch c.selectors.ParkNameAttribute {
		case "text":
			parkName = strings.TrimSpace(e.Text)
		case "title":
			parkName = e.Attr("title")
		case "aria-label":
			parkName = e.Attr("aria-label")
		default:
			parkName = strings.TrimSpace(e.Text)
		}

		fmt.Printf("[Collector] Found park: %s -> %s\n", parkName, parkURL)
		parkURLs = append(parkURLs, parkURL)
	})

	// Start the scraping process
	err := cHomePage.Visit(homepageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to visit homepage: %w", err)
	}

	// Wait for collector to finish
	cHomePage.Wait()

	fmt.Printf("[Collector] Collected %d park URLs total\n", len(parkURLs))

	return parkURLs, nil
}
