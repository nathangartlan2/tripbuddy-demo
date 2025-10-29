package extractors

// ExtractorFactory creates extractors based on state code
type ExtractorFactory struct{}

// NewExtractorFactory creates a new ExtractorFactory
func NewExtractorFactory() *ExtractorFactory {
	return &ExtractorFactory{}
}

// CreateExtractor returns the appropriate extractor for a state code
func (f *ExtractorFactory) CreateExtractor(stateCode string) ParkExtractor {
	switch stateCode {
	case "IL":
		return &ILParkExtractor{}
	case "IN":
		return &INParkExtractor{}
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
