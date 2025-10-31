package scrapers

import (
	"fmt"
	"scraper/extractors"
	"scraper/models"
	"time"

	"github.com/gocolly/colly"
)

type BaseParkScraper struct{
	waitMS int;
	maxRetries int;
	attemptNumber int;
	userAgent string;
	extractor extractors.ParkExtractor;
	urlGatherer ParkUrlGatherer;
	onParkScraped  func(park *models.Park, duration time.Duration, timestamp time.Time);
}


func NewBaseParkScraper(maxRetries int, extractor extractors.ParkExtractor, urlGatherer ParkUrlGatherer, onParkScraped func(park *models.Park, duration time.Duration, timestamp time.Time)) *BaseParkScraper {
	return &BaseParkScraper{
		waitMS: 1,
		maxRetries: maxRetries,
		attemptNumber: 1,
		userAgent: "TripBuddyBot/1.0 (Educational Park Data Scraper; +https://github.com/nathangartlan2/tripbuddy-demo)",
		extractor: extractor,
		urlGatherer: urlGatherer,
		onParkScraped: onParkScraped,
	}
}



func (s *BaseParkScraper) ScrapePark(url string) (*models.Park, time.Duration, error) {

	startTime := time.Now()
	time.Sleep(time.Duration(s.waitMS) * time.Millisecond)

	for i := 0; i < s.maxRetries; i++ {
		fmt.Println("[SCRAPER] park details from:", url)

		Park , err := s.scrapeParkInternal(url)

		if err == nil {
			elapsed := time.Since(startTime)
			if(s.waitMS > 1){
				s.waitMS /= 2
			}

			// Call callback if provided
			if s.onParkScraped != nil {
				s.onParkScraped(Park, elapsed, time.Now())
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

func (s *BaseParkScraper) scrapeParkInternal(url string) (*models.Park, error) {
	cParkPage := colly.NewCollector()

	var scrapedPark *models.Park

	cParkPage.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", s.userAgent)
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

// ScrapeAllParks uses the ParkUrlGatherer to collect all park URLs and then scrapes each one
func (s *BaseParkScraper) ScrapeAllParks(mainPageUrl string) (*[]models.Park, error) {
	if s.urlGatherer == nil {
		return nil, fmt.Errorf("urlGatherer is not set")
	}

	// Gather all park URLs from the main page
	fmt.Printf("[SCRAPER] Gathering park URLs from: %s\n", mainPageUrl)
	urls, err := s.urlGatherer.GatherUrls(mainPageUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to gather URLs: %w", err)
	}

	fmt.Printf("[SCRAPER] Found %d park URLs to scrape\n", len(urls))

	// Scrape each park
	parks := make([]models.Park, 0, len(urls))
	for i, url := range urls {
		fmt.Printf("[%d/%d] ", i+1, len(urls))
		park, duration, err := s.ScrapePark(url)
		if err != nil {
			fmt.Printf("Failed to scrape %s: %v\n", url, err)
			continue
		}

		if(park == nil){
			fmt.Printf("[SCRAPER] Issue scraping park at  %s parks. Skipping \n", url)
			continue;
		}
		parks = append(parks, *park)


		// Call callback if provided
		if s.onParkScraped != nil {
			s.onParkScraped(park, duration, time.Now())
		}
	}

	fmt.Printf("[SCRAPER] Successfully scraped %d/%d parks\n", len(parks), len(urls))
	return &parks, nil
}

