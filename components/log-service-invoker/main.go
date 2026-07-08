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
)

type InvokeResponse struct {
	Status              int             `json:"status"`
	CorrelationID       string          `json:"correlationId"`
	ChoreoCorrelationID string          `json:"choreoCorrelationId"`
	UpstreamStatus      int             `json:"upstreamStatus"`
	Response            json.RawMessage `json:"response"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

const resourcePath = "/test/greeting"

var (
	serviceURL   = os.Getenv("CHOREO_LOG_SERVICE_CONNECTION_ORG_LEVEL_SERVICEURL")
	choreoAPIKey = os.Getenv("CHOREO_LOG_SERVICE_CONNECTION_ORG_LEVEL_CHOREOAPIKEY")

	logClient = &http.Client{Timeout: 10 * time.Second}
)

func main() {
	if serviceURL == "" || choreoAPIKey == "" {
		log.Println("WARNING: missing required env vars: CHOREO_LOG_SERVICE_CONNECTION_ORG_LEVEL_{SERVICEURL,CHOREOAPIKEY}; log service invocations will fail until these are set")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/invoker", invokerHandler)

	log.Println("Log Service Invoker starting on port 9090...")
	log.Printf("Log service URL: %s", serviceURL)
	log.Fatal(http.ListenAndServe(":9090", mux))
}

func invokerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//log request details for debugging
	logRequest(r)

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	correlationID := r.Header.Get("X-Correlation-ID")
	if correlationID == "" {
		correlationID = fmt.Sprintf("corr-invoker-%d", time.Now().UnixNano())
	}
	choreoCorrelationID := r.Header.Get("X-Choreo-Correlation-Id")
	if choreoCorrelationID == "" {
		choreoCorrelationID = correlationID
	}
	requestID := r.Header.Get("X-Request-Id")

	w.Header().Set("X-Correlation-Id", correlationID)
	w.Header().Set("X-Choreo-Correlation-Id", choreoCorrelationID)
	if requestID != "" {
		w.Header().Set("X-Request-Id", requestID)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	upstreamStatus, respBody, err := callLogService(ctx, correlationID, choreoCorrelationID)
	if err != nil {
		log.Printf("log service call failed: %v", err)
		writeError(w, http.StatusBadGateway, "log service call failed: "+err.Error())
		return
	}

	log.Printf("log service invoked: correlationId=%s choreoCorrelationId=%s upstreamStatus=%d", correlationID, choreoCorrelationID, upstreamStatus)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(InvokeResponse{
		Status:              http.StatusOK,
		CorrelationID:       correlationID,
		ChoreoCorrelationID: choreoCorrelationID,
		UpstreamStatus:      upstreamStatus,
		Response:            respBody,
	})
}

func logRequest(r *http.Request) {
	log.Printf("incoming request: %s %s", r.Method, r.URL.Path)
	for name, values := range r.Header {
		for _, v := range values {
			log.Printf("  header: %s: %s", name, v)
		}
	}

	if r.Body == nil {
		log.Printf("  body: <nil>")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("  body: <read error: %v>", err)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	if len(body) == 0 {
		log.Printf("  body: <empty>")
		return
	}
	log.Printf("  body: %s", string(body))
}

func callLogService(ctx context.Context, correlationID, choreoCorrelationID string) (int, json.RawMessage, error) {
	if serviceURL == "" || choreoAPIKey == "" {
		return 0, nil, fmt.Errorf("log service connection not configured")
	}

	url := serviceURL + resourcePath

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Choreo-API-Key", choreoAPIKey)
	req.Header.Set("X-Correlation-ID", correlationID)
	req.Header.Set("X-Choreo-Correlation-Id", choreoCorrelationID)

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
