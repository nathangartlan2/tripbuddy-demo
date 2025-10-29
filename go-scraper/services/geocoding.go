package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// GeocodingService handles geocoding requests to MapBox API
type GeocodingService struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// Coordinates represents a geographic location
type Coordinates struct {
	Latitude  float32
	Longitude float32
}

// MapBoxResponse represents the response from MapBox Geocoding API
type MapBoxResponse struct {
	Type     string `json:"type"`
	Query    []interface{} `json:"query"`
	Features []struct {
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		PlaceType  []string  `json:"place_type"`
		Relevance  float64   `json:"relevance"`
		Properties struct{}  `json:"properties"`
		Text       string    `json:"text"`
		PlaceName  string    `json:"place_name"`
		Center     []float64 `json:"center"`
		Geometry   struct {
			Type        string    `json:"type"`
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
		Context []struct {
			ID        string `json:"id"`
			ShortCode string `json:"short_code,omitempty"`
			WikiData  string `json:"wikidata,omitempty"`
			Text      string `json:"text"`
		} `json:"context"`
	} `json:"features"`
	Attribution string `json:"attribution"`
}

// NewGeocodingService creates a new geocoding service instance
func NewGeocodingService(apiKey string) *GeocodingService {
	return &GeocodingService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://api.mapbox.com/geocoding/v5/mapbox.places",
	}
}

// GeocodeAddress converts an address string to latitude and longitude coordinates
func (g *GeocodingService) GeocodeAddress(address string) (*Coordinates, error) {
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	if g.apiKey == "" {
		return nil, fmt.Errorf("MapBox API key is not configured")
	}

	// URL encode the address
	encodedAddress := url.QueryEscape(address)

	// Build the request URL
	requestURL := fmt.Sprintf("%s/%s.json?access_token=%s&limit=1",
		g.baseURL, encodedAddress, g.apiKey)

	// Make the HTTP request
	resp, err := g.httpClient.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make geocoding request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding API returned status %d", resp.StatusCode)
	}

	// Parse the response
	var mapboxResp MapBoxResponse
	if err := json.NewDecoder(resp.Body).Decode(&mapboxResp); err != nil {
		return nil, fmt.Errorf("failed to decode geocoding response: %w", err)
	}

	// Check if we got any results
	if len(mapboxResp.Features) == 0 {
		return nil, fmt.Errorf("no geocoding results found for address: %s", address)
	}

	// Extract coordinates from the first result
	feature := mapboxResp.Features[0]
	if len(feature.Center) < 2 {
		return nil, fmt.Errorf("invalid coordinates in response")
	}

	// MapBox returns [longitude, latitude]
	coords := &Coordinates{
		Longitude: float32(feature.Center[0]),
		Latitude:  float32(feature.Center[1]),
	}

	return coords, nil
}
