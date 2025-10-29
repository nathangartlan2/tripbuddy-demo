package models

type Park struct {
	Name       string         `json:"name"`
	StateCode  string         `json:"stateCode"`
	Address    string         `json:"address,omitempty"`
	Latitude   float32        `json:"latitude"`
	Longitude  float32        `json:"longitude"`
	Activities []ParkActivity `json:"activities"`
}

type ParkActivity struct {
	Name        string `json:"Name"`
	Description string `json:"description"`
}
