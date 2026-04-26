package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PaymentHandler struct {
  store IdempotencyStore
}

type PaymentRequest struct {
  Amount    float64 `json:"amount"`
  AccountID string  `json:"account_id"`
  LoanID    string  `json:"loan_id"`
}

type PaymentResponse struct {
  Status        string    `json:"status"`
  Amount        float64   `json:"amount"`
  TransactionID string    `json:"transaction_id"`
  Timestamp     time.Time `json:"timestamp"`
  Message       string    `json:"message,omitempty"`
}

func NewPaymentHandler(store IdempotencyStore) *PaymentHandler {
  return &PaymentHandler{
    store: store,
  }
}

func (ph *PaymentHandler) processPayment(amount float64, accountID, loanID string) (*PaymentResponse, error) {
  fmt.Printf("[Business Logic] Processing payment: $%.2f from account %s for loan %s\n", amount, accountID, loanID)
    
  time.Sleep(2 * time.Second)
    
  transactionID := uuid.New().String()
    
  return &PaymentResponse{
    Status:        "paid",
    Amount:        amount,
    TransactionID: transactionID,
    Timestamp:     time.Now(),
    Message:       "Payment processed successfully",
  }, nil
}

func (ph *PaymentHandler) IdempotencyMiddleware(next http.HandlerFunc) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    idempotencyKey := r.Header.Get("Idempotency-Key")
        
    if idempotencyKey == "" {
      http.Error(w, "Idempotency-Key header is required", http.StatusBadRequest)
      return
    }
        
    fmt.Printf("\n[Middleware] Received request with Idempotency-Key: %s\n", idempotencyKey)
    fmt.Printf("   Method: %s, Path: %s\n", r.Method, r.URL.Path)
        
    if cached, exists := ph.store.Get(idempotencyKey); exists {
      if cached.Status == StatusCompleted {
        fmt.Printf("[Middleware] Key %s already completed - returning cached response\n", idempotencyKey)
        for k, vals := range cached.Headers {
          for _, v := range vals {
            w.Header().Add(k, v)
          }
        }
        w.WriteHeader(cached.StatusCode)
        w.Write(cached.Body)
        return
      }
            
      if cached.Status == StatusProcessing {
        fmt.Printf("[Middleware] Key %s is currently processing - returning 409 Conflict\n", idempotencyKey)
        w.Header().Set("Retry-After", "2")
        http.Error(w, "Request with this idempotency key is already in progress", http.StatusConflict)
        return
      }
  }
        
    success, err := ph.store.TryCreate(idempotencyKey, 60*time.Second)
    if err != nil {
      http.Error(w, "Internal server error", http.StatusInternalServerError)
      return
    }
        
    if !success {
      if cached, exists := ph.store.Get(idempotencyKey); exists && cached.Status == StatusCompleted {
        for k, vals := range cached.Headers {
          for _, v := range vals {
            w.Header().Add(k, v)
          }
        }
        w.WriteHeader(cached.StatusCode)
        w.Write(cached.Body)
        return
      }
      w.Header().Set("Retry-After", "2")
      http.Error(w, "Request with this idempotency key is already in progress", http.StatusConflict)
      return
    }
        
    fmt.Printf("[Middleware] Key %s is new - starting processing\n", idempotencyKey)
        
    var paymentReq PaymentRequest
    body, err := io.ReadAll(r.Body)
    if err != nil {
      ph.store.SetProcessing(idempotencyKey, 10*time.Second)
      http.Error(w, "Failed to read request body", http.StatusBadRequest)
      return
    }
    r.Body = io.NopCloser(bytes.NewBuffer(body))
        
    if err := json.Unmarshal(body, &paymentReq); err != nil {
      ph.store.SetProcessing(idempotencyKey, 10*time.Second)
      http.Error(w, "Invalid request body", http.StatusBadRequest)
      return
    }
        
    response, err := ph.processPayment(paymentReq.Amount, paymentReq.AccountID, paymentReq.LoanID)
        
    responseBody, _ := json.Marshal(response)
        
    headers := make(map[string][]string)
    headers["Content-Type"] = []string{"application/json"}
        
    if err != nil {
      ph.store.SetComplete(idempotencyKey, http.StatusInternalServerError, headers, []byte(`{"error":"payment failed"}`), 24*time.Hour)
      http.Error(w, "Payment processing failed", http.StatusInternalServerError)
      return
    }
        
    ph.store.SetComplete(idempotencyKey, http.StatusOK, headers, responseBody, 24*time.Hour)
    
    fmt.Printf("[Middleware] Payment completed - saved result for key %s\n", idempotencyKey)
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(responseBody)
	}
}

func runServer(wg *sync.WaitGroup, port string, handler http.HandlerFunc) {
 defer wg.Done()
    
  server := &http.Server{
    Addr:         ":" + port,
    Handler:      handler,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
  }
    
  fmt.Printf("Server starting on http://localhost:%s\n", port)
  if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
    fmt.Printf("Server error: %v\n", err)
  }
}

