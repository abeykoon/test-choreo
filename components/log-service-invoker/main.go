package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2/clientcredentials"
)

type InvokeResponse struct {
	Status         int             `json:"status"`
	CorrelationID  string          `json:"correlationId"`
	UpstreamStatus int             `json:"upstreamStatus"`
	Payload        json.RawMessage `json:"payload"`
	Response       json.RawMessage `json:"response"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

var (
	consumerKey    = os.Getenv("CHOREO_LOG_SVC_CONNECTION_CONSUMERKEY")
	consumerSecret = os.Getenv("CHOREO_LOG_SVC_CONNECTION_CONSUMERSECRET")
	serviceURL     = os.Getenv("CHOREO_LOG_SVC_CONNECTION_SERVICEURL")
	tokenURL       = os.Getenv("CHOREO_LOG_SVC_CONNECTION_TOKENURL")
	choreoAPIKey   = os.Getenv("CHOREO_LOG_SVC_CONNECTION_APIKEY")

	logClient *http.Client
)

func main() {
	if consumerKey == "" || consumerSecret == "" || serviceURL == "" || tokenURL == "" {
		log.Println("WARNING: missing required env vars: CHOREO_LOG_SVC_CONNECTION_{CONSUMERKEY,CONSUMERSECRET,SERVICEURL,TOKENURL}; log service invocations will fail until these are set")
	} else {
		cfg := clientcredentials.Config{
			ClientID:     consumerKey,
			ClientSecret: consumerSecret,
			TokenURL:     tokenURL,
		}
		logClient = cfg.Client(context.Background())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/invoker", invokerHandler)

	log.Println("Log Service Invoker starting on port 9090...")
	log.Println("Context path: /invoker")
	log.Println("Endpoints: POST /invoker")
	log.Printf("Log service URL: %s", serviceURL)
	log.Printf("OAuth2 token URL: %s", tokenURL)
	log.Fatal(http.ListenAndServe(":9090", mux))
}

func invokerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	correlationID := fmt.Sprintf("corr-invoker-%d", time.Now().UnixNano())

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	upstreamStatus, respBody, err := callLogService(ctx, body, correlationID)
	if err != nil {
		log.Printf("log service call failed: %v", err)
		writeError(w, http.StatusBadGateway, "log service call failed: "+err.Error())
		return
	}

	log.Printf("log service invoked: correlationId=%s upstreamStatus=%d", correlationID, upstreamStatus)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(InvokeResponse{
		Status:         http.StatusOK,
		CorrelationID:  correlationID,
		UpstreamStatus: upstreamStatus,
		Payload:        body,
		Response:       respBody,
	})
}

func callLogService(ctx context.Context, payload []byte, correlationID string) (int, json.RawMessage, error) {
	if logClient == nil || serviceURL == "" {
		return 0, nil, fmt.Errorf("log service connection not configured")
	}
	url := fmt.Sprintf("%s/greeting", serviceURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", correlationID)
	req.Header.Set("X-Source-Service", "log-service-invoker")
	if choreoAPIKey != "" {
		req.Header.Set("Choreo-API-Key", choreoAPIKey)
	}

	resp, err := logClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(b))
	}

	if !json.Valid(b) {
		b, _ = json.Marshal(string(b))
	}
	return resp.StatusCode, json.RawMessage(b), nil
}

func writeError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message, Code: code})
}
