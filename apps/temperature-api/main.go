package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type TemperatureResponse struct {
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	SensorID    string    `json:"sensor_id"`
	SensorType  string    `json:"sensor_type"`
	Description string    `json:"description"`
}

func randomTemp() float64 {
	raw := rand.Float64()*45.0 - 10.0
	return math.Round(raw*10) / 10
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func handleTemperature(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")
	if location == "" {
		location = "unknown"
	}
	writeJSON(w, TemperatureResponse{
		Value:       randomTemp(),
		Unit:        "celsius",
		Timestamp:   time.Now().UTC(),
		Location:    location,
		Status:      "active",
		SensorID:    location,
		SensorType:  "temperature",
		Description: fmt.Sprintf("Temperature sensor at %s", location),
	})
}

func handleTemperatureByID(w http.ResponseWriter, r *http.Request) {
	sensorID := r.PathValue("id")
	writeJSON(w, TemperatureResponse{
		Value:       randomTemp(),
		Unit:        "celsius",
		Timestamp:   time.Now().UTC(),
		Location:    sensorID,
		Status:      "active",
		SensorID:    sensorID,
		SensorType:  "temperature",
		Description: fmt.Sprintf("Temperature sensor %s", sensorID),
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /temperature", handleTemperature)
	mux.HandleFunc("GET /temperature/{id}", handleTemperatureByID)
	mux.HandleFunc("GET /health", handleHealth)

	addr := ":" + port
	log.Printf("temperature-api listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
