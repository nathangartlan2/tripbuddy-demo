# Observer Pattern Implementation

## Overview

The scraper uses the **Observer Pattern** to decouple park scraping from park persistence. When a park is successfully scraped, an event is published to all subscribers.

## Architecture

```
┌─────────────────┐
│   Scraper       │
└────────┬────────┘
         │ scrapes park
         │
         ▼
    ┌────────────────────┐
    │ ParkScrapedEvent   │
    └────────┬───────────┘
             │
             ▼
    ┌────────────────────┐
    │ EventPublisher     │ (with buffered queue)
    └────────┬───────────┘
             │
             ├──────────────────────────┐
             ▼                          ▼
    ┌────────────────────┐    ┌─────────────────┐
    │ ParkJSONWriter     │    │ Other Subscribers│
    │ (saves to JSON)    │    │ (DB, API, etc.) │
    └────────────────────┘    └─────────────────┘
```

## Components

### 1. **ParkScrapedEvent** (`events/park_events.go`)
Event data structure containing:
- `Park` - The scraped park data
- `StateCode` - State code (IL, IN, etc.)
- `URL` - Source URL
- `Duration` - Time taken to scrape
- `Timestamp` - When it was scraped

### 2. **ParkEventPublisher** (`events/park_events.go`)
Manages subscribers and publishes events:
- `Subscribe(subscriber)` - Register a new subscriber
- `Publish(event)` - Send event to all subscribers via buffered queue
- `WaitForQueue()` - Block until all queued events are processed
- Uses a background goroutine to process events asynchronously

### 3. **ParkJSONWriter** (`writers/park_json_writer.go`)
Subscriber that writes parks to JSON files:
- Implements `ParkEventSubscriber` interface
- Creates directory structure: `output/{StateCode}/`
- Generates filenames: `park-name.json` (kebab-case)
- Writes pretty-printed JSON

## Usage

```go
// Create publisher
publisher := events.NewParkEventPublisher()
defer publisher.Close()

// Create and subscribe JSON writer
jsonWriter := writers.NewParkJSONWriter("output")
publisher.Subscribe(jsonWriter)

// Scrape parks - events are published automatically
parks := scrapeParksByState("IL", urls, factory, publisher)

// Wait for all writes to complete
publisher.WaitForQueue()
```

## Adding New Subscribers

Implement the `ParkEventSubscriber` interface:

```go
type MyCustomSubscriber struct {
    // your fields
}

func (s *MyCustomSubscriber) OnParkScraped(event events.ParkScrapedEvent) {
    // Handle the event
    // e.g., save to database, send to API, etc.
}

// Subscribe it
publisher.Subscribe(&MyCustomSubscriber{})
```

## Benefits

✅ **Decoupling** - Scraper doesn't know about persistence
✅ **Async Processing** - Events processed in background via queue
✅ **Extensibility** - Easy to add new subscribers (DB, API, etc.)
✅ **Testability** - Can mock subscribers for testing
✅ **No Blocking** - Scraping continues while files are being written

## Output Structure

```
output/
├── IL/
│   ├── starved-rock-state-park.json
│   ├── chain-o-lakes-state-park.json
│   └── volo-bog-state-natural-area.json
└── IN/
    ├── turkey-run-state-park.json
    ├── brown-county-state-park.json
    └── brookville-lake.json
```

Each JSON file contains the complete park data:
```json
{
  "name": "Starved Rock State Park",
  "stateCode": "IL",
  "latitude": 41.309,
  "longitude": -88.989,
  "activities": [
    {
      "activityName": "Hiking",
      "description": ""
    }
  ]
}
```
