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

type Device struct {
	ID        string    `json:"id"`
	HouseID   string    `json:"houseId"`
	RoomID    string    `json:"roomId"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Vendor    string    `json:"vendor"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

var (
	store = map[string]*Device{}
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

func listDevices(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	list := make([]*Device, 0, len(store))
	for _, d := range store {
		list = append(list, d)
	}
	mu.RUnlock()
	writeJSON(w, http.StatusOK, list)
}

func createDevice(w http.ResponseWriter, r *http.Request) {
	var input struct {
		HouseID string `json:"houseId"`
		RoomID  string `json:"roomId"`
		Name    string `json:"name"`
		Type    string `json:"type"`
		Vendor  string `json:"vendor"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if input.Name == "" || input.Type == "" {
		writeError(w, http.StatusBadRequest, "name and type are required")
		return
	}
	if input.Status == "" {
		input.Status = "online"
	}
	d := &Device{
		ID:        uuid.NewString(),
		HouseID:   input.HouseID,
		RoomID:    input.RoomID,
		Name:      input.Name,
		Type:      input.Type,
		Vendor:    input.Vendor,
		Status:    input.Status,
		CreatedAt: time.Now().UTC(),
	}
	mu.Lock()
	store[d.ID] = d
	mu.Unlock()
	writeJSON(w, http.StatusCreated, d)
}

func getDevice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	mu.RLock()
	d, ok := store[id]
	mu.RUnlock()
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func updateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Status == "" {
		writeError(w, http.StatusBadRequest, "status is required")
		return
	}
	mu.Lock()
	d, ok := store[id]
	if ok {
		d.Status = req.Status
	}
	mu.Unlock()
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /devices", listDevices)
	mux.HandleFunc("POST /devices", createDevice)
	mux.HandleFunc("GET /devices/{id}", getDevice)
	mux.HandleFunc("PATCH /devices/{id}/status", updateStatus)

	addr := ":" + port
	log.Printf("device-service listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
