package storage

import (
	"sync"
	"testing"
)

func TestStoreGetAndSet(t *testing.T) {
	store := NewStore()

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
	store := NewStore()

	value, exists := store.Get("not-a-problem")

	if exists {
		t.Fatal("expected key not to exist")
	}

	if value != "" {
		t.Fatalf("expected empty value, got %s", value)
	}
}
func TestStoreDelete(t *testing.T) {
	store := NewStore()
	store.Set("name", "bilal")
	store.Delete("name")
	_, exists := store.Get("name")
	if exists {
		t.Fatal("expected key not to exist after deletion")
	}
}
func TestStoreConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	store := NewStore()
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
