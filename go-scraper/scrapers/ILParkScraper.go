package scrapers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

type ILParkScraper struct{

}


func NewILParkScraper() *ILParkScraper {
	return &ILParkScraper{
	}
}


func (s *ILParkScraper) extractParkData(e *colly.HTMLElement, scrapedPark **Park) {
	fmt.Println("[Level 2] Extracting park details from:", e.Request.URL)

	// Extract park information
	parkName := e.ChildText("h1")
	latitudeStr := e.ChildText("div.cmp-contentfragment__element--parkLatitude p.cmp-contentfragment__element-value")
	longitudeStr := e.ChildText("div.cmp-contentfragment__element--parkLongitude p.cmp-contentfragment__element-value")

	fmt.Printf("[Level 2] Park Name: %s\n", parkName)
	fmt.Printf("[Level 2] Latitude: %s\n", latitudeStr)
	fmt.Printf("[Level 2] Longitude: %s\n", longitudeStr)

	// Convert lat/long strings to float32
	latitude, err1 := strconv.ParseFloat(latitudeStr, 32)
	longitude, err2 := strconv.ParseFloat(longitudeStr, 32)
	activities := []ParkActivity{}

	e.ForEach("ul.cmp-contentfragment__element-linkList li a", func(_ int, el *colly.HTMLElement) {
		activityName := strings.TrimSpace(el.Text)
		href := el.Attr("href")
		ariaLabel := el.Attr("aria-label")

		activity := ParkActivity{
			Name:        activityName,
			Description: "",
		}

		activities = append(activities, activity)

		fmt.Printf("Activity: %s, URL: %s, Label: %s\n", activity, href, ariaLabel)
	})

	// Only add park if we have valid data
	if parkName != "" && err1 == nil && err2 == nil {
		p := Park{
			Name:       parkName,
			StateCode:  "IL", // Illinois - could be extracted from page if needed
			Latitude:   float32(latitude),
			Longitude:  float32(longitude),
			Activities: activities,
		}

		*scrapedPark = &p
	}
}

func (s *ILParkScraper) ScrapePark(url string) (*Park, error) {
	cParkPage := colly.NewCollector()

	var scrapedPark *Park

	cParkPage.OnRequest(func(r *colly.Request) {
		fmt.Println("[Level 2] Scraping park page:", r.URL)
	})

	// Extract park details from individual park pages
	cParkPage.OnHTML("body", func(e *colly.HTMLElement) {
		s.extractParkData(e, &scrapedPark)
	})

	err := cParkPage.Visit(url)
	if err != nil {
		return nil, err
	}

	return scrapedPark, nil
}

