package scrapers

// ParkActivityScraper scrapes activity details from a park page
type ParkActivityScraper interface {
	// ScrapeActivities scrapes all activities from a park page URL
	ScrapeActivities(url string) ([]ParkActivity, error)
}

// ActivityScraperConfig holds configuration for scraping park activities
type ActivityScraperConfig struct {
	ActivitiesSelector string
}
