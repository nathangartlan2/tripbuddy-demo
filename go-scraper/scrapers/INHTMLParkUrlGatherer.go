package scrapers

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

// INHTMLParkUrlGatherer implements ParkUrlGatherer by parsing HTML from Indiana DNR
type INHTMLParkUrlGatherer struct {
	userAgent string
	baseURL   string // Base URL to prepend to relative URLs
}

// NewINHTMLParkUrlGatherer creates a new INHTMLParkUrlGatherer
func NewINHTMLParkUrlGatherer(baseURL string) *INHTMLParkUrlGatherer {
	return &INHTMLParkUrlGatherer{
		userAgent: "TripBuddyBot/1.0 (Educational Park Data Scraper; +https://github.com/nathangartlan2/tripbuddy-demo)",
		baseURL:   baseURL,
	}
}

// GatherUrls fetches and parses the HTML from the main page URL to extract park URLs
func (g *INHTMLParkUrlGatherer) GatherUrls(mainPageUrl string) ([]string, error) {
	urls := make([]string, 0)

	// Create collector
	c := colly.NewCollector()

	// Set user agent
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", g.userAgent)
	})

	// Find the section with id="564717"
	c.OnHTML("section#564717", func(section *colly.HTMLElement) {
		// Find all links within this section
		section.ForEach("a[href]", func(_ int, link *colly.HTMLElement) {
			href := link.Attr("href")

			// Check if href contains the pattern "/dnr/state-parks/parks-lakes/"
			if strings.Contains(href, "/dnr/state-parks/parks-lakes/") {
				// Build full URL
				fullURL := href
				if g.baseURL != "" && !strings.HasPrefix(href, "http") {
					// Remove leading slash if present to avoid double slashes
					href = strings.TrimPrefix(href, "/")
					fullURL = g.baseURL + href
				}
				urls = append(urls, fullURL)
			}
		})
	})

	// Handle errors
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Error scraping %s: %v\n", r.Request.URL, err)
	})

	// Visit the URL
	err := c.Visit(mainPageUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to visit URL: %w", err)
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("no park URLs found matching pattern '/dnr/state-parks/parks-lakes/'")
	}

	return urls, nil
}
