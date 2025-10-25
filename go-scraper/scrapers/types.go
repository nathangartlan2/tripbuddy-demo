package scrapers

// HomepageSelectors holds selectors for discovering parks on homepage
type HomepageSelectors struct {
	// For JSON API strategy
	APIURLAttribute string           `json:"apiURLAttribute,omitempty"` // CSS selector for JSON API URL
	JSONAPI         JSONAPISelectors `json:"jsonAPI,omitempty"`         // JSON parsing config

	// For Static HTML strategy
	StaticHTML StaticHTMLSelectors `json:"staticHTML,omitempty"` // Static HTML parsing config
}

// JSONAPISelectors for parsing the JSON response
type JSONAPISelectors struct {
	ParksListPath string `json:"parksListPath"` // e.g., "listItems"
	ParkNamePath  string `json:"parkNamePath"`  // e.g., "parkName"
	ParkURLPath   string `json:"parkURLPath"`   // e.g., "meta.dynamicPageLink"
}

// StaticHTMLSelectors for parsing static HTML park lists
type StaticHTMLSelectors struct {
	Section    HTMLSection `json:"section"`    // Parent section containing park links
	URLElement URLElement  `json:"urlElement"` // URL element configuration
}

// HTMLSection identifies the parent container for park links
type HTMLSection struct {
	ID       string `json:"id,omitempty"`       // HTML id attribute
	Class    string `json:"class,omitempty"`    // HTML class attribute
	Selector string `json:"selector,omitempty"` // CSS selector (most flexible option)
}

// URLElement configures how to identify and extract park URLs
type URLElement struct {
	HrefPattern       string `json:"hrefPattern"`                 // Pattern to match href (e.g., "/dnr/state-parks/parks-lakes/*")
	ParkNameAttribute string `json:"parkNameAttribute,omitempty"` // Attribute to get park name (e.g., "text", "title", "aria-label")
}

// ParkPageSelectors for extracting park details from HTML
type ParkPageSelectors struct {
	NameSelector       string `json:"nameSelector"`       // CSS selector for park name
	LatitudeSelector   string `json:"latitudeSelector"`   // CSS selector for latitude
	LongitudeSelector  string `json:"longitudeSelector"`  // CSS selector for longitude
	ActivitiesSelector string `json:"activitiesSelector"` // CSS selector for activities list
}
