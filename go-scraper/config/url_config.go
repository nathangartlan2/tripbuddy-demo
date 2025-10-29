package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// StateURLs represents URLs for a specific state
type StateURLs struct {
	StateCode string   `json:"state_code"`
	URLs      []string `json:"urls"`
}

// URLConfig holds the configuration of URLs by state
type URLConfig struct {
	stateURLMap map[string][]string
}

// LoadURLConfig reads urls.json and returns a URLConfig
func LoadURLConfig(filepath string) (*URLConfig, error) {
	// Read the JSON file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON into slice of StateURLs
	var stateURLs []StateURLs
	if err := json.Unmarshal(data, &stateURLs); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Build the map
	stateURLMap := make(map[string][]string)
	for _, state := range stateURLs {
		stateURLMap[state.StateCode] = state.URLs
	}

	return &URLConfig{
		stateURLMap: stateURLMap,
	}, nil
}

// GetURLsByState returns the list of URLs for a given state code
func (c *URLConfig) GetURLsByState(stateCode string) ([]string, bool) {
	urls, ok := c.stateURLMap[stateCode]
	return urls, ok
}

// GetAllStates returns all state codes in the config
func (c *URLConfig) GetAllStates() []string {
	states := make([]string, 0, len(c.stateURLMap))
	for state := range c.stateURLMap {
		states = append(states, state)
	}
	return states
}

// GetAllURLs returns all URLs across all states
func (c *URLConfig) GetAllURLs() map[string][]string {
	return c.stateURLMap
}
