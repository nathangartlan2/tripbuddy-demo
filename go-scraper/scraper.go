package main

import (
	"fmt"
	"scraper/extractors"
	"scraper/scrapers"
)


func main() {
	extractor := &extractors.ILParkExtractor{}
	il := scrapers.NewILParkScraper(5, extractor)

	output, _, _ := il.ScrapePark("https://dnr.illinois.gov/parks/park.starvedrock.html")

	for i := 0; i < 10; i++{
		il.ScrapePark("https://dnr.illinois.gov/parks/park.starvedrock.html")
		if i % 50 == 0{
			fmt.Printf("Scraped %d", i)
		}
	}

	fmt.Println(output)
}