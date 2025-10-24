package scrapers

type ParkScraper interface {
	// ScrapeAll scrapes all parks from the state's park system
	ScrapeAll() ([]Park, error)
}


