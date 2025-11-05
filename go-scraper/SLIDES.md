# Go Software Design Patterns

## Building a State Park Web Scraper

**Nathan Gartlan**
November 6, 2025

---

## The Problem

Finding outdoor activities across multiple state park websites

<table>
  <tr>
    <td><img src="./images/starved_rock_page.png" alt="Starved Rock Page" width="280"/></td>
    <td><img src="./images/hiking_page_il.png" alt="Hiking Page IL" width="280"/></td>
    <td><img src="./images/boating_page_il.png" alt="Boating Page IL" width="280"/></td>
  </tr>
</table>

**Goal:** Scrape park data → Store in Searchable API

---

## System Architecture

```mermaid
flowchart LR
    A[State Park Website] -->|HTTP GET Raw HTML| C[Go Scraper<br/> <b>We are here</b>]
    B[Client] <-->|/Search| E[TripBuddy API<br/>.NET/C#]
    C -->|HTTP POST JSON| E
    E <-->|SQL CRUD| F[(PostgreSQL Database)]

    style C fill:#00ADD8,color:#fff,stroke:#000,stroke-width:3px
    style E fill:#512BD4,color:#fff
    style F fill:#336791,color:#fff
```

---

## Why Go for Web Scraping?

✅ **Fast compilation** - Quick iteration during development
✅ **Single binary** - Easy deployment (Docker, no dependencies)
✅ **Good Scraping Library** - `github.com/gocolly/colly`
✅ **Strong concurrency** - Goroutines for parallelism/async operations
✅ **Type safety** - Catch errors at compile time

---

## Pattern #1: Stragegy Patterm

**Problem:** Scraping park pages from different domains requires knowledge of different structures and data.

**Latitude Illinois State Parks HTML Structure**

```html
  <div cmp-contentfragment__element--parkLatitude">
      <dt class="cmp-contentfragment__element-title hidden"> Park Latitude </dt>
         <dt class="cmp-contentfragment__element-title hidden"> Park Latitude </dt>
        <p class="cmp-contentfragment__element-value paragraph">41.309</p>
  </div>
```

**Address in State Parks HTML Structure**

```html
<p>
  <strong>Address:</strong>
  <br />1600 N. 25 E. <br />
  Chesterton, IN 46304
</p>
```

---

## Solution: Strategy Pattern

**Solution**: Create an interface that abstracts the specific details of scraping a park

```mermaid
classDiagram
    class ParkExtractor {
        <<interface>>
        +ExtractParkData(e *colly.HTMLElement) *Park
    }

    class ILParkExtractor {
        +ExtractParkData(e *colly.HTMLElement) *Park
    }

    class INParkExtractor {
        +geocodingService GeocodingService
        +ExtractParkData(e *colly.HTMLElement) *Park
    }

    class BaseParkScraper {
        +extractor ParkExtractor
        +ScrapePark(url string) *Park
    }

    ParkExtractor <|.. ILParkExtractor : implements
    ParkExtractor <|.. INParkExtractor : implements
    BaseParkScraper --> ParkExtractor : uses
```

---

## Implicit Interfaces in Go

**Interface in Go**

```go
type ILParkExtractor struct {
}

func (s *ILParkExtractor) ExtractParkData(e *colly.HTMLElement) *models.Park{...}
```

**Interface in Java**

```java
public class ILParkExtractor : ParkExtractor{
    public models.Park ExtractParkData(colly.HTMLElement e){...}
}
```

---

## Implicit Interfaces: Pros & Cons

| **Pros ✅**                                                                                                                                                                          | **Cons ❌**                                                                                    |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------- |
| **Small Interface** design encouraged. Only require a few methods                                                                                                                    | **Less Explicit**: It's less clear to the developer if a class in fact implements an interface |
| **Flexibility**: Use structs from external packages that implement the interface methods                                                                                             |                                                                                                |
| **Refactoring**: If you need an interface and already have concrete types that satisfy the interface, you don't need to modify those types to explicitly refer to that new interface |                                                                                                |

**Bottom Line**: Tough to get used to, clearly offers functional code advantages

---

## Pattern #2: Observer Pattern

