package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func printerServiceURL() string {
	if value := os.Getenv("PRINTER_SERVICE_URL"); value != "" {
		return strings.TrimRight(value, "/")
	}
	return "http://printer-service:8083"
}

func proxyPrinter(c *gin.Context, method, path string, payload any) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid printer payload"})
			return
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(c.Request.Context(), method, printerServiceURL()+path, body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "printer service unavailable"})
		return
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "printer service unavailable"})
		return
	}
	defer response.Body.Close()
	var result any
	if json.NewDecoder(response.Body).Decode(&result) != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid printer service response"})
		return
	}
	c.JSON(response.StatusCode, result)
}

func registerPrinter(c *gin.Context) {
	var payload map[string]any
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid printer registration"})
		return
	}
	proxyPrinter(c, http.MethodPost, "/v1/printers", payload)
}
func getPrinter(c *gin.Context)           { proxyPrinter(c, http.MethodGet, "/v1/printers/"+c.Param("id"), nil) }
func getAvailablePrinters(c *gin.Context) { proxyPrinter(c, http.MethodGet, "/v1/printers", nil) }
