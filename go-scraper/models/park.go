package models

type Park struct {
	Name       string         `json:"name"`
	StateCode  string         `json:"stateCode"`
	Latitude   float32        `json:"latitude"`
	Longitude  float32        `json:"longitude"`
	Activities []ParkActivity `json:"activities"`
}

type ParkActivity struct {
	Name        string `json:"activityName"`
	Description string `json:"description"`
}