**Problem:** Scraper shouldn't care about persistence logic

**Solution:** Publish events when parks are scraped

```go
// Scraper publishes events
publisher.Publish(ParkScrapedEvent{
    Park:      park,
    StateCode: "IL",
    URL:       url,
    Duration:  elapsed,
})

// Subscribers handle persistence
type ParkEventSubscriber interface {
    OnParkScraped(event ParkScrapedEvent)
}
```

---

## Observer Pattern Architecture

```mermaid
classDiagram
    class BaseParkScraper {
        -onParkScraped func()
        -extractor ParkExtractor
        +ScrapePark(url string) Park
        +ScrapeAllParks(url string) []Park
    }

    class ParkEventSubscriber {
        <<interface>>
        +OnParkScraped(event ParkScrapedEvent)
    }

    class ParkEventPublisher {
        -subscribers []ParkEventSubscriber
        -eventQueue chan ParkScrapedEvent
        +Subscribe(subscriber ParkEventSubscriber)
        +Publish(event ParkScrapedEvent)
        +Close()
    }

    class FileParkWriter {
        -outputDir string
        +OnParkScraped(event ParkScrapedEvent)
    }

    class APIParkWriter {
        -apiURL string
        +OnParkScraped(event ParkScrapedEvent)
    }

    class ParkScrapedEvent {
        +Park *models.Park
        +StateCode string
        +URL string
        +Duration time.Duration
    }

    BaseParkScraper --> ParkEventPublisher : indirectly calls Publish()
    ParkEventPublisher --> ParkEventSubscriber : notifies
    ParkEventPublisher --> ParkScrapedEvent : publishes
    FileParkWriter ..|> ParkEventSubscriber : implements
    APIParkWriter ..|> ParkEventSubscriber : implements
```

---

## Event Queue Processing

```mermaid
sequenceDiagram
    participant S as BaseParkScraper
    participant CB as onParkScraped callback
    participant P as ParkEventPublisher
    participant Q as eventQueue (buffered chan)
    participant G as processEvents goroutine
    participant Sub1 as FileParkWriter
    participant Sub2 as APIParkWriter

    S->>CB: Scrapes park, calls callback
    CB->>P: Publish(event)
    P->>Q: event → queue (non-blocking)
    Note over Q: Buffered channel<br/>(100 events)

    Q->>G: event received
    G->>Sub1: OnParkScraped(event)
    S->>CB: Scrapes another park

    G->>Sub2: OnParkScraped(event)

    Note over G,Sub2: Events processed<br/>asynchronously

    CB->>P: Publish(event)
    P->>Q: event → queue

    Note over S,Q: Scraping continues<br/>without blocking
```

---

## Observer pattern in code Example

**Publisher with buffered queue:**

```go
publisher := events.NewParkEventPublisher()
defer publisher.Close()

jsonWriter := writers.NewParkJSONWriter("output")
publisher.Subscribe(jsonWriter)

publisher.WaitForQueue()

```

---

## Observer Pattern Benefits

1. **Decoupling** - Scraper doesn't know about storage
2. **Async Processing** - Events processed in background
3. **Extensibility** - Easy to add new subscribers
4. **Testability** - Mock subscribers for testing
5. **Performance** - Non-blocking scraping

---

## Demo Time

Let's see it in action!

```bash
# Start the system
docker-compose up

# Run the scraper
docker run --network tripbuddy-demo_tripbuddy-network \
  tripbuddy-scraper

# Query for parks
curl "http://localhost:8080/park/search?\
latitude=41.8789&longitude=-87.6359&\
activity=ski&radiusKm=1000"
```

---

## Summary of Key Patterns

1. **Strategy Pattern** - Handling different problems with common interfaces
2. **Observer Pattern** - Decouple scraping from persistence, asynchronously

---

## Resources

**Code:** [github.com/nathangartlan2/tripbuddy-demo](https://github.com/nathangartlan2/tripbuddy-demo)

**Documentation:**

- `go-scraper/README.md` - Full implementation details
- `go-scraper/OBSERVER_PATTERN.md` - Observer pattern deep-dive

**Questions?**

---

# Thank You!

**Nathan Gartlan**

Check out the code!
