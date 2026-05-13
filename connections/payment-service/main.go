package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type PaymentRequest struct {
	OrderID string  `json:"orderId"`
	Amount  float64 `json:"amount"`
}

type PaymentResponse struct {
	OrderID       string `json:"orderId"`
	Status        string `json:"status"`
	TransactionID string `json:"transactionId"`
	ProcessedAt   string `json:"processedAt"`
	CorrelationID string `json:"correlationId,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/payment/approved", approvedHandler)
	mux.HandleFunc("/payment/declined", declinedHandler)

	log.Println("Payment Service starting on port 9091...")
	log.Println("Context path: /payment")
	log.Println("Endpoints: POST /payment/approved, POST /payment/declined")
	log.Fatal(http.ListenAndServe(":9091", mux))
}

func approvedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	correlationID := r.Header.Get("X-Correlation-ID")
	source := r.Header.Get("X-Source-Service")
	log.Printf("[approved] received call: correlationId=%s source=%s", correlationID, source)

	var req PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp := PaymentResponse{
		OrderID:       req.OrderID,
		Status:        "APPROVED",
		TransactionID: "txn-" + req.OrderID,
		ProcessedAt:   time.Now().UTC().Format(time.RFC3339),
		CorrelationID: correlationID,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func declinedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	correlationID := r.Header.Get("X-Correlation-ID")
	log.Printf("[declined] received call: correlationId=%s", correlationID)

	var req PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp := PaymentResponse{
		OrderID:       req.OrderID,
		Status:        "DECLINED",
		TransactionID: "txn-" + req.OrderID,
		ProcessedAt:   time.Now().UTC().Format(time.RFC3339),
		CorrelationID: correlationID,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message, Code: code})
}
