# Go Scraper Architecture

This document describes the interface structure and design patterns used in the TripBuddy park scraper.

## Architecture Diagram

```mermaid
classDiagram
    %% Core Interfaces
    <!-- class ParkScraper {
        <<interface>>
        +ScrapeAll()
    }

    class ParkCollectorScraper {
        <<interface>>
        +CollectParkURLs(homepageURL string)
    } -->

    class ParkActivityScraper {
        <<interface>>
        +ScrapeActivities(url string)
    }

    %% Main Orchestrator
    class MultiStateScraper {
        -configs map
        +ScrapeStates(stateCodes, concurrent)
        -scrapeSequential(stateCodes)
        -scrapeConcurrent(stateCodes)
        -scrapeState(stateCode)
        -scrapeParkPages(stateCode, config, parkURLs)
    }

    %% Factory
    class ParkCollectorFactory {
        <<factory>>
        +NewParkCollector(strategy, selectors)
    }

    %% Domain Models
    class Park {
        +Name string
        +StateCode string
        +Latitude float32
        +Longitude float32
        +Activities []ParkActivity
    }

    class ParkActivity {
        +Name string
        +Description string
    }


    %% Relationships

    ParkCollectorFactory ..> ParkCollectorScraper : creates

    MultiStateScraper ..> ParkCollectorFactory : uses
    MultiStateScraper ..> ParkCollectorScraper : uses
    MultiStateScraper --> Park : produces


    Park --> ParkActivity : contains
```

## Key Components

### Interfaces

**ParkActivityScraper** - Interface for scraping activities from individual park pages.

### Core Classes

**MultiStateScraper** - Orchestrates the scraping process:
- Manages state configurations
- Supports both sequential (`scrapeSequential`) and concurrent (`scrapeConcurrent`) execution modes
- Uses factory to create collectors
- Scrapes individual park pages (`scrapeParkPages`)
- Produces `Park` objects as output

**ParkCollectorFactory** - Creates collector implementations based on strategy configuration.

### Domain Models

**Park** - Represents a state park with name, state code, coordinates, and a list of activities.

**ParkActivity** - Represents an activity available at a park with name and description.
