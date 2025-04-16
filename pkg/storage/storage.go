package storage

import (
	"sync"
)

// Storage is an interface for key-value storage
type Storage interface {
	// Get retrieves a value by key
	Get(key string) (interface{}, bool)
	// Set stores a value by key
	Set(key string, value interface{})
	// Delete removes a value by key
	Delete(key string)
	// Clear removes all values
	Clear()
}

// InMemoryStorage is an in-memory implementation of Storage
type InMemoryStorage struct {
	mu    sync.RWMutex
	items map[string]interface{}
}

// NewInMemoryStorage creates a new in-memory storage
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		items: make(map[string]interface{}),
	}
}

// Get retrieves a value by key
func (s *InMemoryStorage) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.items[key]
	return value, ok
}

// Set stores a value by key
func (s *InMemoryStorage) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = value
}

// Delete removes a value by key
func (s *InMemoryStorage) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
}

// Clear removes all values
func (s *InMemoryStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]interface{})
}
