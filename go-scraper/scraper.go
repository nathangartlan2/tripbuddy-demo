package main

import (
	"fmt"
	"scraper/scrapers"
)


func main() {
	il := scrapers.NewILParkScraper()

	output, _ := il.ScrapePark("https://dnr.illinois.gov/parks/park.starvedrock.html")

	for i := 0; i < 1000; i++{
		il.ScrapePark("https://dnr.illinois.gov/parks/park.starvedrock.html")
	}

	fmt.Println(output)
}