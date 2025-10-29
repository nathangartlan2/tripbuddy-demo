package extractors

import (
	"scraper/models"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

type ILParkExtractor struct {
}

func (s *ILParkExtractor) ExtractParkData(e *colly.HTMLElement) *models.Park{
	// Extract park information
	parkName := e.ChildText("h1")
	latitudeStr := e.ChildText("div.cmp-contentfragment__element--parkLatitude p.cmp-contentfragment__element-value")
	longitudeStr := e.ChildText("div.cmp-contentfragment__element--parkLongitude p.cmp-contentfragment__element-value")

	// Convert lat/long strings to float32
	latitude, err1 := strconv.ParseFloat(latitudeStr, 32)
	longitude, err2 := strconv.ParseFloat(longitudeStr, 32)
	activities := []models.ParkActivity{}

	e.ForEach("ul.cmp-contentfragment__element-linkList li a", func(_ int, el *colly.HTMLElement) {
		activityName := strings.TrimSpace(el.Text)

		activity := models.ParkActivity{
			Name:        activityName,
			Description: "",
		}

		activities = append(activities, activity)

	})

	// Only return park if we have valid data
	if parkName != "" && err1 == nil && err2 == nil {
		return &models.Park{
			Name:       parkName,
			StateCode:  "IL", // Illinois - could be extracted from page if needed
			Latitude:   float32(latitude),
			Longitude:  float32(longitude),
			Activities: activities,
		}
	}

	return nil
}
