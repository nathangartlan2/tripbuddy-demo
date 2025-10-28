package scrapers

// ILConfig holds Illinois-specific scraper configuration
type ILConfig struct {
	BaseURL  string
	Homepage HomepageConfig
	ParkPage ParkPageConfig
}

// HomepageConfig holds configuration for scraping the homepage
type HomepageConfig struct {
	Strategy        string
	APIURLAttribute string
	JSONAPIConfig   JSONAPIConfig
}

// JSONAPIConfig holds JSON API scraping configuration
type JSONAPIConfig struct {
	ParksListPath string // e.g., "listItems"
	ParkNamePath  string // e.g., "parkName"
	ParkURLPath   string // e.g., "meta.dynamicPageLink"
}

// ParkPageConfig holds configuration for scraping individual park pages
type ParkPageConfig struct {
	Strategy           string
	NameSelector       string
	LatitudeSelector   string
	LongitudeSelector  string
	ActivitiesSelector string
}
