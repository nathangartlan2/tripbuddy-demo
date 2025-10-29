package extractors

import (
	"fmt"
	"scraper/models"
	"scraper/services"
	"strings"

	"github.com/gocolly/colly"
)

type INParkExtractor struct {
	geocoder *services.GeocodingService
}

func NewINParkExtractor(geocoder *services.GeocodingService) *INParkExtractor {
	return &INParkExtractor{
		geocoder: geocoder,
	}
}

func (s *INParkExtractor) ExtractParkData(e *colly.HTMLElement) *models.Park{
	// Extract park information
	parkName := e.ChildText("h1")

	// Extract address from the page
	var streetAddress string
	var cityStateZip string

	// Look for address in div#property-add p tags containing "Address:"
	e.ForEach("div#property-add p", func(_ int, el *colly.HTMLElement) {
		// Check if this paragraph contains "Address:"
		if strings.Contains(el.Text, "Address:") {
			// Replace <br> tags with newlines before getting text
			html, _ := el.DOM.Html()
			// Replace <br>, <br/>, and <br /> with newline
			html = strings.ReplaceAll(html, "<br>", "\n")
			html = strings.ReplaceAll(html, "<br/>", "\n")
			html = strings.ReplaceAll(html, "<br />", "\n")

			// Remove HTML tags to preserve line breaks
			cleanText := strings.NewReplacer(
				"<strong>", "",
				"</strong>", "",
			).Replace(html)

			// Remove remaining HTML tags
			cleanText = strings.TrimSpace(cleanText)

			// Split by newlines and extract address components
			lines := strings.Split(cleanText, "\n")
			var addressParts []string

			if(len(lines) > 2){
				for i := 1; i <= 2; i++{
					line := lines[i]
					line = strings.TrimSpace(line)

					// Skip the "Address:" label itself
					if strings.Contains(line, "Address:") {
						continue
					}

					// Collect non-empty lines
					if line != "" {
						addressParts = append(addressParts, line)

						// Check if this line contains city, state, zip (e.g., "Chesterton, IN 46304")
						if strings.Contains(line, ", IN ") {
							cityStateZip = line
						}
					}
				}
			}
			
			

			// First part is street address, last part is city/state/zip
			if len(addressParts) > 0 {
				streetAddress = addressParts[0]
			}
			if len(addressParts) > 1 {
				cityStateZip = addressParts[len(addressParts)-1]
			}
		}
	})

	// Build full address for geocoding
	var fullAddress string
	if streetAddress != "" && cityStateZip != "" {
		fullAddress = streetAddress + ", " + cityStateZip
	} else if streetAddress != "" {
		fullAddress = streetAddress + ", Indiana"
	} else {
		fullAddress = ""
	}

	// Default coordinates (center of Indiana)
	latitude := float32(41.0)
	longitude := float32(-86.0)

	// Try to geocode the address if we found one
	if fullAddress != "" && s.geocoder != nil {
		coords, err := s.geocoder.GeocodeAddress(fullAddress)
		if err != nil {
			fmt.Printf("[GEOCODING ERROR] Failed to geocode address '%s': %v\n", fullAddress, err)
		} else {
			latitude = coords.Latitude
			longitude = coords.Longitude
			fmt.Printf("[GEOCODING SUCCESS] %s -> (%.6f, %.6f)\n", fullAddress, latitude, longitude)
		}
	}

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
			StateCode:  "IN",
			Address:    fullAddress,
			Latitude:   latitude,
			Longitude:  longitude,
			Activities: activities,
		}
	}

	return nil
}