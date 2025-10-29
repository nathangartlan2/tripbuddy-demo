package main

import (
	"fmt"
	"log"
	"scraper/config"
)

func main() {
	// Load the URL configuration
	urlConfig, err := config.LoadURLConfig("../urls.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Example 1: Get URLs for a specific state
	fmt.Println("=== Example 1: Get URLs by State ===")
	ilURLs, ok := urlConfig.GetURLsByState("IL")
	if ok {
		fmt.Printf("Illinois Parks (%d):\n", len(ilURLs))
		for i, url := range ilURLs {
			fmt.Printf("  %d. %s\n", i+1, url)
		}
	}

	fmt.Println()

	// Example 2: Get all states
	fmt.Println("=== Example 2: Get All States ===")
	states := urlConfig.GetAllStates()
	fmt.Printf("States in config: %v\n", states)

	fmt.Println()

	// Example 3: Iterate over all states and their URLs
	fmt.Println("=== Example 3: All States and URLs ===")
	allURLs := urlConfig.GetAllURLs()
	for state, urls := range allURLs {
		fmt.Printf("%s: %d parks\n", state, len(urls))
		for _, url := range urls {
			fmt.Printf("  - %s\n", url)
		}
		fmt.Println()
	}

	// Example 4: Check if a state exists
	fmt.Println("=== Example 4: Check State Exists ===")
	if _, ok := urlConfig.GetURLsByState("CA"); ok {
		fmt.Println("California found!")
	} else {
		fmt.Println("California not in config")
	}
}
