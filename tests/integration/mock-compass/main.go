package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type Policy struct {
	PolicyEngineName string `json:"policyEngineName"`
	PolicyRuleId     string `json:"policyRuleId"`
}

type EnrichmentRequest struct {
	Policy Policy `json:"policy"`
}

type ComplianceControl struct {
	Id                     string   `json:"id"`
	CatalogId              string   `json:"catalogId"`
	Category               string   `json:"category"`
	Applicability          []string `json:"applicability,omitempty"`
	RemediationDescription *string  `json:"remediationDescription,omitempty"`
}

type ComplianceFrameworks struct {
	Frameworks   []string `json:"frameworks"`
	Requirements []string `json:"requirements"`
}

type ComplianceRisk struct {
	Level string `json:"level"`
}

type Compliance struct {
	Control          ComplianceControl    `json:"control"`
	EnrichmentStatus string               `json:"enrichmentStatus"`
	Frameworks       ComplianceFrameworks `json:"frameworks"`
	Risk             *ComplianceRisk      `json:"risk,omitempty"`
}

type EnrichmentResponse struct {
	Compliance Compliance `json:"compliance"`
}

func main() {
	healthcheck := flag.Bool("healthcheck", false, "Run TCP healthcheck against addr and exit")
	fixturesPath := flag.String("fixtures", "/fixtures/compass-responses.json", "Path to fixtures JSON file")
	certPath := flag.String("cert", "/certs/compass.crt", "Path to TLS certificate")
	keyPath := flag.String("key", "/certs/compass.key", "Path to TLS private key")
	addr := flag.String("addr", ":8081", "Address to listen on")
	flag.Parse()

	if *healthcheck {
		conn, err := net.DialTimeout("tcp", *addr, 3*time.Second)
		if err != nil {
			os.Exit(1)
		}
		conn.Close()
		os.Exit(0)
	}

	fixtures, err := loadFixtures(*fixturesPath)
	if err != nil {
		log.Fatalf("Failed to load fixtures: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/enrich", handleEnrich(fixtures))
	mux.HandleFunc("GET /healthz", handleHealth)

	cert, err := tls.LoadX509KeyPair(*certPath, *keyPath)
	if err != nil {
		log.Fatalf("Failed to load TLS certificate: %v", err)
	}

	server := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	log.Printf("Starting mock Compass server on %s", *addr)
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func loadFixtures(path string) (map[string]EnrichmentResponse, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading fixtures file: %w", err)
	}

	var fixtures map[string]EnrichmentResponse
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return nil, fmt.Errorf("parsing fixtures JSON: %w", err)
	}

	return fixtures, nil
}

func handleEnrich(fixtures map[string]EnrichmentResponse) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req EnrichmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		response, ok := fixtures[req.Policy.PolicyRuleId]
		if !ok {
			response = EnrichmentResponse{
				Compliance: Compliance{
					Control: ComplianceControl{
						Id:        "",
						CatalogId: "",
						Category:  "",
					},
					EnrichmentStatus: "Unmapped",
					Frameworks: ComplianceFrameworks{
						Frameworks:   []string{},
						Requirements: []string{},
					},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok"}`)
}
