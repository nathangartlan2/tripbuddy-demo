package scrapers

// ParkUrlGatherer defines the interface for gathering park URLs from a main state park page
type ParkUrlGatherer interface {
	// GatherUrls takes a main state park page URL and returns a list of individual park page URLs
	GatherUrls(mainPageUrl string) ([]string, error)
}
