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

func NewStore(p persistence.Persistence) *Store {
	return &Store{
		data:        make(map[string]string),
		persistence: p,
	}
}

func (s *Store) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Write-ahead: durable log first, then memory.
	if err := s.persistence.Append(persistence.Entry{
		Op:    persistence.OpSet,
		Key:   key,
		Value: value,
	}); err != nil {
		return err
	}

	s.data[key] = value
	return nil
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

	if err := s.persistence.Append(persistence.Entry{
		Op:  persistence.OpDelete,
		Key: key,
	}); err != nil {
		return err
	}

	delete(s.data, key)
	return nil
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
