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
	"sync"
	"time"

	"golang.org/x/oauth2/clientcredentials"
)

type Order struct {
	ID       string  `json:"id"`
	Item     string  `json:"item"`
	Quantity int     `json:"quantity"`
	Amount   float64 `json:"amount"`
	Status   string  `json:"status"`
}

type PaymentResponse struct {
	OrderID       string `json:"orderId"`
	Status        string `json:"status"`
	TransactionID string `json:"transactionId"`
	ProcessedAt   string `json:"processedAt"`
	CorrelationID string `json:"correlationId,omitempty"`
}

type PlaceOrderResponse struct {
	Order   Order           `json:"order"`
	Payment PaymentResponse `json:"payment"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

var (
	consumerKey    = os.Getenv("CHOREO_PAYMENT_SVC_CONNECTION_CONSUMERKEY")
	consumerSecret = os.Getenv("CHOREO_PAYMENT_SVC_CONNECTION_CONSUMERSECRET")
	serviceURL     = os.Getenv("CHOREO_PAYMENT_SVC_CONNECTION_SERVICEURL")
	tokenURL       = os.Getenv("CHOREO_PAYMENT_SVC_CONNECTION_TOKENURL")
	choreoAPIKey   = os.Getenv("CHOREO_PAYMENT_SVC_CONNECTION_APIKEY")

	paymentClient *http.Client

	mu       sync.Mutex
	orderSeq = 0
)

func main() {
	if consumerKey == "" || consumerSecret == "" || serviceURL == "" || tokenURL == "" {
		log.Println("WARNING: missing required env vars: CHOREO_PAYMENT_SVC_CONNECTION_{CONSUMERKEY,CONSUMERSECRET,SERVICEURL,TOKENURL}; payment service invocations will fail until these are set")
	} else {
		cfg := clientcredentials.Config{
			ClientID:     consumerKey,
			ClientSecret: consumerSecret,
			TokenURL:     tokenURL,
		}
		paymentClient = cfg.Client(context.Background())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ordering/orders", ordersHandler)

	log.Println("Ordering Service starting on port 9090...")
	log.Println("Context path: /ordering")
	log.Println("Endpoints: POST /ordering/orders")
	log.Printf("Payment service URL: %s", serviceURL)
	log.Printf("OAuth2 token URL: %s", tokenURL)
	log.Fatal(http.ListenAndServe(":9090", mux))
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if order.Item == "" || order.Quantity <= 0 || order.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "item, quantity, and amount are required")
		return
	}

	order.ID = nextOrderID()
	correlationID := fmt.Sprintf("corr-%s-%d", order.ID, time.Now().UnixNano())

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	payment, err := callPayment(ctx, "approved", order, correlationID)
	if err != nil {
		log.Printf("payment call failed: %v", err)
		writeError(w, http.StatusBadGateway, "payment service call failed: "+err.Error())
		return
	}

	order.Status = payment.Status
	log.Printf("order placed: id=%s status=%s correlationId=%s", order.ID, order.Status, correlationID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(PlaceOrderResponse{Order: order, Payment: payment})
}

func callPayment(ctx context.Context, action string, order Order, correlationID string) (PaymentResponse, error) {
	if paymentClient == nil || serviceURL == "" {
		return PaymentResponse{}, fmt.Errorf("payment service connection not configured")
	}
	url := fmt.Sprintf("%s/%s", serviceURL, action)

	body, _ := json.Marshal(map[string]interface{}{
		"orderId": order.ID,
		"amount":  order.Amount,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return PaymentResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", correlationID)
	req.Header.Set("X-Source-Service", "ordering-service")
	req.Header.Set("X-Order-ID", order.ID)
	if choreoAPIKey != "" {
		req.Header.Set("Choreo-API-Key", choreoAPIKey)
	}

	resp, err := paymentClient.Do(req)
	if err != nil {
		return PaymentResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return PaymentResponse{}, fmt.Errorf("status %d: %s", resp.StatusCode, string(b))
	}

	var pr PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return PaymentResponse{}, err
	}
	return pr, nil
}

func nextOrderID() string {
	mu.Lock()
	defer mu.Unlock()
	orderSeq++
	return fmt.Sprintf("ord-%d", orderSeq)
}

func writeError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message, Code: code})
}
