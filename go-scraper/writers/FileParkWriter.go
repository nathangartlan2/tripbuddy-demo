package writers

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"scraper/events"
	"strings"
)

// FileParkWriter subscribes to park events and writes them to JSON files
type FileParkWriter struct {
	outputDir string
}

// NewParkJSONWriter creates a new JSON writer that writes to the specified output directory
func NewParkJSONWriter(outputDir string) *FileParkWriter {
	return &FileParkWriter{
		outputDir: outputDir,
	}
}

// OnParkScraped is called when a park is scraped - writes it to a JSON file
func (w *FileParkWriter) OnParkScraped(event events.ParkScrapedEvent) {
	if event.Park == nil {
		log.Printf("[JSONWriter] Received nil park in event")
		return
	}

	// Create directory structure: output/{StateCode}/
	stateDir := filepath.Join(w.outputDir, event.StateCode)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		log.Printf("[JSONWriter] Failed to create directory %s: %v", stateDir, err)
		return
	}

	// Generate filename from park name: "Starved Rock State Park" -> "starved-rock-state-park.json"
	filename := w.generateFilename(event.Park.Name)
	filepath := filepath.Join(stateDir, filename)

	// Marshal park to JSON with indentation
	jsonData, err := json.MarshalIndent(event.Park, "", "  ")
	if err != nil {
		log.Printf("[JSONWriter] Failed to marshal park %s: %v", event.Park.Name, err)
		return
	}

	// Write to file
	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		log.Printf("[JSONWriter] Failed to write file %s: %v", filepath, err)
		return
	}

	log.Printf("[JSONWriter] ✓ Wrote %s to %s (%d bytes)", event.Park.Name, filepath, len(jsonData))
}

// generateFilename creates a kebab-case filename from park name
func (w *FileParkWriter) generateFilename(parkName string) string {
	// Convert to lowercase
	filename := strings.ToLower(parkName)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	filename = reg.ReplaceAllString(filename, "-")

	// Remove leading/trailing hyphens
	filename = strings.Trim(filename, "-")

	// Add .json extension
	return filename + ".json"
}
