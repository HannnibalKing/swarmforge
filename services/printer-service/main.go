package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// PrinterProfile represents known printer specifications
type PrinterProfile struct {
	Manufacturer       string
	Model              string
	DefaultTier        string
	PrintVolume        [3]int // x, y, z in mm
	Materials          []string
	NozzleDiameter     float64
	LayerHeightRange   [2]float64 // min, max
	ExpectedTolerance  float64    // mm
	CertificationTests []string
}

// Predefined printer profiles for popular models
var knownPrinters = map[string]PrinterProfile{
	"bambu_x1c": {
		Manufacturer:       "Bambu Lab",
		Model:              "X1 Carbon",
		DefaultTier:        "platinum",
		PrintVolume:        [3]int{256, 256, 256},
		Materials:          []string{"PLA", "PETG", "ABS", "ASA", "TPU", "PA", "PC"},
		NozzleDiameter:     0.4,
		LayerHeightRange:   [2]float64{0.08, 0.28},
		ExpectedTolerance:  0.05,
		CertificationTests: []string{"calibration_cube", "dimensional_accuracy", "bridging", "overhang"},
	},
	"bambu_p1s": {
		Manufacturer:       "Bambu Lab",
		Model:              "P1S",
		DefaultTier:        "gold",
		PrintVolume:        [3]int{256, 256, 256},
		Materials:          []string{"PLA", "PETG", "ABS", "TPU"},
		NozzleDiameter:     0.4,
		LayerHeightRange:   [2]float64{0.1, 0.28},
		ExpectedTolerance:  0.08,
		CertificationTests: []string{"calibration_cube", "dimensional_accuracy"},
	},
	"bambu_p1p": {
		Manufacturer:       "Bambu Lab",
		Model:              "P1P",
		DefaultTier:        "gold",
		PrintVolume:        [3]int{256, 256, 256},
		Materials:          []string{"PLA", "PETG", "TPU"},
		NozzleDiameter:     0.4,
		LayerHeightRange:   [2]float64{0.1, 0.28},
		ExpectedTolerance:  0.08,
		CertificationTests: []string{"calibration_cube", "dimensional_accuracy"},
	},
	"bambu_a1": {
		Manufacturer:       "Bambu Lab",
		Model:              "A1",
		DefaultTier:        "silver",
		PrintVolume:        [3]int{256, 256, 256},
		Materials:          []string{"PLA", "PETG", "TPU"},
		NozzleDiameter:     0.4,
		LayerHeightRange:   [2]float64{0.1, 0.28},
		ExpectedTolerance:  0.1,
		CertificationTests: []string{"calibration_cube"},
	},
	"prusa_xl": {
		Manufacturer:       "Prusa Research",
		Model:              "XL",
		DefaultTier:        "platinum",
		PrintVolume:        [3]int{360, 360, 360},
		Materials:          []string{"PLA", "PETG", "ABS", "ASA", "TPU", "PA", "PC", "PVB"},
		NozzleDiameter:     0.4,
		LayerHeightRange:   [2]float64{0.05, 0.35},
		ExpectedTolerance:  0.05,
		CertificationTests: []string{"calibration_cube", "dimensional_accuracy", "bridging", "overhang"},
	},
	"prusa_mk4": {
		Manufacturer:       "Prusa Research",
		Model:              "MK4",
		DefaultTier:        "gold",
		PrintVolume:        [3]int{250, 210, 220},
		Materials:          []string{"PLA", "PETG", "ABS", "ASA", "TPU", "PA"},
		NozzleDiameter:     0.4,
		LayerHeightRange:   [2]float64{0.05, 0.30},
		ExpectedTolerance:  0.05,
		CertificationTests: []string{"calibration_cube", "dimensional_accuracy", "bridging"},
	},
	"prusa_mk3s": {
		Manufacturer:       "Prusa Research",
		Model:              "MK3S+",
		DefaultTier:        "gold",
		PrintVolume:        [3]int{250, 210, 210},
		Materials:          []string{"PLA", "PETG", "ABS", "ASA", "TPU"},
		NozzleDiameter:     0.4,
		LayerHeightRange:   [2]float64{0.05, 0.30},
		ExpectedTolerance:  0.08,
		CertificationTests: []string{"calibration_cube", "dimensional_accuracy"},
	},
	"prusa_mini": {
		Manufacturer:       "Prusa Research",
		Model:              "Mini+",
		DefaultTier:        "silver",
		PrintVolume:        [3]int{180, 180, 180},
		Materials:          []string{"PLA", "PETG", "TPU"},
		NozzleDiameter:     0.4,
		LayerHeightRange:   [2]float64{0.05, 0.25},
		ExpectedTolerance:  0.1,
		CertificationTests: []string{"calibration_cube"},
	},
}

