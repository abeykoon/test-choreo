package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/test/greeting", greetingHandler)

	log.Println("Log Service starting on port 9090...")
	log.Println("Context path: /test")
	log.Fatal(http.ListenAndServe(":9090", mux))
}

// greetingHandler handles GET and POST requests on /test/greeting
func greetingHandler(w http.ResponseWriter, r *http.Request) {
	logHeaders(r)

	switch r.Method {
	case http.MethodGet:
		handleGet(w, r)
	case http.MethodPost:
		handlePost(w, r)
	default:
		w.Header().Set("Content-Type", "application/json")
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleGet returns a static Hello World string
func handleGet(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Hello World")); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

// handlePost echoes the incoming JSON payload back as JSON
func handlePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	log.Printf("Incoming payload: %s", string(body))

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

// logHeaders logs all headers of the incoming request
func logHeaders(r *http.Request) {
	log.Printf("Incoming %s %s", r.Method, r.URL.Path)
	for name, values := range r.Header {
		for _, value := range values {
			log.Printf("Header: %s = %s", name, value)
		}
	}
}

// writeError writes a JSON error response
func writeError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(ErrorResponse{
		Message: message,
		Code:    code,
	}); err != nil {
		log.Printf("failed to encode error response: %v", err)
	}
}
