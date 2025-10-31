package writers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"scraper/events"
	"time"
)

type APIParkWriter struct{
	baseUrl string
	client  *http.Client
}

func NewAPIParkWriter (url string) *APIParkWriter{
	return &APIParkWriter{baseUrl: url, client: &http.Client{
			Timeout: 10 * time.Second,
		},}
}

func (w *APIParkWriter) OnParkScraped(event events.ParkScrapedEvent){
	// Build the request URL
	requestURL := fmt.Sprintf("%s/park", w.baseUrl)


  	jsonData, _ := json.Marshal(event.Park)
   	bodyReader := bytes.NewReader(jsonData)
	// Make the HTTP request
	log.Printf("[APIWriter] Writing park %s to API", event.Park.Name)
	resp, err := w.client.Post(requestURL, "Application/JSON", bodyReader)
	if err != nil {
		fmt.Printf("[APIWriter] failed to post park %s \n Error : %w", event.Park.Name, err)
	}

	// Check response status
	if resp.StatusCode == 201 || resp.StatusCode == 200 {
		fmt.Printf("[APIWriter] POST successful. API returned status %d", resp.StatusCode)
	}else{
		fmt.Printf("[APIWriter] POST Failed: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	
}