package scrapers

type ParkScraper interface {
	 ScrapePark(url string) (*Park, error)
}