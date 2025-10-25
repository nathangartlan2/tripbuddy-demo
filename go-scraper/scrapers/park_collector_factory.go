package scrapers

import "fmt"

// NewParkCollector creates a ParkCollectorScraper based on the strategy and selectors
func NewParkCollector(strategy string, selectors HomepageSelectors) (ParkCollectorScraper, error) {
	switch strategy {
	case "json_api":
		return NewJSONAPIParkCollector(
			selectors.APIURLAttribute,
			selectors.JSONAPI,
		), nil
	case "static_html":
		return NewStaticHTMLParkCollector(selectors.StaticHTML), nil
	default:
		return nil, fmt.Errorf("unknown homepage strategy: %s", strategy)
	}
}
