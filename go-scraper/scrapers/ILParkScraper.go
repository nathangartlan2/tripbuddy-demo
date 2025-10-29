package scrapers

import (
	"fmt"
	"scraper/extractors"
	"scraper/models"
	"time"

	"github.com/gocolly/colly"
)

type ILParkScraper struct{
	waitMS int;
	maxRetries int;
	attemptNumber int;
	userAgent string;
	extractor extractors.ParkExtractor;
}


func NewILParkScraper(maxRetries int, extractor extractors.ParkExtractor) *ILParkScraper {
	return &ILParkScraper{
		waitMS: 1,
		maxRetries: maxRetries,
		attemptNumber: 1,
		userAgent: "TripBuddyBot/1.0 (Educational Park Data Scraper; +https://github.com/nathangartlan2/tripbuddy-demo)",
		extractor: extractor,
	}
}



func (s *ILParkScraper) ScrapePark(url string) (*models.Park, time.Duration, error) {

	startTime := time.Now()
	time.Sleep(time.Duration(s.waitMS) * time.Millisecond)

	for i := 0; i < s.maxRetries; i++ {
		Park , err := s.scrapeParkInternal(url)

		if err == nil {
			elapsed := time.Since(startTime)
			if(s.waitMS > 1){
				s.waitMS /= 2
			}
			return Park, elapsed, nil
		} else {
			fmt.Printf("[Retry %d/%d] Error scraping URL: %s\n", i+1, s.maxRetries, url)
			fmt.Printf("  Error: %v\n", err)
			fmt.Printf("  Waiting %dms before retry...\n\n", s.waitMS)
			s.waitMS *= 2
		}
	}

	elapsed := time.Since(startTime)
	return nil, elapsed, fmt.Errorf("failed to scrape park after %d retries", s.maxRetries)
}

func (s *ILParkScraper) scrapeParkInternal(url string) (*models.Park, error) {
	cParkPage := colly.NewCollector()

	var scrapedPark *models.Park

	cParkPage.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", s.userAgent)
		fmt.Println("[Level 2] Scraping park page:", r.URL)
	})

	// Extract park details from individual park pages
	cParkPage.OnHTML("body", func(e *colly.HTMLElement) {
		scrapedPark = s.extractor.ExtractParkData(e)
	})

	err := cParkPage.Visit(url)
	if err != nil {
		return nil, err
	}

	return scrapedPark, nil
}

