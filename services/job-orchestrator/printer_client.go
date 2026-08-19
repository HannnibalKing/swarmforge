package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type RegisteredPrinter struct {
	ID        string   `json:"id"`
	IP        string   `json:"ip"`
	Port      int      `json:"port"`
	Model     string   `json:"model"`
	Materials []string `json:"materials"`
	Status    string   `json:"status"`
}

type PrinterClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewPrinterClient() *PrinterClient {
	base := os.Getenv("PRINTER_SERVICE_URL")
	if base == "" {
		base = "http://printer-service:8083"
	}
	return &PrinterClient{BaseURL: strings.TrimRight(base, "/"), HTTPClient: http.DefaultClient}
}

func (client *PrinterClient) ListPrinters(ctx context.Context) ([]RegisteredPrinter, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, client.BaseURL+"/v1/printers", nil)
	if err != nil {
		return nil, err
	}
	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("printer service returned %s", response.Status)
	}
	var printers []RegisteredPrinter
	if err := json.NewDecoder(response.Body).Decode(&printers); err != nil {
		return nil, err
	}
	return printers, nil
}

func (client *PrinterClient) Dispatch(ctx context.Context, printerID string, job PrinterJobRequest) (map[string]string, error) {
	payload, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.BaseURL+"/v1/printers/"+printerID+"/jobs", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("printer dispatch returned %s", response.Status)
	}
	var result map[string]string
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

type PrinterJobRequest struct {
	JobID       string `json:"job_id"`
	PartID      string `json:"part_id"`
	ArtifactURL string `json:"artifact_url"`
}
