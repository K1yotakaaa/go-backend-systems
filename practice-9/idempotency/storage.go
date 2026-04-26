package main

import (
	"encoding/json"
	"sync"
	"time"
)

type RequestStatus string

const (
  StatusProcessing RequestStatus = "processing"
  StatusCompleted  RequestStatus = "completed"
)

type CachedResponse struct {
  StatusCode int             `json:"status_code"`
  Headers    map[string][]string `json:"headers"`
  Body       json.RawMessage `json:"body"`
  Status     RequestStatus   `json:"status"`
  CreatedAt  time.Time       `json:"created_at"`
  CompletedAt *time.Time     `json:"completed_at,omitempty"`
}

type IdempotencyStore interface {
  TryCreate(key string, ttl time.Duration) (bool, error)
  Get(key string) (*CachedResponse, bool)
  SetComplete(key string, statusCode int, headers map[string][]string, body []byte, cacheTTL time.Duration) error
  SetProcessing(key string, ttl time.Duration) error
}

type MemoryStore struct {
  mu    sync.RWMutex
  store map[string]*CachedResponse
}

func NewMemoryStore() *MemoryStore {
  return &MemoryStore{
    store: make(map[string]*CachedResponse),
  }
}

func (ms *MemoryStore) TryCreate(key string, ttl time.Duration) (bool, error) {
  ms.mu.Lock()
  defer ms.mu.Unlock()
    
  if _, exists := ms.store[key]; exists {
    return false, nil
  }
    
  ms.store[key] = &CachedResponse{
    Status:    StatusProcessing,
    CreatedAt: time.Now(),
  }
    
  go ms.scheduleCleanup(key, ttl)
    
  return true, nil
}

func (ms *MemoryStore) scheduleCleanup(key string, ttl time.Duration) {
  time.Sleep(ttl)
  ms.mu.Lock()
  defer ms.mu.Unlock()
    
  if cached, exists := ms.store[key]; exists && cached.Status == StatusProcessing {
    delete(ms.store, key)
  }
}

func (ms *MemoryStore) Get(key string) (*CachedResponse, bool) {
  ms.mu.RLock()
  defer ms.mu.RUnlock()
    
  cached, exists := ms.store[key]
  return cached, exists
}

func (ms *MemoryStore) SetComplete(key string, statusCode int, headers map[string][]string, body []byte, cacheTTL time.Duration) error {
  ms.mu.Lock()
  defer ms.mu.Unlock()
    
  now := time.Now()
  ms.store[key] = &CachedResponse{
    StatusCode:  statusCode,
    Headers:     headers,
    Body:        json.RawMessage(body),
    Status:      StatusCompleted,
    CreatedAt:   ms.store[key].CreatedAt,
    CompletedAt: &now,
  }
    
  go ms.scheduleCleanup(key, cacheTTL)
    
  return nil
}

func (ms *MemoryStore) SetProcessing(key string, ttl time.Duration) error {
  ms.mu.Lock()
  defer ms.mu.Unlock()
    
  ms.store[key] = &CachedResponse{
    Status:    StatusProcessing,
    CreatedAt: time.Now(),
  }
    
 	go ms.scheduleCleanup(key, ttl)
    
  return nil
}