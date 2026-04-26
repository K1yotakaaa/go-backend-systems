package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"time"
)

type PaymentRequest struct {
  Amount   float64 `json:"amount"`
  Currency string  `json:"currency"`
  CardID   string  `json:"card_id"`
}

type PaymentResponse struct {
  Status        string `json:"status"`
  TransactionID string `json:"transaction_id,omitempty"`
  Message       string `json:"message,omitempty"`
}

func IsRetryable(resp *http.Response, err error) bool {
  if err != nil {
    return true
  }
  if resp == nil {
    return true
  }
  defer resp.Body.Close()
    
  switch resp.StatusCode {
  	case http.StatusTooManyRequests,
  	  http.StatusInternalServerError,
  	  http.StatusBadGateway,
  	  http.StatusServiceUnavailable,
  	  http.StatusGatewayTimeout:
  	  return true
  	case http.StatusUnauthorized,
  	  http.StatusNotFound:
  	  return false
  	default:
  	  return resp.StatusCode >= 500
  }
}

func CalculateBackoff(attempt int) time.Duration {
  baseDelay := 100 * time.Millisecond
  maxDelay := 5 * time.Second
    
  backoff := baseDelay * time.Duration(math.Pow(2, float64(attempt)))
  if backoff > maxDelay {
    backoff = maxDelay
  }
    
  jitter := time.Duration(rand.Int63n(int64(backoff)))
    
  return jitter
}

type RetryableClient struct {
  client     *http.Client
  maxRetries int
  baseDelay  time.Duration
  maxDelay   time.Duration
}

func NewRetryableClient(maxRetries int, baseDelay, maxDelay time.Duration) *RetryableClient {
  rand.Seed(time.Now().UnixNano())
  return &RetryableClient{
    client:     &http.Client{Timeout: 30 * time.Second},
    maxRetries: maxRetries,
    baseDelay:  baseDelay,
    maxDelay:   maxDelay,
  }
}

func (rc *RetryableClient) ExecutePayment(ctx context.Context, url string, reqBody PaymentRequest) (*PaymentResponse, error) {
  var lastErr error
  var lastResp *http.Response
    
  for attempt := 0; attempt < rc.maxRetries; attempt++ {
    select {
    case <-ctx.Done():
      return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
    default:
    }
        
    jsonBody, err := json.Marshal(reqBody)
    if err != nil {
      return nil, fmt.Errorf("failed to marshal request: %w", err)
    }
        
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
    if err != nil {
      return nil, fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
        
    lastResp, lastErr = rc.client.Do(req)
        
    if lastErr == nil && lastResp != nil && lastResp.StatusCode == http.StatusOK {
      defer lastResp.Body.Close()
      body, err := io.ReadAll(lastResp.Body)
      if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
      }
      var paymentResp PaymentResponse
      if err := json.Unmarshal(body, &paymentResp); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
      }
      fmt.Printf("Attempt %d: Success!\n", attempt+1)
      return &paymentResp, nil
    }
        
    if lastResp != nil && !IsRetryable(lastResp, lastErr) {
      defer lastResp.Body.Close()
      body, _ := io.ReadAll(lastResp.Body)
      return nil, fmt.Errorf("non-retryable error: status %d, body: %s", lastResp.StatusCode, string(body))
    }
        
    if attempt == rc.maxRetries-1 {
      break
    }
        
    waitTime := CalculateBackoff(attempt)
    fmt.Printf("Attempt %d failed, waiting %v before next retry...\n", attempt+1, waitTime)
        
    select {
    case <-time.After(waitTime):
    case <-ctx.Done():
      return nil, fmt.Errorf("context cancelled while waiting: %w", ctx.Err())
    }
  }
    
  if lastResp != nil {
    defer lastResp.Body.Close()
    body, _ := io.ReadAll(lastResp.Body)
    return nil, fmt.Errorf("failed after %d retries: status %d, body: %s", rc.maxRetries, lastResp.StatusCode, string(body))
  }
  return nil, fmt.Errorf("failed after %d retries: %w", rc.maxRetries, lastErr)
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
  fmt.Println("PART 1: Resilient HTTP Client - Payment Processing")
  fmt.Println(repeat("=", 60))
    
  requestCount := 0
    
  testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    requestCount++
    fmt.Printf("[Server] Received request #%d\n", requestCount)
        
	  if requestCount < 4 {
	    fmt.Printf("[Server] Returning 503 Service Unavailable (request #%d)\n", requestCount)
	    w.WriteHeader(http.StatusServiceUnavailable)
	    json.NewEncoder(w).Encode(PaymentResponse{
	      Status:  "failed",
	      Message: "Service temporarily unavailable",
	    })
	    return
	  }
        
    fmt.Printf("[Server] Returning 200 OK with success response (request #%d)\n", requestCount)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(PaymentResponse{
      Status:        "success",
      TransactionID: fmt.Sprintf("txn_%d", time.Now().UnixNano()),
      Message:       "Payment processed successfully",
    })
  }))
  defer testServer.Close()
    
  client := NewRetryableClient(5, 500*time.Millisecond, 5*time.Second)
    
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
    
  paymentReq := PaymentRequest{
    Amount:   1000.00,
    Currency: "USD",
    CardID:   "card_123456789",
  }
    
  fmt.Printf("\nSending payment request: Amount=$%.2f, Currency=%s\n", paymentReq.Amount, paymentReq.Currency)
  fmt.Printf("Global timeout: 10 seconds\n")
  fmt.Printf("Max retries: 5, Base delay: 500ms\n\n")
    
  startTime := time.Now()
  response, err := client.ExecutePayment(ctx, testServer.URL+"/pay", paymentReq)
  elapsed := time.Since(startTime)
    
  fmt.Println("\n" + repeat("-", 40))
  if err != nil {
    fmt.Printf("Payment failed: %v\n", err)
  } else {
    fmt.Printf("Payment successful!\n")
    fmt.Printf("   Status: %s\n", response.Status)
    fmt.Printf("   Transaction ID: %s\n", response.TransactionID)
    fmt.Printf("   Message: %s\n", response.Message)
  }
  fmt.Printf("   Total time: %v\n", elapsed)
  fmt.Println(repeat("-", 40))
  fmt.Printf("\nTotal server requests received: %d\n", requestCount)
}