package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type NetworkPrinter struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	IP            string    `json:"ip"`
	Port          int       `json:"port"`
	Model         string    `json:"model"`
	Materials     []string  `json:"materials"`
	Status        string    `json:"status"`
	CurrentJobID  string    `json:"current_job_id,omitempty"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	RegisteredAt  time.Time `json:"registered_at"`
}

type PrinterRegistration struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	IP        string   `json:"ip"`
	Port      int      `json:"port"`
	Model     string   `json:"model"`
	Materials []string `json:"materials"`
}

type PrinterHeartbeat struct {
	Status       string `json:"status"`
	CurrentJobID string `json:"current_job_id,omitempty"`
}

type PrinterJob struct {
	JobID       string `json:"job_id"`
	PartID      string `json:"part_id"`
	ArtifactURL string `json:"artifact_url"`
}

type PrinterRegistry struct {
	mu       sync.RWMutex
	printers map[string]NetworkPrinter
	clock    func() time.Time
}

func NewPrinterRegistry() *PrinterRegistry {
	return &PrinterRegistry{printers: make(map[string]NetworkPrinter), clock: time.Now}
}

func (registry *PrinterRegistry) Register(input PrinterRegistration) (NetworkPrinter, error) {
	if strings.TrimSpace(input.ID) == "" || strings.TrimSpace(input.IP) == "" || input.Port < 1 || input.Port > 65535 {
		return NetworkPrinter{}, fmt.Errorf("id, ip, and a valid port are required")
	}
	now := registry.clock()
	printer := NetworkPrinter{ID: input.ID, Name: input.Name, IP: input.IP, Port: input.Port, Model: input.Model, Materials: input.Materials, Status: "online", LastHeartbeat: now, RegisteredAt: now}
	registry.mu.Lock()
	registry.printers[input.ID] = printer
	registry.mu.Unlock()
	return printer, nil
}

func (registry *PrinterRegistry) Heartbeat(id string, input PrinterHeartbeat) (NetworkPrinter, error) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	printer, ok := registry.printers[id]
	if !ok {
		return NetworkPrinter{}, fmt.Errorf("printer not registered")
	}
	if input.Status == "" {
		input.Status = "online"
	}
	printer.Status = input.Status
	printer.CurrentJobID = input.CurrentJobID
	printer.LastHeartbeat = registry.clock()
	registry.printers[id] = printer
	return printer, nil
}

func (registry *PrinterRegistry) List() []NetworkPrinter {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	result := make([]NetworkPrinter, 0, len(registry.printers))
	for _, printer := range registry.printers {
		result = append(result, printer)
	}
	return result
}

func (registry *PrinterRegistry) Dispatch(id string, job PrinterJob) (map[string]string, error) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	printer, ok := registry.printers[id]
	if !ok {
		return nil, fmt.Errorf("printer not registered")
	}
	if printer.Status != "online" && printer.Status != "idle" {
		return nil, fmt.Errorf("printer is not available")
	}
	printer.Status = "printing"
	printer.CurrentJobID = job.JobID
	printer.LastHeartbeat = registry.clock()
	registry.printers[id] = printer
	return map[string]string{"status": "accepted", "printer_id": id, "job_id": job.JobID, "target": fmt.Sprintf("http://%s:%d/api/jobs", printer.IP, printer.Port)}, nil
}

func writeNetworkJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func NewPrinterNetworkHandler(registry *PrinterRegistry) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeNetworkJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/v1/printers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			writeNetworkJSON(w, http.StatusOK, registry.List())
			return
		}
		if r.Method != http.MethodPost {
			writeNetworkJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var input PrinterRegistration
		if json.NewDecoder(r.Body).Decode(&input) != nil {
			writeNetworkJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid registration"})
			return
		}
		printer, err := registry.Register(input)
		if err != nil {
			writeNetworkJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeNetworkJSON(w, http.StatusCreated, printer)
	})
	mux.HandleFunc("/v1/printers/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			writeNetworkJSON(w, http.StatusNotFound, map[string]string{"error": "printer route not found"})
			return
		}
		id, action := parts[2], ""
		if len(parts) > 3 {
			action = parts[3]
		}
		if action == "" && r.Method == http.MethodGet {
			for _, printer := range registry.List() {
				if printer.ID == id {
					writeNetworkJSON(w, http.StatusOK, printer)
					return
				}
			}
			writeNetworkJSON(w, http.StatusNotFound, map[string]string{"error": "printer not registered"})
			return
		}
		if action == "heartbeat" && r.Method == http.MethodPost {
			var input PrinterHeartbeat
			if json.NewDecoder(r.Body).Decode(&input) != nil {
				writeNetworkJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid heartbeat"})
				return
			}
			printer, err := registry.Heartbeat(id, input)
			if err != nil {
				writeNetworkJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
				return
			}
			writeNetworkJSON(w, http.StatusOK, printer)
			return
		}
		if action == "jobs" && r.Method == http.MethodPost {
			var job PrinterJob
			if json.NewDecoder(r.Body).Decode(&job) != nil {
				writeNetworkJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid job"})
				return
			}
			result, err := registry.Dispatch(id, job)
			if err != nil {
				writeNetworkJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
				return
			}
			writeNetworkJSON(w, http.StatusAccepted, result)
			return
		}
		writeNetworkJSON(w, http.StatusNotFound, map[string]string{"error": "printer route not found"})
	})
	return mux
}
