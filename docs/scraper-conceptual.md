## Web Scraper Conceptual Flow

```mermaid
flowchart TD
      Homepage["Homepage URL
      Strategy: JSON API / Static HTML
      Uses: Selectors Config"] -->|collect| Parks[Park URLs]
      Parks -->|scrape| Info[Park Info + Activities]
```
