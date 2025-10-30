package extractors

import "scraper/services"

// ExtractorFactory creates extractors based on state code
type ExtractorFactory struct{
	geocodingService *services.GeocodingService
}

// NewExtractorFactory creates a new ExtractorFactory
func NewExtractorFactory(geocodingService *services.GeocodingService) *ExtractorFactory {
	return &ExtractorFactory{
		geocodingService: geocodingService,
	}
}

// CreateExtractor returns the appropriate extractor for a state code
func (f *ExtractorFactory) CreateExtractor(stateCode string) ParkExtractor {
	switch stateCode {
	case "IL":
		return &ILParkExtractor{}
	case "IN":
		return NewINParkExtractor(f.geocodingService)
	default:
		return nil
	}
}

// GetSupportedStates returns a list of all supported state codes
func (f *ExtractorFactory) GetSupportedStates() []string {
	return []string{"IL", "IN"}
}

// IsStateSupported checks if a state code is supported
func (f *ExtractorFactory) IsStateSupported(stateCode string) bool {
	return f.CreateExtractor(stateCode) != nil
}
