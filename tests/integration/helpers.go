// Package integration provides shared helpers for integration test suites.
package integration

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// HTTPClient is a shared client with a generous timeout for polling operations.
var HTTPClient = &http.Client{Timeout: 10 * time.Second}

// PostEvidence reads a fixture file and POSTs it to the webhook endpoint.
// Returns the HTTP response for status code assertions.
func PostEvidence(webhookURL, fixturePath string) (*http.Response, error) {
	body, err := os.ReadFile(fixturePath)
	if err != nil {
		return nil, fmt.Errorf("reading fixture %s: %w", fixturePath, err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL+"/eventsource/receiver", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return HTTPClient.Do(req)
}

// LokiQueryResponse is the envelope returned by Loki's query_range API.
type LokiQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		Result []struct {
			Values [][]string `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// QueryLoki queries Loki's query_range endpoint and returns the log line strings.
// The start time is set to 1 hour ago to capture recent logs. Returns an empty
// slice (not an error) when no results match — this works with Eventually/ShouldNot(BeEmpty()).
func QueryLoki(lokiURL, query string) ([]string, error) {
	startTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339Nano)

	req, err := http.NewRequest(http.MethodGet, lokiURL+"/loki/api/v1/query_range", nil)
	if err != nil {
		return nil, fmt.Errorf("creating loki request: %w", err)
	}
	q := req.URL.Query()
	q.Set("query", query)
	q.Set("start", startTime)
	q.Set("limit", "10")
	req.URL.RawQuery = q.Encode()

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("querying loki: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("loki returned status %d: %s", resp.StatusCode, string(body))
	}

	var result LokiQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding loki response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("loki query status: %s", result.Status)
	}

	var lines []string
	for _, stream := range result.Data.Result {
		for _, entry := range stream.Values {
			if len(entry) >= 2 {
				lines = append(lines, entry[1])
			}
		}
	}
	return lines, nil
}

// S3ListBucketResult is the XML envelope for S3 ListObjectsV2.
type S3ListBucketResult struct {
	XMLName  xml.Name   `xml:"ListBucketResult"`
	Contents []S3Object `xml:"Contents"`
}

// S3Object represents a single object in an S3 listing.
type S3Object struct {
	Key string `xml:"Key"`
}

// ListS3Objects queries the S3 ListObjectsV2 API via plain HTTP (anonymous access)
// and returns the object keys matching the given prefix.
func ListS3Objects(s3URL, bucket, prefix string) ([]string, error) {
	url := fmt.Sprintf("%s/%s?list-type=2&prefix=%s", s3URL, bucket, prefix)

	resp, err := HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("querying S3: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("S3 returned status %d: %s", resp.StatusCode, string(body))
	}

	var result S3ListBucketResult
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding S3 XML: %w", err)
	}

	var keys []string
	for _, obj := range result.Contents {
		keys = append(keys, obj.Key)
	}
	return keys, nil
}

// ExecInContainer runs a command inside a running compose service container using
// podman-compose exec. Returns combined stdout+stderr output. The compose project
// is determined by the working directory (repo root is expected).
func ExecInContainer(service string, command ...string) (string, error) {
	composeBin, composeArgs := DetectComposeTool()

	args := append(composeArgs, "exec", "-T", service)
	args = append(args, command...)

	cmd := exec.Command(composeBin, args...)
	cmd.Dir = RepoRoot()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("exec in %s failed: %w\noutput: %s", service, err, string(out))
	}
	return string(out), nil
}

// DetectComposeTool returns the binary name and base args for the compose tool.
func DetectComposeTool() (string, []string) {
	if _, err := exec.LookPath("podman-compose"); err == nil {
		return "podman-compose", []string{"-f", "compose.yaml"}
	}
	return "docker", []string{"compose", "-f", "compose.yaml"}
}

// RepoRoot walks up from the test directory to find the repo root (where compose.yaml lives).
func RepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "../../.."
	}

	if _, err := os.Stat(wd + "/compose.yaml"); err == nil {
		return wd
	}

	parts := strings.Split(wd, string(os.PathSeparator))
	for i := len(parts); i > 0; i-- {
		candidate := string(os.PathSeparator) + strings.Join(parts[1:i], string(os.PathSeparator))
		if _, err := os.Stat(candidate + "/compose.yaml"); err == nil {
			return candidate
		}
	}

	return "../../.."
}

// EnvOrDefault returns the value of the environment variable named by key,
// or fallback if the variable is empty or unset.
func EnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// CheckStackRunning verifies the compose stack is reachable by hitting the
// webhook healthcheck endpoint. Returns an error describing the failure and
// which profile to start if the stack is not running.
func CheckStackRunning(webhookURL, profile string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	endpoint := webhookURL + "/eventreceiver/healthcheck"

	resp, err := client.Get(endpoint)
	if err != nil {
		return fmt.Errorf(
			"stack not running — webhook healthcheck at %s failed.\n"+
				"Start it with: task integration:up PROFILE=%s\nError: %v",
			endpoint, profile, err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"stack not running — webhook healthcheck at %s returned %d.\n"+
				"Start it with: task integration:up PROFILE=%s",
			endpoint, resp.StatusCode, profile,
		)
	}
	return nil
}
