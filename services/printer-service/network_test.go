package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPrinterRegistryRegistrationHeartbeatAndDispatch(t *testing.T) {
	registry := NewPrinterRegistry()
	printer, err := registry.Register(PrinterRegistration{ID: "printer-1", Name: "Workshop X1C", IP: "192.168.1.44", Port: 7125, Model: "bambu_x1c", Materials: []string{"PLA", "PETG"}})
	if err != nil || printer.Status != "online" {
		t.Fatalf("registration failed: %#v %v", printer, err)
	}
	updated, err := registry.Heartbeat("printer-1", PrinterHeartbeat{Status: "idle"})
	if err != nil || updated.Status != "idle" {
		t.Fatalf("heartbeat failed: %#v %v", updated, err)
	}
	result, err := registry.Dispatch("printer-1", PrinterJob{JobID: "job-1", PartID: "part-1", ArtifactURL: "http://storage/part-1.3mf"})
	if err != nil || result["status"] != "accepted" {
		t.Fatalf("dispatch failed: %#v %v", result, err)
	}
}

func TestPrinterNetworkHTTPContract(t *testing.T) {
	server := httptest.NewServer(NewPrinterNetworkHandler(NewPrinterRegistry()))
	defer server.Close()
	payload, _ := json.Marshal(PrinterRegistration{ID: "printer-2", IP: "10.0.0.8", Port: 8080, Model: "prusa_mk4"})
	response, err := http.Post(server.URL+"/v1/printers", "application/json", bytes.NewReader(payload))
	if err != nil || response.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected registration response: %v", err)
	}
	response.Body.Close()
	list, err := http.Get(server.URL + "/v1/printers")
	if err != nil || list.StatusCode != http.StatusOK {
		t.Fatalf("unexpected list response: %v", err)
	}
	list.Body.Close()
}
