package store

import (
	"sync"
	"sync/atomic"
)

type Item struct {
	Value   []byte
	Version int64
}

type Store struct {
	mu      sync.RWMutex
	data    map[string]*Item
	version int64
}

func New() *Store {
	return &Store{
		data: make(map[string]*Item),
	}
}

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

func (s *Store) Set(key string, value []byte) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := atomic.AddInt64(&s.version, 1)
	val := make([]byte, len(value))
	copy(val, value)
	s.data[key] = &Item{Value: val, Version: v}
	return v
}

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


func (s *Store) Delete(key string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := atomic.AddInt64(&s.version, 1)
	s.data[key] = &Item{Value: nil, Version: v} // tombstone
	return v
}

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

func (s *Store) Version() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.version
}

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
