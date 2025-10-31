package scrapers

import "scraper/models"

type ParkScraper interface {
	 ScrapePark(url string) (*models.Park, error)

	 ScrapeAllParks(url string) (*[] models.Park, error)
}