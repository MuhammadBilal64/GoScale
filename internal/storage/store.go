package storage

import (
	"sync"

	"github.com/MuhammadBilal64/go-scale/internal/persistence"
)

type Store struct {
	data        map[string]string
	mu          sync.RWMutex
	persistence persistence.Persistence
}

// var mu sync.RWMutex

func NewStore(p persistence.Persistence) *Store {
	return &Store{
		data:        make(map[string]string),
		persistence: p,
	}
}
func (s *Store) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return s.persistence.Save(s.data)
}
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.data[key]
	return value, exists
}
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return s.persistence.Save(s.data)
}
func (s *Store) Load() error {
	data, err := s.persistence.Load()
	if err != nil {
		return err
	}
	if data == nil {
		data = make(map[string]string)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = data
	return nil
}
