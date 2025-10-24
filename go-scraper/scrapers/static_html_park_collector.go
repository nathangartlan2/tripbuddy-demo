package scrapers

import (
	"fmt"
	"path/filepath"
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

// buildSectionSelector builds a CSS selector from the HTMLSection config
func (c *StaticHTMLParkCollector) buildSectionSelector() string {
	section := c.selectors.Section

	// Use explicit selector if provided
	if section.Selector != "" {
		return section.Selector
	}

	// Build selector from ID or Class
	if section.ID != "" {
		return fmt.Sprintf("#%s", section.ID)
	}

	if section.Class != "" {
		return fmt.Sprintf(".%s", section.Class)
	}

	// Default to body if nothing specified
	return "body"
}

// matchesHrefPattern checks if a URL matches the configured pattern
func (c *StaticHTMLParkCollector) matchesHrefPattern(href string) bool {
	pattern := c.selectors.URLElement.HrefPattern

	// If no pattern specified, match all
	if pattern == "" {
		return true
	}

	// Use filepath.Match for glob-style pattern matching
	matched, err := filepath.Match(pattern, href)
	if err != nil {
		// If pattern is invalid, try simple prefix/contains matching
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			return strings.HasPrefix(href, prefix)
		}
		return strings.Contains(href, pattern)
	}

	return matched
}

// CollectParkURLs implements the ParkCollectorScraper interface
func (c *StaticHTMLParkCollector) CollectParkURLs(homepageURL string) ([]string, error) {
	var parkURLs []string

	// Create collector
	cHomePage := colly.NewCollector()

	// Build the section selector
	sectionSelector := c.buildSectionSelector()
	fmt.Printf("[Collector] Using section selector: %s\n", sectionSelector)
	fmt.Printf("[Collector] Using href pattern: %s\n", c.selectors.URLElement.HrefPattern)

	// ===== Extract park links from static HTML =======
	cHomePage.OnRequest(func(r *colly.Request) {
		fmt.Println("[Collector] Visiting homepage:", r.URL)
	})

	// Find the section first, then look for links within it
	cHomePage.OnHTML(sectionSelector, func(section *colly.HTMLElement) {
		fmt.Printf("[Collector] Found section, searching for park links...\n")

		// Find all links within this section
		section.ForEach("a[href]", func(_ int, e *colly.HTMLElement) {
			href := e.Attr("href")

			// Check if this link matches our pattern
			if !c.matchesHrefPattern(href) {
				return // Skip this link
			}

			parkURL := e.Request.AbsoluteURL(href)
			parkName := ""

			// Extract park name based on configured attribute
			nameAttr := c.selectors.URLElement.ParkNameAttribute
			if nameAttr == "" {
				nameAttr = "text" // Default to text content
			}

			switch nameAttr {
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
