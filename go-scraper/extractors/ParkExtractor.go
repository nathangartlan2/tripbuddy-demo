package extractors

import (
	"scraper/models"

	"github.com/gocolly/colly"
)

type ParkExtractor interface{
	ExtractParkData(e *colly.HTMLElement) *models.Park
}