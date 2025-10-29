package events

import (
	"scraper/models"
	"time"
)

// ParkScrapedEvent represents a park that has been successfully scraped
type ParkScrapedEvent struct {
	Park      *models.Park
	StateCode string
	URL       string
	Duration  time.Duration
	Timestamp time.Time
}

// ParkEventSubscriber is the interface for park event subscribers
type ParkEventSubscriber interface {
	OnParkScraped(event ParkScrapedEvent)
}

// ParkEventPublisher manages subscribers and publishes events
type ParkEventPublisher struct {
	subscribers []ParkEventSubscriber
	eventQueue  chan ParkScrapedEvent
	done        chan bool
}

// NewParkEventPublisher creates a new event publisher
func NewParkEventPublisher() *ParkEventPublisher {
	p := &ParkEventPublisher{
		subscribers: make([]ParkEventSubscriber, 0),
		eventQueue:  make(chan ParkScrapedEvent, 100), // Buffer 100 events
		done:        make(chan bool),
	}

	// Start event processing goroutine
	go p.processEvents()

	return p
}

// Subscribe adds a subscriber to receive events
func (p *ParkEventPublisher) Subscribe(subscriber ParkEventSubscriber) {
	p.subscribers = append(p.subscribers, subscriber)
}

// Publish sends an event to all subscribers via the queue
func (p *ParkEventPublisher) Publish(event ParkScrapedEvent) {
	p.eventQueue <- event
}

// processEvents processes events from the queue in the background
func (p *ParkEventPublisher) processEvents() {
	for {
		select {
		case event := <-p.eventQueue:
			// Notify all subscribers
			for _, subscriber := range p.subscribers {
				subscriber.OnParkScraped(event)
			}
		case <-p.done:
			// Drain remaining events before exiting
			for len(p.eventQueue) > 0 {
				event := <-p.eventQueue
				for _, subscriber := range p.subscribers {
					subscriber.OnParkScraped(event)
				}
			}
			return
		}
	}
}

// Close stops the event publisher and drains the queue
func (p *ParkEventPublisher) Close() {
	p.done <- true
}

// WaitForQueue blocks until all events in the queue are processed
func (p *ParkEventPublisher) WaitForQueue() {
	for len(p.eventQueue) > 0 {
		time.Sleep(10 * time.Millisecond)
	}
}
