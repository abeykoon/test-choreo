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
	"sort"
	"strings"
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

// Env var prefixes of the payment connections this service may be injected with,
// in order of preference.
var connectionPrefixes = []string{
	"CHOREO_FIRST_PAYMENT_CONN",
	"CHOREO_SECOND_PAYMENT_CONN",
}

var (
	connectionPrefix string
	consumerKey      string
	consumerSecret   string
	serviceURL       string
	tokenURL         string
	choreoAPIKey     string

	paymentClient *http.Client

	mu       sync.Mutex
	orderSeq = 0
)

// resolveConnection picks the first connection prefix that has any env var
// injected and loads its values, logging which connection is referred.
func resolveConnection() {
	var found []string
	for _, p := range connectionPrefixes {
		if os.Getenv(p+"_SERVICEURL") != "" || os.Getenv(p+"_CONSUMERKEY") != "" {
			found = append(found, p)
		}
	}

	switch len(found) {
	case 0:
		connectionPrefix = connectionPrefixes[0]
		log.Printf("Payment connection: none found (looked for %v); defaulting to %s", connectionPrefixes, connectionPrefix)
	case 1:
		connectionPrefix = found[0]
		log.Printf("Payment connection: using %s", connectionPrefix)
	default:
		connectionPrefix = found[0]
		log.Printf("Payment connection: multiple found %v; using %s", found, connectionPrefix)
	}

	consumerKey = os.Getenv(connectionPrefix + "_CONSUMERKEY")
	consumerSecret = os.Getenv(connectionPrefix + "_CONSUMERSECRET")
	serviceURL = os.Getenv(connectionPrefix + "_SERVICEURL")
	tokenURL = os.Getenv(connectionPrefix + "_TOKENURL")
	choreoAPIKey = os.Getenv(connectionPrefix + "_CHOREOAPIKEY")
}

// logEnvCheck prints the env vars used to verify whether the custom env var
// mapping (configurations.env in component.yaml) was applied on deploy.
func logEnvCheck() {
	for _, name := range []string{"CHOREO_FIRST_PAYMENT_CONN_SERVICEURL", "MY_CONN_SERVICEURL"} {
		if v, ok := os.LookupEnv(name); ok {
			log.Printf("Env check: %s=%s", name, v)
		} else {
			log.Printf("Env check: %s=<not set>", name)
		}
	}

	var names []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(name, "CHOREO_") || strings.HasPrefix(name, "MY_CONN_") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	log.Printf("Env check: connection-related env var names: %v", names)
}

func main() {
	logEnvCheck()
	resolveConnection()

	if consumerKey == "" || consumerSecret == "" || serviceURL == "" || tokenURL == "" {
		log.Printf("WARNING: missing required env vars: %s_{CONSUMERKEY,CONSUMERSECRET,SERVICEURL,TOKENURL}; payment service invocations will fail until these are set", connectionPrefix)
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
	log.Printf("OAuth2 consumer key: %s", consumerKey)
	log.Printf("OAuth2 consumer secret: %s", mask(consumerSecret))
	log.Printf("Choreo API key: %s", mask(choreoAPIKey))
	log.Fatal(http.ListenAndServe(":9090", mux))
}

// mask hides all but the last 4 characters of a secret so it can be logged safely.
func mask(s string) string {
	if s == "" {
		return "<not set>"
	}
	if len(s) <= 4 {
		return "****"
	}
	return fmt.Sprintf("****%s (len=%d)", s[len(s)-4:], len(s))
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

	payment, err := callPayment(ctx, "payment/approved", order, correlationID)
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
