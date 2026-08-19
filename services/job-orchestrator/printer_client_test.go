package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPrinterClientListAndDispatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			writer.Header().Set("Content-Type", "application/json")
			writer.Write([]byte(`[{"id":"p1","ip":"10.0.0.2","port":8080,"status":"idle"}]`))
			return
		}
		writer.WriteHeader(http.StatusAccepted)
		writer.Write([]byte(`{"status":"accepted","printer_id":"p1","job_id":"j1"}`))
	}))
	defer server.Close()
	client := &PrinterClient{BaseURL: server.URL, HTTPClient: server.Client()}
	printers, err := client.ListPrinters(context.Background())
	if err != nil || len(printers) != 1 {
		t.Fatalf("list failed: %#v %v", printers, err)
	}
	result, err := client.Dispatch(context.Background(), "p1", PrinterJobRequest{JobID: "j1", PartID: "part1", ArtifactURL: "http://storage/part1"})
	if err != nil || result["status"] != "accepted" {
		t.Fatalf("dispatch failed: %#v %v", result, err)
	}
}
