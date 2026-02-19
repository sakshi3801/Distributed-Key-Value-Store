package store

import (
	"sync"
	"sync/atomic"
)

// Item holds a value and its version for ordering.
type Item struct {
	Value   []byte
	Version int64
}

// Store is an in-memory key-value store with versioning.
type Store struct {
	mu      sync.RWMutex
	data    map[string]*Item
	version int64
}

// New creates a new in-memory store.
func New() *Store {
	return &Store{
		data: make(map[string]*Item),
	}
}

// Get returns the value and version for key, and whether it existed.
func (s *Store) Get(key string) (value []byte, version int64, found bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.data[key]
	if !ok || item == nil || item.Value == nil {
		return nil, 0, false
	}
	// Return a copy so caller cannot mutate
	val := make([]byte, len(item.Value))
	copy(val, item.Value)
	return val, item.Version, true
}

// Set stores value for key with the next version. Returns the version used.
func (s *Store) Set(key string, value []byte) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := atomic.AddInt64(&s.version, 1)
	val := make([]byte, len(value))
	copy(val, value)
	s.data[key] = &Item{Value: val, Version: v}
	return v
}

// SetVersion stores value with an explicit version (for replication).
// Only applies if version is greater than existing.
func (s *Store) SetVersion(key string, value []byte, version int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.data[key]; ok && existing != nil && existing.Version >= version {
		return false
	}
	val := make([]byte, len(value))
	copy(val, value)
	s.data[key] = &Item{Value: val, Version: version}
	if version > s.version {
		s.version = version
	}
	return true
}

// Delete removes the key. Uses version for ordering.
func (s *Store) Delete(key string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := atomic.AddInt64(&s.version, 1)
	s.data[key] = &Item{Value: nil, Version: v} // tombstone
	return v
}

// DeleteVersion removes the key with explicit version (for replication).
func (s *Store) DeleteVersion(key string, version int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.data[key]; ok && existing != nil && existing.Version >= version {
		return false
	}
	s.data[key] = &Item{Value: nil, Version: version}
	if version > s.version {
		s.version = version
	}
	return true
}

// Version returns the current logical version counter.
func (s *Store) Version() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.version
}

// Snapshot returns a copy of all key-value pairs (for rebalancing/debug). Tombstones are skipped.
func (s *Store) Snapshot() map[string]*Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]*Item, len(s.data))
	for k, v := range s.data {
		if v == nil || v.Value == nil {
			continue
		}
		val := make([]byte, len(v.Value))
		copy(val, v.Value)
		out[k] = &Item{Value: val, Version: v.Version}
	}
	return out
}
