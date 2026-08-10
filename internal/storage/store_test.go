package storage

import (
	"sync"
	"testing"

	"github.com/MuhammadBilal64/go-scale/internal/persistence"
)

type fakePersistence struct {
	entries []persistence.Entry
}

func (f *fakePersistence) Append(entry persistence.Entry) error {
	f.entries = append(f.entries, entry)
	return nil
}

func (f *fakePersistence) Load() (map[string]string, error) {
	data := make(map[string]string)
	for _, entry := range f.entries {
		switch entry.Op {
		case persistence.OpSet:
			data[entry.Key] = entry.Value
		case persistence.OpDelete:
			delete(data, entry.Key)
		}
	}
	return data, nil
}

func (f *fakePersistence) Compact(data map[string]string) error {
	return nil
}

func TestStoreGetAndSet(t *testing.T) {
	store := NewStore(&fakePersistence{})

	store.Set("name", "Bilal")

	value, exists := store.Get("name")
	if !exists {
		t.Fatal("expected key to exist")
	}
	if value != "Bilal" {
		t.Fatalf("expected Bilal, got %s", value)
	}
}

func TestStoreGetMissing(t *testing.T) {
	store := NewStore(&fakePersistence{})
	value, exists := store.Get("not-a-problem")

	if exists {
		t.Fatal("expected key not to exist")
	}

	if value != "" {
		t.Fatalf("expected empty value, got %s", value)
	}
}

func TestStoreDelete(t *testing.T) {
	store := NewStore(&fakePersistence{})
	store.Set("name", "bilal")
	store.Delete("name")
	_, exists := store.Get("name")
	if exists {
		t.Fatal("expected key not to exist after deletion")
	}
}

func TestStoreConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	store := NewStore(&fakePersistence{})
	wg.Add(2)
	go func() {
		defer wg.Done()
		store.Set("Name", "Bilal")
	}()

	go func() {
		defer wg.Done()
		store.Set("Name", "Malik")
	}()
	wg.Wait()
}

func TestStoreLoad(t *testing.T) {
	persistenceLayer := &fakePersistence{
		entries: []persistence.Entry{
			{Op: persistence.OpSet, Key: "name", Value: "Bilal"},
		},
	}
	store := NewStore(persistenceLayer)

	if err := store.Load(); err != nil {
		t.Fatalf("expected Load to succeed, got %v", err)
	}

	value, exists := store.Get("name")
	if !exists {
		t.Fatal("expected key to exist after Load")
	}
	if value != "Bilal" {
		t.Fatalf("expected Bilal, got %s", value)
	}
}

func TestStoreSetAppendsBeforeMemory(t *testing.T) {
	p := &fakePersistence{}
	store := NewStore(p)

	if err := store.Set("name", "Bilal"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if len(p.entries) != 1 {
		t.Fatalf("expected 1 wal entry, got %d", len(p.entries))
	}
	if p.entries[0].Op != persistence.OpSet || p.entries[0].Key != "name" {
		t.Fatalf("unexpected entry: %+v", p.entries[0])
	}
}

type failingCompactPersistence struct {
	fakePersistence
}

func (f *failingCompactPersistence) Compact(data map[string]string) error {
	return errCompactFailed
}

var errCompactFailed = &compactError{"compact failed"}

type compactError struct{ msg string }

func (e *compactError) Error() string { return e.msg }

func TestStoreSetSucceedsWhenCompactFails(t *testing.T) {
	store := NewStore(&failingCompactPersistence{})

	if err := store.Set("name", "Bilal"); err != nil {
		t.Fatalf("expected Set to succeed when compact fails, got %v", err)
	}
	value, exists := store.Get("name")
	if !exists || value != "Bilal" {
		t.Fatalf("expected Bilal in memory, got %q exists=%v", value, exists)
	}
}