// CertificationScore calculates the overall certification score
type CertificationScore struct {
	DimensionalAccuracy float64 // 0-100
	SurfaceQuality      float64 // 0-100
	Consistency         float64 // 0-100
	OverallScore        float64 // 0-100
}

func calculateCertificationScore(testResults map[string]interface{}) CertificationScore {
	// Extract test results
	expectedDims := testResults["expected_dimensions"].(map[string]float64)
	actualDims := testResults["actual_dimensions"].(map[string]float64)

	// Calculate dimensional accuracy
	var totalError float64
	dimensions := []string{"x", "y", "z"}
	for _, dim := range dimensions {
		error := abs(expectedDims[dim] - actualDims[dim])
		totalError += error
	}
	avgError := totalError / 3.0

	// Score based on tolerance
	// ±0.05mm = 100 points, ±0.1mm = 95 points, ±0.2mm = 85 points, etc.
	dimensionalScore := 100.0
	if avgError > 0.05 {
		dimensionalScore = 100.0 - ((avgError - 0.05) * 100)
	}
	if dimensionalScore < 0 {
		dimensionalScore = 0
	}

	// Surface quality score (from manual inspection)
	surfaceScore := testResults["surface_quality_score"].(float64) * 10.0

	// Consistency (if multiple test prints available)
	consistencyScore := 90.0 // Default, improved with more test data

	overallScore := (dimensionalScore*0.5 + surfaceScore*0.3 + consistencyScore*0.2)

	return CertificationScore{
		DimensionalAccuracy: dimensionalScore,
		SurfaceQuality:      surfaceScore,
		Consistency:         consistencyScore,
		OverallScore:        overallScore,
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func determineTier(score CertificationScore, printerModel string) string {
	profile, exists := knownPrinters[printerModel]

	// Start with default tier for the model
	defaultTier := "silver"
	if exists {
		defaultTier = profile.DefaultTier
	}

	// Adjust based on actual performance
	if score.OverallScore >= 95 {
		return "platinum"
	} else if score.OverallScore >= 90 {
		if defaultTier == "platinum" {
			return "gold" // Downgrade if underperforming
		}
		return "gold"
	} else if score.OverallScore >= 85 {
		return "silver"
	}

	return "unverified"
}

// PrinterMatcher finds suitable printers for a job part
type PrinterMatcher struct {
	// Database connection would go here
}

type PartRequirements struct {
	Dimensions        [3]float64 // x, y, z in mm
	Material          string
	RequiredTier      string
	RequiredTolerance float64
	Location          [2]float64 // lat, lng
}

type MatchedPrinter struct {
	PrinterID  string
	Score      float64
	Distance   float64 // km
	Reputation float64
}

func (pm *PrinterMatcher) findBestPrinters(req PartRequirements, limit int) []MatchedPrinter {
	// TODO: Query database for printers matching:
	// 1. Print volume >= required dimensions
	// 2. Supports required material
	// 3. Tier >= required tier
	// 4. Tolerance <= required tolerance
	// 5. Currently available

	// Score each printer:
	// - Location proximity (30%)
	// - Reputation/success rate (30%)
	// - Tier/certification level (25%)
	// - Availability/turnaround time (15%)

	// Return top N matches

	log.Println("Finding best printers for requirements:", req)
	return []MatchedPrinter{}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	log.Println("Printer Service starting...")

	// Example: Print known printer profiles
	profiles, _ := json.MarshalIndent(knownPrinters, "", "  ")
	log.Printf("Loaded printer profiles:\n%s\n", string(profiles))

	port := os.Getenv("PRINTER_PORT")
	if port == "" {
		port = "8083"
	}
	registry := NewPrinterRegistry()
	log.Printf("Printer network service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, NewPrinterNetworkHandler(registry)))
}