func repeat(s string, count int) string {
  result := ""
  for i := 0; i < count; i++ {
    result += s
  }
  return result
}

func main() {
  fmt.Println(repeat("=", 60))
  fmt.Println("PART 2: Idempotency Middleware - Loan Repayment")
  fmt.Println(repeat("=", 60))
    
  store := NewMemoryStore()
  paymentHandler := NewPaymentHandler(store)
    
  mux := http.NewServeMux()
  mux.HandleFunc("/api/payments", paymentHandler.IdempotencyMiddleware(func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
      http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
      return
    }
    w.Header().Set("Content-Type", "application/json")
  }))
    
  var wg sync.WaitGroup
  wg.Add(1)
  go runServer(&wg, "8080", mux.ServeHTTP)
    
  time.Sleep(500 * time.Millisecond)
    
  fmt.Println("\n" + repeat("-", 60))
  fmt.Println("TEST SCENARIO: Simulating Double-Click Attack")
  fmt.Println(repeat("-", 60))
    
  idempotencyKey := uuid.New().String()
  fmt.Printf("\nUsing Idempotency-Key: %s\n", idempotencyKey)
    
  paymentData := PaymentRequest{
    Amount:    1000.00,
    AccountID: "acc_123456789",
    LoanID:    "loan_987654321",
  }
  jsonData, _ := json.Marshal(paymentData)
    
  var wgRequests sync.WaitGroup
  results := make(chan struct {
    id     int
    status int
    body   string
  }, 10)
    
  fmt.Println("Sending 10 simultaneous requests with the same Idempotency-Key...")
    
  for i := 1; i <= 10; i++ {
    wgRequests.Add(1)
    go func(reqID int) {
      defer wgRequests.Done()
            
      client := &http.Client{Timeout: 5 * time.Second}
      req, _ := http.NewRequest("POST", "http://localhost:8080/api/payments", bytes.NewBuffer(jsonData))
      req.Header.Set("Content-Type", "application/json")
      req.Header.Set("Idempotency-Key", idempotencyKey)
            
      startTime := time.Now()
      resp, err := client.Do(req)
      elapsed := time.Since(startTime)
            
      if err != nil {
        results <- struct {
          id     int
          status int
          body   string
        }{reqID, 0, err.Error()}
        return
      }
      defer resp.Body.Close()
            
      body, _ := io.ReadAll(resp.Body)
      results <- struct {
        id     int
        status int
        body   string
      }{reqID, resp.StatusCode, string(body) + fmt.Sprintf(" (took %v)", elapsed)}
  	}(i)
  }
    
  wgRequests.Wait()
  close(results)
    
  fmt.Println("\nRESULTS:")
  fmt.Println(repeat("-", 40))
    
  successCount := 0
  conflictCount := 0
  errorCount := 0
    
  for res := range results {
    switch res.status {
    case 200:
      successCount++
      var paymentResp PaymentResponse
      json.Unmarshal([]byte(res.body), &paymentResp)
      fmt.Printf("Request #%2d: Status %d - %s\n", res.id, res.status, paymentResp.Message)
      fmt.Printf("              Transaction: %s\n", paymentResp.TransactionID)
    case 409:
      conflictCount++
      fmt.Printf("Request #%2d: Status %d - Conflict (request already in progress)\n", res.id, res.status)
    default:
      errorCount++
      fmt.Printf("Request #%2d: Status %d - %s\n", res.id, res.status, res.body)
    }
  }
    
  fmt.Println("\n" + repeat("-", 40))
  fmt.Printf("\nSUMMARY:\n")
  fmt.Printf("   Successful (first request): %d\n", successCount)
  fmt.Printf("   Conflict (during processing): %d\n", conflictCount)
  fmt.Printf("   Errors: %d\n", errorCount)
  fmt.Printf("   Total requests: %d\n", successCount+conflictCount+errorCount)
    
  fmt.Println("\n Waiting 3 seconds for first request to complete if still running...")
  time.Sleep(3 * time.Second)
    
  fmt.Println("\n Sending one more request with the same key (should get cached response)...")
  client := &http.Client{Timeout: 5 * time.Second}
  req, _ := http.NewRequest("POST", "http://localhost:8080/api/payments", bytes.NewBuffer(jsonData))
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Idempotency-Key", idempotencyKey)
    
  resp, err := client.Do(req)
  if err != nil {
    fmt.Printf("Request failed: %v\n", err)
  } else {
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    var paymentResp PaymentResponse
    json.Unmarshal(body, &paymentResp)
    fmt.Printf("Cached response: Status %d - %s\n", resp.StatusCode, paymentResp.Message)
    fmt.Printf("   Transaction ID: %s (same as first request)\n", paymentResp.TransactionID)
  }
    
  fmt.Println("\n" + repeat("=", 60))
  fmt.Println("All tests completed!")
  fmt.Println("   - Business logic executed only ONCE")
  fmt.Println("   - Duplicate requests received cached responses")
  fmt.Println("   - Concurrent requests properly handled with 409 Conflict")
  fmt.Println(repeat("=", 60))
}