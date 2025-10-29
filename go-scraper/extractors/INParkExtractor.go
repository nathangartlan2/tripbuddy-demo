package extractors

import (
	"fmt"
	"scraper/models"
	"strings"

	"github.com/gocolly/colly"
)

type INParkExtractor struct {
}

func (s *INParkExtractor) ExtractParkData(e *colly.HTMLElement) *models.Park{
	fmt.Println("[Level 2] Extracting park details from:", e.Request.URL)

	// Extract park information
	parkName := e.ChildText("h1")
	latitude := 41
	longitude := -86

	fmt.Printf("[Level 2] Park Name: %s\n", parkName)

	activities := []models.ParkActivity{}


	e.ForEach("div#Activities li", func(_ int, el *colly.HTMLElement) {

          activityName := strings.TrimSpace(el.Text)
          // href := el.Attr("href")

          if activityName != "" {
              activity := models.ParkActivity{
                  Name:        activityName,
                  Description: "",
              }
              activities = append(activities, activity)
          }
      })

	// Only return park if we have valid data
	if parkName != ""  {
		return &models.Park{
			Name:       parkName,
			StateCode:  "IN", // Illinois - could be extracted from page if needed
			Latitude:   float32(latitude),
			Longitude:  float32(longitude),
			Activities: activities,
		}
	}

	return nil
}