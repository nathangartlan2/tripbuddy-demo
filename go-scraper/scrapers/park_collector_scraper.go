package scrapers

// ParkCollectorScraper discovers park page URLs from a homepage
type ParkCollectorScraper interface {
	// CollectParkURLs scrapes the homepage and returns a list of park page URLs
	CollectParkURLs(homepageURL string) ([]string, error)
}
