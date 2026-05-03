package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TelemetryEvent struct {
	ID         string    `json:"id"`
	DeviceID   string    `json:"deviceId"`
	Capability string    `json:"capability"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
	OccurredAt time.Time `json:"occurredAt"`
	ReceivedAt time.Time `json:"receivedAt"`
}

var (
	store []TelemetryEvent
	mu    sync.RWMutex
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func ingestTelemetry(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DeviceID   string    `json:"deviceId"`
		Capability string    `json:"capability"`
		Value      float64   `json:"value"`
		Unit       string    `json:"unit"`
		OccurredAt time.Time `json:"occurredAt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if input.DeviceID == "" || input.Capability == "" {
		writeError(w, http.StatusBadRequest, "deviceId and capability are required")
		return
	}
	if input.OccurredAt.IsZero() {
		input.OccurredAt = time.Now().UTC()
	}
	event := TelemetryEvent{
		ID:         uuid.NewString(),
		DeviceID:   input.DeviceID,
		Capability: input.Capability,
		Value:      input.Value,
		Unit:       input.Unit,
		OccurredAt: input.OccurredAt,
		ReceivedAt: time.Now().UTC(),
	}
	mu.Lock()
	store = append(store, event)
	mu.Unlock()
	writeJSON(w, http.StatusCreated, event)
}

func listTelemetry(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	snapshot := make([]TelemetryEvent, len(store))
	copy(snapshot, store)
	mu.RUnlock()
	writeJSON(w, http.StatusOK, snapshot)
}

func telemetryByDevice(w http.ResponseWriter, r *http.Request) {
	deviceID := r.PathValue("deviceId")
	mu.RLock()
	result := []TelemetryEvent{}
	for _, e := range store {
		if e.DeviceID == deviceID {
			result = append(result, e)
		}
	}
	mu.RUnlock()
	writeJSON(w, http.StatusOK, result)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /telemetry", ingestTelemetry)
	mux.HandleFunc("GET /telemetry", listTelemetry)
	mux.HandleFunc("GET /telemetry/device/{deviceId}", telemetryByDevice)

	addr := ":" + port
	log.Printf("telemetry-service listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
